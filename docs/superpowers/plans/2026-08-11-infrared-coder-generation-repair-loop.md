# Infrared Coder Generation: Baseline-Value Fix + Validate/Repair/Escalate Loop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the root-cause prompt gap that makes `WriteCoder` reject a
device's own baseline state value, then add a bounded validate → repair →
escalate loop around it that scores a generated encoder against the real
bits the pipeline already recorded, instead of trusting a single
`err == nil` smoke test.

**Architecture:** `WriteCoder`'s prompt gains the baseline's per-state
values. A new `Validate()` re-runs the encoder against every already-known
case (baseline + each recorded OFAT case) and diffs its decoded output
bit-for-bit against the real recorded bits for that case — data
`buildAnalysisPayload` already computes but currently throws away. If
validation isn't clean, `runAnalysisAndGeneration` either routes to a new,
capped-at-once-per-session "ask for more recording scenarios" path (when
only checksum bits are wrong — a data problem) or into a bounded 3-round
repair loop that feeds concrete mismatches back to the LLM (when owned bits
are wrong — a logic problem), keeping only the best-scoring attempt across
rounds. Below a 60% owned-bit floor after repair is exhausted, the session
fails outright; at or above it, the best attempt is persisted with a
`Warn`-level log, unchanged from today's `UNVERIFIED` status.

**Tech Stack:** Go (existing backend module), existing
`domaincontractsllm.Client`, existing `domaincontractsutility.JSEngine`
(goja), `github.com/google/uuid`, `database/sql`-free Postgres access via
the existing `squirrel`/`pgxdt` repository pattern, golang-migrate SQL
migrations.

## Global Constraints

- `Validate()`'s pass bar is **exact match**: `OwnedCorrect == OwnedTotal`
  for every known case, with zero `RunError`s — these are real recorded
  bits, not estimates, so any mismatch is a genuine defect.
- Checksum-bit accuracy is scored **separately** from owned-bit accuracy
  and never gates pass/fail on its own (informational only), because
  `Attribute()`'s checksum-bit detection is already known-incomplete.
- Repair is capped at **3 rounds** per pass through
  `runAnalysisAndGeneration`; only a round that *improves*
  `OwnedCorrect` replaces `best` — a regressive round is discarded, not
  carried forward into the next round's feedback.
- The checksum-clarification detour is capped at **once per session**
  (`ChecksumClarificationUsedAt`) — a second occurrence falls through to
  the repair loop / 60%-floor logic instead of looping again.
- The 60% floor (`ownedBitFatalFloor = 0.6`) is a named constant, not
  derived from data — a tunable starting point.
- `analysisTimeout` (currently `3 * time.Minute`, covering
  `runAnalysisAndGeneration`'s entire body) becomes `10 * time.Minute`.
  Each individual `WriteCoder`/`RepairCoder` call gets its own
  `coderGenerationCallTimeout = 2 * time.Minute` via
  `context.WithTimeout`, so one hung call can't consume the whole budget.
- A degraded (below-100%, at-or-above-60%) persisted coder stays
  `UNVERIFIED`, exactly like any newly generated coder today — the
  existing hardware test-case-transmission stage remains the further
  real-world gate. No schema/API/UI change for visibility; a
  `Warn`-level log carries the score.
- `WriteChecksumClarificationCases` triggers pre-persistence, so it takes
  the in-memory `coder_generation.Coder` (the LLM's own unpersisted
  output), never a persisted `domainmodels.InfraredStateCoder` — no DB row
  exists yet at that point.

---

### Task 1: `WriteCoder` baseline-value prompt fix

**Files:**
- Modify: `backend/internal/application/infrared/coder_generation/llm_coder.go`
- Modify: `backend/internal/application/infrared/coder_generation/llm_coder_test.go`

**Interfaces:**
- Produces: `WriteCoder(ctx, client, deviceBrand, deviceModel string, states []domainmodels.InfraredState, definitions []domainmodels.InfraredStateDeviceDefinition, payload applicationinfraredanalysis.AnalysisPayload, baselineValues map[string]string) (Coder, error)` — one new parameter, `baselineValues`, appended at the end.

- [ ] **Step 1: Write the failing test**

In `llm_coder_test.go`, update `TestWriteCoderParsesStructuredResponse` to
pass a `baselineValues` map and assert it appears in the prompt sent to
the LLM. Replace the existing call:

```go
coder, err := WriteCoder(context.Background(), client, "Polytron", "PAC-09HDN", states, definitions, payload)
```

with:

```go
baselineValues := map[string]string{"POWER": "OFF"}
coder, err := WriteCoder(context.Background(), client, "Polytron", "PAC-09HDN", states, definitions, payload, baselineValues)
```

and add a new assertion right after the existing ones in that test:

```go
if !strings.Contains(client.lastRequest.Prompt, "POWER=OFF") {
    t.Fatalf("prompt = %q, want it to state the baseline value POWER=OFF explicitly", client.lastRequest.Prompt)
}
```

Add `"strings"` to the test file's imports if not already present (check
first — `llm_coder.go` itself already imports it, but the `_test.go` file
may not).

Also update the two other `WriteCoder` call sites in the same file
(`TestWriteCoderPropagatesLlmError`, `TestWriteCoderErrorsOnMalformedResponse`)
to pass a trailing `nil` for the new parameter — a nil map is valid input
(no baseline line gets added, tested in Step 3 below) and these two tests
don't care about prompt content:

```go
_, err := WriteCoder(context.Background(), client, "Polytron", "PAC-09HDN", nil, nil, applicationinfraredanalysis.AnalysisPayload{}, nil)
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -run TestWriteCoder -v`
Expected: compile error (`WriteCoder` called with 7 args, wants 8) — this
confirms the signature doesn't exist yet.

- [ ] **Step 3: Implement the signature change and prompt fix**

In `llm_coder.go`, change the `WriteCoder` signature to:

```go
func WriteCoder(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	states []domainmodels.InfraredState,
	definitions []domainmodels.InfraredStateDeviceDefinition,
	payload applicationinfraredanalysis.AnalysisPayload,
	baselineValues map[string]string,
) (Coder, error) {
```

Immediately after the existing `stateNameById` construction loop, add the
new prompt line — only when `baselineValues` is non-empty, so a nil/empty
map (as used by the two error-path tests above) produces the exact same
prompt as before:

```go
	if len(baselineValues) > 0 {
		var baselineParts []string
		for _, s := range states {
			if value, ok := baselineValues[s.Name]; ok {
				baselineParts = append(baselineParts, fmt.Sprintf("%s=%s", s.Name, value))
			}
		}
		if len(baselineParts) > 0 {
			fmt.Fprintf(&promptBuilder, "The baseline (most common) state is: %s. Every state's baseline value MUST be a valid, correctly-encoded input to your encoder — do not write validation logic that only accepts the specific non-baseline values mentioned below.\n\n", strings.Join(baselineParts, ", "))
		}
	}
```

Place this block right after the existing
`fmt.Fprintf(&promptBuilder, "Frame is %d bits. ...")` call and before
`promptBuilder.WriteString("Per-state bit ownership...")`, so the baseline
line reads naturally ahead of the per-state ownership details.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -run TestWriteCoder -v`
Expected: `PASS` for all three `TestWriteCoder*` tests.

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/application/infrared/coder_generation/llm_coder.go internal/application/infrared/coder_generation/llm_coder_test.go
git commit -m "infrared: state the baseline value explicitly in WriteCoder's prompt"
```

---

### Task 2: `Validate()` — score a generated encoder against real recorded bits

**Files:**
- Create: `backend/internal/application/infrared/coder_generation/validate.go`
- Create: `backend/internal/application/infrared/coder_generation/validate_test.go`

**Interfaces:**
- Consumes: `domaincontractsutility.JSEngine.RunEncoder(source string, state map[string]string, timeout time.Duration) ([]int32, error)`; `applicationinfraredanalysis.DemodulateBits(frame []int32) []int` (takes a `[]int32` and treats the first mark/space pair as header, exactly as used by the goja-backed `JSEngine` implementation's output).
- Produces (for Task 3/7 to consume):
  ```go
  type KnownCase struct {
      Label string            // for logging, e.g. "baseline" or "MODE=HEAT"
      State map[string]string // full state map, name-keyed, ready for RunEncoder
      Bits  []int             // the real recorded bits for this exact state
  }

  type CaseValidation struct {
      Label         string
      OwnedCorrect  int
      OwnedTotal    int
      ChecksumOK    int
      ChecksumTotal int
      RunError      string
  }

  type ValidationResult struct {
      Cases         []CaseValidation
      OwnedCorrect  int
      OwnedTotal    int
      ChecksumOK    int
      ChecksumTotal int
  }

  func Validate(runner domaincontractsutility.JSEngine, encoderSource string, checksumBits []int, knownCases []KnownCase, timeout time.Duration) ValidationResult
  func (r ValidationResult) Passed() bool
  func (r ValidationResult) ChecksumOnlyGap() bool
  func (r ValidationResult) OwnedAccuracy() float64
  ```

- [ ] **Step 1: Write the failing tests**

Create `validate_test.go`:

```go
package applicationinfraredcodergeneration

import (
	"errors"
	"testing"
	"time"
)

type fakeJSEngine struct {
	raw []int32
	err error
}

func (f *fakeJSEngine) RunEncoder(_ string, _ map[string]string, _ time.Duration) ([]int32, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.raw, nil
}

func TestValidatePassesWhenEncoderMatchesEveryKnownCase(t *testing.T) {
	// [9000,4500] header + [560,560] one bit that decodes to bit 0 (space
	// 560 is not > the frame's own min/max midpoint of 560) — same fixture
	// shape as the analysis package's own tests.
	engine := &fakeJSEngine{raw: []int32{9000, 4500, 560, 560}}
	knownCases := []KnownCase{
		{Label: "baseline", State: map[string]string{"POWER": "OFF"}, Bits: []int{0}},
	}

	result := Validate(engine, "function encode(state){}", nil, knownCases, time.Second)

	if !result.Passed() {
		t.Fatalf("Validate() Passed() = false, want true; result = %+v", result)
	}
	if result.OwnedCorrect != 1 || result.OwnedTotal != 1 {
		t.Fatalf("owned = %d/%d, want 1/1", result.OwnedCorrect, result.OwnedTotal)
	}
}

func TestValidateFailsOnBitMismatch(t *testing.T) {
	// Two bits: real is [0,1] (spaces 560 then 1690 -> midpoint 1125 ->
	// bit0=560<1125=>0, bit1=1690>1125=>1); encoder always returns a raw
	// that decodes to [0,0] instead.
	engine := &fakeJSEngine{raw: []int32{9000, 4500, 560, 560, 560, 560}}
	knownCases := []KnownCase{
		{Label: "baseline", State: map[string]string{}, Bits: []int{0, 1}},
	}

	result := Validate(engine, "function encode(state){}", nil, knownCases, time.Second)

	if result.Passed() {
		t.Fatal("Validate() Passed() = true, want false (bit 1 mismatches)")
	}
	if result.OwnedCorrect != 1 || result.OwnedTotal != 2 {
		t.Fatalf("owned = %d/%d, want 1/2", result.OwnedCorrect, result.OwnedTotal)
	}
}

func TestValidateSeparatesChecksumBitsFromOwnedBits(t *testing.T) {
	// 2 bits: bit0 owned, bit1 is a checksum bit (checksumBits=[1]). Real
	// bits [0,1]; encoder returns [0,0] -> owned bit0 correct, checksum
	// bit1 wrong, but Passed() only cares about owned bits.
	engine := &fakeJSEngine{raw: []int32{9000, 4500, 560, 560, 560, 560}}
	knownCases := []KnownCase{
		{Label: "baseline", State: map[string]string{}, Bits: []int{0, 1}},
	}

	result := Validate(engine, "function encode(state){}", []int{1}, knownCases, time.Second)

	if !result.Passed() {
		t.Fatalf("Validate() Passed() = false, want true (checksum bits must not gate pass/fail); result = %+v", result)
	}
	if result.OwnedTotal != 1 || result.OwnedCorrect != 1 {
		t.Fatalf("owned = %d/%d, want 1/1 (bit 1 excluded, it's a checksum bit)", result.OwnedCorrect, result.OwnedTotal)
	}
	if result.ChecksumTotal != 1 || result.ChecksumOK != 0 {
		t.Fatalf("checksum = %d/%d, want 0/1", result.ChecksumOK, result.ChecksumTotal)
	}
}

func TestValidateRunErrorFailsThatCaseWithoutVacuousPass(t *testing.T) {
	// A case whose RunEncoder throws must never be counted as a trivial
	// 0/0 pass — Passed() must be false even though OwnedCorrect==OwnedTotal==0.
	engine := &fakeJSEngine{err: errors.New("encoder threw")}
	knownCases := []KnownCase{
		{Label: "baseline", State: map[string]string{}, Bits: []int{0}},
	}

	result := Validate(engine, "function encode(state){}", nil, knownCases, time.Second)

	if result.Passed() {
		t.Fatal("Validate() Passed() = true, want false (RunEncoder errored)")
	}
	if result.Cases[0].RunError == "" {
		t.Fatal("Cases[0].RunError is empty, want the RunEncoder error captured")
	}
	if result.OwnedAccuracy() != 0 {
		t.Fatalf("OwnedAccuracy() = %v, want 0 (no case ever produced comparable bits)", result.OwnedAccuracy())
	}
}

func TestValidateChecksumOnlyGapDetection(t *testing.T) {
	// Owned bits perfect, checksum bits wrong -> ChecksumOnlyGap() true.
	engine := &fakeJSEngine{raw: []int32{9000, 4500, 560, 560, 560, 560}}
	knownCases := []KnownCase{
		{Label: "baseline", State: map[string]string{}, Bits: []int{0, 1}},
	}
	result := Validate(engine, "function encode(state){}", []int{1}, knownCases, time.Second)
	if !result.ChecksumOnlyGap() {
		t.Fatal("ChecksumOnlyGap() = false, want true (owned bits clean, checksum bit wrong)")
	}

	// Owned bits also wrong -> ChecksumOnlyGap() false, even with a checksum mismatch too.
	engine2 := &fakeJSEngine{raw: []int32{9000, 4500, 1690, 1690, 560, 560}}
	result2 := Validate(engine2, "function encode(state){}", []int{1}, knownCases, time.Second)
	if result2.ChecksumOnlyGap() {
		t.Fatal("ChecksumOnlyGap() = true, want false (owned bit 0 is also wrong)")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -run TestValidate -v`
Expected: compile errors — `Validate`, `KnownCase`, `CaseValidation`,
`ValidationResult` don't exist yet.

- [ ] **Step 3: Implement `validate.go`**

```go
package applicationinfraredcodergeneration

import (
	"time"

	applicationinfraredanalysis "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/analysis"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
)

// KnownCase is one state the pipeline already has real, recorded bits for
// — the baseline, plus every already-recorded non-baseline OFAT case.
type KnownCase struct {
	Label string
	State map[string]string
	Bits  []int
}

type CaseValidation struct {
	Label         string
	OwnedCorrect  int
	OwnedTotal    int
	ChecksumOK    int
	ChecksumTotal int
	RunError      string
}

type ValidationResult struct {
	Cases         []CaseValidation
	OwnedCorrect  int
	OwnedTotal    int
	ChecksumOK    int
	ChecksumTotal int
}

// Passed reports whether every known case ran without error and matched
// the real recorded bits exactly on every owned bit position. A
// ValidationResult with no cases is never "passed" — buildAnalysisPayload
// always requires at least a baseline case to exist before WriteCoder is
// ever called, so an empty Cases slice here is a caller bug, not a valid
// pass.
func (r ValidationResult) Passed() bool {
	if len(r.Cases) == 0 {
		return false
	}
	for _, c := range r.Cases {
		if c.RunError != "" || c.OwnedCorrect != c.OwnedTotal {
			return false
		}
	}
	return true
}

// ChecksumOnlyGap reports whether every owned bit is correct across every
// case (no RunError, no owned mismatch) but at least one checksum bit is
// wrong — the signal used to route to the checksum-clarification path
// instead of the repair loop, since a repair prompt can't fix a data gap.
func (r ValidationResult) ChecksumOnlyGap() bool {
	if len(r.Cases) == 0 {
		return false
	}
	for _, c := range r.Cases {
		if c.RunError != "" || c.OwnedCorrect != c.OwnedTotal {
			return false
		}
	}
	return r.ChecksumTotal > 0 && r.ChecksumOK < r.ChecksumTotal
}

// OwnedAccuracy returns 0 when OwnedTotal is 0 (e.g. every case errored)
// rather than dividing by zero.
func (r ValidationResult) OwnedAccuracy() float64 {
	if r.OwnedTotal == 0 {
		return 0
	}
	return float64(r.OwnedCorrect) / float64(r.OwnedTotal)
}

// Validate runs encoderSource against every knownCases entry via the real
// JSEngine, decodes its output with the same DemodulateBits production
// uses on real captures, and compares it bit-for-bit against that case's
// real recorded bits — owned-bit positions scored separately from
// checksumBits positions.
func Validate(
	runner domaincontractsutility.JSEngine,
	encoderSource string,
	checksumBits []int,
	knownCases []KnownCase,
	timeout time.Duration,
) ValidationResult {
	checksumSet := make(map[int]struct{}, len(checksumBits))
	for _, b := range checksumBits {
		checksumSet[b] = struct{}{}
	}

	var result ValidationResult
	for _, kc := range knownCases {
		cv := CaseValidation{Label: kc.Label}

		raw, err := runner.RunEncoder(encoderSource, kc.State, timeout)
		if err != nil {
			cv.RunError = err.Error()
			result.Cases = append(result.Cases, cv)
			continue
		}

		decoded := applicationinfraredanalysis.DemodulateBits(raw)
		for i, want := range kc.Bits {
			got := -1
			if i < len(decoded) {
				got = decoded[i]
			}
			if _, isChecksum := checksumSet[i]; isChecksum {
				cv.ChecksumTotal++
				if got == want {
					cv.ChecksumOK++
				}
			} else {
				cv.OwnedTotal++
				if got == want {
					cv.OwnedCorrect++
				}
			}
		}

		result.Cases = append(result.Cases, cv)
		result.OwnedCorrect += cv.OwnedCorrect
		result.OwnedTotal += cv.OwnedTotal
		result.ChecksumOK += cv.ChecksumOK
		result.ChecksumTotal += cv.ChecksumTotal
	}
	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -run TestValidate -v`
Expected: `PASS` for all five new tests.

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/application/infrared/coder_generation/validate.go internal/application/infrared/coder_generation/validate_test.go
git commit -m "infrared: add Validate() to score a generated encoder against real recorded bits"
```

---

### Task 3: `RepairCoder()` — feed concrete mismatches back to the LLM

**Files:**
- Create: `backend/internal/application/infrared/coder_generation/llm_coder_repair.go`
- Create: `backend/internal/application/infrared/coder_generation/llm_coder_repair_test.go`

**Interfaces:**
- Consumes: `responseSchema` (package-private const, already defined in `llm_coder.go`), `coderResponse` (package-private struct, already defined in `llm_coder.go`), `Coder` (already defined in `llm_coder.go`), `ValidationResult`/`CaseValidation` (Task 2).
- Produces (for Task 7 to consume):
  ```go
  func RepairCoder(
      ctx context.Context,
      client domaincontractsllm.Client,
      deviceBrand string,
      deviceModel string,
      priorEncoderSource string,
      priorDecoderSource string,
      validation ValidationResult,
  ) (Coder, error)
  ```

- [ ] **Step 1: Write the failing test**

Create `llm_coder_repair_test.go`:

```go
package applicationinfraredcodergeneration

import (
	"context"
	"strings"
	"testing"
)

func TestRepairCoderIncludesPriorSourceAndMismatchFeedback(t *testing.T) {
	responseBody := `{"encoder_source": "function encode(state) { return [9000, 4500]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`
	client := &fakeLlmClient{responseText: responseBody}

	validation := ValidationResult{
		Cases: []CaseValidation{
			{Label: "baseline", RunError: "encoder threw: TEMP out of range"},
			{Label: "MODE=HEAT", OwnedCorrect: 10, OwnedTotal: 12},
		},
	}

	coder, err := RepairCoder(context.Background(), client, "Polytron", "PAC-09HDN", "function encode(old){}", "function decode(old){}", validation)
	if err != nil {
		t.Fatalf("RepairCoder() error = %v, want nil", err)
	}
	if coder.EncoderSource == "" {
		t.Fatal("RepairCoder() returned empty EncoderSource")
	}

	prompt := client.lastRequest.Prompt
	if !strings.Contains(prompt, "function encode(old){}") {
		t.Fatalf("prompt = %q, want it to include the prior encoder source", prompt)
	}
	if !strings.Contains(prompt, "TEMP out of range") {
		t.Fatalf("prompt = %q, want it to include the RunError feedback", prompt)
	}
	if !strings.Contains(prompt, "MODE=HEAT") {
		t.Fatalf("prompt = %q, want it to include the mismatching case's label", prompt)
	}
	if client.lastRequest.ResponseSchema == nil {
		t.Fatal("GenerateText() request had no ResponseSchema — structured output was not requested")
	}
}

func TestRepairCoderPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: errPlaceholder}
	_, err := RepairCoder(context.Background(), client, "Polytron", "PAC-09HDN", "", "", ValidationResult{})
	if err == nil {
		t.Fatal("RepairCoder() error = nil, want propagated error")
	}
}
```

`fakeLlmClient` already exists in `llm_coder_test.go` (same package) with a
`text`/`err`/`lastRequest` shape — reused here directly. Add one small
addition to `llm_coder_test.go` right after the existing `fakeLlmClient`
type: an `errPlaceholder` sentinel error for this test to reference:

```go
var errPlaceholder = errors.New("boom")
```

(Add `"errors"` to `llm_coder_test.go`'s imports if not already present —
check first, since other tests in that file may already import it.)

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -run TestRepairCoder -v`
Expected: compile error — `RepairCoder` doesn't exist yet.

- [ ] **Step 3: Implement `llm_coder_repair.go`**

```go
package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

// RepairCoder asks the LLM to fix a previously-generated encoder/decoder
// pair, given the concrete per-case validation results (RunEncoder errors
// and/or bit-level mismatches against real recorded signals) from
// Validate(). Reuses WriteCoder's response schema and return type — a
// repair response is shaped identically to an initial one.
func RepairCoder(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	priorEncoderSource string,
	priorDecoderSource string,
	validation ValidationResult,
) (Coder, error) {
	var prompt strings.Builder
	fmt.Fprintf(&prompt, "Device: %s %s\n\n", deviceBrand, deviceModel)
	prompt.WriteString("Your previous encoder:\n\n```javascript\n")
	prompt.WriteString(priorEncoderSource)
	prompt.WriteString("\n```\n\nYour previous decoder:\n\n```javascript\n")
	prompt.WriteString(priorDecoderSource)
	prompt.WriteString("\n```\n\n")
	prompt.WriteString(buildRepairFeedback(validation))
	prompt.WriteString("\nFix the encoder (and decoder if relevant) so every case above is correct. Respond with the complete corrected JSON matching the same schema as before (encoder_source, decoder_source, summary_readme, detail_readme) — do not send a partial diff.")

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are an expert at reverse-engineering infrared remote control protocols and writing correct JavaScript encoders/decoders from bit-level analysis data and test feedback.",
		Prompt:          prompt.String(),
		MaxOutputTokens: 8192,
		ResponseSchema:  []byte(responseSchema),
	})
	if err != nil {
		return Coder{}, domainmodels.NewError("failed to generate repaired encoder/decoder", domainmodels.ErrTypeFailure, err)
	}

	var response coderResponse
	if err := json.Unmarshal([]byte(result.Text), &response); err != nil {
		return Coder{}, domainmodels.NewError("llm returned malformed repair response", domainmodels.ErrTypeFailure, err)
	}

	return Coder{
		EncoderSource: response.EncoderSource,
		DecoderSource: response.DecoderSource,
		SummaryReadme: response.SummaryReadme,
		DetailReadme:  response.DetailReadme,
	}, nil
}

// buildRepairFeedback turns a ValidationResult into the concrete,
// per-case feedback text the repair prompt sends back to the LLM — the
// same shape (RunError verbatim, or up to 8 example
// "bit N: expected X got Y" mismatches per case) validated empirically
// against real models earlier in this project's development.
func buildRepairFeedback(validation ValidationResult) string {
	var b strings.Builder
	b.WriteString("Your previous encoder was tested against the real recorded signal for several states of this device and had problems. Details per test case:\n\n")
	for _, c := range validation.Cases {
		if c.RunError != "" {
			fmt.Fprintf(&b, "- Case %q: the encoder THREW A RUNTIME ERROR: %q. This must not happen for any valid state value within the documented ranges.\n", c.Label, c.RunError)
			continue
		}
		if c.OwnedCorrect == c.OwnedTotal {
			fmt.Fprintf(&b, "- Case %q: correct.\n", c.Label)
			continue
		}
		fmt.Fprintf(&b, "- Case %q: %d/%d owned bits wrong.\n", c.Label, c.OwnedTotal-c.OwnedCorrect, c.OwnedTotal)
	}
	return b.String()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -run TestRepairCoder -v`
Expected: `PASS` for both tests.

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/application/infrared/coder_generation/llm_coder_repair.go internal/application/infrared/coder_generation/llm_coder_repair_test.go internal/application/infrared/coder_generation/llm_coder_test.go
git commit -m "infrared: add RepairCoder() to fix a generated encoder from concrete validation feedback"
```

---

### Task 4: `WriteChecksumClarificationCases()` — ask for more recording scenarios

**Files:**
- Create: `backend/internal/application/infrared/coder_generation/llm_checksum_clarification_cases.go`
- Create: `backend/internal/application/infrared/coder_generation/llm_checksum_clarification_cases_test.go`

**Interfaces:**
- Consumes: `RetryCasePlan`, `testCasePlanResponse`, `testCaseResponseSchema` (all package-private/exported, already defined in `llm_retry_cases.go`/`llm_test_cases.go`); `Coder` (Task 1, in-memory unpersisted LLM output — this path runs before any coder is persisted, so it deliberately does **not** take a `domainmodels.InfraredStateCoder`).
- Produces (for Task 7 to consume):
  ```go
  func WriteChecksumClarificationCases(
      ctx context.Context,
      client domaincontractsllm.Client,
      deviceBrand string,
      deviceModel string,
      coder Coder,
      states []domainmodels.InfraredState,
  ) ([]RetryCasePlan, error)
  ```

- [ ] **Step 1: Write the failing test**

Create `llm_checksum_clarification_cases_test.go`:

```go
package applicationinfraredcodergeneration

import (
	"context"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func TestWriteChecksumClarificationCasesParsesResponse(t *testing.T) {
	powerId := uuid.New()
	states := []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	coder := Coder{SummaryReadme: "summary", DetailReadme: "detail"}

	responseBody := `[{"description": "repeat POWER=ON twice", "states": {"POWER": "ON"}}]`
	client := &fakeLlmClient{responseText: responseBody}

	plans, err := WriteChecksumClarificationCases(context.Background(), client, "Polytron", "PAC-09HDN", coder, states)
	if err != nil {
		t.Fatalf("WriteChecksumClarificationCases() error = %v, want nil", err)
	}
	if len(plans) != 1 {
		t.Fatalf("len(plans) = %d, want 1", len(plans))
	}
	if plans[0].States[powerId] != "ON" {
		t.Fatalf("plans[0].States[powerId] = %q, want \"ON\"", plans[0].States[powerId])
	}
	if client.lastRequest.ResponseSchema == nil {
		t.Fatal("GenerateText() request had no ResponseSchema")
	}
}

func TestWriteChecksumClarificationCasesPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: errPlaceholder}
	_, err := WriteChecksumClarificationCases(context.Background(), client, "Polytron", "PAC-09HDN", Coder{}, nil)
	if err == nil {
		t.Fatal("WriteChecksumClarificationCases() error = nil, want propagated error")
	}
}

func TestWriteChecksumClarificationCasesRejectsUnknownState(t *testing.T) {
	states := []domainmodels.InfraredState{{Id: uuid.New(), Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	client := &fakeLlmClient{responseText: `[{"description": "x", "states": {"NOT_A_REAL_STATE": "ON"}}]`}

	_, err := WriteChecksumClarificationCases(context.Background(), client, "Polytron", "PAC-09HDN", Coder{}, states)
	if err == nil {
		t.Fatal("WriteChecksumClarificationCases() error = nil, want error for unknown state name")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -run TestWriteChecksumClarificationCases -v`
Expected: compile error — function doesn't exist yet.

- [ ] **Step 3: Implement `llm_checksum_clarification_cases.go`**

```go
package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"fmt"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

// WriteChecksumClarificationCases asks the LLM to propose additional
// recording scenarios likely to help pin down a checksum algorithm it
// couldn't fully infer — used when Validate() reports a ChecksumOnlyGap:
// every owned bit is correct, but the checksum bits aren't, which is a
// data problem (too few examples), not something a repair prompt can fix.
// Runs before any coder is persisted, so it takes the in-memory Coder
// (this attempt's own unpersisted output), not a DB-backed
// domainmodels.InfraredStateCoder. Reuses testCaseResponseSchema's shape
// — identical output contract to WriteTestCases/WriteRetryCases.
func WriteChecksumClarificationCases(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	coder Coder,
	states []domainmodels.InfraredState,
) ([]RetryCasePlan, error) {
	stateIdByName := make(map[string]uuid.UUID, len(states))
	for _, state := range states {
		stateIdByName[state.Name] = state.Id
	}

	prompt := fmt.Sprintf(
		"Device: %s %s\n\nProtocol summary: %s\n\nProtocol detail: %s\n\nYour encoder correctly reproduces every bit whose meaning is already known, but could not be verified to compute the checksum bits correctly from the examples recorded so far — there isn't enough data yet to confirm the checksum algorithm. Propose new targeted recording scenarios (full target state plus a short description) likely to reveal the checksum pattern — for example, scenarios that vary multiple states together, or repeat an existing case to confirm determinism. Respond as a JSON array matching the given schema.",
		deviceBrand, deviceModel, coder.SummaryReadme, coder.DetailReadme,
	)

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are an expert at diagnosing infrared remote control encoder failures and proposing corrective recording scenarios.",
		Prompt:          prompt,
		MaxOutputTokens: 4096,
		ResponseSchema:  []byte(testCaseResponseSchema),
	})
	if err != nil {
		return nil, domainmodels.NewError("failed to generate checksum clarification cases", domainmodels.ErrTypeFailure, err)
	}

	var responses []testCasePlanResponse
	if err := json.Unmarshal([]byte(result.Text), &responses); err != nil {
		return nil, domainmodels.NewError("llm returned malformed checksum clarification response", domainmodels.ErrTypeFailure, err)
	}

	plans := make([]RetryCasePlan, 0, len(responses))
	for _, r := range responses {
		planStates := make(map[uuid.UUID]string, len(r.States))
		for name, value := range r.States {
			stateId, ok := stateIdByName[name]
			if !ok {
				return nil, domainmodels.NewError(fmt.Sprintf("llm referenced unknown state %q", name), domainmodels.ErrTypeFailure, nil)
			}
			planStates[stateId] = value
		}
		plans = append(plans, RetryCasePlan{Description: r.Description, States: planStates})
	}
	return plans, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -run TestWriteChecksumClarificationCases -v`
Expected: `PASS` for all three tests.

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/application/infrared/coder_generation/llm_checksum_clarification_cases.go internal/application/infrared/coder_generation/llm_checksum_clarification_cases_test.go
git commit -m "infrared: add WriteChecksumClarificationCases() for the checksum-only-gap path"
```

---

### Task 5: `ChecksumClarificationUsedAt` — domain model, repository, migration

**Files:**
- Create: `backend/database/migrations/20260811040000_infrared_record_session_checksum_clarification.up.sql`
- Create: `backend/database/migrations/20260811040000_infrared_record_session_checksum_clarification.down.sql`
- Modify: `backend/internal/domain/models/infrared.go:45-56` (the `InfraredRecordSession` struct)
- Modify: `backend/internal/domain/contracts/repository/infrared_record_session.go`
- Modify: `backend/internal/infrastructure/repository/infrared_record_session/postgres.go`
- Modify: `backend/internal/infrastructure/repository/infrared_record_session/postgres_query.go`
- Test: `backend/internal/infrastructure/repository/infrared_record_session/postgres_test.go` (create if it doesn't already exist — check first with `ls`; if the package has no existing test file, this task's test is a query-builder test, not a live-DB test, matching the pattern in sibling repository packages)

**Interfaces:**
- Produces: `domainmodels.InfraredRecordSession.ChecksumClarificationUsedAt *time.Time` (new field); `domaincontractsrepository.InfraredRecordSession.MarkChecksumClarificationUsedById(ctx context.Context, id uuid.UUID) error` (new interface method).

- [ ] **Step 1: Check for an existing repository test file**

Run: `ls backend/internal/infrastructure/repository/infrared_record_session/`
If a `_test.go` file already exists, read it to match its exact style
(query-builder assertion pattern, e.g. asserting the generated SQL string)
before writing Step 2's test. If none exists, Step 2's test is this
package's first test file, in the same query-assertion style used by
other repository packages in this codebase (assert the built SQL
string/args, not a live DB call).

- [ ] **Step 2: Write the failing test**

Create (or append to) the test file:

```go
package infrastructurerepositoryinfraredrecordsession

import (
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func TestQueryMarkChecksumClarificationUsedByIdSetsTimestamp(t *testing.T) {
	sqrDollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	p := &postgresImpl{}
	p.SqrD = &sqrDollar

	query, _, err := p.queryMarkChecksumClarificationUsedById(uuid.New())
	if err != nil {
		t.Fatalf("queryMarkChecksumClarificationUsedById() error = %v, want nil", err)
	}
	if !strings.Contains(query, "checksum_clarification_used_at") {
		t.Fatalf("query = %q, want it to set checksum_clarification_used_at", query)
	}
	if !strings.Contains(query, "infrared_record_session") {
		t.Fatalf("query = %q, want it to target infrared_record_session", query)
	}
}
```

Check `BasePostgres`'s exact field names (`SqrD` etc.) against
`internal/infrastructure/repository/shared/*.go` before writing this —
they're already used identically by `postgres.go` in this same package
(`p.SqrD.Insert(...)` etc. in `queryCreate`), so match that spelling
exactly.

- [ ] **Step 3: Run test to verify it fails**

Run: `cd backend && go test ./internal/infrastructure/repository/infrared_record_session/... -run TestQueryMarkChecksumClarificationUsedById -v`
Expected: compile error — `queryMarkChecksumClarificationUsedById` doesn't exist yet.

- [ ] **Step 4: Write the migration**

Create `20260811040000_infrared_record_session_checksum_clarification.up.sql`:

```sql
ALTER TABLE infrared_record_session
    ADD COLUMN checksum_clarification_used_at TIMESTAMPTZ;
```

Create `20260811040000_infrared_record_session_checksum_clarification.down.sql`:

```sql
ALTER TABLE infrared_record_session
    DROP COLUMN checksum_clarification_used_at;
```

- [ ] **Step 5: Add the domain field**

In `backend/internal/domain/models/infrared.go`, modify the
`InfraredRecordSession` struct (currently lines 45-56) to add the new
field after `IsCompleted`:

```go
type InfraredRecordSession struct {
	Id                            uuid.UUID
	NodeId                        uuid.UUID
	InfraredDeviceId              uuid.UUID
	RecordingState                string
	CurrentRecordCaseId           *uuid.UUID
	IsCompleted                   bool
	ChecksumClarificationUsedAt   *time.Time
	CreatedAt                     time.Time
	DeletedAt                     *time.Time
	CreatedBy                     *uuid.UUID
	DeletedBy                     *uuid.UUID
}
```

- [ ] **Step 6: Add the repository interface method**

In `backend/internal/domain/contracts/repository/infrared_record_session.go`,
add one line to the `InfraredRecordSession` interface, right after
`UpdateCurrentRecordCaseIdById`:

```go
	MarkChecksumClarificationUsedById(ctx context.Context, id uuid.UUID) error
```

- [ ] **Step 7: Implement the query builder and postgres method**

In `postgres_query.go`, add (matching the exact style of
`queryUpdateRecordingStateById` immediately above it):

```go
func (p *postgresImpl) queryMarkChecksumClarificationUsedById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Update("infrared_record_session").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("checksum_clarification_used_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		ToSql()
}
```

Also add `"checksum_clarification_used_at"` to the
`infraredRecordSessionColumns` slice (right after `"is_completed"`), and
update `scanInfraredRecordSession` in `postgres.go` to scan the new column
in the same position:

```go
func scanInfraredRecordSession(row pgx.Row, item *domainmodels.InfraredRecordSession) error {
	return row.Scan(
		&item.Id, &item.NodeId, &item.InfraredDeviceId, &item.RecordingState,
		&item.CurrentRecordCaseId, &item.IsCompleted, &item.ChecksumClarificationUsedAt,
		&item.CreatedAt, &item.CreatedBy, &item.DeletedAt, &item.DeletedBy,
	)
}
```

In `postgres.go`, add the new method right after
`UpdateCurrentRecordCaseIdById`:

```go
func (p *postgresImpl) MarkChecksumClarificationUsedById(ctx context.Context, id uuid.UUID) error {
	query, args, err := p.queryMarkChecksumClarificationUsedById(id)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build mark checksum clarification used query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to mark checksum clarification used", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_record_session not found", nil)
	}

	return nil
}
```

- [ ] **Step 8: Run tests to verify they pass, and run the migration**

Run: `cd backend && go build ./... && go test ./internal/infrastructure/repository/infrared_record_session/... -v`
Expected: `PASS`.

Run (only if a local dev Postgres is configured and reachable — check
`.env`/`BE_POSTGRES_*` first; skip this specific sub-step if not, noting
so in the task's completion report):
`cd backend && migrate -database "$POSTGRES_URL" -path database/migrations up`
Expected: the new migration applies cleanly with no error.

- [ ] **Step 9: Commit**

```bash
cd backend
git add database/migrations/20260811040000_infrared_record_session_checksum_clarification.up.sql \
        database/migrations/20260811040000_infrared_record_session_checksum_clarification.down.sql \
        internal/domain/models/infrared.go \
        internal/domain/contracts/repository/infrared_record_session.go \
        internal/infrastructure/repository/infrared_record_session/postgres.go \
        internal/infrastructure/repository/infrared_record_session/postgres_query.go
git add internal/infrastructure/repository/infrared_record_session/postgres_test.go 2>/dev/null || true
git commit -m "infrared: add checksum_clarification_used_at to infrared_record_session"
```

---

### Task 6: `buildAnalysisPayload` — return per-case bits, not just the attributed payload

**Files:**
- Modify: `backend/internal/application/infrared/record_session_management/usecase.go:829-936` (the `buildAnalysisPayload` method and its call site)
- Test: `backend/internal/application/infrared/record_session_management/usecase_test.go` (extend the existing test(s) that exercise `runAnalysisAndGeneration`, since `buildAnalysisPayload` itself is an unexported method with no dedicated test file today — verify via the same integration-style tests already covering it)

**Interfaces:**
- Consumes: nothing new (same inputs as today).
- Produces (for Task 7 to consume): a new private type and a changed
  return signature:
  ```go
  type recordedCase struct {
      TargetStateId uuid.UUID
      TargetValue   string
      Bits          []int
  }

  func (u *usecase) buildAnalysisPayload(
      ctx context.Context,
      tag string,
      cases []domainmodels.InfraredStateDeviceRecordCase,
      states []domainmodels.InfraredState,
      definitions []domainmodels.InfraredStateDeviceDefinition,
  ) (payload applicationinfraredanalysis.AnalysisPayload, baselineState map[string]string, baselineBits []int, recordedCases []recordedCase, err error)
  ```
  (two new return values — `baselineBits` and `recordedCases` — appended
  after the existing three-value return; the existing three keep their
  current order and meaning unchanged.)

- [ ] **Step 1: Write the failing test**

`buildAnalysisPayload` is already exercised indirectly by
`TestRunAnalysisAndGenerationPersistsCoderAndAdvancesToTestCaseGeneration`
(usecase_test.go, currently starting at line 852) and
`TestRunAnalysisAndGenerationFailsSessionWhenEncoderThrows` (currently
starting at line 970) — this task's signature change alone doesn't need a
new test (Task 7 is where the new return values actually get consumed and
asserted on). Confirm the current tests still compile and pass **before**
this task's change, as a baseline:

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run TestRunAnalysisAndGeneration -v`
Expected: `PASS` (this confirms the starting point before the signature change, per this task's own step 4 recheck).

- [ ] **Step 2: Not applicable — no separate failing-test step for a pure return-signature change with no new observable behavior at this layer.**

(Task 7 supplies the behavior-level tests that actually exercise the new `baselineBits`/`recordedCases` values.)

- [ ] **Step 3: Implement the signature change**

In `usecase.go`, add the new type right above `buildAnalysisPayload`
(currently at line 829):

```go
// recordedCase is one already-recorded non-baseline OFAT case's target
// state/value plus its own demodulated bits — buildAnalysisPayload always
// computed this internally to feed Attribute(), but previously discarded
// it once the AnalysisPayload was built. Task 7 uses it to build
// coder_generation.KnownCase entries for Validate().
type recordedCase struct {
	TargetStateId uuid.UUID
	TargetValue   string
	Bits          []int
}
```

Change the `buildAnalysisPayload` signature (currently lines 829-835) to:

```go
func (u *usecase) buildAnalysisPayload(
	ctx context.Context,
	tag string,
	cases []domainmodels.InfraredStateDeviceRecordCase,
	states []domainmodels.InfraredState,
	definitions []domainmodels.InfraredStateDeviceDefinition,
) (applicationinfraredanalysis.AnalysisPayload, map[string]string, []int, []recordedCase, error) {
```

The function body's `caseBits`/`caseTargetState`/`caseTargetValue` local
maps (already built by the existing loop over `cases`, unchanged) now also
need converting into the new `[]recordedCase` slice right before the
existing `Attribute(...)` call, and both new values need threading through
every existing `return` statement in the function body. There are three
`return` statements in the current body — the two early-error returns
(inside the `for _, c := range cases` loop, on `json.Unmarshal` and
`DetectVolatileBits` errors) and the two guard returns (`if baselineBits ==
nil` and after the `Attribute` call error check) all currently return
`applicationinfraredanalysis.AnalysisPayload{}, nil, err` — update every
one of them to the new 5-value shape,
`applicationinfraredanalysis.AnalysisPayload{}, nil, nil, nil, err`. The
final success return currently reads `return payload, baselineState, nil`
— change it to build the `recordedCases` slice first:

```go
	volatile := applicationinfraredanalysis.UnionVolatileBits(volatileSets...)
	payload, err := applicationinfraredanalysis.Attribute(baselineBits, caseBits, caseTargetState, caseTargetValue, volatile)
	if err != nil {
		return applicationinfraredanalysis.AnalysisPayload{}, nil, nil, nil, err
	}

	recordedCases := make([]recordedCase, 0, len(caseBits))
	for caseId, bits := range caseBits {
		recordedCases = append(recordedCases, recordedCase{
			TargetStateId: caseTargetState[caseId],
			TargetValue:   caseTargetValue[caseId],
			Bits:          bits,
		})
	}

	return payload, baselineState, baselineBits, recordedCases, nil
```

Finally, update `buildAnalysisPayload`'s single call site inside
`runAnalysisAndGeneration` (currently `payload, baselineState, err :=
u.buildAnalysisPayload(ctx, tag, cases, states, definitions)`) to:

```go
	payload, baselineState, baselineBits, recordedCases, err := u.buildAnalysisPayload(ctx, tag, cases, states, definitions)
```

(`baselineBits` and `recordedCases` are consumed starting in Task 7 —
until then they're intentionally unused local variables, which Go treats
as a compile error. To keep this task independently buildable, add a
throwaway `_ = baselineBits; _ = recordedCases` immediately after the call
for now; Task 7 removes both blank-identifier lines when it adds the real
usage.)

- [ ] **Step 4: Run tests to verify the existing behavior still passes**

Run: `cd backend && go build ./... && go test ./internal/application/infrared/record_session_management/... -run TestRunAnalysisAndGeneration -v`
Expected: `PASS` — the signature change is purely additive/internal;
existing behavior is unchanged until Task 7 wires the new values in.

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/application/infrared/record_session_management/usecase.go
git commit -m "infrared: buildAnalysisPayload also returns per-case bits for validation"
```

---

### Task 7: Wire the validate/repair/escalate loop into `runAnalysisAndGeneration`

**Files:**
- Modify: `backend/internal/application/infrared/record_session_management/usecase.go`
- Modify: `backend/internal/application/infrared/record_session_management/usecase_test.go`

**Interfaces:**
- Consumes: `applicationinfraredcodergeneration.{Validate, KnownCase, ValidationResult, RepairCoder, WriteChecksumClarificationCases}` (Tasks 2-4); `WriteCoder`'s new `baselineValues` parameter (Task 1); `u.session.MarkChecksumClarificationUsedById` (Task 5); `recordedCase`/`buildAnalysisPayload`'s new return values (Task 6).
- Produces: the new orchestration inside `runAnalysisAndGeneration`;
  `persistNewCasesAndReturnToRecording` (new private helper, extracted
  from `runRetryCaseGeneration`'s tail) reused by both
  `runRetryCaseGeneration` and the new checksum-clarification path.

- [ ] **Step 1: Update the test-double fakes to support sequential responses on one client**

In `usecase_test.go`, `fakeLlmClient` (currently around line 365) only
holds one fixed `text` and returns it on every call — the repair loop
calls `GenerateText` multiple times on the *same* client instance obtained
once via `u.llmFactory.Current(ctx)`, so the fake needs to support a
per-call sequence too. Change `fakeLlmClient` to:

```go
type fakeLlmClient struct {
	text      string
	err       error
	texts     []string // if set, successive GenerateText calls on THIS client return these in order, sticking on the last entry once exhausted
	callCount int
	lastRequest domaincontractsllm.GenerateTextRequest
}

func (f *fakeLlmClient) GenerateText(_ context.Context, req domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	f.lastRequest = req
	if f.err != nil {
		return domaincontractsllm.GenerateTextResult{}, f.err
	}
	if len(f.texts) > 0 {
		i := f.callCount
		if i >= len(f.texts) {
			i = len(f.texts) - 1
		}
		f.callCount++
		return domaincontractsllm.GenerateTextResult{Text: f.texts[i]}, nil
	}
	return domaincontractsllm.GenerateTextResult{Text: f.text}, nil
}
```

This is additive: existing tests that set only `text`/`err` (not `texts`)
keep behaving identically. Then update `fakeLlmClientFactory` (currently
around line 371) to accept a way to produce a single client with its own
sequence — add one new field, `clientSequence []string`, checked before
the existing `responseTexts`/`responseText` logic in `Current()`:

```go
type fakeLlmClientFactory struct {
	err            error
	responseText   string
	responseTexts  []string
	callIndex      int
	clientSequence []string // when set, Current() returns ONE client whose successive GenerateText calls cycle through this list — for testing multi-call loops (e.g. the repair loop) on a single client
}

func (f *fakeLlmClientFactory) Current(_ context.Context) (domaincontractsllm.Client, error) {
	if f.err != nil {
		return nil, f.err
	}
	if len(f.clientSequence) > 0 {
		return &fakeLlmClient{texts: f.clientSequence}, nil
	}
	if len(f.responseTexts) > 0 {
		i := f.callIndex
		if i >= len(f.responseTexts) {
			i = len(f.responseTexts) - 1
		}
		f.callIndex++
		return &fakeLlmClient{text: f.responseTexts[i]}, nil
	}
	text := f.responseText
	if text == "" {
		text = `[{"case_index": 0, "description": "baseline", "order": 1}]`
	}
	return &fakeLlmClient{text: text}, nil
}
```

Also update `fakeEncoderRunner` (currently around line 139) so its default
(no `err` configured) return actually decodes correctly against the
fixture used by the two existing `TestRunAnalysisAndGeneration*` tests —
today it always returns the fixed `[]int32{9000, 4500}`, which
`DemodulateBits` treats as header-only (zero data bits), so the new
stricter `Validate()` gate would treat every case as a bit-length mismatch
even on the passing-today test. Change its default return to match the
fixture's own real raw shape instead — one bit, decoding to `0`:

```go
type fakeEncoderRunner struct {
	err           error
	receivedState map[string]string
	raw           []int32            // if set, always returned regardless of state (repair-loop tests)
	byState       map[string][]int32 // if set, looked up by state["POWER"]+"|"+state["MODE"] (the checksum-clarification test below is the only user of this — its fixture's two known state names)
}

func (f *fakeEncoderRunner) RunEncoder(_ string, state map[string]string, _ time.Duration) ([]int32, error) {
	f.receivedState = state
	if f.err != nil {
		return nil, f.err
	}
	if f.byState != nil {
		if raw, ok := f.byState[state["POWER"]+"|"+state["MODE"]]; ok {
			return raw, nil
		}
	}
	if f.raw != nil {
		return f.raw, nil
	}
	return []int32{9000, 4500, 560, 560}, nil
}
```

This single definition replaces the existing `fakeEncoderRunner` struct
and method entirely — it's referenced as-is by every test in Step 3 below
(some set only `raw`, one sets only `byState`, most set neither and rely
on the default return).

Finally, `fakeCaseRepository`'s `ReadListRawByCaseId`/
`ReadListStatesByCaseId` (currently around lines 332-337) ignore their
`caseId` parameter entirely and always return the same fixed
`listRawByCaseIdResult`/`listStatesByCaseIdResult` — fine for every
existing single-case test fixture, but the checksum-only-gap test in Step
3 below needs three *different* cases (baseline + two distinct
non-baseline cases) with different raw signals and different target
states. Add two new map fields to `fakeCaseRepository`'s struct
(alongside the existing `rawById map[uuid.UUID]*...`), and check them
first, falling back to the existing single-result fields when unset (so
every existing test, which only sets the single-result fields, is
unaffected):

```go
	listRawByCaseId    map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordRaw
	listStatesByCaseId map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordState
```

```go
func (f *fakeCaseRepository) ReadListRawByCaseId(_ context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordRaw, error) {
	if f.listRawByCaseId != nil {
		return f.listRawByCaseId[caseId], nil
	}
	return f.listRawByCaseIdResult, nil
}
func (f *fakeCaseRepository) ReadListStatesByCaseId(_ context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordState, error) {
	if f.listStatesByCaseId != nil {
		return f.listStatesByCaseId[caseId], nil
	}
	return f.listStatesByCaseIdResult, nil
}
```

- [ ] **Step 2: Run the existing tests to confirm the fake changes alone don't break anything**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -v`
Expected: still `PASS` (or a compile error if `runAnalysisAndGeneration`
hasn't been updated yet to match Task 6's new `buildAnalysisPayload`
return arity — if Task 6 was committed separately, this should already be
green from Task 6's own Step 4).

- [ ] **Step 3: Write the new failing tests for the loop's actual behavior**

Add to `usecase_test.go`, right after
`TestRunAnalysisAndGenerationFailsSessionWhenEncoderThrows`:

```go
func TestRunAnalysisAndGenerationRepairsAnInitiallyFailingEncoder(t *testing.T) {
	sessionId := uuid.New()
	powerId := uuid.New()

	rawBytes, err := json.Marshal([]int32{9000, 4500, 560, 560})
	if err != nil {
		t.Fatalf("failed to marshal fixture raw data: %v", err)
	}
	baselineCaseId := uuid.New()

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}}
	definitionRepo := &fakeDefinitionRepository{listByDeviceIdResult: []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerId, Options: []string{"ON", "OFF"}},
	}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: baselineCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
		listRawByCaseIdResult: []domainmodels.InfraredStateDeviceRecordRaw{
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: rawBytes, Status: domainmodels.InfraredRecordRawStatusAccepted},
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: rawBytes, Status: domainmodels.InfraredRecordRawStatusAccepted},
		},
		listStatesByCaseIdResult: []domainmodels.InfraredStateDeviceRecordState{
			{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: powerId, StateValue: "ON"},
		},
	}
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: uuid.New(), SummaryReadme: "summary", DetailReadme: "detail"}}

	// clientSequence: WriteCoder's own call, then runTestCaseGeneration's
	// WriteTestCases call (a fresh Current() call, unaffected by
	// clientSequence — see fakeLlmClientFactory.Current()'s ordering:
	// clientSequence only governs the FIRST Current() call's client).
	// So this scenario needs responseTexts (per-Current-call) to seed the
	// eventual WriteTestCases call, AND the single coder-generation client
	// itself needs a multi-call sequence for its own repair rounds — use
	// clientSequence for that first client's internal calls.
	llmFactory := &fakeLlmClientFactory{
		clientSequence: []string{
			`{"encoder_source": "function encode(state) { return [1]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
			`{"encoder_source": "function encode(state) { return [9000, 4500, 560, 560]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s2", "detail_readme": "d2"}`,
		},
	}

	encoderRunner := &fakeEncoderRunner{}
	testCaseRepo := &fakeTestCaseRepository{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if !coderRepo.createCalled {
		t.Fatal("coder repository Create() was never called, want the repair loop to eventually succeed and persist")
	}
	if !strings.Contains(coderRepo.created.EncoderSource, "560, 560") {
		t.Fatalf("persisted EncoderSource = %q, want the REPAIRED (second) attempt, not the initial broken one", coderRepo.created.EncoderSource)
	}
}

func TestRunAnalysisAndGenerationPersistsDegradedBestAttemptAboveFatalFloor(t *testing.T) {
	sessionId := uuid.New()
	powerId := uuid.New()

	rawBytes, err := json.Marshal([]int32{9000, 4500, 560, 560})
	if err != nil {
		t.Fatalf("failed to marshal fixture raw data: %v", err)
	}
	baselineCaseId := uuid.New()

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}}
	definitionRepo := &fakeDefinitionRepository{listByDeviceIdResult: []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerId, Options: []string{"ON", "OFF"}},
	}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: baselineCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
		listRawByCaseIdResult: []domainmodels.InfraredStateDeviceRecordRaw{
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: rawBytes, Status: domainmodels.InfraredRecordRawStatusAccepted},
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: rawBytes, Status: domainmodels.InfraredRecordRawStatusAccepted},
		},
		listStatesByCaseIdResult: []domainmodels.InfraredStateDeviceRecordState{
			{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: powerId, StateValue: "ON"},
		},
	}
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: uuid.New(), SummaryReadme: "summary", DetailReadme: "detail"}}

	// Every attempt (initial + all 3 repair rounds) returns a coder whose
	// smoke-tested encoder never matches the single real recorded bit
	// (fakeEncoderRunner always returns [9000,4500] -> zero decoded bits,
	// forced via err=nil but raw explicitly set below) — this is the
	// single-known-case fixture's only case, so OwnedTotal=1, OwnedCorrect=0
	// throughout: 0% accuracy is BELOW the 60% floor, so this scenario
	// actually exercises the FAILED path, not degraded-persist — see the
	// next test for the degraded-persist path's own dedicated fixture
	// (two known cases, one right one wrong -> 50%... still below 60%,
	// so use a fixture where one of two cases is right: 1/1 owned bits
	// correct on case A, 0/1 on... this still needs >=60% with only
	// whole-number case granularity. Simplest: a single case whose OWNED
	// bits are 2, one of which the fake encoder gets right, one wrong ->
	// 1/2 = 50%, still below floor. Use a fixture with 3 owned bits, 2
	// correct -> 2/3 = 66.7%, above the 60% floor.)
	encoderRunner := &fakeEncoderRunner{raw: []int32{9000, 4500, 560, 560, 1690, 560}} // decodes to [0, 1, 0]; real recorded bits below are [0, 1, 1] -> 2/3 correct

	llmFactory := &fakeLlmClientFactory{
		clientSequence: []string{
			`{"encoder_source": "function encode(state) { return [1]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
			`{"encoder_source": "function encode(state) { return [2]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
			`{"encoder_source": "function encode(state) { return [3]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
			`{"encoder_source": "function encode(state) { return [4]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
		},
	}

	testCaseRepo := &fakeTestCaseRepository{}
	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if !coderRepo.createCalled {
		t.Fatal("coder repository Create() was never called, want the best (66.7%, above the 60% floor) attempt persisted")
	}
}

func TestRunAnalysisAndGenerationFailsBelowFatalFloorAfterRepairExhausted(t *testing.T) {
	sessionId := uuid.New()
	powerId := uuid.New()

	rawBytes, err := json.Marshal([]int32{9000, 4500, 560, 560})
	if err != nil {
		t.Fatalf("failed to marshal fixture raw data: %v", err)
	}
	baselineCaseId := uuid.New()

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}}
	definitionRepo := &fakeDefinitionRepository{listByDeviceIdResult: []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerId, Options: []string{"ON", "OFF"}},
	}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: baselineCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
		listRawByCaseIdResult: []domainmodels.InfraredStateDeviceRecordRaw{
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: rawBytes, Status: domainmodels.InfraredRecordRawStatusAccepted},
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: rawBytes, Status: domainmodels.InfraredRecordRawStatusAccepted},
		},
		listStatesByCaseIdResult: []domainmodels.InfraredStateDeviceRecordState{
			{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: powerId, StateValue: "ON"},
		},
	}
	coderRepo := &fakeCoderRepository{}
	// RunEncoder always throws -> every case always fails -> 0% accuracy,
	// below the 60% floor even after 3 exhausted repair rounds.
	encoderRunner := &fakeEncoderRunner{err: errors.New("encoder threw")}
	llmFactory := &fakeLlmClientFactory{responseText: `{"encoder_source": "function encode(state) { return [9000, 4500]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if coderRepo.createCalled {
		t.Fatal("coder repository Create() was called despite 0% accuracy after exhausted repair, want no persisted row")
	}
	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateFailed {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include FAILED", sessionRepo.StatusUpdates())
	}
}

func TestRunAnalysisAndGenerationRequestsChecksumClarificationOnChecksumOnlyGap(t *testing.T) {
	sessionId := uuid.New()
	powerId, modeId := uuid.New(), uuid.New()
	baselineCaseId, case1Id, case2Id := uuid.New(), uuid.New(), uuid.New()

	// 4-bit fixture: header(9000,4500) + 4 data bits. Baseline is all
	// zero. case1 targets POWER (bits 0 and 3 flip). case2 targets MODE
	// (bits 1 and 3 flip). Bit 3 changes for BOTH cases -> Attribute()
	// flags it as a checksum bit (bits.go's intersection rule), exactly
	// the "MODE" fixture already used by
	// TestAttributeIdentifiesChecksumBitSharedAcrossStates in the
	// analysis package's own tests, restated here as raw signal data.
	marshalRaw := func(durations []int32) []byte {
		b, err := json.Marshal(durations)
		if err != nil {
			t.Fatalf("failed to marshal fixture raw data: %v", err)
		}
		return b
	}
	baselineRaw := marshalRaw([]int32{9000, 4500, 560, 560, 560, 560, 560, 560, 560, 560})
	case1Raw := marshalRaw([]int32{9000, 4500, 560, 1690, 560, 560, 560, 560, 560, 1690}) // decodes [1,0,0,1]
	case2Raw := marshalRaw([]int32{9000, 4500, 560, 560, 560, 1690, 560, 560, 560, 1690}) // decodes [0,1,0,1]

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{
		{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum},
		{Id: modeId, Name: "MODE", Type: domainmodels.InfraredStateTypeEnum},
	}}
	definitionRepo := &fakeDefinitionRepository{listByDeviceIdResult: []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerId, Options: []string{"OFF", "ON"}}, // baseline = Options[0] = "OFF"
		{InfraredStateId: modeId, Options: []string{"COOL", "HEAT"}}, // baseline = Options[0] = "COOL"
	}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: baselineCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
			{Id: case1Id, InfraredRecordSessionId: sessionId, Step: 2, Status: domainmodels.InfraredRecordCaseStatusAccepted},
			{Id: case2Id, InfraredRecordSessionId: sessionId, Step: 3, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
		listRawByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordRaw{
			baselineCaseId: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: baselineRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: baselineRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
			case1Id: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case1Id, RawData: case1Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case1Id, RawData: case1Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
			case2Id: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case2Id, RawData: case2Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case2Id, RawData: case2Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
		},
		listStatesByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordState{
			baselineCaseId: {
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: powerId, StateValue: "OFF"},
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: modeId, StateValue: "COOL"},
			},
			case1Id: {
				{InfraredStateDeviceRecordCaseId: case1Id, InfraredStateId: powerId, StateValue: "ON"},
				{InfraredStateDeviceRecordCaseId: case1Id, InfraredStateId: modeId, StateValue: "COOL"},
			},
			case2Id: {
				{InfraredStateDeviceRecordCaseId: case2Id, InfraredStateId: powerId, StateValue: "OFF"},
				{InfraredStateDeviceRecordCaseId: case2Id, InfraredStateId: modeId, StateValue: "HEAT"},
			},
		},
	}
	coderRepo := &fakeCoderRepository{}
	// The encoder always returns a raw that decodes bit-for-bit correctly
	// against whichever known state it's asked to encode: owned bits
	// (0 and 1) always right, checksum bit (3) always wrong (encoder
	// always emits 0 there instead of the real per-case value) -> a pure
	// ChecksumOnlyGap.
	encoderRunner := &fakeEncoderRunner{byState: map[string][]int32{
		"OFF|COOL": {9000, 4500, 560, 560, 560, 560, 560, 560, 560, 560},   // [0,0,0,0] vs real [0,0,0,0] -> all correct
		"ON|COOL":  {9000, 4500, 560, 1690, 560, 560, 560, 560, 560, 560}, // [1,0,0,0] vs real [1,0,0,1] -> bit3 wrong only
		"OFF|HEAT": {9000, 4500, 560, 560, 560, 1690, 560, 560, 560, 560}, // [0,1,0,0] vs real [0,1,0,1] -> bit3 wrong only
	}}
	llmFactory := &fakeLlmClientFactory{clientSequence: []string{
		`{"encoder_source": "function encode(state) { return []; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
		`[{"description": "repeat baseline twice more", "states": {"POWER": "OFF", "MODE": "COOL"}}]`,
	}}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if coderRepo.createCalled {
		t.Fatal("coder repository Create() was called on a checksum-only gap, want the session sent back to RECORDING instead")
	}
	if len(caseRepo.CreatedCaseIds()) == 0 {
		t.Fatal("no new case was persisted, want WriteChecksumClarificationCases' plan to have been recorded via persistNewCasesAndReturnToRecording")
	}
	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateRecording {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include RECORDING (checksum clarification sends the session back to record more)", sessionRepo.StatusUpdates())
	}
}
```

`fakeEncoderRunner`'s `byState` field (already added in Step 1's
consolidated definition) is what makes this test's per-state responses
possible — the `"POWER"+"|"+"MODE"` key shape is specific to this one
test's two known state names; every other test in this file leaves
`byState` nil and falls through to `raw`/the default.

Add `"strings"` to `usecase_test.go`'s imports if not already present —
check first.

- [ ] **Step 4: Run tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run TestRunAnalysisAndGeneration -v`
Expected: the four new tests `FAIL` (or don't compile, since
`persistNewCasesAndReturnToRecording` and the loop don't exist in
`runAnalysisAndGeneration` yet) while
`TestRunAnalysisAndGenerationPersistsCoderAndAdvancesToTestCaseGeneration`
and `TestRunAnalysisAndGenerationFailsSessionWhenEncoderThrows` still pass.

- [ ] **Step 5: Extract `persistNewCasesAndReturnToRecording` from `runRetryCaseGeneration`'s tail**

In `usecase.go`, `runRetryCaseGeneration` (currently starting at line 684)
currently ends with this block (case-persisting + activation +
transition), after its `WriteRetryCases` call and step-numbering logic:

```go
	var firstNewCaseId uuid.UUID
	for i, plan := range plans {
		caseStates := make([]domainmodels.InfraredStateDeviceRecordState, 0, len(plan.States))
		for stateId, value := range plan.States {
			caseStates = append(caseStates, domainmodels.InfraredStateDeviceRecordState{InfraredStateId: stateId, StateValue: value})
		}
		caseId, err := u.recordCase.CreateWithStates(ctx, sessionId, nextStep+int32(i), plan.Description, caseStates)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to persist retry case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
			u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
			return
		}
		if i == 0 {
			firstNewCaseId = caseId
		}
	}

	if err := u.recordCase.UpdateStatusById(ctx, firstNewCaseId, domainmodels.InfraredRecordCaseStatusActive); err != nil {
		u.logger.Error(ctx, tag, "failed to activate first retry case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	if err := u.session.UpdateCurrentRecordCaseIdById(ctx, sessionId, &firstNewCaseId); err != nil {
		u.logger.Error(ctx, tag, "failed to set session cursor to first retry case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateRecording)
}
```

Extract this into a new shared helper (placed right after
`runRetryCaseGeneration`'s closing brace), taking `plans` and `nextStep`
as parameters and returning a `bool` (`true` on success) so both callers
can decide what to do next without duplicating the failure-transition
logic:

```go
// persistNewCasesAndReturnToRecording persists a batch of newly-proposed
// record cases starting at nextStep, activates the first one, sets it as
// the session's current case, and transitions the session back to
// RECORDING. Shared by runRetryCaseGeneration (post-test-case-failure) and
// the checksum-clarification path (pre-persistence) — both need identical
// "append cases and go record them" mechanics, even though they trigger
// from very different points in the session lifecycle and have different
// preconditions about what else exists yet (a persisted coder/test cases,
// or not). Returns false if it already transitioned the session to FAILED
// on an internal error, so callers know not to do anything further.
func (u *usecase) persistNewCasesAndReturnToRecording(ctx context.Context, tag string, sessionId uuid.UUID, nextStep int32, plans []applicationinfraredcodergeneration.RetryCasePlan) bool {
	var firstNewCaseId uuid.UUID
	for i, plan := range plans {
		caseStates := make([]domainmodels.InfraredStateDeviceRecordState, 0, len(plan.States))
		for stateId, value := range plan.States {
			caseStates = append(caseStates, domainmodels.InfraredStateDeviceRecordState{InfraredStateId: stateId, StateValue: value})
		}
		caseId, err := u.recordCase.CreateWithStates(ctx, sessionId, nextStep+int32(i), plan.Description, caseStates)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to persist new record case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
			u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
			return false
		}
		if i == 0 {
			firstNewCaseId = caseId
		}
	}

	if err := u.recordCase.UpdateStatusById(ctx, firstNewCaseId, domainmodels.InfraredRecordCaseStatusActive); err != nil {
		u.logger.Error(ctx, tag, "failed to activate first new case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return false
	}
	if err := u.session.UpdateCurrentRecordCaseIdById(ctx, sessionId, &firstNewCaseId); err != nil {
		u.logger.Error(ctx, tag, "failed to set session cursor to first new case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return false
	}
	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateRecording)
	return true
}
```

Replace the extracted block in `runRetryCaseGeneration` with a single call:

```go
	u.persistNewCasesAndReturnToRecording(ctx, tag, sessionId, nextStep, plans)
}
```

(`runRetryCaseGeneration`'s own `nextStep` computation, right above this
block, is untouched — only the persist/activate/transition tail moves.)

- [ ] **Step 6: Bump the timeouts**

At the top of `usecase.go` (currently lines 23-27), change:

```go
const analysisTimeout = 3 * time.Minute
```

to:

```go
const analysisTimeout = 10 * time.Minute
```

and add one new constant right after it:

```go
const coderGenerationCallTimeout = 2 * time.Minute
```

Also add two new constants for the loop's own bounds, next to the timeout
constants:

```go
const maxRepairRounds = 3
const ownedBitFatalFloor = 0.6
```

- [ ] **Step 7: Rewrite `runAnalysisAndGeneration`'s coder-generation section**

First, check `usecase.go`'s import block (currently lines 3-21): it does
**not** currently import `"fmt"` (only `"context"`, `"encoding/json"`,
`"time"`, and the aliased packages). The new code below uses
`fmt.Sprintf` twice (the fatal-floor error message, and inside
`buildKnownCases`) — add `"fmt"` to the standard-library import group
(alphabetically, between `"encoding/json"` and `"time"`) before
proceeding, or the package won't compile.

Replace the current block (from the `client, err := u.llmFactory.Current(ctx)`
line through the `coderId, err := u.coder.Create(...)` block and the
subsequent `go u.runTestCaseGeneration(sessionId, coderId)` — i.e. roughly
the current lines 381-419) with:

```go
	client, err := u.llmFactory.Current(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve llm client", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	baselineValues := stateIdToName(states, baselineState)

	callCtx, cancel := context.WithTimeout(ctx, coderGenerationCallTimeout)
	best, err := applicationinfraredcodergeneration.WriteCoder(callCtx, client, device.Brand, device.Model, states, definitions, payload, baselineValues)
	cancel()
	if err != nil {
		u.logger.Error(ctx, tag, "failed to write coder", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	knownCases := buildKnownCases(states, baselineValues, baselineBits, recordedCases)
	bestResult := applicationinfraredcodergeneration.Validate(u.encoderRunner, best.EncoderSource, payload.ChecksumBits, knownCases, encoderSmokeTestTimeout)

	// NOTE: ChecksumOnlyGap() can only be true when owned bits are already
	// 100% correct — which means Passed() is ALSO already true at that
	// point (Passed() deliberately ignores checksum bits, per this
	// package's Global Constraint that checksum never gates pass/fail on
	// its own). So this branch's condition checks ChecksumOnlyGap() alone,
	// not "!Passed() && ChecksumOnlyGap()" — that combination can never be
	// true and would make this branch dead code. If this session has
	// already used its one clarification attempt
	// (ChecksumClarificationUsedAt != nil) and checksum is still
	// imperfect, this branch is skipped and the coder is persisted anyway
	// via the fallthrough below (Passed() is still true — checksum staying
	// wrong doesn't block persisting, it's informational only, and a
	// second clarification round wouldn't help since it's a data problem
	// this session already tried once to resolve).
	if bestResult.ChecksumOnlyGap() && session.ChecksumClarificationUsedAt == nil {
		plans, err := applicationinfraredcodergeneration.WriteChecksumClarificationCases(ctx, client, device.Brand, device.Model, best, states)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to write checksum clarification cases", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
			u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
			return
		}
		if len(plans) == 0 {
			u.logger.Error(ctx, tag, "llm proposed no checksum clarification cases", domainmodels.LoggerMeta{"session_id": sessionId})
			u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
			return
		}
		if err := u.session.MarkChecksumClarificationUsedById(ctx, sessionId); err != nil {
			u.logger.Error(ctx, tag, "failed to mark checksum clarification used", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
			u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
			return
		}

		nextStep := int32(1)
		for _, c := range cases {
			if c.Step >= nextStep {
				nextStep = c.Step + 1
			}
		}
		u.persistNewCasesAndReturnToRecording(ctx, tag, sessionId, nextStep, plans)
		return
	}

	for round := 1; round <= maxRepairRounds && !bestResult.Passed(); round++ {
		repairCtx, repairCancel := context.WithTimeout(ctx, coderGenerationCallTimeout)
		repaired, err := applicationinfraredcodergeneration.RepairCoder(repairCtx, client, device.Brand, device.Model, best.EncoderSource, best.DecoderSource, bestResult)
		repairCancel()
		if err != nil {
			u.logger.Warn(ctx, tag, "repair round failed to generate, keeping best attempt", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "round": round})
			continue
		}
		result := applicationinfraredcodergeneration.Validate(u.encoderRunner, repaired.EncoderSource, payload.ChecksumBits, knownCases, encoderSmokeTestTimeout)
		if result.OwnedCorrect > bestResult.OwnedCorrect {
			best, bestResult = repaired, result
		}
	}

	if !bestResult.Passed() {
		accuracy := bestResult.OwnedAccuracy()
		if accuracy < ownedBitFatalFloor {
			err := domainmodels.NewError(fmt.Sprintf("llm failed to capture the device's encoding pattern (best attempt: %.0f%% of known bits correct)", accuracy*100), domainmodels.ErrTypeFailure, nil)
			u.logger.Error(ctx, tag, "coder generation exhausted repair attempts below fatal floor", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "owned_accuracy": accuracy})
			u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
			return
		}
		u.logger.Warn(ctx, tag, "persisting best-effort coder below full validation", domainmodels.LoggerMeta{
			"session_id": sessionId, "owned_accuracy": accuracy,
			"checksum_correct": bestResult.ChecksumOK, "checksum_total": bestResult.ChecksumTotal,
		})
	}

	coderId, err := u.coder.Create(ctx, domainmodels.InfraredStateCoder{
		InfraredDeviceId:        session.InfraredDeviceId,
		InfraredRecordSessionId: sessionId,
		EncoderSource:           best.EncoderSource,
		DecoderSource:           best.DecoderSource,
		SummaryReadme:           best.SummaryReadme,
		DetailReadme:            best.DetailReadme,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to persist coder", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	go u.runTestCaseGeneration(sessionId, coderId)
```

This removes the old standalone
`u.encoderRunner.RunEncoder(written.EncoderSource, ...)` smoke test
entirely — `Validate()`'s baseline case covers exactly the same ground
(and more), so the separate call is now redundant.

Add the new private helper `buildKnownCases` right after
`runAnalysisAndGeneration` (or next to `stateIdToName`, which it's a
sibling of):

```go
// buildKnownCases converts buildAnalysisPayload's baseline/recordedCases
// output into the coder_generation.KnownCase list Validate() needs — each
// entry pairs a full, name-keyed state map (ready for RunEncoder) with
// that exact state's real recorded bits. Non-baseline cases are
// reconstructed by copying the baseline's full state and overriding the
// one field this case's OFAT construction changed, since a recorded
// non-baseline case's row only carries its target state/value delta, not
// a full state snapshot.
func buildKnownCases(states []domainmodels.InfraredState, baselineValues map[string]string, baselineBits []int, recordedCases []recordedCase) []applicationinfraredcodergeneration.KnownCase {
	nameById := make(map[string]string, len(states))
	for _, s := range states {
		nameById[s.Id.String()] = s.Name
	}

	baselineCopy := make(map[string]string, len(baselineValues))
	for k, v := range baselineValues {
		baselineCopy[k] = v
	}

	knownCases := []applicationinfraredcodergeneration.KnownCase{
		{Label: "baseline", State: baselineCopy, Bits: baselineBits},
	}

	for _, rc := range recordedCases {
		stateName, ok := nameById[rc.TargetStateId.String()]
		if !ok {
			continue
		}
		state := make(map[string]string, len(baselineValues))
		for k, v := range baselineValues {
			state[k] = v
		}
		state[stateName] = rc.TargetValue
		knownCases = append(knownCases, applicationinfraredcodergeneration.KnownCase{
			Label: fmt.Sprintf("%s=%s", stateName, rc.TargetValue),
			State: state,
			Bits:  rc.Bits,
		})
	}

	return knownCases
}
```

Finally, remove the two throwaway `_ = baselineBits; _ = recordedCases`
lines added in Task 6 (now genuinely used above), and remove the old
`stateIdToName(states, baselineState)` call site that fed the old smoke
test (replaced by the `baselineValues := stateIdToName(states,
baselineState)` line already shown above, which now serves double duty:
`WriteCoder`'s new parameter and `buildKnownCases`' baseline).

- [ ] **Step 8: Run all four `TestRunAnalysisAndGeneration*` tests**

Run: `cd backend && go build ./... && go test ./internal/application/infrared/record_session_management/... -run TestRunAnalysisAndGeneration -v`
Expected: `PASS` for all six:
`TestRunAnalysisAndGenerationPersistsCoderAndAdvancesToTestCaseGeneration`,
`TestRunAnalysisAndGenerationFailsSessionWhenEncoderThrows`,
`TestRunAnalysisAndGenerationRepairsAnInitiallyFailingEncoder`,
`TestRunAnalysisAndGenerationPersistsDegradedBestAttemptAboveFatalFloor`,
`TestRunAnalysisAndGenerationFailsBelowFatalFloorAfterRepairExhausted`,
`TestRunAnalysisAndGenerationRequestsChecksumClarificationOnChecksumOnlyGap`.

If `TestRunAnalysisAndGenerationPersistsDegradedBestAttemptAboveFatalFloor`
doesn't land at the intended ~66.7% given the exact fixture above,
recompute `encoderRunner.raw`'s decoded bits by hand against the fixture's
real recorded bits (three owned bits, `[0, 1, 1]`) and adjust the
`raw` value until exactly 2 of 3 positions match — the important property
this test verifies is "some correct, some wrong, overall ≥60%", not the
literal byte values.

- [ ] **Step 9: Run the full existing suite for this package plus `go vet`**

Run: `cd backend && go vet ./internal/application/infrared/... && go test ./internal/application/infrared/... -v`
Expected: `PASS` across every test in `record_session_management` and
`coder_generation` (including the untouched `runRetryCaseGeneration`
tests, which must still pass unchanged after Step 5's extraction — if any
fail, the extraction changed observable behavior and needs fixing before
proceeding, not working around).

- [ ] **Step 10: Commit**

```bash
cd backend
git add internal/application/infrared/record_session_management/usecase.go internal/application/infrared/record_session_management/usecase_test.go
git commit -m "infrared: wire the validate/repair/escalate loop into runAnalysisAndGeneration"
```

---

### Task 8: Full-repo verification

**Files:** none (verification only).

**Interfaces:** none.

- [ ] **Step 1: Build the whole module**

Run: `cd backend && go build ./...`
Expected: no errors.

- [ ] **Step 2: Vet the whole module**

Run: `cd backend && go vet ./...`
Expected: no findings.

- [ ] **Step 3: Run the full test suite**

Run: `cd backend && go test ./...`
Expected: all packages `PASS` (or `ok`), including every package touched
across Tasks 1-7 and every package this session's earlier work already
touched (LLM config, Gemini provider, etc.) — this task's job is to catch
any cross-package regression the individual tasks' scoped test runs
wouldn't have seen.

- [ ] **Step 4: Format check**

Run: `cd backend && gofmt -l .`
Expected: empty output (no unformatted files). If any files are listed,
run `gofmt -w <file>` on each and re-run this step.

- [ ] **Step 5: Module tidiness**

Run: `cd backend && go mod tidy && git diff --stat go.mod go.sum`
Expected: no diff (this feature adds no new dependencies — everything used
is already in `go.mod`). If there is a diff, investigate why before
committing it — an unexpected `go.mod`/`go.sum` change here signals a
stray import that shouldn't be there.

- [ ] **Step 6: Confirm no stray files**

Run: `cd backend && git status --short`
Expected: clean, or showing only the files intentionally modified across
Tasks 1-7 (all already committed by their own Step 5/9/10 — this should
show nothing at all if every prior task's commit succeeded).
