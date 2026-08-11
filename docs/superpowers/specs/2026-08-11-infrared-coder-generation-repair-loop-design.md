# Infrared coder generation: baseline-value prompt fix + validate/repair/escalate loop

## Context

This session's real-protocol testing (Daikin64 and Samsung AC, using ground
truth verified against IRremoteESP8266's own source and a real captured
signal) found that `WriteCoder()`
(`backend/internal/application/infrared/coder_generation/llm_coder.go`) has
a real, reproducible gap: its prompt sends the baseline case only as a raw
bit array, never as named state values. Models systematically wrote encoders
whose validation/switch logic only recognized the delta-tested values (the
ones that appear in `AnalysisPayload.States[].ValueBits`) and rejected the
baseline value itself — even though the baseline is the single most common
case in real usage. A one-round "feed back the concrete bit mismatch and
ask for a fix" repair loop measurably helped one failure (0% → 76.6% owned
bits) but was inconsistent: it did nothing for a second model, and actively
broke a third model's previously-parseable JavaScript. This spec covers
fixing the root-cause prompt gap, then adding a bounded, evidence-based
validate → repair → escalate loop as a safety net for whatever the prompt
fix doesn't catch on its own.

## 1. Prompt fix: state baseline values explicitly

`runAnalysisAndGeneration` (`record_session_management/usecase.go`) already
computes `baselineState` (a `map[string]string` keyed by state ID, built by
`buildAnalysisPayload`) — it's just never passed to `WriteCoder`. Convert it
to a name-keyed map (the same `stateIdToName` helper already used for the
goja smoke test) and pass it through.

`WriteCoder`'s signature gains one parameter:

```go
func WriteCoder(
    ctx context.Context,
    client domaincontractsllm.Client,
    deviceBrand string,
    deviceModel string,
    states []domainmodels.InfraredState,
    definitions []domainmodels.InfraredStateDeviceDefinition,
    payload applicationinfraredanalysis.AnalysisPayload,
    baselineValues map[string]string, // NEW: state name -> its value in the baseline case
) (Coder, error)
```

The prompt gains one line, e.g.:

```
The baseline (most common) state is: POWER=ON, MODE=COOL, TEMP=16. Every
state's baseline value MUST be a valid, correctly-encoded input to your
encoder — do not write validation logic that only accepts the specific
non-baseline values mentioned below.
```

## 2. Validation: compare against real recorded bits, not just "did it throw"

New file `coder_generation/validate.go`. Production already has real ground
truth for every case it has ever recorded — `buildAnalysisPayload` computes
`baselineBits` and a `caseBits` map (keyed by case ID) with
`caseTargetState`/`caseTargetValue` alongside it. Nothing new needs to be
captured; this just uses data the pipeline already threw away after
building `AnalysisPayload`.

```go
type CaseValidation struct {
    Label         string   // e.g. state name + target value, for logging
    OwnedCorrect  int
    OwnedTotal    int
    ChecksumOK    int
    ChecksumTotal int
    RunError      string   // non-empty if RunEncoder itself failed
}

type ValidationResult struct {
    Cases         []CaseValidation
    OwnedCorrect, OwnedTotal     int // summed across Cases
    ChecksumOK, ChecksumTotal   int // summed across Cases
}

func Validate(
    runner domaincontractsutility.JSEngine,
    encoderSource string,
    checksumBits []int,
    knownCases []KnownCase, // {Label string; State map[string]string; Bits []int}
) ValidationResult
```

For each known case (baseline + every already-recorded non-baseline case,
built by the caller from `buildAnalysisPayload`'s existing outputs): run
`RunEncoder(encoderSource, state, encoderSmokeTestTimeout)`, decode the
result via the existing `DemodulateBits`, and compare position-by-position
against that case's real bits — owned-bit positions (from
`payload.States[].BitOffsets` unioned) scored separately from
`payload.ChecksumBits` positions, exactly as this session's test scripts
did. A `RunError` short-circuits that case's comparison (0 owned, 0
checksum — it contributes to the denominator as a full miss).

**Pass bar:** `OwnedCorrect == OwnedTotal` (100% exact match) — these are
real recorded bits, not estimates, so any mismatch is a genuine defect, not
a threshold call.

## 2a. Timeout budget

`runAnalysisAndGeneration` currently runs its whole body — including the
single `WriteCoder` call — under `analysisTimeout = 3 * time.Minute`. This
session's testing saw individual generation calls take anywhere from ~20s
to over 70s. With up to 4 sequential generation calls in the repair loop
(1 initial + 3 repair) plus a possible checksum-clarification call, the old
3-minute ceiling is no longer enough — hitting it mid-loop today surfaces
as a raw `context deadline exceeded`, not a clean escalation.

Two changes:

- Bump `analysisTimeout` from `3 * time.Minute` to `10 * time.Minute` — a
  generous ceiling for the whole sequence (bit analysis is fast; the
  budget is almost entirely LLM call time).
- Add a new `coderGenerationCallTimeout = 2 * time.Minute`, applied via
  `context.WithTimeout(ctx, coderGenerationCallTimeout)` around each
  individual `WriteCoder`/`RepairCoder` call. This bounds any single call
  so one hung request can't consume the entire 10-minute budget by itself
  — a timed-out call is treated the same as any other error from that
  round (contributes nothing, `best`/`bestResult` unchanged, loop
  continues to the next round or exhausts normally).

## 3. Repair loop

In `runAnalysisAndGeneration`, after the (now baseline-aware) `WriteCoder`
call succeeds and before the existing single goja smoke test:

```
attempt := initial Coder from WriteCoder
best := attempt
bestResult := Validate(attempt)

for round := 1; round <= 3 && bestResult.OwnedCorrect < bestResult.OwnedTotal; round++ {
    if bestResult.OwnedCorrect == bestResult.OwnedTotal && bestResult checksum-only gap {
        break // handled by section 4, not repair
    }
    feedback := buildFeedback(bestResult) // concrete per-case mismatches/errors, same shape validated this session
    repaired, err := RepairCoder(ctx, client, brand, model, attempt.EncoderSource, attempt.DecoderSource, feedback)
    if err != nil { continue } // this round's attempt failed to even generate; best/bestResult unchanged
    attempt = repaired
    result := Validate(attempt)
    if result.OwnedCorrect > bestResult.OwnedCorrect { // only replace on improvement — repair isn't monotonic
        best, bestResult = attempt, result
    }
}
```

`RepairCoder` (new function in `coder_generation`, e.g.
`llm_coder_repair.go`) mirrors `WriteCoder`'s response contract (same
`responseSchema`, same `Coder` return type) but its prompt includes the
prior encoder/decoder source plus a structured feedback string built from
`ValidationResult` (per case: either `RunError`, or up to N example
`"bit %d: expected %d got %d"` mismatches) — this is the same feedback shape
already validated empirically this session (helped once, inert once, but
never made scoring *worse* when the code stayed syntactically valid — the
one regression we saw is exactly why `best`/`bestResult` only ever move
forward on measured improvement).

**Why "improvement" and not "did it pass": to correctly implement the
early-exit, the loop condition re-checks `bestResult` each iteration, not
the latest `result` — a regressive round is silently discarded rather than
becoming the new baseline for the next round's feedback. (The next round's
feedback is still built from `bestResult`/`best`, not the discarded
attempt, so a bad round doesn't poison the next one either.)

## 4. Checksum-only gap → new pre-persistence case-request path

After validation (initial or any repair round), if `OwnedCorrect ==
OwnedTotal` but `ChecksumOK < ChecksumTotal`, that's scoped as a data
problem (too few examples to reverse-engineer the checksum algorithm — a
plausible failure mode independent of model quality), not something a
repair prompt can fix. Skip the repair loop (or abandon it if already
mid-loop) and take a new path instead of persisting a coder yet:

- New `coder_generation.WriteChecksumClarificationCases(ctx, client,
  brand, model, coder Coder, states []InfraredState) ([]RetryCasePlan,
  error)` — same shape as the existing `WriteRetryCases`
  (`RetryCasePlan{Description string; States map[uuid.UUID]string}`, same
  `testCaseResponseSchema`), new prompt asking specifically for recording
  scenarios likely to help pin down the checksum algorithm (e.g. varying
  multiple states together, or repeating a case to confirm determinism).
- Extract the shared tail of `runRetryCaseGeneration` (persist new cases via
  `recordCase.CreateWithStates`, activate the first one, set it as the
  session's current case, transition to `RECORDING`) into a private helper
  `persistNewCasesAndReturnToRecording(ctx, tag, sessionId, plans
  []RetryCasePlan)` used by both the existing post-test-case-failure path
  and this new pre-persistence path. `runRetryCaseGeneration` itself is
  **not** reused directly — its preconditions (a persisted coder, real
  hardware-verified test case failures) don't hold here, and bending them
  to fit would make both call sites harder to reason about. Only the
  "persist cases and go back to RECORDING" mechanics are shared.
- **Capped at once per session.** New nullable
  `ChecksumClarificationUsedAt *time.Time` column on
  `infrared_record_session` (migration +
  `InfraredRecordSession.ChecksumClarificationUsedAt` domain field + a new
  repository method `MarkChecksumClarificationUsedById(ctx, id)` +
  postgres implementation). If this path would trigger again on a later
  pass through `runAnalysisAndGeneration` for the same session, it's
  treated as a repair-loop case instead (falls through to section 5's
  60%-floor logic on whatever owned/checksum split it has) rather than
  looping indefinitely.

## 5. Exhausted repair loop: 60% floor

If the loop in section 3 finishes 3 rounds without reaching
`OwnedCorrect == OwnedTotal` (and it wasn't a checksum-only gap routed to
section 4):

- **`best`'s owned-bit accuracy ≥ 60%:** persist `best` as the coder via
  the existing `u.coder.Create(...)` call (status defaults to
  `UNVERIFIED`, unchanged from today — the existing hardware
  test-case-transmission stage remains the further real-world gate on it).
  Log at `Warn` level: the final owned-bit score, which round produced
  `best` (0 = the initial attempt, no repair helped), and the checksum
  score for visibility. No schema/API/UI change — this is deliberately
  log-only for this iteration; a persisted/surfaced signal can be added
  later if operators need it without digging through logs.
- **`best`'s owned-bit accuracy < 60%:** do not persist a coder. Transition
  the session to `InfraredRecordingStateFailed` with a
  `domainmodels.NewError("llm failed to capture the device's encoding
  pattern (best attempt: N% of known bits correct)", ErrTypeFailure, nil)`
  — same shape as every other failure path in this function.

`60%` is a named constant (`const ownedBitFatalFloor = 0.6` or similar),
not derived from data — flagged as a tunable starting point.

## Interfaces touched (summary)

- `coder_generation.WriteCoder` — new `baselineValues map[string]string`
  parameter; prompt text gains one line.
- `coder_generation.Validate` (new), `coder_generation.CaseValidation`
  (new), `coder_generation.ValidationResult` (new), `coder_generation.KnownCase` (new).
- `coder_generation.RepairCoder` (new).
- `coder_generation.WriteChecksumClarificationCases` (new), reusing
  `RetryCasePlan`/`testCasePlanResponse`/`testCaseResponseSchema` from
  `llm_retry_cases.go`.
- `record_session_management.usecase.runAnalysisAndGeneration` — gains the
  full validate/repair/escalate flow after `WriteCoder` and before the
  existing single goja smoke test (which becomes redundant with `Validate`'s
  baseline case and can be removed, since `Validate` always checks the
  baseline first).
- `record_session_management.usecase.persistNewCasesAndReturnToRecording`
  (new private helper, extracted from `runRetryCaseGeneration`'s tail).
- `domainmodels.InfraredRecordSession` — new
  `ChecksumClarificationUsedAt *time.Time` field.
- `domaincontractsrepository.InfraredRecordSession` — new
  `MarkChecksumClarificationUsedById(ctx, id) error` method.
- New migration: `infrared_record_session` gains nullable
  `checksum_clarification_used_at timestamptz`.

## Safety properties

- **Bounded iteration**: repair is capped at 3 rounds per pass; the
  checksum-clarification detour is capped at once per session via
  `ChecksumClarificationUsedAt`. No infinite loop is possible.
- **Bounded cost**: at most ~4 generation calls per pass (1 initial + 3
  repair), each individually bounded by `coderGenerationCallTimeout`.
- **No new concurrency risk**: the loop is sequential within the existing
  single background goroutine — same execution model as today, no shared
  mutable state introduced.
- **The existing hardware-verification gate is unchanged**: a coder
  persisted via section 5's degraded path stays `UNVERIFIED`, exactly like
  any newly generated coder today. It only ever reaches `ACTIVE` (eligible
  for real transmission) after the existing test-case stage — real
  hardware transmission, human-confirmed pass/fail — succeeds. Nothing in
  this design lets a coder skip that gate.

## Out of scope

- Any change to `Attribute()`'s checksum-bit detection heuristic itself
  (the incompleteness found this session is a separate, pre-existing gap —
  this spec works around it by scoring checksum bits separately and never
  gating on them, not by fixing detection).
- The two `SegmentFrames` bugs (doubled-leader, multi-section
  misalignment) found earlier this session — separate spec if/when
  prioritized.
- Any UI/API surface for the new validation scores or the degraded-persist
  log — log-only per section 5.
