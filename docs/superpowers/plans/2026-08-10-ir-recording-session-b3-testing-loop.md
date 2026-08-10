# IR Recording Session B3 (Test Generation, MQTT Transmit, Testing Loop) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Take a session sitting in `FUNCTION_GENERATING` with a persisted `UNVERIFIED` `InfraredStateCoder` (built by Plan B2) all the way through hardware-verified test cases to `COMPLETED` (coder flips `ACTIVE`) or back through a retry loop (new targeted `RecordCase` rows appended to `RECORDING`, producing a new coder attempt). This is the final plan in the IR recording session series — after this, the feature is end-to-end complete.

**Architecture:** Two more LLM-driven structured-output calls (`WriteTestCases`, `WriteRetryCases`), using the same `domaincontractsllm.Client`/`ResponseSchema` pattern Plans B1/B2 already established. Two more background-goroutine jobs on the same `RecordSessionManagement` usecase (`runTestCaseGeneration`, `runRetryCaseGeneration`), continuing the exact async-job convention B1/B2 built. A new MQTT publish direction (backend → node) for transmitting a test case's encoded IR signal, mirroring B1's `ir_capture` (node → backend) topic/handler/DTO shape exactly, but in reverse. Staff observe the physical unit and record PASS/FAIL via one new HTTP endpoint per test case — there is no automated round-trip verification anywhere in this plan; the MQTT ack only confirms the node successfully drove its IR transmitter, never that the target device responded correctly.

**Tech Stack:** Go, the same `domaincontractsllm.Client`/`infrastructurellm.ClientFactory` abstraction, the same `infrastructurejsengine.RunEncoder` (Plan B2) for running the coder's `EncoderSource` against a test case's declared state before transmitting, PostgreSQL via the existing `BasePostgres`/squirrel pattern, MQTT via the existing `domaincontractsnode.Publish`/`Subscriptions` pattern.

## Global Constraints

- No new websocket broadcaster — reuses `domaincontractsbroadcaster.InfraredRecordSession`/`domainmodels.InfraredRecordSessionEvent` for every status update in this plan, same as B1/B2.
- No new permissions — every new HTTP endpoint reuses the existing `infrared_record_session:get`/`add`/`set` permissions from Plan B1, since every new action in this plan (viewing test cases, transmitting a test, recording a result) is the same read/write scope as the session it belongs to.
- Staff-observed hardware verification is the only source of truth for pass/fail. The MQTT `ir_transmit_ack` this plan adds is a diagnostic, log-only signal ("did the node's IR transmitter fire without erroring") — it never sets a test case's status, and no code in this plan infers PASS/FAIL from it.
- A session can accumulate multiple `InfraredStateCoder` rows over its lifetime (one per `FUNCTION_GENERATING` pass — the very first, plus one per retry round). `InfraredStateCoder.GetBySessionId`'s query (Plan B2) had no `ORDER BY`/`LIMIT`, which Plan B2's own final review flagged as a latent, deferred gap since B2 could only ever produce at most one coder per session. This plan makes multiple coders per session a real, expected occurrence — fixing that query is Task 4 here, not optional cleanup.
- `FUNCTION_GENERATING → TEST_CASES_GENERATING` was deliberately left untriggered by Plan B2 ("nothing here auto-advances past `FUNCTION_GENERATING`" — B2's own stated scope boundary, not an oversight). This plan closes that boundary: Task 8 extends B2's `runAnalysisAndGeneration` to launch this plan's `runTestCaseGeneration` job immediately after a coder is successfully persisted, continuing the fully-automatic pipeline B1/B2 already established (the only staff-driven steps in the whole feature remain: submitting the initial recording form, accepting/discarding raw captures, and — new in this plan — transmitting a test case and recording its observed result).
- Retry rounds never delete or overwrite prior work. A failed test case appends new `RecordCase` rows to the *same* session (continuing the monotonic `Step` sequence Plan B1 established) and produces a *new* `InfraredStateCoder` row on the next analysis pass — every attempt stays inspectable.

---

### Task 1: Domain model — `InfraredTestCase`, `InfraredTestCaseState`

**Files:**
- Modify: `backend/internal/domain/models/infrared.go`

**Interfaces:**
- Produces: `domainmodels.InfraredTestCaseStatus` (`PENDING`/`PASSED`/`FAILED`) and `domainmodels.InfraredTestCase{Id, InfraredStateCoderId, Step, Description, Status}`, `domainmodels.InfraredTestCaseState{Id, InfraredTestCaseId, InfraredStateId, StateValue}`.

`InfraredTestCase` is keyed to the coder it tests (`InfraredStateCoderId`), not the session directly — the session is reachable via `coder.InfraredRecordSessionId` when needed, exactly how the design spec frames test cases as "the LLM's own judgment about which combinations best exercise the encoder... not necessarily the same set as the recording cases."

- [ ] **Step 1: Add the types**

Append to `backend/internal/domain/models/infrared.go`:

```go
type InfraredTestCaseStatus string

const (
	InfraredTestCaseStatusPending InfraredTestCaseStatus = "PENDING"
	InfraredTestCaseStatusPassed  InfraredTestCaseStatus = "PASSED"
	InfraredTestCaseStatusFailed  InfraredTestCaseStatus = "FAILED"
)

type InfraredTestCase struct {
	Id                   uuid.UUID
	InfraredStateCoderId uuid.UUID
	Step                 int32
	Description          string
	Status               InfraredTestCaseStatus
}

type InfraredTestCaseState struct {
	Id                 uuid.UUID
	InfraredTestCaseId uuid.UUID
	InfraredStateId    uuid.UUID
	StateValue         string
}
```

- [ ] **Step 2: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/models/`
Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/domain/models/infrared.go
git commit -m "feat: add InfraredTestCase and InfraredTestCaseState domain models"
```

---

### Task 2: Migration — `infrared_test_case`, `infrared_test_case_state` tables

**Files:**
- Create: `backend/database/migrations/20260810090003_infrared_test_case.up.sql`
- Create: `backend/database/migrations/20260810090003_infrared_test_case.down.sql`

**Interfaces:**
- Produces: `infrared_test_case`, `infrared_test_case_state` tables.

- [ ] **Step 1: Write the migration**

`backend/database/migrations/20260810090003_infrared_test_case.up.sql`:

```sql
CREATE TABLE infrared_test_case (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_state_coder_id UUID NOT NULL REFERENCES infrared_state_coder (id) ON DELETE CASCADE,
    step INTEGER NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'PENDING',
    CONSTRAINT chk_infrared_test_case_status CHECK (status IN ('PENDING', 'PASSED', 'FAILED')),
    CONSTRAINT uq_infrared_test_case_coder_step UNIQUE (infrared_state_coder_id, step)
);

CREATE INDEX idx_infrared_test_case_coder_id ON infrared_test_case (infrared_state_coder_id);

CREATE TABLE infrared_test_case_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_test_case_id UUID NOT NULL REFERENCES infrared_test_case (id) ON DELETE CASCADE,
    infrared_state_id UUID NOT NULL REFERENCES infrared_state (id) ON DELETE CASCADE,
    state_value TEXT NOT NULL,
    CONSTRAINT uq_infrared_test_case_state_case_state UNIQUE (infrared_test_case_id, infrared_state_id)
);

CREATE INDEX idx_infrared_test_case_state_test_case_id ON infrared_test_case_state (infrared_test_case_id);
```

`backend/database/migrations/20260810090003_infrared_test_case.down.sql`:

```sql
DROP TABLE IF EXISTS infrared_test_case_state;
DROP TABLE IF EXISTS infrared_test_case;
```

- [ ] **Step 2: Apply and verify**

Run: `migrate -path backend/database/migrations -database "$DATABASE_URL" up`
Expected: no error. If no database is reachable in this environment, skip this step and say so explicitly — do not fabricate output — but confirm the SQL's filename/format matches sibling migrations and that `infrared_state_coder`/`infrared_state` (the two tables referenced) genuinely exist in an earlier migration.

- [ ] **Step 3: Verify build**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l database/`
Expected: no output (this task has no Go code).

- [ ] **Step 4: Commit**

```bash
git add backend/database/migrations/20260810090003_infrared_test_case.*
git commit -m "feat: add infrared_test_case and infrared_test_case_state tables"
```

---

### Task 3: `InfraredTestCase` repository

**Files:**
- Create: `backend/internal/domain/contracts/repository/infrared_test_case.go`
- Create: `backend/internal/infrastructure/repository/infrared_test_case/postgres.go`, `postgres_query.go`

**Interfaces:**
- Consumes: `domainmodels.InfraredTestCase`, `InfraredTestCaseState` (Task 1).
- Produces: `domaincontractsrepository.InfraredTestCase` with:
  - `CreateWithStates(ctx, coderId uuid.UUID, step int32, description string, states []domainmodels.InfraredTestCaseState) (id uuid.UUID, err error)`
  - `ListByCoderId(ctx, coderId uuid.UUID) ([]domainmodels.InfraredTestCase, error)`
  - `GetById(ctx, id uuid.UUID) (*domainmodels.InfraredTestCase, error)`
  - `UpdateStatusById(ctx, id uuid.UUID, status domainmodels.InfraredTestCaseStatus) error`
  - `ListStatesByTestCaseId(ctx, testCaseId uuid.UUID) ([]domainmodels.InfraredTestCaseState, error)`

One combined contract covering both tables, following the exact same "always used together" rationale Plan B1's `InfraredStateDeviceRecordCase` repository already established for its case+state+raw trio. No dedicated test, per this codebase's established repository-layer convention.

- [ ] **Step 1: Write the domain contract**

```go
// backend/internal/domain/contracts/repository/infrared_test_case.go
package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredTestCase interface {
	CreateWithStates(ctx context.Context, coderId uuid.UUID, step int32, description string, states []domainmodels.InfraredTestCaseState) (id uuid.UUID, err error)
	ListByCoderId(ctx context.Context, coderId uuid.UUID) ([]domainmodels.InfraredTestCase, error)
	GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredTestCase, error)
	UpdateStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredTestCaseStatus) error
	ListStatesByTestCaseId(ctx context.Context, testCaseId uuid.UUID) ([]domainmodels.InfraredTestCaseState, error)
}
```

- [ ] **Step 2: Write the Postgres implementation**

Same `BasePostgres`/squirrel pattern as every prior repository task in this feature. `CreateWithStates` inserts the test-case row (`RETURNING id`), then loops `states` inserting one `infrared_test_case_state` row per entry with the returned test-case id — sequential inserts, same reasoning already accepted for `InfraredStateDeviceRecordCase.CreateWithStates` in Plan B1. `UpdateStatusById` follows the established rows-affected → `NotFound`-on-zero convention (compare against `infrared_state_device_record_case/postgres.go`'s `UpdateStatusById`).

```go
// backend/internal/infrastructure/repository/infrared_test_case/postgres.go
package infrastructurerepositoryinfraredtestcase

import (
	"context"
	"errors"

	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type postgresImpl struct {
	infrastructurerepositoryshared.BasePostgres
}

func NewPostgresImpl(
	dt pgxdt.Pgxdt,
	sqrQuestion *squirrel.StatementBuilderType,
	sqrDollar *squirrel.StatementBuilderType,
) domaincontractsrepository.InfraredTestCase {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) CreateWithStates(ctx context.Context, coderId uuid.UUID, step int32, description string, states []domainmodels.InfraredTestCaseState) (uuid.UUID, error) {
	query, args, err := p.SqrD.Insert("infrared_test_case").
		Columns("infrared_state_coder_id", "step", "description").
		Values(coderId, step, description).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_test_case query", err)
	}

	var testCaseId uuid.UUID
	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&testCaseId); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_test_case", err)
	}

	for _, s := range states {
		stateQuery, stateArgs, err := p.SqrD.Insert("infrared_test_case_state").
			Columns("infrared_test_case_id", "infrared_state_id", "state_value").
			Values(testCaseId, s.InfraredStateId, s.StateValue).
			ToSql()
		if err != nil {
			return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_test_case_state query", err)
		}
		if _, err := p.Dt.Exec(ctx, stateQuery, stateArgs...); err != nil {
			return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_test_case_state", err)
		}
	}

	return testCaseId, nil
}

func (p *postgresImpl) ListByCoderId(ctx context.Context, coderId uuid.UUID) ([]domainmodels.InfraredTestCase, error) {
	query, args, err := p.SqrD.Select("id", "infrared_state_coder_id", "step", "description", "status").
		From("infrared_test_case").
		Where(squirrel.Eq{"infrared_state_coder_id": coderId}).
		OrderBy("step").
		ToSql()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build list infrared_test_case query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to list infrared_test_case", err)
	}
	defer rows.Close()

	var result []domainmodels.InfraredTestCase
	for rows.Next() {
		var item domainmodels.InfraredTestCase
		var status string
		if err := rows.Scan(&item.Id, &item.InfraredStateCoderId, &item.Step, &item.Description, &status); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_test_case", err)
		}
		item.Status = domainmodels.InfraredTestCaseStatus(status)
		result = append(result, item)
	}
	return result, nil
}

func (p *postgresImpl) GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredTestCase, error) {
	query, args, err := p.SqrD.Select("id", "infrared_state_coder_id", "step", "description", "status").
		From("infrared_test_case").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build get infrared_test_case query", err)
	}

	var item domainmodels.InfraredTestCase
	var status string
	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&item.Id, &item.InfraredStateCoderId, &item.Step, &item.Description, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_test_case not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to get infrared_test_case", err)
	}
	item.Status = domainmodels.InfraredTestCaseStatus(status)
	return &item, nil
}

func (p *postgresImpl) UpdateStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredTestCaseStatus) error {
	query, args, err := p.SqrD.Update("infrared_test_case").
		Set("status", status).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build update infrared_test_case status query", err)
	}

	tag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to update infrared_test_case status", err)
	}
	if tag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("infrared_test_case not found", nil)
	}
	return nil
}

func (p *postgresImpl) ListStatesByTestCaseId(ctx context.Context, testCaseId uuid.UUID) ([]domainmodels.InfraredTestCaseState, error) {
	query, args, err := p.SqrD.Select("id", "infrared_test_case_id", "infrared_state_id", "state_value").
		From("infrared_test_case_state").
		Where(squirrel.Eq{"infrared_test_case_id": testCaseId}).
		ToSql()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build list infrared_test_case_state query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to list infrared_test_case_state", err)
	}
	defer rows.Close()

	var result []domainmodels.InfraredTestCaseState
	for rows.Next() {
		var item domainmodels.InfraredTestCaseState
		if err := rows.Scan(&item.Id, &item.InfraredTestCaseId, &item.InfraredStateId, &item.StateValue); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_test_case_state", err)
		}
		result = append(result, item)
	}
	return result, nil
}
```

Confirm `p.Dt` (the `pgxdt.Pgxdt` interface) exposes `Exec` — every prior repository task in this feature has used `Query`/`QueryRow`; `UpdateStatusById`'s rows-affected check and `CreateWithStates`'s per-state insert both need `Exec`'s command tag. If `Exec` isn't already part of the interface this codebase uses, check how `infrared_state_device_record_case/postgres.go`'s `UpdateStatusById` gets its rows-affected count and match that exact mechanism instead.

Move the inline `Select(...)`/`Insert(...)`/`Update(...)` calls into `postgres_query.go` methods if the reviewer of this task's PR prefers matching the two-file split exactly (same latitude every repository task in this feature has been given).

- [ ] **Step 3: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/repository/ internal/infrastructure/repository/infrared_test_case/`
Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/domain/contracts/repository/infrared_test_case.go \
        backend/internal/infrastructure/repository/infrared_test_case/
git commit -m "feat: add InfraredTestCase repository"
```

---

### Task 4: Fix coder ordering and add activation to the `InfraredStateCoder` repository

**Files:**
- Modify: `backend/internal/domain/contracts/repository/infrared_state_coder.go`
- Modify: `backend/internal/infrastructure/repository/infrared_state_coder/postgres.go`, `postgres_query.go`

**Interfaces:**
- Produces: `GetBySessionId`'s query gains `ORDER BY created_at DESC LIMIT 1` (returns the LATEST coder for a session, not an arbitrary one — this plan is the first to make multiple coders per session a normal occurrence); a new `domaincontractsrepository.InfraredStateCoder.Activate(ctx, coderId uuid.UUID, deviceId uuid.UUID) error` that sets the given coder `ACTIVE` and every other coder for that device `SUPERSEDED`, in that order, as two sequential statements (no transaction wrapper — matches this codebase's established precedent for multi-statement repository operations elsewhere in this feature; a crash between the two statements leaves at most one coder without a superseding write, which the next `Activate` call self-heals by re-running the same two statements).

**Why this task exists:** Plan B2's `GetBySessionId` had no `ORDER BY`/`LIMIT` because B2 could only ever produce one coder per session — Postgres was free to return whichever single row existed. This plan is the first to make a session accumulate multiple coders (one per retry round), so an unordered, unlimited query would now return a nondeterministic result if it ever matched more than one row. B2's own final review flagged this exact gap and deferred it here.

- [ ] **Step 1: Update the domain contract**

`backend/internal/domain/contracts/repository/infrared_state_coder.go` — add one method:

```go
Activate(ctx context.Context, coderId uuid.UUID, deviceId uuid.UUID) error
```

- [ ] **Step 2: Fix `GetBySessionId`'s query and add `Activate`**

In `postgres_query.go`, add `.OrderBy("created_at DESC").Limit(1)` to the existing `GetBySessionId` query builder chain.

Add to `postgres.go`:

```go
func (p *postgresImpl) Activate(ctx context.Context, coderId uuid.UUID, deviceId uuid.UUID) error {
	activateQuery, activateArgs, err := p.SqrD.Update("infrared_state_coder").
		Set("status", domainmodels.InfraredStateCoderStatusActive).
		Where(squirrel.Eq{"id": coderId}).
		ToSql()
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build activate infrared_state_coder query", err)
	}
	if _, err := p.Dt.Exec(ctx, activateQuery, activateArgs...); err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to activate infrared_state_coder", err)
	}

	supersedeQuery, supersedeArgs, err := p.SqrD.Update("infrared_state_coder").
		Set("status", domainmodels.InfraredStateCoderStatusSuperseded).
		Where(squirrel.Eq{"infrared_device_id": deviceId, "status": domainmodels.InfraredStateCoderStatusActive}).
		Where(squirrel.NotEq{"id": coderId}).
		ToSql()
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build supersede infrared_state_coder query", err)
	}
	if _, err := p.Dt.Exec(ctx, supersedeQuery, supersedeArgs...); err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to supersede prior infrared_state_coder", err)
	}
	return nil
}
```

Note the `Where` filter on the supersede statement: `status = 'ACTIVE'` — this only ever demotes a PRIOR active coder, never touches other `UNVERIFIED`/`SUPERSEDED` rows for the same device (there is at most one `ACTIVE` coder per device at any time, by construction of this same method always being the only writer of that status).

- [ ] **Step 3: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/repository/ internal/infrastructure/repository/infrared_state_coder/`
Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/domain/contracts/repository/infrared_state_coder.go \
        backend/internal/infrastructure/repository/infrared_state_coder/
git commit -m "fix: order InfraredStateCoder.GetBySessionId by latest, add Activate"
```

---

### Task 5: MQTT transmit (publish, ack subscribe, DTO, handler)

**Files:**
- Modify: `backend/internal/domain/contracts/node/publish.go`
- Modify: `backend/internal/domain/contracts/node/subscriptions.go`
- Modify: `backend/internal/infrastructure/node/publish/mqtt.go`
- Modify: `backend/internal/infrastructure/node/subscriptions/mqtt.go`
- Create: `backend/internal/presentation/mqtt/dto/ir_transmit_ack.go`
- Create: `backend/internal/presentation/mqtt/handler/ir_transmit_ack.go`
- Modify: `backend/internal/presentation/mqtt/event/message.go`

**Interfaces:**
- Produces: `Publish.IrTransmit(ctx, nodeDeviceId string, executionId uuid.UUID, rawData []int32) error`; `Subscriptions.IrTransmitAck(ctx, nodeDeviceId string) error`; `presentationmqttdto.DecodeIrTransmitAck(deviceId string, payload []byte) (domainmodels.IrTransmitAckEvent, error)`; a new `Handler.IrTransmitAck(ctx, msg mqtt.Message, deviceId string)` method; topic routing for the new topic.

This is the transmit-direction mirror of Plan B1's `ir_capture` (receive-direction) topic — same `NodeSubTopic`/`NodePubTopic` helpers, same DTO/handler/routing shape, in reverse. The ack is diagnostic-only (see Global Constraints) — its handler logs and returns, it never calls back into the session usecase or changes any persisted state.

- [ ] **Step 1: Add the publish method**

Add to `backend/internal/domain/contracts/node/publish.go`:

```go
IrTransmit(
	ctx context.Context,
	nodeDeviceId string,
	executionId uuid.UUID,
	rawData []int32,
) (err error)
```

Implement in `backend/internal/infrastructure/node/publish/mqtt.go`, following `Action`'s exact shape (read it first) — publishes to `infrastructurenodeshared.NodeSubTopic(nodeDeviceId, "ir_transmit")` (backend publishes, node subscribes — same direction as every other `Publish` method in this file) with a JSON payload `{"execution_id": "<uuid>", "raw_data": [...]}`. Define `irTransmitQos` as a `byte` constant alongside the file's existing `actionQos`, matching its value.

- [ ] **Step 2: Add the ack subscription**

Add to `backend/internal/domain/contracts/node/subscriptions.go`:

```go
IrTransmitAck(ctx context.Context, nodeDeviceId string) (err error)
```

Implement in `backend/internal/infrastructure/node/subscriptions/mqtt.go` following the exact shape of the existing `IrCapture` method — subscribes to `infrastructurenodeshared.NodePubTopic(nodeDeviceId, "ir_transmit_ack")` (node publishes, backend subscribes). Define `irTransmitAckQos` alongside the file's existing `irCaptureQos`, matching its value.

- [ ] **Step 3: Write the DTO decoder**

`backend/internal/presentation/mqtt/dto/ir_transmit_ack.go` — the node publishes `{"execution_id": "<uuid>", "status": "SUCCESS"|"FAILED", "message": "..."}` once it has attempted to drive its IR transmitter (`message` optional, populated on failure):

```go
package presentationmqttdto

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type IrTransmitAck struct {
	ExecutionId uuid.UUID `json:"execution_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message,omitempty"`
}

type IrTransmitAckEvent struct {
	DeviceId    string
	ExecutionId uuid.UUID
	Status      string
	Message     string
}

func DecodeIrTransmitAck(deviceId string, payload []byte) (IrTransmitAckEvent, error) {
	var dto IrTransmitAck
	if err := json.Unmarshal(payload, &dto); err != nil {
		return IrTransmitAckEvent{}, domainmodels.NewError("invalid ir transmit ack payload", domainmodels.ErrTypeValidation, err)
	}
	return IrTransmitAckEvent{
		DeviceId:    deviceId,
		ExecutionId: dto.ExecutionId,
		Status:      dto.Status,
		Message:     dto.Message,
	}, nil
}
```

- [ ] **Step 4: Write the MQTT handler method**

`backend/internal/presentation/mqtt/handler/ir_transmit_ack.go` — log-only, no usecase call (see Global Constraints on why this ack never drives state):

```go
package presentationmqtthandler

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationmqttdto "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/dto"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (h *Handler) IrTransmitAck(ctx context.Context, msg mqtt.Message, deviceId string) {
	const tag = path + "/IrTransmitAck"

	event, err := presentationmqttdto.DecodeIrTransmitAck(deviceId, msg.Payload())
	if err != nil {
		h.Logger.Warn(ctx, tag, "invalid ir transmit ack payload", domainmodels.LoggerMeta{"err": err, "device_id": deviceId})
		return
	}

	if event.Status != "SUCCESS" {
		h.Logger.Warn(ctx, tag, "node reported ir transmit failure", domainmodels.LoggerMeta{
			"device_id": deviceId, "execution_id": event.ExecutionId, "message": event.Message,
		})
		return
	}
	h.Logger.Debug(ctx, tag, "ir transmit acknowledged", domainmodels.LoggerMeta{"device_id": deviceId, "execution_id": event.ExecutionId})
}
```

- [ ] **Step 5: Register the topic route and the reconnect subscription**

In `backend/internal/presentation/mqtt/event/message.go`, add a case for `"ir_transmit_ack"` calling `handler.IrTransmitAck(ctx, msg, deviceId)`, following the exact shape of the existing `"ir_capture"` entry.

In `backend/internal/application/node/messaging_callback/usecase.go`'s `subscribeNode` (the same method Plan B2's final review already extended to also subscribe `IrCapture` on every connect/reconnect), add `u.subscriptions.IrTransmitAck(ctx, deviceId)` alongside the existing calls, same error-handling shape.

- [ ] **Step 6: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/node/ internal/infrastructure/node/ internal/presentation/mqtt/ internal/application/node/messaging_callback/`
Expected: no output.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/domain/contracts/node/ \
        backend/internal/infrastructure/node/ \
        backend/internal/presentation/mqtt/dto/ir_transmit_ack.go \
        backend/internal/presentation/mqtt/handler/ir_transmit_ack.go \
        backend/internal/presentation/mqtt/event/message.go \
        backend/internal/application/node/messaging_callback/usecase.go
git commit -m "feat: add MQTT IR transmit publish and diagnostic ack subscription"
```

---

### Task 6: LLM test-case generation (`WriteTestCases`)

**Files:**
- Create: `backend/internal/application/infrared/coder_generation/llm_test_cases.go`
- Create: `backend/internal/application/infrared/coder_generation/llm_test_cases_test.go`

**Interfaces:**
- Consumes: `domaincontractsllm.Client` (merged prior plan); `domainmodels.InfraredStateCoder`, `InfraredState`, `InfraredStateDeviceDefinition`.
- Produces: `applicationinfraredcodergeneration.WriteTestCases(ctx context.Context, client domaincontractsllm.Client, deviceBrand string, deviceModel string, coder domainmodels.InfraredStateCoder, states []domainmodels.InfraredState, definitions []domainmodels.InfraredStateDeviceDefinition) ([]TestCasePlan, error)`, where `TestCasePlan{Description string, States map[uuid.UUID]string}`.

This plan's third consumer of the merged LLM provider abstraction, living alongside `WriteCoder` in the same package since both are LLM calls informed by the coder/analysis context. Unlike `WriteScript`/`WriteCoder`'s bit-level or press-order concerns, this call's only inputs are the coder's own README fields (its self-description) plus the state/definition list — the LLM proposes which state combinations are worth exercising against the *finished* encoder, which may be a different set than the recording cases (per the design spec: "its own judgment about which combinations best exercise the encoder").

- [ ] **Step 1: Write the failing test**

`backend/internal/application/infrared/coder_generation/llm_test_cases_test.go`:

```go
package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"testing"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func TestWriteTestCasesParsesStructuredResponse(t *testing.T) {
	powerId := uuid.New()
	states := []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	definitions := []domainmodels.InfraredStateDeviceDefinition{{InfraredStateId: powerId, Options: []string{"ON", "OFF"}}}
	coder := domainmodels.InfraredStateCoder{SummaryReadme: "summary", DetailReadme: "detail"}

	responseBody, _ := json.Marshal([]map[string]interface{}{
		{"description": "Turn the unit on.", "states": map[string]string{"POWER": "ON"}},
		{"description": "Turn the unit off.", "states": map[string]string{"POWER": "OFF"}},
	})
	client := &fakeLlmClient{responseText: string(responseBody)}

	plans, err := WriteTestCases(context.Background(), client, "Polytron", "PAC-09HDN", coder, states, definitions)
	if err != nil {
		t.Fatalf("WriteTestCases() error = %v, want nil", err)
	}
	if len(plans) != 2 {
		t.Fatalf("WriteTestCases() returned %d plans, want 2", len(plans))
	}
	if plans[0].Description == "" {
		t.Fatal("WriteTestCases() plan has empty description")
	}
	if plans[0].States[powerId] != "ON" {
		t.Fatalf("WriteTestCases() plan[0].States[POWER] = %q, want ON", plans[0].States[powerId])
	}

	if client.lastRequest.ResponseSchema == nil {
		t.Fatal("GenerateText() request had no ResponseSchema — structured output was not requested")
	}
}

func TestWriteTestCasesPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: context.DeadlineExceeded}
	_, err := WriteTestCases(context.Background(), client, "Polytron", "PAC-09HDN", domainmodels.InfraredStateCoder{}, nil, nil)
	if err == nil {
		t.Fatal("WriteTestCases() error = nil, want propagated error")
	}
}

func TestWriteTestCasesErrorsOnUnknownStateName(t *testing.T) {
	states := []domainmodels.InfraredState{{Id: uuid.New(), Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	responseBody, _ := json.Marshal([]map[string]interface{}{
		{"description": "bad", "states": map[string]string{"NONEXISTENT": "X"}},
	})
	client := &fakeLlmClient{responseText: string(responseBody)}

	_, err := WriteTestCases(context.Background(), client, "Polytron", "PAC-09HDN", domainmodels.InfraredStateCoder{}, states, nil)
	if err == nil {
		t.Fatal("WriteTestCases() error = nil, want an error for a state name the LLM invented")
	}
}
```

`fakeLlmClient` already exists in this package's test files from Task 8 of Plan B2 (`llm_coder_test.go`) — reuse it, don't redeclare it.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -run TestWriteTestCases -v`
Expected: FAIL — `WriteTestCases` undefined.

- [ ] **Step 3: Write the implementation**

`backend/internal/application/infrared/coder_generation/llm_test_cases.go`:

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

const testCaseResponseSchema = `{
	"type": "array",
	"items": {
		"type": "object",
		"properties": {
			"description": {"type": "string"},
			"states": {"type": "object", "additionalProperties": {"type": "string"}}
		},
		"required": ["description", "states"],
		"additionalProperties": false
	}
}`

type TestCasePlan struct {
	Description string
	States      map[uuid.UUID]string
}

type testCasePlanResponse struct {
	Description string            `json:"description"`
	States      map[string]string `json:"states"`
}

// WriteTestCases asks the LLM to propose a minimal-but-sufficient set of
// test scenarios for the finished encoder, given the coder's own
// self-description (its README fields) and the device's state/definition
// list. This is judgment about WHICH combinations to exercise, not a
// re-derivation of bit layout — the encoder itself is treated as a black
// box here.
func WriteTestCases(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	coder domainmodels.InfraredStateCoder,
	states []domainmodels.InfraredState,
	definitions []domainmodels.InfraredStateDeviceDefinition,
) ([]TestCasePlan, error) {
	stateIdByName := make(map[string]uuid.UUID, len(states))
	for _, state := range states {
		stateIdByName[state.Name] = state.Id
	}

	prompt := fmt.Sprintf(
		"Device: %s %s\n\nProtocol summary: %s\n\nProtocol detail: %s\n\nPropose a minimal but sufficient set of test scenarios to verify this encoder works correctly on real hardware. Each scenario names a full target state (every field) and a short description of what a technician should observe. Respond as a JSON array matching the given schema.",
		deviceBrand, deviceModel, coder.SummaryReadme, coder.DetailReadme,
	)

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are an expert at designing minimal, high-coverage test plans for infrared remote control encoders.",
		Prompt:          prompt,
		MaxOutputTokens: 4096,
		ResponseSchema:  []byte(testCaseResponseSchema),
	})
	if err != nil {
		return nil, domainmodels.NewError("failed to generate test cases", domainmodels.ErrTypeFailure, err)
	}

	var responses []testCasePlanResponse
	if err := json.Unmarshal([]byte(result.Text), &responses); err != nil {
		return nil, domainmodels.NewError("llm returned malformed test case response", domainmodels.ErrTypeFailure, err)
	}

	plans := make([]TestCasePlan, 0, len(responses))
	for _, r := range responses {
		states := make(map[uuid.UUID]string, len(r.States))
		for name, value := range r.States {
			stateId, ok := stateIdByName[name]
			if !ok {
				return nil, domainmodels.NewError(fmt.Sprintf("llm referenced unknown state %q", name), domainmodels.ErrTypeFailure, nil)
			}
			states[stateId] = value
		}
		plans = append(plans, TestCasePlan{Description: r.Description, States: states})
	}
	return plans, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -v`
Expected: PASS on every test in the package, including Plan B2's Task 8 tests.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/application/infrared/coder_generation/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/infrared/coder_generation/llm_test_cases.go backend/internal/application/infrared/coder_generation/llm_test_cases_test.go
git commit -m "feat: add LLM-driven test case generation"
```

---

### Task 7: LLM retry-case generation (`WriteRetryCases`)

**Files:**
- Create: `backend/internal/application/infrared/coder_generation/llm_retry_cases.go`
- Create: `backend/internal/application/infrared/coder_generation/llm_retry_cases_test.go`

**Interfaces:**
- Consumes: `domaincontractsllm.Client`; `domainmodels.InfraredStateCoder`, `InfraredState`.
- Produces: `applicationinfraredcodergeneration.WriteRetryCases(ctx context.Context, client domaincontractsllm.Client, deviceBrand string, deviceModel string, coder domainmodels.InfraredStateCoder, failedTestCaseStates map[uuid.UUID]string, states []domainmodels.InfraredState) ([]RetryCasePlan, error)`, where `RetryCasePlan{Description string, States map[uuid.UUID]string}` (deliberately the same shape as `TestCasePlan` — both represent "a target state plus a human-facing description," just consumed differently downstream — kept as two distinct named types rather than one shared type since they represent different domain concepts that happen to coincide in shape today, and diverging independently later should not require a breaking rename of a shared type).

This plan's fourth and final LLM-abstraction consumer, resolving the design spec's own explicitly-deferred open item: "How `InfraredTestCase` failures translate into which new `RecordCase` rows get generated — likely an LLM call given the failing test case's declared state and the current coder's README/analysis context." Given a failed test case's target state and the coder's own self-description, the LLM proposes new targeted recording scenarios likely to explain the discrepancy (e.g. a state combination near the failure that wasn't in the original OFAT sweep, or a repeat of the exact failing state for a cleaner capture).

- [ ] **Step 1: Write the failing test**

`backend/internal/application/infrared/coder_generation/llm_retry_cases_test.go`:

```go
package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func TestWriteRetryCasesParsesStructuredResponse(t *testing.T) {
	powerId := uuid.New()
	states := []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	coder := domainmodels.InfraredStateCoder{SummaryReadme: "summary", DetailReadme: "detail"}
	failedState := map[uuid.UUID]string{powerId: "OFF"}

	responseBody, _ := json.Marshal([]map[string]interface{}{
		{"description": "Re-record POWER OFF.", "states": map[string]string{"POWER": "OFF"}},
	})
	client := &fakeLlmClient{responseText: string(responseBody)}

	plans, err := WriteRetryCases(context.Background(), client, "Polytron", "PAC-09HDN", coder, failedState, states)
	if err != nil {
		t.Fatalf("WriteRetryCases() error = %v, want nil", err)
	}
	if len(plans) != 1 {
		t.Fatalf("WriteRetryCases() returned %d plans, want 1", len(plans))
	}
	if plans[0].States[powerId] != "OFF" {
		t.Fatalf("WriteRetryCases() plan[0].States[POWER] = %q, want OFF", plans[0].States[powerId])
	}
}

func TestWriteRetryCasesPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: context.DeadlineExceeded}
	_, err := WriteRetryCases(context.Background(), client, "Polytron", "PAC-09HDN", domainmodels.InfraredStateCoder{}, nil, nil)
	if err == nil {
		t.Fatal("WriteRetryCases() error = nil, want propagated error")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -run TestWriteRetryCases -v`
Expected: FAIL — `WriteRetryCases` undefined.

- [ ] **Step 3: Write the implementation**

`backend/internal/application/infrared/coder_generation/llm_retry_cases.go`:

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

type RetryCasePlan struct {
	Description string
	States      map[uuid.UUID]string
}

// WriteRetryCases asks the LLM to propose new targeted recording scenarios
// after a test case failed hardware verification, given the failing
// scenario's target state and the current coder's own self-description.
// Reuses testCaseResponseSchema's shape (description + states object) —
// the LLM's output contract is identical to WriteTestCases', even though
// the two calls serve different points in the session lifecycle.
func WriteRetryCases(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	coder domainmodels.InfraredStateCoder,
	failedTestCaseStates map[uuid.UUID]string,
	states []domainmodels.InfraredState,
) ([]RetryCasePlan, error) {
	stateNameById := make(map[string]string, len(states))
	stateIdByName := make(map[string]uuid.UUID, len(states))
	for _, state := range states {
		stateNameById[state.Id.String()] = state.Name
		stateIdByName[state.Name] = state.Id
	}

	failedDescription := ""
	for stateId, value := range failedTestCaseStates {
		failedDescription += fmt.Sprintf("%s=%s ", stateNameById[stateId.String()], value)
	}

	prompt := fmt.Sprintf(
		"Device: %s %s\n\nProtocol summary: %s\n\nProtocol detail: %s\n\nA hardware test transmitting this target state failed: %s\n\nPropose new targeted recording scenarios (full target state plus a short description) likely to explain the discrepancy and improve the encoder. Respond as a JSON array matching the given schema.",
		deviceBrand, deviceModel, coder.SummaryReadme, coder.DetailReadme, failedDescription,
	)

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are an expert at diagnosing infrared remote control encoder failures and proposing corrective recording scenarios.",
		Prompt:          prompt,
		MaxOutputTokens: 4096,
		ResponseSchema:  []byte(testCaseResponseSchema),
	})
	if err != nil {
		return nil, domainmodels.NewError("failed to generate retry cases", domainmodels.ErrTypeFailure, err)
	}

	var responses []testCasePlanResponse
	if err := json.Unmarshal([]byte(result.Text), &responses); err != nil {
		return nil, domainmodels.NewError("llm returned malformed retry case response", domainmodels.ErrTypeFailure, err)
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

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/coder_generation/... -v`
Expected: PASS on every test in the package.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/application/infrared/coder_generation/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/infrared/coder_generation/llm_retry_cases.go backend/internal/application/infrared/coder_generation/llm_retry_cases_test.go
git commit -m "feat: add LLM-driven retry case generation"
```

---

### Task 8: Test-case generation async job (extends B2's `runAnalysisAndGeneration`)

**Files:**
- Modify: `backend/internal/application/infrared/record_session_management/usecase.go`
- Modify: `backend/internal/application/infrared/record_session_management/usecase_test.go`

**Interfaces:**
- Consumes: `applicationinfraredcodergeneration.WriteTestCases` (Task 6); `domaincontractsrepository.InfraredTestCase` (Task 3).
- Produces: a new `runTestCaseGeneration(sessionId uuid.UUID)` async job, launched from the end of Plan B2's `runAnalysisAndGeneration` immediately after a coder is successfully persisted (closing the `FUNCTION_GENERATING → TEST_CASES_GENERATING` boundary B2 deliberately left open).

**Read `runAnalysisAndGeneration`'s current body in full before editing** — this task adds exactly one call at its very end (right after `u.coder.Create(...)` succeeds) and one new method; it does not change anything about B2's analysis/generation logic itself.

- [ ] **Step 1: Add the constructor dependency and interface method**

Add `testCase domaincontractsrepository.InfraredTestCase` to the `usecase` struct and `NewUsecaseImpl`'s parameter list, appended after the existing `coder` parameter and before `logger`. Update every existing `NewUsecaseImpl(...)` call site in `usecase_test.go` — there are many, from this feature's two prior plans — with a new `&fakeTestCaseRepository{}` at that exact position; do not reorder any existing argument.

Add to `backend/internal/domain/usecases/infrared/record_session_management.go`:

```go
ListTestCases(ctx context.Context, sessionId uuid.UUID) ([]TestCaseWithStates, error)
```

with a new DTO alongside the file's existing `CaseWithStatesAndRaw`:

```go
type TestCaseWithStates struct {
	TestCase domainmodels.InfraredTestCase
	States   []domainmodels.InfraredTestCaseState
}
```

- [ ] **Step 2: Write the failing tests**

Add to `usecase_test.go` — a `fakeTestCaseRepository` fake (following this file's established mutex+snapshot-accessor convention for any field a background goroutine might write concurrently with a test read) and tests covering: `runTestCaseGeneration` persists one test case per `WriteTestCases` plan and transitions the session to `TESTING`; it transitions to `FAILED` if `WriteTestCases` errors; `runAnalysisAndGeneration`'s existing successful-persist test (from Plan B2) now ALSO ends with a `TESTING` transition, not stopping at `FUNCTION_GENERATING` (update that existing test's assertions — it is no longer testing B2's old scope boundary, since this task deliberately closes it).

```go
type fakeTestCaseRepository struct {
	domaincontractsrepository.InfraredTestCase
	mu             sync.Mutex
	createdTestCaseIds []uuid.UUID
}

func (f *fakeTestCaseRepository) CreateWithStates(_ context.Context, _ uuid.UUID, _ int32, _ string, _ []domainmodels.InfraredTestCaseState) (uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := uuid.New()
	f.createdTestCaseIds = append(f.createdTestCaseIds, id)
	return id, nil
}

func (f *fakeTestCaseRepository) CreatedTestCaseIds() []uuid.UUID {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]uuid.UUID(nil), f.createdTestCaseIds...)
}

func TestRunTestCaseGenerationPersistsOneTestCasePerPlanAndTransitionsToTesting(t *testing.T) {
	sessionId := uuid.New()
	coderId := uuid.New()
	powerId := uuid.New()

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}}
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: coderId, SummaryReadme: "s", DetailReadme: "d"}}
	testCaseRepo := &fakeTestCaseRepository{}
	llmFactory := &fakeLlmClientFactory{responseText: `[{"description": "d1", "states": {"POWER": "ON"}}, {"description": "d2", "states": {"POWER": "OFF"}}]`}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		stateRepo, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &noopLogger{},
	).(*usecase)

	impl.runTestCaseGeneration(sessionId, coderId)

	if len(testCaseRepo.CreatedTestCaseIds()) != 2 {
		t.Fatalf("created test case ids = %v, want 2", testCaseRepo.CreatedTestCaseIds())
	}
	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateTesting {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include TESTING", sessionRepo.StatusUpdates())
	}
}

func TestRunTestCaseGenerationFailsSessionWhenLlmErrors(t *testing.T) {
	sessionId := uuid.New()
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: uuid.New()}}
	testCaseRepo := &fakeTestCaseRepository{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{err: errors.New("llm down")}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &noopLogger{},
	).(*usecase)

	impl.runTestCaseGeneration(sessionId, uuid.New())

	if len(testCaseRepo.CreatedTestCaseIds()) != 0 {
		t.Fatal("test cases were created despite the llm call failing")
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
```

**`fakeLlmClientFactory` needs a real extension, not just a note to "handle it":** `runAnalysisAndGeneration` and `runTestCaseGeneration` are now two separate calls to `u.llmFactory.Current(ctx)` in the same test (one per LLM step), and the real `fakeLlmClientFactory.Current` (`usecase_test.go`) currently has a single `responseText string` field returned identically on every call — insufficient once a test needs `WriteCoder` and `WriteTestCases` to see two different JSON shapes in sequence. Extend it with a queue, keeping every existing single-response call site unchanged:

```go
type fakeLlmClientFactory struct {
	err           error
	responseText  string   // existing field — single-response tests keep using this
	responseTexts []string // new: if set, Current() returns these in order, one per call
	callIndex     int
}

func (f *fakeLlmClientFactory) Current(_ context.Context) (domaincontractsllm.Client, error) {
	if f.err != nil {
		return nil, f.err
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

Then update Plan B2's existing `TestRunAnalysisAndGenerationPersistsCoderAndStaysInFunctionGenerating` test (rename it to reflect the new behavior, e.g. `TestRunAnalysisAndGenerationPersistsCoderAndAdvancesToTestCaseGeneration`): give it a `testCaseRepo := &fakeTestCaseRepository{}` passed into `NewUsecaseImpl` at the new position, change its `llmFactory` to `&fakeLlmClientFactory{responseTexts: []string{<the existing WriteCoder-shaped JSON object>, <a WriteTestCases-shaped JSON array, e.g. \`[{"description": "d1", "states": {"POWER": "ON"}}]\`>}}`, and add an assertion that status updates include `TESTING`, not just `FUNCTION_GENERATING`.

- [ ] **Step 3: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -v`
Expected: FAIL — compile errors from the new constructor parameter, and the updated B2 test's new assertion failing against the old behavior.

- [ ] **Step 4: Write the implementation**

At the very end of `runAnalysisAndGeneration`, replace the function's final line (the `u.coder.Create(...)` call's success path, which previously just fell off the end of the function) with a launch of the new job:

```go
	coderId, err := u.coder.Create(ctx, domainmodels.InfraredStateCoder{
		InfraredDeviceId:        session.InfraredDeviceId,
		InfraredRecordSessionId: sessionId,
		EncoderSource:           written.EncoderSource,
		DecoderSource:           written.DecoderSource,
		SummaryReadme:           written.SummaryReadme,
		DetailReadme:            written.DetailReadme,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to persist coder", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	go u.runTestCaseGeneration(sessionId, coderId)
```

(Find the exact existing `u.coder.Create(ctx, domainmodels.InfraredStateCoder{...})` call and its error-handling block in `usecase.go` — if it doesn't already capture the returned id into a variable, since B2's original code discarded it, add that; then append the `go u.runTestCaseGeneration(...)` line after the success branch.)

Add the new job:

```go
const testCaseGenerationTimeout = 2 * time.Minute

func (u *usecase) runTestCaseGeneration(sessionId uuid.UUID, coderId uuid.UUID) {
	const tag = "infrared/record_session_management/runTestCaseGeneration"

	ctx, cancel := context.WithTimeout(context.Background(), testCaseGenerationTimeout)
	defer cancel()

	coder, err := u.coder.GetBySessionId(ctx, sessionId)
	if err != nil || coder == nil {
		u.logger.Error(ctx, tag, "failed to look up coder", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	session, err := u.session.GetById(ctx, sessionId)
	if err != nil || session == nil {
		u.logger.Error(ctx, tag, "failed to look up session", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	device, err := u.device.GetById(ctx, session.InfraredDeviceId)
	if err != nil || device == nil {
		u.logger.Error(ctx, tag, "failed to look up device", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	states, err := u.state.ListByDeviceTypeId(ctx, device.InfraredDeviceTypeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list states", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	definitions, err := u.definition.ListByDeviceId(ctx, session.InfraredDeviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list definitions", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	client, err := u.llmFactory.Current(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve llm client", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	plans, err := applicationinfraredcodergeneration.WriteTestCases(ctx, client, device.Brand, device.Model, *coder, states, definitions)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to write test cases", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	for i, plan := range plans {
		testCaseStates := make([]domainmodels.InfraredTestCaseState, 0, len(plan.States))
		for stateId, value := range plan.States {
			testCaseStates = append(testCaseStates, domainmodels.InfraredTestCaseState{InfraredStateId: stateId, StateValue: value})
		}
		if _, err := u.testCase.CreateWithStates(ctx, coderId, int32(i+1), plan.Description, testCaseStates); err != nil {
			u.logger.Error(ctx, tag, "failed to persist test case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "step": i + 1})
			u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
			return
		}
	}

	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateTesting)
}

func (u *usecase) ListTestCases(ctx context.Context, sessionId uuid.UUID) ([]domainusecasesinfrared.TestCaseWithStates, error) {
	coder, err := u.coder.GetBySessionId(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	if coder == nil {
		return nil, nil
	}
	testCases, err := u.testCase.ListByCoderId(ctx, coder.Id)
	if err != nil {
		return nil, err
	}
	result := make([]domainusecasesinfrared.TestCaseWithStates, 0, len(testCases))
	for _, tc := range testCases {
		states, err := u.testCase.ListStatesByTestCaseId(ctx, tc.Id)
		if err != nil {
			return nil, err
		}
		result = append(result, domainusecasesinfrared.TestCaseWithStates{TestCase: tc, States: states})
	}
	return result, nil
}
```

Add `applicationinfraredcodergeneration "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/coder_generation"` to `usecase.go`'s imports if not already present (Task 8 of Plan B2 already imports it for `WriteCoder` — reuse the same import line, don't duplicate).

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -v`
Expected: PASS on every test in the package, including every test from Plans B1/B2 (except the one you deliberately updated in Step 2) and the new ones above.

- [ ] **Step 6: Verify**

Run: `cd backend && go build ./... 2>&1 | grep -v "composition/main"` (composition wiring is Task 12; an isolated failure there is expected) `&& go vet ./internal/domain/usecases/infrared/... ./internal/application/infrared/record_session_management/... && gofmt -l internal/domain/usecases/infrared/ internal/application/infrared/record_session_management/`
Expected: no output beyond the expected composition/main isolation.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/domain/usecases/infrared/record_session_management.go \
        backend/internal/application/infrared/record_session_management/
git commit -m "feat: add test case generation async job, close FUNCTION_GENERATING to TEST_CASES_GENERATING boundary"
```

---

### Task 9: Transmit a test case (`TransmitTestCase`)

**Files:**
- Modify: `backend/internal/domain/usecases/infrared/record_session_management.go`
- Modify: `backend/internal/application/infrared/record_session_management/usecase.go`
- Modify: `backend/internal/application/infrared/record_session_management/usecase_test.go`

**Interfaces:**
- Consumes: `infrastructurejsengine.RunEncoder` (via the existing `EncoderRunner` interface, Plan B2), `domaincontractsnode.Publish.IrTransmit` (Task 5).
- Produces: `RecordSessionManagement.TransmitTestCase(ctx context.Context, testCaseId uuid.UUID) error` — staff-triggered: runs the coder's `EncoderSource` against the test case's declared state (via goja, same smoke-test mechanism Plan B2 already built), then publishes the resulting raw duration array to the session's node over MQTT. Does not wait for or require the `ir_transmit_ack` (Task 5) — that ack is diagnostic-only (see Global Constraints).

This is staff-triggered (an explicit HTTP call, wired in Task 11), not automatic, because staff need to be physically ready to observe the target device before the signal fires — unlike every other transition in this feature, there is no sensible "as soon as possible" moment for a hardware action a human needs to watch.

- [ ] **Step 1: Add the `Publish` dependency and interface method**

Add `TransmitTestCase(ctx context.Context, testCaseId uuid.UUID) error` to `RecordSessionManagement` in `record_session_management.go`.

Add `publish domaincontractsnode.Publish` to the `usecase` struct and `NewUsecaseImpl`'s parameter list in `usecase.go`, appended after the existing `testCase` parameter (Task 8) and before `logger`. Update every `NewUsecaseImpl(...)` call site in `usecase_test.go` with a new `&fakePublish{}` at that position.

- [ ] **Step 2: Write the failing tests**

Add to `usecase_test.go`:

```go
type fakePublish struct {
	domaincontractsnode.Publish
	mu                sync.Mutex
	transmittedRawData []int32
	transmitCalls      int
}

func (f *fakePublish) IrTransmit(_ context.Context, _ string, _ uuid.UUID, rawData []int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.transmittedRawData = rawData
	f.transmitCalls++
	return nil
}

func TestTransmitTestCaseRunsEncoderAndPublishesResult(t *testing.T) {
	testCaseId := uuid.New()
	coderId := uuid.New()
	powerId := uuid.New()
	nodeId := uuid.New()
	sessionId := uuid.New()
	deviceId := "AC276E5E030C"

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: testCaseId, InfraredStateCoderId: coderId},
		listStatesByTestCaseIdResult: []domainmodels.InfraredTestCaseState{
			{InfraredTestCaseId: testCaseId, InfraredStateId: powerId, StateValue: "ON"},
		},
	}
	coderRepo := &fakeCoderRepository{getByIdResult: &domainmodels.InfraredStateCoder{Id: coderId, InfraredRecordSessionId: sessionId, EncoderSource: "function encode(state) { return [9000, 4500]; }"}}
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, NodeId: nodeId}}
	nodeRepo := &fakeNodeRepository{result: &domainmodels.Node{Id: nodeId, DeviceId: deviceId}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER"}}}
	publish := &fakePublish{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		stateRepo, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, nodeRepo, &fakeEncoderRunner{}, coderRepo, testCaseRepo, publish, &noopLogger{},
	)

	if err := impl.TransmitTestCase(context.Background(), testCaseId); err != nil {
		t.Fatalf("TransmitTestCase() error = %v, want nil", err)
	}
	if publish.transmitCalls != 1 {
		t.Fatalf("IrTransmit() calls = %d, want 1", publish.transmitCalls)
	}
}
```

`fakeCoderRepository` (Plan B2) currently only has a `getResult` field wired to `GetBySessionId` — it has no `GetById` method at all (Task 9 is the first task to add `GetById` to the real `InfraredStateCoder` repository contract). Add a distinct `getByIdResult *domainmodels.InfraredStateCoder` field and a new method:

```go
func (f *fakeCoderRepository) GetById(_ context.Context, _ uuid.UUID) (*domainmodels.InfraredStateCoder, error) {
	return f.getByIdResult, nil
}
```

Do not repurpose the existing `getResult`/`GetBySessionId` field for this — they serve a different lookup path and a future test may need both set to different values on the same fake instance.

Similarly extend `fakeTestCaseRepository` (Task 8) with `getResult *domainmodels.InfraredTestCase` + a `GetById` method, and `listStatesByTestCaseIdResult []domainmodels.InfraredTestCaseState` + a `ListStatesByTestCaseId` method, following the same pattern.

- [ ] **Step 3: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run TestTransmitTestCase -v`
Expected: FAIL — compile errors from the new constructor parameter and undefined method.

- [ ] **Step 4: Write the implementation**

```go
func (u *usecase) TransmitTestCase(ctx context.Context, testCaseId uuid.UUID) error {
	const tag = "infrared/record_session_management/TransmitTestCase"

	testCase, err := u.testCase.GetById(ctx, testCaseId)
	if err != nil {
		return err
	}
	if testCase == nil {
		return domainmodels.NewError("test case not found", domainmodels.ErrTypeNotFound, nil)
	}

	testCaseStates, err := u.testCase.ListStatesByTestCaseId(ctx, testCaseId)
	if err != nil {
		return err
	}

	coder, err := u.coderById(ctx, testCase.InfraredStateCoderId)
	if err != nil {
		return err
	}

	session, err := u.session.GetById(ctx, coder.InfraredRecordSessionId)
	if err != nil || session == nil {
		u.logger.Error(ctx, tag, "failed to look up session for test case", domainmodels.LoggerMeta{"err": err, "test_case_id": testCaseId})
		return err
	}
	node, err := u.node.ReadById(ctx, session.NodeId)
	if err != nil || node == nil {
		u.logger.Error(ctx, tag, "failed to look up node for test case", domainmodels.LoggerMeta{"err": err, "test_case_id": testCaseId})
		return err
	}

	device, err := u.device.GetById(ctx, session.InfraredDeviceId)
	if err != nil || device == nil {
		return err
	}
	states, err := u.state.ListByDeviceTypeId(ctx, device.InfraredDeviceTypeId)
	if err != nil {
		return err
	}

	stateByName := make(map[string]string, len(testCaseStates))
	for _, s := range testCaseStates {
		stateByName[stateIdToNameLookup(states, s.InfraredStateId)] = s.StateValue
	}

	rawData, err := u.encoderRunner.RunEncoder(coder.EncoderSource, stateByName, encoderSmokeTestTimeout)
	if err != nil {
		u.logger.Error(ctx, tag, "encoder failed while transmitting test case", domainmodels.LoggerMeta{"err": err, "test_case_id": testCaseId})
		return err
	}

	executionId := uuid.New()
	if err := u.publish.IrTransmit(ctx, node.DeviceId, executionId, rawData); err != nil {
		u.logger.Error(ctx, tag, "failed to publish ir transmit", domainmodels.LoggerMeta{"err": err, "test_case_id": testCaseId})
		return err
	}
	return nil
}

// coderById is a small helper since InfraredStateCoder's repository only
// exposes GetBySessionId (Task 3 of this plan deliberately did not add a
// GetById, since nothing before this task ever needed to look a coder up
// by its own id) — TransmitTestCase is the first caller that has a coder id
// (from the test case) but not yet a session id, so it must go by coder id
// specifically. Add GetById to the InfraredStateCoder repository contract
// and Postgres implementation as part of this task (same shape as every
// other repository's GetById in this feature), rather than routing through
// GetBySessionId a second time.
func (u *usecase) coderById(ctx context.Context, coderId uuid.UUID) (*domainmodels.InfraredStateCoder, error) {
	return u.coder.GetById(ctx, coderId)
}
```

**This task also adds `GetById(ctx, id uuid.UUID) (*domainmodels.InfraredStateCoder, error)` to `domaincontractsrepository.InfraredStateCoder`** (contract in `backend/internal/domain/contracts/repository/infrared_state_coder.go`, implementation in `backend/internal/infrastructure/repository/infrared_state_coder/postgres.go`/`postgres_query.go`, same pattern as every other repository's `GetById` in this feature) — flagged explicitly here rather than silently added, since Task 4 already touched this same repository file for a different reason and a later reader diffing Task 4 against this task should know the second change is deliberate, not a merge artifact.

`stateIdToNameLookup` is a small helper (`func stateIdToNameLookup(states []domainmodels.InfraredState, id uuid.UUID) string`) scanning `states` for a matching `Id` and returning its `Name` — write it as a small loop; this is a different shape from the existing `stateIdToName` (Plan B2) which converts a whole map at once, so don't try to force-reuse that one, but keep both colocated in the same file since they solve the same underlying UUID-vs-name problem.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -v`
Expected: PASS on every test in the package.

- [ ] **Step 6: Verify**

Run: `cd backend && go build ./... 2>&1 | grep -v "composition/main" && go vet ./internal/domain/contracts/repository/... ./internal/infrastructure/repository/infrared_state_coder/... ./internal/domain/usecases/infrared/... ./internal/application/infrared/record_session_management/... && gofmt -l internal/domain/contracts/repository/ internal/infrastructure/repository/infrared_state_coder/ internal/domain/usecases/infrared/ internal/application/infrared/record_session_management/`
Expected: no output beyond the expected composition/main isolation.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/domain/contracts/repository/infrared_state_coder.go \
        backend/internal/infrastructure/repository/infrared_state_coder/ \
        backend/internal/domain/usecases/infrared/record_session_management.go \
        backend/internal/application/infrared/record_session_management/
git commit -m "feat: add TransmitTestCase (encoder run + MQTT publish)"
```

---

### Task 10: Record a test result and the retry-loop trigger (`RecordTestCaseResult`)

**Files:**
- Modify: `backend/internal/domain/usecases/infrared/record_session_management.go`
- Modify: `backend/internal/application/infrared/record_session_management/usecase.go`
- Modify: `backend/internal/application/infrared/record_session_management/usecase_test.go`

**Interfaces:**
- Consumes: `applicationinfraredcodergeneration.WriteRetryCases` (Task 7).
- Produces: `RecordSessionManagement.RecordTestCaseResult(ctx context.Context, testCaseId uuid.UUID, passed bool) error`. Marks the given test case `PASSED`/`FAILED`. If any `PENDING` test case remains for the same coder, that's all this call does (staff works through them one at a time, same pattern as `AcceptRaw`'s cursor logic — except test cases have no ordering cursor, staff can transmit/record any of them in any order). Once every test case for the coder has a terminal status: if ALL passed, transition the session to `COMPLETED` and activate the coder (superseding any prior active coder for the device); if ANY failed, launch this task's new async job (`runRetryCaseGeneration`) which appends new targeted `RecordCase` rows and transitions the session back to `RECORDING`.

- [ ] **Step 1: Add the interface method**

Add to `record_session_management.go`:

```go
RecordTestCaseResult(ctx context.Context, testCaseId uuid.UUID, passed bool) error
```

- [ ] **Step 2: Write the failing tests**

Add to `usecase_test.go` three tests: (a) recording a result when other test cases are still `PENDING` only updates that one case's status and does nothing else; (b) recording the last remaining result as PASSED, with every other case already PASSED, transitions to `COMPLETED` and calls `Activate`; (c) recording the last remaining result as FAILED, with every other case PASSED, launches the retry job (assert the session eventually re-enters `RECORDING` via `sessionRepo.StatusUpdates()`, polling with a timeout the same way Plan B1's async-job tests already do, since the retry job runs on its own goroutine).

Extend `fakeTestCaseRepository` with a `listByCoderIdResult []domainmodels.InfraredTestCase` field and `ListByCoderId` override (following the file's established pattern), and extend `fakeCoderRepository` with an `activateCalls int`/`Activate` override.

```go
func (f *fakeTestCaseRepository) ListByCoderId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredTestCase, error) {
	return f.listByCoderIdResult, nil
}

func (f *fakeCoderRepository) Activate(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	f.activateCalls++
	return nil
}
```

```go
func TestRecordTestCaseResultDoesNothingElseWhilePendingCasesRemain(t *testing.T) {
	coderId := uuid.New()
	firstCaseId, secondCaseId := uuid.New(), uuid.New()

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: firstCaseId, InfraredStateCoderId: coderId},
		listByCoderIdResult: []domainmodels.InfraredTestCase{
			{Id: firstCaseId, InfraredStateCoderId: coderId, Status: domainmodels.InfraredTestCaseStatusPassed},
			{Id: secondCaseId, InfraredStateCoderId: coderId, Status: domainmodels.InfraredTestCaseStatusPending},
		},
	}
	coderRepo := &fakeCoderRepository{}
	sessionRepo := &fakeSessionRepository{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	)

	if err := impl.RecordTestCaseResult(context.Background(), firstCaseId, true); err != nil {
		t.Fatalf("RecordTestCaseResult() error = %v, want nil", err)
	}
	if coderRepo.activateCalls != 0 {
		t.Fatal("Activate() was called despite a pending test case remaining")
	}
	if len(sessionRepo.StatusUpdates()) != 0 {
		t.Fatalf("session status updates = %v, want none (session should not transition while cases are still pending)", sessionRepo.StatusUpdates())
	}
}

func TestRecordTestCaseResultCompletesSessionWhenAllPass(t *testing.T) {
	coderId := uuid.New()
	deviceId := uuid.New()
	sessionId := uuid.New()
	onlyCaseId := uuid.New()

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: onlyCaseId, InfraredStateCoderId: coderId},
		listByCoderIdResult: []domainmodels.InfraredTestCase{
			{Id: onlyCaseId, InfraredStateCoderId: coderId, Status: domainmodels.InfraredTestCaseStatusPassed},
		},
	}
	coderRepo := &fakeCoderRepository{getByIdResult: &domainmodels.InfraredStateCoder{Id: coderId, InfraredDeviceId: deviceId, InfraredRecordSessionId: sessionId}}
	sessionRepo := &fakeSessionRepository{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	)

	if err := impl.RecordTestCaseResult(context.Background(), onlyCaseId, true); err != nil {
		t.Fatalf("RecordTestCaseResult() error = %v, want nil", err)
	}
	if coderRepo.activateCalls != 1 {
		t.Fatalf("Activate() calls = %d, want 1", coderRepo.activateCalls)
	}
	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateCompleted {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include COMPLETED", sessionRepo.StatusUpdates())
	}
}

func TestRecordTestCaseResultTriggersRetryLoopWhenAnyFail(t *testing.T) {
	coderId := uuid.New()
	deviceId := uuid.New()
	sessionId := uuid.New()
	powerId := uuid.New()
	failedCaseId := uuid.New()

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: failedCaseId, InfraredStateCoderId: coderId},
		listByCoderIdResult: []domainmodels.InfraredTestCase{
			{Id: failedCaseId, InfraredStateCoderId: coderId, Status: domainmodels.InfraredTestCaseStatusFailed},
		},
		listStatesByTestCaseIdResult: []domainmodels.InfraredTestCaseState{
			{InfraredTestCaseId: failedCaseId, InfraredStateId: powerId, StateValue: "OFF"},
		},
	}
	coderRepo := &fakeCoderRepository{getByIdResult: &domainmodels.InfraredStateCoder{Id: coderId, InfraredDeviceId: deviceId, InfraredRecordSessionId: sessionId}}
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: deviceId}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER"}}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: uuid.New(), InfraredRecordSessionId: sessionId, Step: 3, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
	}
	llmFactory := &fakeLlmClientFactory{responseText: `[{"description": "Re-record POWER OFF.", "states": {"POWER": "OFF"}}]`}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	)

	if err := impl.RecordTestCaseResult(context.Background(), failedCaseId, false); err != nil {
		t.Fatalf("RecordTestCaseResult() error = %v, want nil", err)
	}

	// runRetryCaseGeneration runs on its own goroutine — poll with a timeout,
	// the same pattern Plan B1's async-job tests already established.
	deadline := time.Now().Add(2 * time.Second)
	found := false
	for time.Now().Before(deadline) && !found {
		for _, s := range sessionRepo.StatusUpdates() {
			if s == domainmodels.InfraredRecordingStateRecording {
				found = true
			}
		}
		if !found {
			time.Sleep(5 * time.Millisecond)
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to eventually include RECORDING", sessionRepo.StatusUpdates())
	}
	if coderRepo.activateCalls != 0 {
		t.Fatal("Activate() was called despite a failed test case")
	}
}
```

Every `NewUsecaseImpl(...)` call above already includes the `testCase`/`publish` arguments Tasks 8-9 added — this task (Task 10) adds no further constructor parameters of its own, only new interface methods on the existing usecase.

- [ ] **Step 3: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run TestRecordTestCaseResult -v`
Expected: FAIL — undefined method.

- [ ] **Step 4: Write the implementation**

```go
func (u *usecase) RecordTestCaseResult(ctx context.Context, testCaseId uuid.UUID, passed bool) error {
	const tag = "infrared/record_session_management/RecordTestCaseResult"

	status := domainmodels.InfraredTestCaseStatusFailed
	if passed {
		status = domainmodels.InfraredTestCaseStatusPassed
	}
	if err := u.testCase.UpdateStatusById(ctx, testCaseId, status); err != nil {
		u.logger.Error(ctx, tag, "failed to update test case status", domainmodels.LoggerMeta{"err": err, "test_case_id": testCaseId})
		return err
	}

	testCase, err := u.testCase.GetById(ctx, testCaseId)
	if err != nil || testCase == nil {
		u.logger.Error(ctx, tag, "failed to look up updated test case", domainmodels.LoggerMeta{"err": err, "test_case_id": testCaseId})
		return err
	}

	allTestCases, err := u.testCase.ListByCoderId(ctx, testCase.InfraredStateCoderId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list test cases for coder", domainmodels.LoggerMeta{"err": err, "coder_id": testCase.InfraredStateCoderId})
		return err
	}

	allTerminal := true
	anyFailed := false
	for _, tc := range allTestCases {
		if tc.Status == domainmodels.InfraredTestCaseStatusPending {
			allTerminal = false
			break
		}
		if tc.Status == domainmodels.InfraredTestCaseStatusFailed {
			anyFailed = true
		}
	}
	if !allTerminal {
		return nil
	}

	coder, err := u.coderById(ctx, testCase.InfraredStateCoderId)
	if err != nil || coder == nil {
		u.logger.Error(ctx, tag, "failed to look up coder for completion check", domainmodels.LoggerMeta{"err": err, "coder_id": testCase.InfraredStateCoderId})
		return err
	}

	if !anyFailed {
		if err := u.coder.Activate(ctx, coder.Id, coder.InfraredDeviceId); err != nil {
			u.logger.Error(ctx, tag, "failed to activate coder", domainmodels.LoggerMeta{"err": err, "coder_id": coder.Id})
			return err
		}
		u.transition(ctx, tag, coder.InfraredRecordSessionId, domainmodels.InfraredRecordingStateCompleted)
		return nil
	}

	failedStates := make(map[uuid.UUID]string)
	for _, tc := range allTestCases {
		if tc.Status != domainmodels.InfraredTestCaseStatusFailed {
			continue
		}
		states, err := u.testCase.ListStatesByTestCaseId(ctx, tc.Id)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to list states for failed test case", domainmodels.LoggerMeta{"err": err, "test_case_id": tc.Id})
			return err
		}
		for _, s := range states {
			failedStates[s.InfraredStateId] = s.StateValue
		}
	}

	go u.runRetryCaseGeneration(coder.InfraredRecordSessionId, coder.Id, failedStates)
	return nil
}
```

`failedStates` deliberately merges every failed test case's states into one map (rather than calling `WriteRetryCases` once per failure) — a single LLM call reasoning about every failure together can spot a shared root cause across them, matching the design spec's own framing of retry-case generation as diagnostic judgment, not a mechanical per-case translation.

Add the retry-loop job:

```go
const retryCaseGenerationTimeout = 2 * time.Minute

func (u *usecase) runRetryCaseGeneration(sessionId uuid.UUID, coderId uuid.UUID, failedStates map[uuid.UUID]string) {
	const tag = "infrared/record_session_management/runRetryCaseGeneration"

	ctx, cancel := context.WithTimeout(context.Background(), retryCaseGenerationTimeout)
	defer cancel()

	coder, err := u.coderById(ctx, coderId)
	if err != nil || coder == nil {
		u.logger.Error(ctx, tag, "failed to look up coder", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	session, err := u.session.GetById(ctx, sessionId)
	if err != nil || session == nil {
		u.logger.Error(ctx, tag, "failed to look up session", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	device, err := u.device.GetById(ctx, session.InfraredDeviceId)
	if err != nil || device == nil {
		u.logger.Error(ctx, tag, "failed to look up device", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	states, err := u.state.ListByDeviceTypeId(ctx, device.InfraredDeviceTypeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list states", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	client, err := u.llmFactory.Current(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve llm client", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	plans, err := applicationinfraredcodergeneration.WriteRetryCases(ctx, client, device.Brand, device.Model, *coder, failedStates, states)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to write retry cases", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	if len(plans) == 0 {
		u.logger.Error(ctx, tag, "llm proposed no retry cases", domainmodels.LoggerMeta{"session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	existingCases, err := u.recordCase.ListBySessionId(ctx, sessionId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list existing cases for step numbering", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	nextStep := int32(1)
	for _, c := range existingCases {
		if c.Step >= nextStep {
			nextStep = c.Step + 1
		}
	}

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

The `Step` continuation logic (`nextStep` derived from the max existing `Step` plus one) is what the design spec calls out explicitly: "`Step` is append-only and monotonic for the lifetime of the session... any later round of cases appended after a test failure takes `N+1..N+k`." This is the only place in the whole feature that needs to compute a starting `Step` from existing data, rather than starting fresh at 1 — every case-creation path before this task (Plan B1's `runCaseGeneration`) only ever ran once per session.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -v`
Expected: PASS on every test in the package.

- [ ] **Step 6: Verify**

Run: `cd backend && go build ./... 2>&1 | grep -v "composition/main" && go vet ./internal/domain/usecases/infrared/... ./internal/application/infrared/record_session_management/... && gofmt -l internal/domain/usecases/infrared/ internal/application/infrared/record_session_management/`
Expected: no output beyond the expected composition/main isolation.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/domain/usecases/infrared/record_session_management.go \
        backend/internal/application/infrared/record_session_management/
git commit -m "feat: add RecordTestCaseResult, coder activation, and the retry loop"
```

---

### Task 11: HTTP presentation layer

**Files:**
- Modify: `backend/internal/presentation/http/request/infrared.go`
- Modify: `backend/internal/presentation/http/response/infrared.go`
- Modify: `backend/internal/presentation/http/handler/infrared/handler.go`
- Modify: `backend/internal/presentation/http/route/route.go`

**Interfaces:**
- Consumes: `domainusecasesinfrared.RecordSessionManagement.ListTestCases`/`TransmitTestCase`/`RecordTestCaseResult` (Tasks 8-10).
- Produces: `GET /v1/infrared/record-sessions/:id/test-cases`, `POST /v1/infrared/test-cases/:id/transmit`, `POST /v1/infrared/test-cases/:id/result` — all gated by the existing `infrared_record_session:get`/`set` permissions (no new permissions).

- [ ] **Step 1: Add request/response DTOs**

Append to `backend/internal/presentation/http/request/infrared.go`:

```go
type RecordTestCaseResultRequest struct {
	Passed bool `json:"passed" example:"true"`
}
```

Append to `backend/internal/presentation/http/response/infrared.go`, following `InfraredStateDeviceRecordCaseResponse`'s exact nested-states shape:

```go
type InfraredTestCaseStateResponse struct {
	Id              string `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	InfraredStateId string `json:"infrared_state_id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	StateValue      string `json:"state_value" example:"ON"`
}

func InfraredTestCaseStates(models []domainmodels.InfraredTestCaseState) []InfraredTestCaseStateResponse {
	responses := make([]InfraredTestCaseStateResponse, len(models))
	for i, model := range models {
		responses[i] = InfraredTestCaseStateResponse{
			Id:              UUIDString(model.Id),
			InfraredStateId: UUIDString(model.InfraredStateId),
			StateValue:      model.StateValue,
		}
	}
	return responses
}

type InfraredTestCaseResponse struct {
	Id          string                          `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	Step        int32                           `json:"step" example:"1"`
	Description string                          `json:"description" example:"Confirm the unit powers on."`
	Status      string                          `json:"status" example:"PENDING"`
	States      []InfraredTestCaseStateResponse `json:"states"`
}

func InfraredTestCases(models []domainusecasesinfrared.TestCaseWithStates) []InfraredTestCaseResponse {
	responses := make([]InfraredTestCaseResponse, len(models))
	for i, m := range models {
		responses[i] = InfraredTestCaseResponse{
			Id:          UUIDString(m.TestCase.Id),
			Step:        m.TestCase.Step,
			Description: m.TestCase.Description,
			Status:      string(m.TestCase.Status),
			States:      InfraredTestCaseStates(m.States),
		}
	}
	return responses
}
```

Add `domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"` to `response/infrared.go`'s imports — check first whether it's already there (Task 11 of Plan B1 already needed a usecase-layer type for its own case-bundle mapper; if so, reuse that import line).

- [ ] **Step 2: Add the handler methods**

Append to `backend/internal/presentation/http/handler/infrared/handler.go`, following `RecordSessionCasesGetList`'s exact shape for the list endpoint and `RecordCaseRetry`'s exact shape for the two action endpoints:

```go
// RecordSessionTestCasesGetList godoc
//
// @Summary Infrared Record Session Test Cases List
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {array} presentationhttpresponse.InfraredTestCaseResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-sessions/{id}/test-cases [get]
func (h *handler) RecordSessionTestCasesGetList(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	testCases, err := h.recordSessionUseCase.ListTestCases(c.Request().Context(), id)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.InfraredTestCases(testCases))
}

// TestCaseTransmitPost godoc
//
// @Summary Infrared Test Case Transmit
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/test-cases/{id}/transmit [post]
func (h *handler) TestCaseTransmitPost(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.TransmitTestCase(c.Request().Context(), id); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// TestCaseResultPost godoc
//
// @Summary Infrared Test Case Result
// @Tags Infrared
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.RecordTestCaseResultRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/test-cases/{id}/result [post]
func (h *handler) TestCaseResultPost(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	var req presentationhttprequest.RecordTestCaseResultRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	if err := h.recordSessionUseCase.RecordTestCaseResult(c.Request().Context(), id, req.Passed); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
```

- [ ] **Step 3: Register the routes**

Add all three method names to the `InfraredHandler` interface in `route.go`, and add to `routeInfrared`:

```go
v1.GET("/infrared/record-sessions/:id/test-cases", handler.RecordSessionTestCasesGetList, permission("infrared_record_session:get"))
v1.POST("/infrared/test-cases/:id/transmit", handler.TestCaseTransmitPost, permission("infrared_record_session:set"))
v1.POST("/infrared/test-cases/:id/result", handler.TestCaseResultPost, permission("infrared_record_session:set"))
```

- [ ] **Step 4: Verify**

Run: `cd backend && go build ./... 2>&1 | grep -v "composition/main" && go vet ./internal/presentation/... && gofmt -l internal/presentation/http/`
Expected: no output beyond the expected composition/main isolation.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/presentation/http/request/infrared.go \
        backend/internal/presentation/http/response/infrared.go \
        backend/internal/presentation/http/handler/infrared/handler.go \
        backend/internal/presentation/http/route/route.go
git commit -m "feat: add HTTP endpoints for test cases, transmit, and results"
```

---

### Task 12: Composition wiring

**Files:**
- Modify: `backend/internal/composition/main/infrastructure.go`
- Modify: `backend/internal/composition/main/application.go`

**Interfaces:**
- Consumes: everything from Tasks 3, 5, 8-10.

Pure wiring — no business logic. This is the task that resolves every isolated `composition/main` build failure Tasks 8-11 leave behind.

- [ ] **Step 1: Wire the new repository**

In `infrastructure.go`, construct `infraredTestCaseRepository := infrastructurerepositoryinfraredtestcase.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)`, following the exact one-field-plus-one-constructor-line pattern used for every existing entity.

- [ ] **Step 2: Wire the usecase**

In `application.go`, update the existing `applicationinfraredrecordsessionmanagement.NewUsecaseImpl(...)` call site (from Plan B2's Task 12) to add the two new arguments (`l.infra.infraredTestCaseRepository`, `l.infra.nodePublish` — check the exact existing field name this composition root already uses for the `domaincontractsnode.Publish` implementation, since it's already constructed for the messaging-callback usecase and should be reused, not duplicated) in the exact position Tasks 8-9's constructor signature places them — after `coder`, before `logger`.

- [ ] **Step 3: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l . && go test -count=1 ./...`
Expected: no build/vet/gofmt output; every test passes, including every test from this plan and from Plans B1/B2 — this is the task that makes the whole repo build clean again, and the final task of the entire IR recording session feature.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/composition/main/
git commit -m "feat: wire test case generation, transmit, and retry loop into composition root"
```

---

## Self-review notes

- **Spec coverage:** every element of the design spec's remaining pipeline (`TEST_CASES_GENERATING` → `TESTING` → `COMPLETED`/retry-to-`RECORDING`, MQTT transmit, coder activation/supersession) has a task. Decoder verification, automated round-trip signal checking, and exhaustive pairwise state-interaction coverage remain explicitly out of scope, matching the design spec's own stated non-goals.
- **Placeholder scan:** no "TBD"/"implement later" language in any task. `TransmitTestCase`'s helper `coderById` and `stateIdToNameLookup` are both fully specified, not stubs.
- **Two design-spec open items resolved, with rationale:** (1) the MQTT transmit topic/message shape — resolved by mirroring Plan B1's `ir_capture` shape exactly in reverse (Task 5), for consistency within this same feature rather than inventing a different convention; (2) how test-case failures translate into new `RecordCase` rows — resolved via `WriteRetryCases` (Task 7), a fourth structured LLM call following the exact pattern every other LLM step in this feature already uses, given the failing state(s) and the coder's own README as context.
- **A load-bearing fix to already-merged code, not new-feature scope creep:** Task 4 fixes `InfraredStateCoder.GetBySessionId`'s missing `ORDER BY`/`LIMIT` — a gap Plan B2's own final review explicitly identified and deferred specifically because B2 could never actually trigger it (at most one coder per session in B2's scope). This plan is what makes multiple coders per session real, so the fix belongs here, first, before anything else in this plan can rely on "the latest coder" being well-defined.
- **Type consistency:** `domainmodels.InfraredTestCase`/`InfraredTestCaseState` (Task 1) are used identically across the repository (Task 3), the two new usecase methods (Tasks 8-10), and the response mappers (Task 11). `applicationinfraredcodergeneration.TestCasePlan`/`RetryCasePlan` (Tasks 6-7) are deliberately kept as separate types with identical shape rather than unified, with the reasoning stated inline in Task 7 rather than left implicit. `RecordSessionManagement`'s final method set (after this plan) is `Start, GetById, ListCases, AcceptRaw, DiscardRaw, RetryCase, SetCurrentCase, CaptureIrRaw, GetCoderBySessionId, ListTestCases, TransmitTestCase, RecordTestCaseResult` — every one of the last three, added in this plan, is consumed by exactly one HTTP handler method in Task 11, one-to-one.
