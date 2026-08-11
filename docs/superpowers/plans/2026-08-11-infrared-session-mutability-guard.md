# Infrared Session Mutability Guard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a backend-level guard, enforced regardless of which client calls the API, that rejects `RetryCase`, `SetCurrentCase`, `AcceptRaw`, `DiscardRaw`, `TransmitTestCase`, and `RecordTestCaseResult` once a record session has already reached a terminal state (`COMPLETED` or `FAILED`) — closing the gap the earlier frontend-only fix left open (a direct API call could still destroy captures on a finished session).

**Architecture:** One private helper, `requireMutableSession`, added to the existing `record_session_management` usecase. Each of the six mutation methods is reordered (where necessary) to resolve the target session *before* performing any mutation, then calls the guard (or, where the session is already fetched mid-function, an equivalent inline check using the same shared error) before proceeding. No new types, no schema/route/response changes — this is a pure application-layer addition whose rejection already flows through the existing `ApiError`-based error handling on both backend and frontend.

**Tech Stack:** Go (existing backend module), existing `domainmodels.NewError`/`ErrType*` error pattern, existing fake-repository test doubles in `usecase_test.go`.

## Global Constraints

- The rejection error uses `domainmodels.ErrTypeValidation` with the message `"record session has already finished and can no longer be modified"` — one shared message via a single helper function, so every caller reports the identical reason.
- A session that cannot be found (nil from `ReadById`) is a separate, distinct error: `domainmodels.ErrTypeNotFound`, message `"record session not found"`.
- `IsCompleted` is the single source of truth for "finished" — already verified (in this session's own prior work) to be set to `true` exactly when `RecordingState` transitions to `COMPLETED` or `FAILED`, and reset to `false` on any later transition back to a live state. No new field, no new logic to compute "is this session done."
- Every guard call must run strictly before the first state-mutating repository call in its method — a session check performed after a mutation has already started defeats the purpose entirely.
- No route, handler, response, or frontend changes in this plan — the existing error-to-`ApiError` mapping and existing frontend UI-level guard (already shipped) are both untouched and remain in place as defense-in-depth alongside this backend enforcement.
- `CaptureIrRaw` is explicitly out of scope — already safe by construction (`ReadActiveByNodeId` only ever returns a session in `RECORDING` state).

---

### Task 1: `requireMutableSession` helper + `SetCurrentCase` and `TransmitTestCase`

**Files:**
- Modify: `backend/internal/application/infrared/record_session_management/usecase.go`
- Modify: `backend/internal/application/infrared/record_session_management/usecase_test.go`

**Interfaces:**
- Consumes: existing `u.session domaincontractsrepository.InfraredRecordSession` field (already has `ReadById(ctx, id) (*domainmodels.InfraredRecordSession, error)`); existing `domainmodels.InfraredRecordSession.IsCompleted bool` field.
- Produces (for Task 2 to consume):
  ```go
  func sessionAlreadyFinishedError() error
  func (u *usecase) requireMutableSession(ctx context.Context, sessionId uuid.UUID) (*domainmodels.InfraredRecordSession, error)
  ```

This task adds the helper and wires it into the two methods where wiring is simplest (`SetCurrentCase` already takes a `sessionId` parameter directly; `TransmitTestCase` already fetches the session mid-function) — establishing the pattern before Tasks 2-4 apply it to the four methods that need a resolve-then-guard reorder.

- [ ] **Step 1: Write the failing tests**

In `usecase_test.go`, first update the two existing `SetCurrentCase` tests (currently at lines 736 and 760) to give their `sessionRepo` a non-nil session — without this, both tests will start failing once the guard requires a real session to exist. Change:

```go
	sessionRepo := &fakeSessionRepository{}
```

to (in `TestSetCurrentCaseRejectsCaseFromDifferentSession`, using the test's existing `sessionId` variable):

```go
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId}}
```

and identically in `TestSetCurrentCaseSetsSessionCurrentCase` (same `sessionId` variable already in scope there).

Then add two new tests, right after `TestSetCurrentCaseSetsSessionCurrentCase` (currently ending at line 789) and right after `TestTransmitTestCaseRunsEncoderAndPublishesResult` (currently ending at line 1925):

```go
func TestSetCurrentCaseRejectsCompletedSession(t *testing.T) {
	sessionId := uuid.New()
	caseId := uuid.New()
	caseRepo := &fakeCaseRepository{getResult: &domainmodels.InfraredStateDeviceRecordCase{
		Id: caseId, InfraredRecordSessionId: sessionId,
	}}
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, IsCompleted: true}}

	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	)

	err := usecase.SetCurrentCase(context.Background(), sessionId, caseId)
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("SetCurrentCase() error = %v, want validation error", err)
	}
	if len(caseRepo.statusUpdateIds) != 0 {
		t.Fatalf("case status updates = %v, want none (session already finished)", caseRepo.statusUpdateIds)
	}
	if sessionRepo.CurrentCaseId() != nil {
		t.Fatalf("session current case id = %v, want nil (session already finished)", sessionRepo.CurrentCaseId())
	}
}
```

```go
func TestTransmitTestCaseRejectsCompletedSession(t *testing.T) {
	testCaseId := uuid.New()
	coderId := uuid.New()
	sessionId := uuid.New()

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: testCaseId, InfraredStateCoderId: coderId},
	}
	coderRepo := &fakeCoderRepository{getByIdResult: &domainmodels.InfraredStateCoder{Id: coderId, InfraredRecordSessionId: sessionId}}
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, IsCompleted: true}}
	publish := &fakePublish{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, publish, &noopLogger{},
	)

	err := impl.TransmitTestCase(context.Background(), testCaseId)
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("TransmitTestCase() error = %v, want validation error", err)
	}
	if publish.transmitCalls != 0 {
		t.Fatalf("IrTransmit() calls = %d, want 0 (session already finished)", publish.transmitCalls)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run 'TestSetCurrentCase|TestTransmitTestCase' -v`
Expected: `TestSetCurrentCaseRejectsCaseFromDifferentSession` and `TestSetCurrentCaseSetsSessionCurrentCase` still `PASS` (fixture-only change, no behavior change yet). `TestSetCurrentCaseRejectsCompletedSession` and `TestTransmitTestCaseRejectsCompletedSession` `FAIL` — `SetCurrentCase`/`TransmitTestCase` don't reject a completed session yet, so both new tests get `err = nil` where a validation error was expected.

- [ ] **Step 3: Add the helper**

In `backend/internal/application/infrared/record_session_management/usecase.go`, add right after `coderById` (currently ending at line 921, right before the `stateIdToNameLookup` comment block):

```go
// sessionAlreadyFinishedError is returned by every mutation entrypoint
// below once a session has reached a terminal state (COMPLETED or
// FAILED) — kept as a single function so every caller reports the
// identical reason.
func sessionAlreadyFinishedError() error {
	return domainmodels.NewError("record session has already finished and can no longer be modified", domainmodels.ErrTypeValidation, nil)
}

// requireMutableSession reads the session and rejects the caller if it has
// already reached a terminal state (COMPLETED or FAILED, both captured by
// IsCompleted) — without this guard, a mutation could destroy
// already-accepted captures with no way to recapture them, or leave a
// finished session's case/cursor state internally inconsistent (e.g. a
// non-nil CurrentRecordCaseId on a COMPLETED session). The frontend
// already hides the controls that would trigger these calls once a
// session is done; this is the guard that holds regardless of which
// client is calling.
func (u *usecase) requireMutableSession(ctx context.Context, sessionId uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
	session, err := u.session.ReadById(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, domainmodels.NewError("record session not found", domainmodels.ErrTypeNotFound, nil)
	}
	if session.IsCompleted {
		return nil, sessionAlreadyFinishedError()
	}
	return session, nil
}
```

- [ ] **Step 4: Wire the guard into `SetCurrentCase`**

In the same file, `SetCurrentCase` (currently starting at line 1167) currently begins:

```go
func (u *usecase) SetCurrentCase(ctx context.Context, sessionId uuid.UUID, caseId uuid.UUID) error {
	const tag = "infrared/record_session_management/SetCurrentCase"

	recordCase, err := u.recordCase.ReadById(ctx, caseId)
```

Change to:

```go
func (u *usecase) SetCurrentCase(ctx context.Context, sessionId uuid.UUID, caseId uuid.UUID) error {
	const tag = "infrared/record_session_management/SetCurrentCase"

	if _, err := u.requireMutableSession(ctx, sessionId); err != nil {
		return err
	}

	recordCase, err := u.recordCase.ReadById(ctx, caseId)
```

The rest of the function is unchanged.

- [ ] **Step 5: Wire the guard into `TransmitTestCase`**

`TransmitTestCase` (currently starting at line 645) already fetches the session mid-function:

```go
	session, err := u.session.ReadById(ctx, coder.InfraredRecordSessionId)
	if err != nil || session == nil {
		u.logger.Error(ctx, tag, "failed to look up session for test case", domainmodels.LoggerMeta{"err": err, "test_case_id": testCaseId})
		return err
	}
	node, err := u.node.ReadById(ctx, session.NodeId)
```

Change to (inserting the check between the two existing blocks, reusing the already-fetched `session` rather than re-fetching via the helper):

```go
	session, err := u.session.ReadById(ctx, coder.InfraredRecordSessionId)
	if err != nil || session == nil {
		u.logger.Error(ctx, tag, "failed to look up session for test case", domainmodels.LoggerMeta{"err": err, "test_case_id": testCaseId})
		return err
	}
	if session.IsCompleted {
		return sessionAlreadyFinishedError()
	}
	node, err := u.node.ReadById(ctx, session.NodeId)
```

The rest of the function is unchanged.

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run 'TestSetCurrentCase|TestTransmitTestCase' -v`
Expected: `PASS` for all four (`TestSetCurrentCaseRejectsCaseFromDifferentSession`, `TestSetCurrentCaseSetsSessionCurrentCase`, `TestSetCurrentCaseRejectsCompletedSession`, `TestTransmitTestCaseRunsEncoderAndPublishesResult`, `TestTransmitTestCaseRejectsCompletedSession`).

- [ ] **Step 7: Run the whole package to confirm nothing else broke**

Run: `cd backend && go build ./... && go vet ./... && go test ./internal/application/infrared/record_session_management/... -v`
Expected: full package `PASS` — this task hasn't touched `RetryCase`/`AcceptRaw`/`DiscardRaw`/`RecordTestCaseResult` yet, so their existing tests (still using session-less fixtures) are expected to still pass unchanged at this point.

- [ ] **Step 8: Commit**

```bash
cd backend
git add internal/application/infrared/record_session_management/usecase.go internal/application/infrared/record_session_management/usecase_test.go
git commit -m "infrared: add requireMutableSession guard, wire into SetCurrentCase and TransmitTestCase"
```

---

### Task 2: `RetryCase` and `AcceptRaw`

**Files:**
- Modify: `backend/internal/application/infrared/record_session_management/usecase.go`
- Modify: `backend/internal/application/infrared/record_session_management/usecase_test.go`

**Interfaces:**
- Consumes: `requireMutableSession` (Task 1).
- Produces: nothing new consumed by later tasks — `RetryCase`/`AcceptRaw`'s signatures are unchanged.

Both methods currently mutate before resolving the session (`RetryCase` discards raws and reactivates the case before ever looking up which session the case belongs to; `AcceptRaw` marks a raw accepted before looking up anything). Both need reordering, not just an inserted check, and both restructurings incidentally eliminate a now-redundant duplicate lookup the original code performed after its mutation.

- [ ] **Step 1: Write the failing tests**

First update the four existing tests that will break once these two methods require a real session. In `TestRetryCaseSetsCurrentCase` (currently at line 791), change:

```go
	sessionRepo := &fakeSessionRepository{}
```

to (the test already declares `sessionId` and uses it for the case fixture):

```go
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId}}
```

In `TestAcceptRawAdvancesCursorToNextPendingCase` (line 818), `TestAcceptRawTriggersAnalyzingWhenNoPendingCaseRemains` (line 858), and `TestAcceptRawDoesNotAdvanceCursorBeforeSecondRawAccepted` (line 899) — all three already declare `sessionId` and use it in their case fixtures — apply the identical change: `sessionRepo := &fakeSessionRepository{}` → `sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId}}`.

Then add two new tests, right after `TestRetryCaseSetsCurrentCase` (currently ending at line 816) and right after `TestAcceptRawDoesNotAdvanceCursorBeforeSecondRawAccepted` (currently ending at line 930):

```go
func TestRetryCaseRejectsCompletedSession(t *testing.T) {
	sessionId := uuid.New()
	caseId := uuid.New()
	caseRepo := &fakeCaseRepository{
		getResult: &domainmodels.InfraredStateDeviceRecordCase{Id: caseId, InfraredRecordSessionId: sessionId},
	}
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, IsCompleted: true}}

	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	)

	err := usecase.RetryCase(context.Background(), caseId)
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RetryCase() error = %v, want validation error", err)
	}
	if len(caseRepo.statusUpdateIds) != 0 {
		t.Fatalf("case status updates = %v, want none (session already finished)", caseRepo.statusUpdateIds)
	}
	if sessionRepo.CurrentCaseId() != nil {
		t.Fatalf("session current case id = %v, want nil (session already finished)", sessionRepo.CurrentCaseId())
	}
}
```

```go
func TestAcceptRawRejectsCompletedSession(t *testing.T) {
	sessionId := uuid.New()
	caseId := uuid.New()
	rawId := uuid.New()
	caseRepo := &fakeCaseRepository{
		getResult: &domainmodels.InfraredStateDeviceRecordCase{Id: caseId, InfraredRecordSessionId: sessionId},
		rawById: map[uuid.UUID]*domainmodels.InfraredStateDeviceRecordRaw{
			rawId: {Id: rawId, InfraredStateDeviceRecordCaseId: caseId},
		},
	}
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, IsCompleted: true}}

	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	)

	err := usecase.AcceptRaw(context.Background(), rawId)
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("AcceptRaw() error = %v, want validation error", err)
	}
	if len(caseRepo.statusUpdateIds) != 0 {
		t.Fatalf("case status updates = %v, want none (session already finished)", caseRepo.statusUpdateIds)
	}
	if sessionRepo.CurrentCaseId() != nil {
		t.Fatalf("session current case id = %v, want nil (session already finished)", sessionRepo.CurrentCaseId())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run 'TestRetryCase|TestAcceptRaw' -v`
Expected: the four pre-existing tests still `PASS` (fixture-only change so far). `TestRetryCaseRejectsCompletedSession` and `TestAcceptRawRejectsCompletedSession` `FAIL`.

- [ ] **Step 3: Rewrite `RetryCase`**

`RetryCase` (currently starting at line 1134) is:

```go
func (u *usecase) RetryCase(ctx context.Context, caseId uuid.UUID) error {
	const tag = "infrared/record_session_management/RetryCase"

	raw, err := u.recordCase.ReadListRawByCaseId(ctx, caseId)
	if err != nil {
		return err
	}
	for _, r := range raw {
		if r.Status == domainmodels.InfraredRecordRawStatusAccepted {
			reason := "case marked for redo"
			if err := u.recordCase.UpdateRawStatusById(ctx, r.Id, domainmodels.InfraredRecordRawStatusDiscarded, &reason); err != nil {
				u.logger.Error(ctx, tag, "failed to discard accepted raw for retry", domainmodels.LoggerMeta{"err": err, "raw_id": r.Id})
				return err
			}
		}
	}
	if err := u.recordCase.UpdateStatusById(ctx, caseId, domainmodels.InfraredRecordCaseStatusActive); err != nil {
		u.logger.Error(ctx, tag, "failed to activate case for retry", domainmodels.LoggerMeta{"err": err, "case_id": caseId})
		return err
	}

	recordCase, err := u.recordCase.ReadById(ctx, caseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up case for retry", domainmodels.LoggerMeta{"err": err, "case_id": caseId})
		return err
	}
	if err := u.session.UpdateCurrentRecordCaseIdById(ctx, recordCase.InfraredRecordSessionId, &caseId); err != nil {
		u.logger.Error(ctx, tag, "failed to set session's current record case for retry", domainmodels.LoggerMeta{"err": err, "case_id": caseId})
		return err
	}
	return nil
}
```

Replace it entirely with:

```go
func (u *usecase) RetryCase(ctx context.Context, caseId uuid.UUID) error {
	const tag = "infrared/record_session_management/RetryCase"

	recordCase, err := u.recordCase.ReadById(ctx, caseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up case for retry", domainmodels.LoggerMeta{"err": err, "case_id": caseId})
		return err
	}
	if recordCase == nil {
		return domainmodels.NewError("record case not found", domainmodels.ErrTypeNotFound, nil)
	}
	if _, err := u.requireMutableSession(ctx, recordCase.InfraredRecordSessionId); err != nil {
		return err
	}

	raw, err := u.recordCase.ReadListRawByCaseId(ctx, caseId)
	if err != nil {
		return err
	}
	for _, r := range raw {
		if r.Status == domainmodels.InfraredRecordRawStatusAccepted {
			reason := "case marked for redo"
			if err := u.recordCase.UpdateRawStatusById(ctx, r.Id, domainmodels.InfraredRecordRawStatusDiscarded, &reason); err != nil {
				u.logger.Error(ctx, tag, "failed to discard accepted raw for retry", domainmodels.LoggerMeta{"err": err, "raw_id": r.Id})
				return err
			}
		}
	}
	if err := u.recordCase.UpdateStatusById(ctx, caseId, domainmodels.InfraredRecordCaseStatusActive); err != nil {
		u.logger.Error(ctx, tag, "failed to activate case for retry", domainmodels.LoggerMeta{"err": err, "case_id": caseId})
		return err
	}
	if err := u.session.UpdateCurrentRecordCaseIdById(ctx, recordCase.InfraredRecordSessionId, &caseId); err != nil {
		u.logger.Error(ctx, tag, "failed to set session's current record case for retry", domainmodels.LoggerMeta{"err": err, "case_id": caseId})
		return err
	}
	return nil
}
```

Note this both adds the guard *and* removes the second, now-redundant `u.recordCase.ReadById(ctx, caseId)` call the original performed after mutating (replaced by reusing the `recordCase` fetched up front) — as a side benefit, this also means a case that can't be found is now caught *before* any mutation starts, instead of after the raw-discard-and-reactivate work had already run.

- [ ] **Step 4: Rewrite `AcceptRaw`**

`AcceptRaw` (currently starting at line 247) is:

```go
func (u *usecase) AcceptRaw(ctx context.Context, rawId uuid.UUID) error {
	const tag = "infrared/record_session_management/AcceptRaw"

	if err := u.recordCase.UpdateRawStatusById(ctx, rawId, domainmodels.InfraredRecordRawStatusAccepted, nil); err != nil {
		return err
	}

	raw, err := u.recordCase.ReadRawById(ctx, rawId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up accepted raw", domainmodels.LoggerMeta{"err": err, "raw_id": rawId})
		return err
	}

	acceptedCount, err := u.recordCase.CountAcceptedRawByCaseId(ctx, raw.InfraredStateDeviceRecordCaseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to count accepted raws", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}
	if acceptedCount < 2 {
		return nil
	}

	if err := u.recordCase.UpdateStatusById(ctx, raw.InfraredStateDeviceRecordCaseId, domainmodels.InfraredRecordCaseStatusAccepted); err != nil {
		u.logger.Error(ctx, tag, "failed to mark case accepted", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}

	// InfraredStateDeviceRecordRaw has no session id of its own (only a case
	// id) — look the case up to get the session id from a field that's
	// genuinely backed by a column, rather than adding a synthetic
	// join-only field to the raw model.
	acceptedCase, err := u.recordCase.ReadById(ctx, raw.InfraredStateDeviceRecordCaseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up accepted case", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}
	if acceptedCase == nil {
		err := domainmodels.NewError("accepted case not found", domainmodels.ErrTypeNotFound, nil)
		u.logger.Error(ctx, tag, "accepted case not found", domainmodels.LoggerMeta{"case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}
	sessionId := acceptedCase.InfraredRecordSessionId

	cases, err := u.recordCase.ReadListBySessionId(ctx, sessionId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list cases for cursor advancement", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		return err
	}

	var next *domainmodels.InfraredStateDeviceRecordCase
	for i := range cases {
		if cases[i].Status != domainmodels.InfraredRecordCaseStatusPending {
			continue
		}
		if next == nil || cases[i].Step < next.Step {
			next = &cases[i]
		}
	}

	if next != nil {
		if err := u.recordCase.UpdateStatusById(ctx, next.Id, domainmodels.InfraredRecordCaseStatusActive); err != nil {
			u.logger.Error(ctx, tag, "failed to activate next case", domainmodels.LoggerMeta{"err": err, "case_id": next.Id})
			return err
		}
		if err := u.session.UpdateCurrentRecordCaseIdById(ctx, sessionId, &next.Id); err != nil {
			u.logger.Error(ctx, tag, "failed to advance session cursor", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
			return err
		}
		u.broadcastBestEffort(ctx, tag, sessionId, domainmodels.InfraredRecordingStateRecording, &next.Id)
		return nil
	}

	if err := u.session.UpdateCurrentRecordCaseIdById(ctx, sessionId, nil); err != nil {
		u.logger.Error(ctx, tag, "failed to clear session cursor before analysis", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		return err
	}
	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateAnalyzing)

	go u.runAnalysisAndGeneration(sessionId)

	return nil
}
```

Replace it entirely with:

```go
func (u *usecase) AcceptRaw(ctx context.Context, rawId uuid.UUID) error {
	const tag = "infrared/record_session_management/AcceptRaw"

	raw, err := u.recordCase.ReadRawById(ctx, rawId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up raw before accept", domainmodels.LoggerMeta{"err": err, "raw_id": rawId})
		return err
	}
	if raw == nil {
		return domainmodels.NewError("raw capture not found", domainmodels.ErrTypeNotFound, nil)
	}

	// InfraredStateDeviceRecordRaw has no session id of its own (only a case
	// id) — look the case up to get the session id from a field that's
	// genuinely backed by a column, rather than adding a synthetic
	// join-only field to the raw model.
	recordCase, err := u.recordCase.ReadById(ctx, raw.InfraredStateDeviceRecordCaseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up case before accept", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}
	if recordCase == nil {
		return domainmodels.NewError("record case not found", domainmodels.ErrTypeNotFound, nil)
	}
	if _, err := u.requireMutableSession(ctx, recordCase.InfraredRecordSessionId); err != nil {
		return err
	}
	sessionId := recordCase.InfraredRecordSessionId

	if err := u.recordCase.UpdateRawStatusById(ctx, rawId, domainmodels.InfraredRecordRawStatusAccepted, nil); err != nil {
		return err
	}

	acceptedCount, err := u.recordCase.CountAcceptedRawByCaseId(ctx, raw.InfraredStateDeviceRecordCaseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to count accepted raws", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}
	if acceptedCount < 2 {
		return nil
	}

	if err := u.recordCase.UpdateStatusById(ctx, raw.InfraredStateDeviceRecordCaseId, domainmodels.InfraredRecordCaseStatusAccepted); err != nil {
		u.logger.Error(ctx, tag, "failed to mark case accepted", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}

	cases, err := u.recordCase.ReadListBySessionId(ctx, sessionId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list cases for cursor advancement", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		return err
	}

	var next *domainmodels.InfraredStateDeviceRecordCase
	for i := range cases {
		if cases[i].Status != domainmodels.InfraredRecordCaseStatusPending {
			continue
		}
		if next == nil || cases[i].Step < next.Step {
			next = &cases[i]
		}
	}

	if next != nil {
		if err := u.recordCase.UpdateStatusById(ctx, next.Id, domainmodels.InfraredRecordCaseStatusActive); err != nil {
			u.logger.Error(ctx, tag, "failed to activate next case", domainmodels.LoggerMeta{"err": err, "case_id": next.Id})
			return err
		}
		if err := u.session.UpdateCurrentRecordCaseIdById(ctx, sessionId, &next.Id); err != nil {
			u.logger.Error(ctx, tag, "failed to advance session cursor", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
			return err
		}
		u.broadcastBestEffort(ctx, tag, sessionId, domainmodels.InfraredRecordingStateRecording, &next.Id)
		return nil
	}

	if err := u.session.UpdateCurrentRecordCaseIdById(ctx, sessionId, nil); err != nil {
		u.logger.Error(ctx, tag, "failed to clear session cursor before analysis", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		return err
	}
	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateAnalyzing)

	go u.runAnalysisAndGeneration(sessionId)

	return nil
}
```

Note this both adds the guard *and* removes the original's second, now-redundant `ReadRawById`/`ReadById`-for-`acceptedCase` calls performed after the mutation (both replaced by reusing `raw`/`recordCase` fetched up front, since neither value the original re-fetched — `raw.InfraredStateDeviceRecordCaseId`, `acceptedCase.InfraredRecordSessionId` — could have changed as a result of the accept). As with `RetryCase`, a missing raw or case is now caught before any mutation starts.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run 'TestRetryCase|TestAcceptRaw' -v`
Expected: `PASS` for all six (`TestRetryCaseSetsCurrentCase`, `TestRetryCaseRejectsCompletedSession`, `TestAcceptRawAdvancesCursorToNextPendingCase`, `TestAcceptRawTriggersAnalyzingWhenNoPendingCaseRemains`, `TestAcceptRawDoesNotAdvanceCursorBeforeSecondRawAccepted`, `TestAcceptRawRejectsCompletedSession`).

- [ ] **Step 6: Run the whole package**

Run: `cd backend && go build ./... && go vet ./... && go test ./internal/application/infrared/record_session_management/... -v`
Expected: full package `PASS`.

- [ ] **Step 7: Commit**

```bash
cd backend
git add internal/application/infrared/record_session_management/usecase.go internal/application/infrared/record_session_management/usecase_test.go
git commit -m "infrared: wire requireMutableSession guard into RetryCase and AcceptRaw"
```

---

### Task 3: `DiscardRaw` and `RecordTestCaseResult`

**Files:**
- Modify: `backend/internal/application/infrared/record_session_management/usecase.go`
- Modify: `backend/internal/application/infrared/record_session_management/usecase_test.go`

**Interfaces:**
- Consumes: `requireMutableSession` (Task 1).
- Produces: nothing new consumed by later tasks.

`DiscardRaw` currently does no lookups at all (its only existing check is the empty-reason validation) — it needs the same raw→case→guard resolution `AcceptRaw` now has. `RecordTestCaseResult` already has an early idempotency return; the guard must be added *after* that check (so a legitimate double-submit on an already-terminal test case still short-circuits harmlessly, matching existing behavior) but *before* the status update that follows.

- [ ] **Step 1: Write the failing tests**

First update the three existing `RecordTestCaseResult` tests that will break. In `TestRecordTestCaseResultDoesNothingElseWhilePendingCasesRemain` (currently at line 1927), the fixture is:

```go
	coderRepo := &fakeCoderRepository{}
	sessionRepo := &fakeSessionRepository{}
```

Change to (the test already declares `coderId`; add a `sessionId` declaration and wire both fakes to it):

```go
	sessionId := uuid.New()
	coderRepo := &fakeCoderRepository{getByIdResult: &domainmodels.InfraredStateCoder{Id: coderId, InfraredRecordSessionId: sessionId}}
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId}}
```

(Insert the `sessionId := uuid.New()` line right after the test's existing `coderId := uuid.New()`/case-id declarations, before the fixture construction.)

In `TestRecordTestCaseResultCompletesSessionWhenAllPass` (currently at line 1958), the test already declares `sessionId` and uses it for `fixtureCoder`. Change:

```go
	sessionRepo := &fakeSessionRepository{}
```

to:

```go
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId}}
```

In `TestRecordTestCaseResultIgnoresSupersededCoderRound` (currently at line 2094), the test already declares `sessionId` and uses it for both `coderRepo.getByIdResult`/`getResult`. Apply the identical change: `sessionRepo := &fakeSessionRepository{}` → `sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId}}`.

Then add two new tests, right after `TestDiscardRawRequiresNonEmptyReason` (currently ending at line 644) and right after `TestRecordTestCaseResultIgnoresSupersededCoderRound` (currently ending at line 2128):

```go
func TestDiscardRawRejectsCompletedSession(t *testing.T) {
	sessionId := uuid.New()
	caseId := uuid.New()
	rawId := uuid.New()
	caseRepo := &fakeCaseRepository{
		getResult: &domainmodels.InfraredStateDeviceRecordCase{Id: caseId, InfraredRecordSessionId: sessionId},
		rawById: map[uuid.UUID]*domainmodels.InfraredStateDeviceRecordRaw{
			rawId: {Id: rawId, InfraredStateDeviceRecordCaseId: caseId},
		},
	}
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, IsCompleted: true}}

	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	)

	err := usecase.DiscardRaw(context.Background(), rawId, "changed my mind")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("DiscardRaw() error = %v, want validation error", err)
	}
}
```

```go
func TestRecordTestCaseResultRejectsCompletedSession(t *testing.T) {
	coderId := uuid.New()
	sessionId := uuid.New()
	testCaseId := uuid.New()

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: testCaseId, InfraredStateCoderId: coderId, Status: domainmodels.InfraredTestCaseStatusPending},
	}
	coderRepo := &fakeCoderRepository{getByIdResult: &domainmodels.InfraredStateCoder{Id: coderId, InfraredRecordSessionId: sessionId}}
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, IsCompleted: true}}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	)

	err := impl.RecordTestCaseResult(context.Background(), testCaseId, true)
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RecordTestCaseResult() error = %v, want validation error", err)
	}
	if coderRepo.activateCalls != 0 {
		t.Fatal("Activate() was called despite the session already being finished")
	}
	if len(sessionRepo.StatusUpdates()) != 0 {
		t.Fatalf("session status updates = %v, want none (session already finished)", sessionRepo.StatusUpdates())
	}
}
```

The `existing.Status` in this new test is deliberately `Pending` (not `Passed`/`Failed`) — a terminal status would trigger the *existing* idempotency early-return before ever reaching the new guard, which would make the test pass for the wrong reason.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run 'TestDiscardRaw|TestRecordTestCaseResult' -v`
Expected: the three pre-existing `RecordTestCaseResult` tests updated in Step 1 still `PASS` (fixture-only change so far), along with the two unmodified ones (`TestRecordTestCaseResultTriggersRetryLoopWhenAnyFail`, `TestRecordTestCaseResultIsIdempotentAgainstDoubleSubmit`, both already unaffected as established during planning). `TestDiscardRawRequiresNonEmptyReason` still `PASS`. `TestDiscardRawRejectsCompletedSession` and `TestRecordTestCaseResultRejectsCompletedSession` `FAIL`.

- [ ] **Step 3: Rewrite `DiscardRaw`**

`DiscardRaw` (currently at line 1127) is:

```go
func (u *usecase) DiscardRaw(ctx context.Context, rawId uuid.UUID, reason string) error {
	if reason == "" {
		return domainmodels.NewError("reason is required", domainmodels.ErrTypeValidation, nil)
	}
	return u.recordCase.UpdateRawStatusById(ctx, rawId, domainmodels.InfraredRecordRawStatusDiscarded, &reason)
}
```

Replace it entirely with:

```go
func (u *usecase) DiscardRaw(ctx context.Context, rawId uuid.UUID, reason string) error {
	const tag = "infrared/record_session_management/DiscardRaw"

	if reason == "" {
		return domainmodels.NewError("reason is required", domainmodels.ErrTypeValidation, nil)
	}

	raw, err := u.recordCase.ReadRawById(ctx, rawId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up raw before discard", domainmodels.LoggerMeta{"err": err, "raw_id": rawId})
		return err
	}
	if raw == nil {
		return domainmodels.NewError("raw capture not found", domainmodels.ErrTypeNotFound, nil)
	}
	recordCase, err := u.recordCase.ReadById(ctx, raw.InfraredStateDeviceRecordCaseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up case before discard", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}
	if recordCase == nil {
		return domainmodels.NewError("record case not found", domainmodels.ErrTypeNotFound, nil)
	}
	if _, err := u.requireMutableSession(ctx, recordCase.InfraredRecordSessionId); err != nil {
		return err
	}

	return u.recordCase.UpdateRawStatusById(ctx, rawId, domainmodels.InfraredRecordRawStatusDiscarded, &reason)
}
```

The empty-reason check stays first and unchanged — it needs no session at all, and `TestDiscardRawRequiresNonEmptyReason`'s fixture (an entirely empty `fakeSessionRepository{}`/`fakeCaseRepository{}`) relies on that check firing before anything else runs.

- [ ] **Step 4: Wire the guard into `RecordTestCaseResult`**

`RecordTestCaseResult` (currently starting at line 705) currently reads, right after the idempotency check:

```go
	if existing.Status == domainmodels.InfraredTestCaseStatusPassed || existing.Status == domainmodels.InfraredTestCaseStatusFailed {
		// Already recorded — idempotent no-op against a double-submit.
		return nil
	}

	status := domainmodels.InfraredTestCaseStatusFailed
```

Change to:

```go
	if existing.Status == domainmodels.InfraredTestCaseStatusPassed || existing.Status == domainmodels.InfraredTestCaseStatusFailed {
		// Already recorded — idempotent no-op against a double-submit.
		return nil
	}

	coderForGuard, err := u.coderById(ctx, existing.InfraredStateCoderId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up coder before recording test case result", domainmodels.LoggerMeta{"err": err, "test_case_id": testCaseId})
		return err
	}
	if coderForGuard == nil {
		return domainmodels.NewError("coder not found", domainmodels.ErrTypeNotFound, nil)
	}
	if _, err := u.requireMutableSession(ctx, coderForGuard.InfraredRecordSessionId); err != nil {
		return err
	}

	status := domainmodels.InfraredTestCaseStatusFailed
```

This adds one extra `coderById` lookup beyond what the function already performs later (at its existing `coder, err := u.coderById(ctx, testCase.InfraredStateCoderId)` call, used for the post-update superseded-round check) — deliberately not deduplicated, since that later lookup reads the freshly re-fetched `testCase` (post status-update) for a genuinely different purpose, and collapsing the two adds restructuring risk for a single extra read with no correctness benefit.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run 'TestDiscardRaw|TestRecordTestCaseResult' -v`
Expected: `PASS` for all eight (`TestDiscardRawRequiresNonEmptyReason`, `TestDiscardRawRejectsCompletedSession`, `TestRecordTestCaseResultDoesNothingElseWhilePendingCasesRemain`, `TestRecordTestCaseResultCompletesSessionWhenAllPass`, `TestRecordTestCaseResultTriggersRetryLoopWhenAnyFail`, `TestRecordTestCaseResultIsIdempotentAgainstDoubleSubmit`, `TestRecordTestCaseResultIgnoresSupersededCoderRound`, `TestRecordTestCaseResultRejectsCompletedSession`).

- [ ] **Step 6: Run the whole package**

Run: `cd backend && go build ./... && go vet ./... && go test ./internal/application/infrared/record_session_management/... -v`
Expected: full package `PASS` — every test in the file, not just the ones this task touched.

- [ ] **Step 7: Commit**

```bash
cd backend
git add internal/application/infrared/record_session_management/usecase.go internal/application/infrared/record_session_management/usecase_test.go
git commit -m "infrared: wire requireMutableSession guard into DiscardRaw and RecordTestCaseResult"
```

---

### Task 4: Full-repo verification

**Files:** none (verification only).

- [ ] **Step 1: Build, vet, test, format the whole backend**

Run: `cd backend && go build ./... && go vet ./... && go test ./... && gofmt -l .`
Expected: clean build/vet, every package `ok`, empty `gofmt -l` output.

- [ ] **Step 2: Module tidiness**

Run: `cd backend && go mod tidy && git diff --stat go.mod go.sum`
Expected: no diff — this plan adds no new dependencies.

- [ ] **Step 3: Confirm no stray files**

Run: `cd /home/dodol/Repositories/mate/mate-things && git status --short`
Expected: clean, or showing only files intentionally modified across Tasks 1-3 (all already committed by their own step — this should show nothing).
