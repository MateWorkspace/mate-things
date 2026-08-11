# Infrared record session mutability guard (backend)

## Context

The final review of the infrared Record list/detail pages found that the
UI-level fix for a data-loss bug (Retry/Transmit/Pass-Fail controls hidden
once a session is `COMPLETED` or `FAILED`) has no backend enforcement of
its own — a direct API call bypasses the UI entirely. `RetryCase` discards
already-accepted raw captures and reactivates a case with no check on
whether the session itself has already finished; `SetCurrentCase` moves the
session's current-case cursor with the same gap. Investigating further
during this session found the identical class of gap in four more mutation
methods on `RecordSessionManagement`: `AcceptRaw`, `DiscardRaw`,
`TransmitTestCase`, and `RecordTestCaseResult` — none of them check whether
the session they're mutating has already reached a terminal state.

One method, `CaptureIrRaw` (the MQTT ingestion path a physical node's raw
captures arrive through), is already safe: it resolves the target session
via `ReadActiveByNodeId`, which only ever returns a session currently in
`RECORDING` state — a finished session never matches, so no change is
needed there.

## Fix

A single private helper on `usecase` in
`backend/internal/application/infrared/record_session_management/usecase.go`:

```go
// requireMutableSession reads the session and rejects the caller with a
// validation error if it has already reached a terminal state (COMPLETED
// or FAILED, both captured by IsCompleted) — without this guard, a
// mutation could destroy already-accepted captures with no way to
// recapture them, or leave a finished session's case/cursor state
// internally inconsistent (e.g. a non-nil CurrentRecordCaseId on a
// COMPLETED session). The frontend already hides the controls that would
// trigger these calls once a session is done; this is the guard that
// holds regardless of which client is calling.
func (u *usecase) requireMutableSession(ctx context.Context, sessionId uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
	session, err := u.session.ReadById(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, domainmodels.NewError("record session not found", domainmodels.ErrTypeNotFound, nil)
	}
	if session.IsCompleted {
		return nil, domainmodels.NewError("record session has already finished and can no longer be modified", domainmodels.ErrTypeValidation, nil)
	}
	return session, nil
}
```

`ErrTypeValidation` matches the existing precedent for a rejected mutation
in this same file (`DiscardRaw`'s `reason is required` check) and already
maps to a 4xx at the HTTP layer — no new error-handling plumbing needed;
the frontend's existing `ApiError` → server-action-error path already
surfaces whatever message comes back.

## Per-method wiring

- **`SetCurrentCase(ctx, sessionId, caseId)`** — already has `sessionId`
  directly; call the guard first, before the existing case-ownership
  check.
- **`RetryCase(ctx, caseId)`** — currently reads the case (to learn its
  session id) only *after* already discarding accepted raws and
  reactivating the case. Reorder: read the case first, call the guard
  using its `InfraredRecordSessionId`, and only then run the existing
  raw-discard-and-reactivate logic.
- **`AcceptRaw(ctx, rawId)`** — currently mutates the raw's status as its
  very first action, with no session lookup anywhere in the function.
  Add a raw→case lookup (`ReadRawById` then `ReadById` on the case) to
  resolve the session id, call the guard, and only then proceed with the
  existing accept/cursor-advance logic.
- **`DiscardRaw(ctx, rawId, reason)`** — same shape as `AcceptRaw`: add
  the raw→case→session resolution and guard before the existing
  `UpdateRawStatusById` call.
- **`TransmitTestCase(ctx, testCaseId)`** — already resolves `session`
  mid-function via the coder lookup; add the `session.IsCompleted` check
  immediately after that lookup, before `RunEncoder`/`publish.IrTransmit`.
- **`RecordTestCaseResult(ctx, testCaseId, passed)`** — resolve the
  session via the existing `coderById` lookup path; call the guard right
  after the existing idempotency check (`existing.Status ==
  Passed/Failed → no-op`), so a legitimate double-submit still
  short-circuits harmlessly first, and before the status update that
  follows.

## Testing

Each of the six methods gets one new test asserting that calling it
against a session with `IsCompleted: true` returns an error and performs
no mutation (verified via the existing fake repositories' call-recording
fields — e.g. asserting `UpdateStatusById`/`UpdateRawStatusById` were never
called). Existing tests for these methods must continue to pass unchanged
(they exercise the non-terminal-session path, which this change doesn't
alter).

## Out of scope

- No route/handler/response changes — the existing error path already
  surfaces a validation/not-found error correctly.
- No frontend changes — the UI-level guard from the prior fix stays as
  defense-in-depth; this spec only adds the backend enforcement that
  holds regardless of client.
- `CaptureIrRaw` is unchanged (already safe by construction).
