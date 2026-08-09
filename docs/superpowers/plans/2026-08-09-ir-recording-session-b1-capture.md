# IR Recording Session — B1: Domain, Persistence & Capture Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Take an IR recording session from staff filling out the "start recording" form through every case being physically recorded and accepted — `DRAFT → CASES_GENERATING → RECORDING → ANALYZING` — with live status over a websocket. This is the first of three plans implementing `docs/superpowers/specs/2026-08-09-infrared-recording-session-design.md`; B2 (bit-analysis + LLM-generated encoder/decoder via goja) and B3 (test-case generation, MQTT transmit, TESTING loop) build on what this plan merges.

**Architecture:** Deterministic Go enumeration produces the OFAT case set; one LLM call (through the `llm.Client` abstraction merged in the prior plan) writes human-facing case descriptions and a press-order; an ESP32 node publishes each raw IR capture over MQTT, correlated server-side to the session's current case; a background goroutine (this codebase's first non-request-scoped async job) drives `DRAFT → CASES_GENERATING`; a websocket broadcaster, modeled directly on the existing telemetry broadcaster, pushes every state change to subscribed frontend clients.

## Global Constraints

- Domain model source of truth for this plan is the live, uncommitted `backend/internal/domain/models/infrared.go` as it exists at plan start, adjusted per the "Session model reconciliation" section below — not the earlier committed design spec, which predates several since-made simplifications.
- `InfraredRecordSession.RecordingState` is a plain `string` field (not a typed Go enum) — the valid-value set is enforced in `internal/application/shared/validation.go`, matching this codebase's validation-split convention. Package-level `const` string values are still declared for reuse, just not as a named type.
- `InfraredStateDeviceRecordCase.Status` and `InfraredStateDeviceRecordRaw.Status` ARE typed enums (`InfraredRecordCaseStatus`, `InfraredRecordRawStatus`), following the existing `type X string` + `const` block convention.
- Repository pattern: `BasePostgres` embedding + squirrel query builders in a separate `postgres_query.go`, exactly as established in `internal/infrastructure/repository/llm_config/` (merged in the prior plan) and `internal/infrastructure/repository/node_class_action/`.
- MQTT backend→node publishing follows `internal/infrastructure/node/publish/mqtt.go`'s `Publish` contract shape; node→backend subscriptions follow `internal/infrastructure/node/subscriptions/mqtt.go`'s `Subscriptions` contract shape; topic helpers (`NodePubTopic`/`NodeSubTopic`) live in `internal/infrastructure/node/shared/`.
- Websocket broadcasting follows `internal/infrastructure/broadcaster/telemetry/gorilla.go` exactly: a `Register`/`Send` contract, gorilla-websocket-backed implementation, best-effort (warn-log-only, never returned as an error) sends from the usecase layer.
- InfraredDeviceType and InfraredState (the "Air Conditioner" device type and its 5 states: POWER, MODE, FAN_SPEED, TEMPERATURE, SWING) are seeded once via this codebase's existing JSON seeder mechanism (`database/seeder/`), not admin-managed via CRUD endpoints in this plan — there is exactly one device type in scope, and building a management UI for a set of one is premature.
- Verification bar for every task: `go build ./... && go vet ./... && gofmt -l .` clean, plus `go test -count=1 ./...` for anything with dedicated tests.
- Test doubles are hand-written structs embedding the interface they fake, never a mocking library.

## Session model reconciliation

The live `infrared.go` differs from the earlier committed spec in ways confirmed as intentional simplifications. This plan builds the following final shape (Task 1 applies it):

- `InfraredRecordSession` keeps `RecordingState string` + `IsCompleted bool`, but regains `CurrentRecordCaseId *uuid.UUID` (needed to track which case staff is actively working — an explicit, stated requirement that predates the file's simplification and was never intentionally dropped) and `CreatedAt time.Time` (audit convention used by every other entity in this codebase).
- `InfraredStateDeviceRecordCase` regains `InfraredRecordSessionId uuid.UUID` (replacing `InfraredDeviceId`) — cases belong to the session that recorded them, not the device directly, so a device's multiple recording efforts over time don't collide.
- `InfraredStateDeviceRecordCase` and `InfraredStateDeviceRecordRaw` regain typed `Status` fields (misclick-retry tracking: a case is `PENDING`/`ACTIVE`/`ACCEPTED`; a raw capture is `CAPTURED`/`ACCEPTED`/`DISCARDED`).
- `InfraredState` regains an explicit `Type` field (`InfraredStateType`: `RANGE`/`ENUM`) rather than inferring it from which of `Options`/`Minimum` is populated on `InfraredStateDeviceDefinition`.
- `InfraredDeviceType` and `InfraredDevice` are otherwise unchanged from the live file (no `Type`/`InfraredProtocolType` field — hybrid/stateless protocol support stays out of scope, as the original design's own Non-goals section already stated).

---

## File structure

```
backend/
  database/
    migrations/
      20260810090000_infrared_reference.up.sql       (new — device_type, state tables)
      20260810090000_infrared_reference.down.sql
      20260810090001_infrared_recording.up.sql       (new — device, state_device_definition,
      20260810090001_infrared_recording.down.sql       record_session, record_case, record_state, record_raw)
    seeder/
      infrared_device_type.json                       (new)
      infrared_state.json                              (new)
      seeder.go or equivalent loader                   (modify — register the two new seed files)
  internal/
    domain/
      models/
        infrared.go                                    (rewrite per reconciliation above)
      contracts/
        repository/
          infrared_device_type.go                       (new)
          infrared_state.go                             (new)
          infrared_device.go                             (new)
          infrared_state_device_definition.go             (new)
          infrared_record_session.go                     (new)
          infrared_state_device_record_case.go            (new)
        node/
          subscriptions.go                               (modify — add IrCapture method)
        broadcaster/
          infrared_record_session.go                     (new)
      usecases/
        infrared/
          record_session_management.go                   (new — interface + request/result DTOs)
    application/
      infrared/
        case_generation/
          generator.go                                    (new — pure OFAT enumeration function)
          generator_test.go                                (new)
        record_session_management/
          usecase.go                                       (new)
          usecase_test.go                                  (new)
    infrastructure/
      repository/
        infrared_device_type/postgres.go, postgres_query.go       (new)
        infrared_state/postgres.go, postgres_query.go              (new)
        infrared_device/postgres.go, postgres_query.go             (new)
        infrared_state_device_definition/postgres.go, postgres_query.go (new)
        infrared_record_session/postgres.go, postgres_query.go     (new)
        infrared_state_device_record_case/postgres.go, postgres_query.go (new)
      node/
        subscriptions/mqtt.go                              (modify — add IrCapture)
        shared/                                             (unchanged — reuse NodePubTopic)
      broadcaster/
        infrared_record_session/gorilla.go, gorilla_client.go (new)
    presentation/
      mqtt/
        dto/ir_capture.go                                   (new)
        handler/ir_capture.go                               (new)
        event/message.go                                    (modify — route new topic)
      http/
        request/infrared.go                                 (new)
        response/infrared.go                                 (new)
        handler/infrared/handler.go                          (new)
        route/route.go                                       (modify — add infrared routes)
    composition/
      main/
        infrastructure.go, application.go, presentation.go, driver.go (modify)
```

---

### Task 1: Rewrite the domain model

**Files:**
- Modify: `backend/internal/domain/models/infrared.go`

**Interfaces:**
- Produces: the finalized `InfraredDeviceType`, `InfraredDevice`, `InfraredState`/`InfraredStateType`, `InfraredStateDeviceDefinition`, `InfraredRecordSession`, `InfraredStateDeviceRecordCase`/`InfraredRecordCaseStatus`, `InfraredStateDeviceRecordState`, `InfraredStateDeviceRecordRaw`/`InfraredRecordRawStatus` types every later task in this plan depends on.

- [ ] **Step 1: Write the finalized model**

Replace the full content of `backend/internal/domain/models/infrared.go`:

```go
package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type InfraredDeviceType struct {
	Id   uuid.UUID
	Name string
}

type InfraredDevice struct {
	Id                   uuid.UUID
	InfraredDeviceTypeId uuid.UUID
	Brand                string
	Model                string
}

const (
	InfraredRecordingStateDraft               = "DRAFT"
	InfraredRecordingStateCasesGenerating     = "CASES_GENERATING"
	InfraredRecordingStateRecording           = "RECORDING"
	InfraredRecordingStateAnalyzing           = "ANALYZING"
	InfraredRecordingStateFunctionGenerating  = "FUNCTION_GENERATING"
	InfraredRecordingStateTestCasesGenerating = "TEST_CASES_GENERATING"
	InfraredRecordingStateTesting             = "TESTING"
	InfraredRecordingStateCompleted           = "COMPLETED"
	InfraredRecordingStateFailed              = "FAILED"
)

type InfraredRecordSession struct {
	Id                  uuid.UUID
	NodeId              uuid.UUID
	InfraredDeviceId    uuid.UUID
	RecordingState      string
	CurrentRecordCaseId *uuid.UUID
	IsCompleted         bool
	CreatedAt           time.Time
}

type InfraredStateType string

const (
	InfraredStateTypeRange InfraredStateType = "RANGE"
	InfraredStateTypeEnum  InfraredStateType = "ENUM"
)

type InfraredState struct {
	Id                   uuid.UUID
	InfraredDeviceTypeId uuid.UUID
	Name                 string
	Type                 InfraredStateType
}

type InfraredStateDeviceDefinition struct {
	Id               uuid.UUID
	InfraredDeviceId uuid.UUID
	InfraredStateId  uuid.UUID
	Options          []string
	Minimum          *float64
	Maximum          *float64
	Step             *float64
}

type InfraredRecordCaseStatus string

const (
	InfraredRecordCaseStatusPending  InfraredRecordCaseStatus = "PENDING"
	InfraredRecordCaseStatusActive   InfraredRecordCaseStatus = "ACTIVE"
	InfraredRecordCaseStatusAccepted InfraredRecordCaseStatus = "ACCEPTED"
)

type InfraredStateDeviceRecordCase struct {
	Id                      uuid.UUID
	InfraredRecordSessionId uuid.UUID
	Step                    int32
	Description             string
	Status                  InfraredRecordCaseStatus
}

type InfraredStateDeviceRecordState struct {
	Id                              uuid.UUID
	InfraredStateDeviceRecordCaseId uuid.UUID
	InfraredStateId                 uuid.UUID
	StateValue                      string
}

type InfraredRecordRawStatus string

const (
	InfraredRecordRawStatusCaptured  InfraredRecordRawStatus = "CAPTURED"
	InfraredRecordRawStatusAccepted  InfraredRecordRawStatus = "ACCEPTED"
	InfraredRecordRawStatusDiscarded InfraredRecordRawStatus = "DISCARDED"
)

type InfraredStateDeviceRecordRaw struct {
	Id                              uuid.UUID
	InfraredStateDeviceRecordCaseId uuid.UUID
	RawData                         []byte
	Status                          InfraredRecordRawStatus
	DiscardedReason                 *string
}
```

- [ ] **Step 2: Verify**

Run: `cd backend && go build ./... 2>&1 | head -50`
Expected: build errors ONLY in files that referenced the old `infrared.go` shape (there should be none yet, since nothing else in the codebase references these types — this file is currently unused outside itself). If any build errors appear, read them; every later task in this plan is what makes this model load-bearing.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/domain/models/infrared.go
git commit -m "feat: finalize IR domain model (session status field, case/raw retry status, state type)"
```

---

### Task 2: Reference-data migration and seed data

**Files:**
- Create: `backend/database/migrations/20260810090000_infrared_reference.up.sql`
- Create: `backend/database/migrations/20260810090000_infrared_reference.down.sql`
- Create: `backend/database/seeder/infrared_device_type.json`
- Create: `backend/database/seeder/infrared_state.json`
- Modify: whatever Go file registers/loads existing seed JSON files (find it — see Step 3)

**Interfaces:**
- Produces: `infrared_device_type` and `infrared_state` tables, seeded with one "Air Conditioner" device type and its five states (POWER, MODE, FAN_SPEED, TEMPERATURE, SWING).

- [ ] **Step 1: Write the migration**

`backend/database/migrations/20260810090000_infrared_reference.up.sql`:

```sql
CREATE TABLE infrared_device_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE infrared_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_device_type_id UUID NOT NULL REFERENCES infrared_device_type (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    CONSTRAINT chk_infrared_state_type CHECK (type IN ('RANGE', 'ENUM')),
    CONSTRAINT uq_infrared_state_device_type_id_name UNIQUE (infrared_device_type_id, name)
);

CREATE INDEX idx_infrared_state_infrared_device_type_id ON infrared_state (infrared_device_type_id);
```

`backend/database/migrations/20260810090000_infrared_reference.down.sql`:

```sql
DROP TABLE IF EXISTS infrared_state;
DROP TABLE IF EXISTS infrared_device_type;
```

- [ ] **Step 2: Apply the migration locally**

Run: `migrate -path backend/database/migrations -database "$DATABASE_URL" up`
Expected: no error; `\d infrared_device_type` and `\d infrared_state` in `psql` show the tables above.

- [ ] **Step 3: Find and follow the existing seeder convention**

Read `backend/database/seeder/permission.json` and `role.json` (used in the prior plan's Task 10) to confirm the exact JSON shape this codebase's seeder expects, and find the Go file that lists which JSON files get loaded (grep for `permission.json` or `role.json` as a literal string, or for a `//go:embed` directive, in `backend/database/seeder/` and `backend/internal/composition/seeder/`). Match that file's structure precisely — don't invent a new seed-file shape.

- [ ] **Step 4: Write the seed data**

`backend/database/seeder/infrared_device_type.json` — one entry, `{"name": "Air Conditioner"}`, in whatever exact key/array shape Step 3 revealed `permission.json` uses.

`backend/database/seeder/infrared_state.json` — five entries, each `{"device_type_name": "Air Conditioner", "name": "POWER", "type": "ENUM"}` (repeat for `MODE`, `FAN_SPEED`, `TEMPERATURE` as `RANGE`, `SWING` as `ENUM`) — adjust the exact field names to match whatever cross-referencing convention Step 3's reference files use for foreign-key-by-name seeding (e.g. how `role_permission.json`, if it exists, references a permission by name rather than by a not-yet-known UUID).

- [ ] **Step 5: Register the new seed files**

Add both new JSON files to whatever loader Step 3 found, in the same style as the existing entries. If the seeder's load order matters (e.g. must load `infrared_device_type.json` before `infrared_state.json` since the latter references the former by name), place them in that order.

- [ ] **Step 6: Verify**

Run the seeder against the local database (find the exact command from `AGENTS.md` or the seeder's own `main.go` — likely `go run ./cmd/seeder` or similar). Confirm via `psql`: `SELECT name FROM infrared_device_type;` returns "Air Conditioner"; `SELECT name, type FROM infrared_state ORDER BY name;` returns the five states with correct types.

- [ ] **Step 7: Commit**

```bash
git add backend/database/migrations/20260810090000_infrared_reference.* \
        backend/database/seeder/infrared_device_type.json \
        backend/database/seeder/infrared_state.json
# plus whatever loader file Step 5 modified
git commit -m "feat: add infrared device type/state reference tables and seed Air Conditioner"
```

---

### Task 3: Recording-pipeline migration

**Files:**
- Create: `backend/database/migrations/20260810090001_infrared_recording.up.sql`
- Create: `backend/database/migrations/20260810090001_infrared_recording.down.sql`

**Interfaces:**
- Produces: `infrared_device`, `infrared_state_device_definition`, `infrared_record_session`, `infrared_state_device_record_case`, `infrared_state_device_record_state`, `infrared_state_device_record_raw` tables.

- [ ] **Step 1: Write the migration**

`backend/database/migrations/20260810090001_infrared_recording.up.sql`:

```sql
CREATE TABLE infrared_device (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_device_type_id UUID NOT NULL REFERENCES infrared_device_type (id) ON DELETE CASCADE,
    brand TEXT NOT NULL,
    model TEXT NOT NULL
);

CREATE INDEX idx_infrared_device_infrared_device_type_id ON infrared_device (infrared_device_type_id);

CREATE TABLE infrared_state_device_definition (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_device_id UUID NOT NULL REFERENCES infrared_device (id) ON DELETE CASCADE,
    infrared_state_id UUID NOT NULL REFERENCES infrared_state (id) ON DELETE CASCADE,
    options TEXT[],
    minimum DOUBLE PRECISION,
    maximum DOUBLE PRECISION,
    step DOUBLE PRECISION,
    CONSTRAINT uq_infrared_state_device_definition_device_state UNIQUE (infrared_device_id, infrared_state_id)
);

CREATE TABLE infrared_record_session (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    node_id UUID NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
    infrared_device_id UUID NOT NULL REFERENCES infrared_device (id) ON DELETE CASCADE,
    recording_state TEXT NOT NULL DEFAULT 'DRAFT',
    current_record_case_id UUID,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_infrared_record_session_node_id ON infrared_record_session (node_id);
CREATE INDEX idx_infrared_record_session_infrared_device_id ON infrared_record_session (infrared_device_id);

CREATE TABLE infrared_state_device_record_case (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_record_session_id UUID NOT NULL REFERENCES infrared_record_session (id) ON DELETE CASCADE,
    step INTEGER NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'PENDING',
    CONSTRAINT chk_infrared_record_case_status CHECK (status IN ('PENDING', 'ACTIVE', 'ACCEPTED')),
    CONSTRAINT uq_infrared_record_case_session_step UNIQUE (infrared_record_session_id, step)
);

CREATE INDEX idx_infrared_record_case_session_id ON infrared_state_device_record_case (infrared_record_session_id);

-- current_record_case_id is added as a foreign key here, once the table it
-- references exists, rather than on infrared_record_session's own CREATE.
ALTER TABLE infrared_record_session
    ADD CONSTRAINT fk_infrared_record_session_current_case
    FOREIGN KEY (current_record_case_id) REFERENCES infrared_state_device_record_case (id) ON DELETE SET NULL;

CREATE TABLE infrared_state_device_record_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_state_device_record_case_id UUID NOT NULL REFERENCES infrared_state_device_record_case (id) ON DELETE CASCADE,
    infrared_state_id UUID NOT NULL REFERENCES infrared_state (id) ON DELETE CASCADE,
    state_value TEXT NOT NULL,
    CONSTRAINT uq_infrared_record_state_case_state UNIQUE (infrared_state_device_record_case_id, infrared_state_id)
);

CREATE INDEX idx_infrared_record_state_case_id ON infrared_state_device_record_state (infrared_state_device_record_case_id);

CREATE TABLE infrared_state_device_record_raw (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    infrared_state_device_record_case_id UUID NOT NULL REFERENCES infrared_state_device_record_case (id) ON DELETE CASCADE,
    raw_data BYTEA NOT NULL,
    status TEXT NOT NULL DEFAULT 'CAPTURED',
    discarded_reason TEXT,
    CONSTRAINT chk_infrared_record_raw_status CHECK (status IN ('CAPTURED', 'ACCEPTED', 'DISCARDED'))
);

CREATE INDEX idx_infrared_record_raw_case_id ON infrared_state_device_record_raw (infrared_state_device_record_case_id);
```

**Before writing this file:** confirm the real name of the existing nodes table (`nodes`, `node`, or something else) by checking `internal/infrastructure/repository/node/postgres_query.go`'s table name in its query builders, and adjust the `REFERENCES` clause above to match exactly.

`backend/database/migrations/20260810090001_infrared_recording.down.sql`:

```sql
DROP TABLE IF EXISTS infrared_state_device_record_raw;
DROP TABLE IF EXISTS infrared_state_device_record_state;
ALTER TABLE IF EXISTS infrared_record_session DROP CONSTRAINT IF EXISTS fk_infrared_record_session_current_case;
DROP TABLE IF EXISTS infrared_state_device_record_case;
DROP TABLE IF EXISTS infrared_record_session;
DROP TABLE IF EXISTS infrared_state_device_definition;
DROP TABLE IF EXISTS infrared_device;
```

- [ ] **Step 2: Apply and verify**

Run: `migrate -path backend/database/migrations -database "$DATABASE_URL" up`
Expected: no error; `\d infrared_record_session` in `psql` shows all six columns including the `current_record_case_id` foreign key added by the `ALTER TABLE`.

- [ ] **Step 3: Verify build**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l database/`
Expected: no output (this task has no Go code, but confirms nothing else broke).

- [ ] **Step 4: Commit**

```bash
git add backend/database/migrations/20260810090001_infrared_recording.*
git commit -m "feat: add infrared device/definition/session/case/state/raw tables"
```

---

### Task 4: Reference-data repositories (InfraredDeviceType, InfraredState) and seed wiring

**Files:**
- Create: `backend/internal/domain/contracts/repository/infrared_device_type.go`
- Create: `backend/internal/domain/contracts/repository/infrared_state.go`
- Create: `backend/internal/infrastructure/repository/infrared_device_type/postgres.go`
- Create: `backend/internal/infrastructure/repository/infrared_device_type/postgres_query.go`
- Create: `backend/internal/infrastructure/repository/infrared_state/postgres.go`
- Create: `backend/internal/infrastructure/repository/infrared_state/postgres_query.go`
- Modify: `backend/internal/application/seeder/usecase.go`
- Modify: `backend/internal/composition/seeder/infrastructure.go`, `backend/internal/composition/seeder/application.go`

**Interfaces:**
- Consumes: `domainmodels.InfraredDeviceType`, `domainmodels.InfraredState` (Task 1); `infrastructurerepositoryshared.BasePostgres`/`MapPgxError`/`QueryBuildError` (existing); `seederdata.Data.InfraredDeviceTypes`/`InfraredStates` (Task 2, already loadable but not yet persisted — this task closes that gap).
- Produces: `domaincontractsrepository.InfraredDeviceType` with `List(ctx) ([]domainmodels.InfraredDeviceType, error)`, `ReadByName(ctx, name string) (domainmodels.InfraredDeviceType, error)`, `Create(ctx, name string) (uuid.UUID, error)`; `domaincontractsrepository.InfraredState` with `ListByDeviceTypeId(ctx, deviceTypeId uuid.UUID) ([]domainmodels.InfraredState, error)`, `ReadByDeviceTypeIdAndName(ctx, deviceTypeId uuid.UUID, name string) (domainmodels.InfraredState, error)`, `Create(ctx, deviceTypeId uuid.UUID, name string, stateType domainmodels.InfraredStateType) (uuid.UUID, error)`.

Reference data is read-mostly at runtime (`List`/`ListByDeviceTypeId` are what later tasks consume), but Task 2 only made the seed JSON parseable into `seederdata.Data` — nothing persisted it. This task closes that gap: it adds the minimal `ReadByName`/`Create` pair each repository needs for idempotent seeding (mirroring `permission`'s `ReadByName`+`Create` pattern in `internal/infrastructure/repository/permission/postgres.go`), plus the seed-usecase functions and composition wiring that actually call them. No `Update`/`Delete` — this data is never edited after seeding. No dedicated repository test, per this codebase's convention (matches `internal/infrastructure/repository/llm_config/` — no test file there either); the seed usecase functions ARE covered by the existing `internal/application/seeder` test conventions if one exists for sibling `seedX` functions — check for `usecase_test.go` in that package and follow its pattern if present, otherwise no dedicated test is required (matching `seedPermissions`/`seedRoles`, which have none either).

- [ ] **Step 1: Write the domain contracts**

`backend/internal/domain/contracts/repository/infrared_device_type.go`:

```go
package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredDeviceType interface {
	List(ctx context.Context) ([]domainmodels.InfraredDeviceType, error)
	ReadByName(ctx context.Context, name string) (*domainmodels.InfraredDeviceType, error)
	Create(ctx context.Context, name string) (id uuid.UUID, err error)
}
```

`backend/internal/domain/contracts/repository/infrared_state.go`:

```go
package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredState interface {
	ListByDeviceTypeId(ctx context.Context, infraredDeviceTypeId uuid.UUID) ([]domainmodels.InfraredState, error)
	ReadByDeviceTypeIdAndName(ctx context.Context, infraredDeviceTypeId uuid.UUID, name string) (*domainmodels.InfraredState, error)
	Create(ctx context.Context, infraredDeviceTypeId uuid.UUID, name string, stateType domainmodels.InfraredStateType) (id uuid.UUID, err error)
}
```

`ReadByName`/`ReadByDeviceTypeIdAndName` return `(nil, domainmodels.ErrTypeNotFound-wrapped-error)` when absent — follow `permission.ReadByName`'s exact not-found convention in `backend/internal/infrastructure/repository/permission/postgres.go` (wraps `pgx.ErrNoRows` via `infrastructurerepositoryshared.NotFound(...)`), since the seed usecase (Step 3 below) branches on `errors.Is(err, domainmodels.ErrTypeNotFound)` exactly like `seedPermissions` does.

- [ ] **Step 2: Write both Postgres implementations**

Follow the exact `BasePostgres` + squirrel pattern in `backend/internal/infrastructure/repository/llm_config/postgres.go` and the `ReadByName`/`Create` shapes in `backend/internal/infrastructure/repository/permission/postgres.go` (read both first — constructor signature, `p.SqrD.Select(...)`, `p.Dt.Query`/`QueryRow`, `MapPgxError`/`QueryBuildError`/`NotFound` usage). `infrared_device_type/postgres.go`:

```go
package infrastructurerepositoryinfrareddevicetype

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
) domaincontractsrepository.InfraredDeviceType {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) List(ctx context.Context) ([]domainmodels.InfraredDeviceType, error) {
	query, args, err := p.SqrD.Select("id", "name").From("infrared_device_type").OrderBy("name").ToSql()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_device_type list query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to list infrared_device_type", err)
	}
	defer rows.Close()

	var result []domainmodels.InfraredDeviceType
	for rows.Next() {
		var item domainmodels.InfraredDeviceType
		if err := rows.Scan(&item.Id, &item.Name); err != nil {
			return nil, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_device_type", err)
		}
		result = append(result, item)
	}
	return result, nil
}

func (p *postgresImpl) ReadByName(ctx context.Context, name string) (*domainmodels.InfraredDeviceType, error) {
	query, args, err := p.SqrD.Select("id", "name").From("infrared_device_type").Where(squirrel.Eq{"name": name}).ToSql()
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build infrared_device_type read query", err)
	}

	var item domainmodels.InfraredDeviceType
	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&item.Id, &item.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, infrastructurerepositoryshared.NotFound("infrared_device_type not found", err)
		}
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read infrared_device_type", err)
	}
	return &item, nil
}

func (p *postgresImpl) Create(ctx context.Context, name string) (id uuid.UUID, err error) {
	query, args, err := p.SqrD.Insert("infrared_device_type").Columns("name").Values(name).Suffix("RETURNING id").ToSql()
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create infrared_device_type query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create infrared_device_type", err)
	}
	return id, nil
}
```

`infrared_device_type/postgres_query.go` can stay empty of standalone query-builder functions for this task (the queries are inline above) — if the reviewer of this task's PR prefers matching the two-file split exactly, move the `Select(...).ToSql()`/`Insert(...).ToSql()` calls into `postgres_query.go` methods instead, mirroring `llm_config`'s split precisely.

Write `infrared_state/postgres.go` the same way: `ListByDeviceTypeId` filters `WHERE infrared_device_type_id = ?` via `p.SqrD.Select(...).Where(squirrel.Eq{"infrared_device_type_id": infraredDeviceTypeId})`, scanning the extra `type` column into a `string` and casting to `domainmodels.InfraredStateType`; `ReadByDeviceTypeIdAndName` filters on both `infrared_device_type_id` and `name` with the same not-found convention as above; `Create` inserts `infrared_device_type_id`, `name`, `type` and returns the generated `id`, taking `stateType domainmodels.InfraredStateType` and casting it to `string` for the `type` column value.

- [ ] **Step 3: Add seed-usecase functions**

In `backend/internal/application/seeder/usecase.go`: add `infraredDeviceType domaincontractsrepository.InfraredDeviceType` and `infraredState domaincontractsrepository.InfraredState` fields to the `usecase` struct and `NewUsecaseImpl`'s parameter list (append after the existing `user` parameter, before `password`), threading them through the struct literal exactly like the other repository fields.

Add two new methods, following `seedPermissions`'s exact idempotent-lookup-then-create shape:

```go
func (u *usecase) seedInfraredDeviceTypes(ctx context.Context) (map[string]uuid.UUID, error) {
	const tag = "seeder/seedInfraredDeviceTypes"

	ids := make(map[string]uuid.UUID, len(u.data.InfraredDeviceTypes))
	for _, deviceType := range u.data.InfraredDeviceTypes {
		existing, err := u.infraredDeviceType.ReadByName(ctx, deviceType.Name)
		if err == nil {
			u.logger.Debug(ctx, tag, "infrared device type already exists, skipping", domainmodels.LoggerMeta{"name": deviceType.Name})
			ids[deviceType.Name] = existing.Id
			continue
		}
		if !errors.Is(err, domainmodels.ErrTypeNotFound) {
			u.logger.Error(ctx, tag, "failed to read infrared device type", domainmodels.LoggerMeta{
				"err":  err,
				"name": deviceType.Name,
			})
			return nil, err
		}

		id, err := u.infraredDeviceType.Create(ctx, deviceType.Name)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create infrared device type", domainmodels.LoggerMeta{
				"err":  err,
				"name": deviceType.Name,
			})
			return nil, err
		}

		u.logger.Info(ctx, tag, "infrared device type created", domainmodels.LoggerMeta{"name": deviceType.Name})
		ids[deviceType.Name] = id
	}

	return ids, nil
}

func (u *usecase) seedInfraredStates(ctx context.Context, deviceTypeIds map[string]uuid.UUID) error {
	const tag = "seeder/seedInfraredStates"

	for _, state := range u.data.InfraredStates {
		deviceTypeId, ok := deviceTypeIds[state.DeviceTypeName]
		if !ok {
			err := domainmodels.NewError("infrared device type not found for infrared_state seeding", domainmodels.ErrTypeNotFound, nil)
			u.logger.Error(ctx, tag, "unknown infrared device type", domainmodels.LoggerMeta{
				"err":         err,
				"device_type": state.DeviceTypeName,
			})
			return err
		}

		existing, err := u.infraredState.ReadByDeviceTypeIdAndName(ctx, deviceTypeId, state.Name)
		if err == nil {
			u.logger.Debug(ctx, tag, "infrared state already exists, skipping", domainmodels.LoggerMeta{"name": state.Name})
			_ = existing
			continue
		}
		if !errors.Is(err, domainmodels.ErrTypeNotFound) {
			u.logger.Error(ctx, tag, "failed to read infrared state", domainmodels.LoggerMeta{
				"err":  err,
				"name": state.Name,
			})
			return err
		}

		if _, err := u.infraredState.Create(ctx, deviceTypeId, state.Name, domainmodels.InfraredStateType(state.Type)); err != nil {
			u.logger.Error(ctx, tag, "failed to create infrared state", domainmodels.LoggerMeta{
				"err":  err,
				"name": state.Name,
			})
			return err
		}

		u.logger.Info(ctx, tag, "infrared state created", domainmodels.LoggerMeta{"name": state.Name})
	}

	return nil
}
```

Wire both into `Run(ctx)`, right after `seedUsers` (order doesn't matter relative to the existing chain — infrared reference data has no dependency on permissions/roles/users):

```go
	deviceTypeIds, err := u.seedInfraredDeviceTypes(ctx)
	if err != nil {
		return err
	}

	if err := u.seedInfraredStates(ctx, deviceTypeIds); err != nil {
		return err
	}
```

- [ ] **Step 4: Wire composition**

In `backend/internal/composition/seeder/infrastructure.go`: construct `infraredDeviceTypeRepository := infrastructurerepositoryinfrareddevicetype.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)` and the equivalent for `infrared_state`, add both as fields on the infra struct, following the exact placement/pattern of `permissionRepository`.

In `backend/internal/composition/seeder/application.go`: pass `l.infra.infraredDeviceTypeRepository` and `l.infra.infraredStateRepository` into `applicationseeder.NewUsecaseImpl(...)` at the position matching Step 3's new parameter order (after `l.infra.userRepository`, before `l.infra.password`).

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/repository/ internal/infrastructure/repository/infrared_device_type/ internal/infrastructure/repository/infrared_state/ internal/application/seeder/ internal/composition/seeder/`
Expected: no output.

If a database is reachable in this environment, additionally run the seeder (per Task 2's Step 6 command) and confirm via `psql`: `SELECT name FROM infrared_device_type;` returns "Air Conditioner"; `SELECT name, type FROM infrared_state ORDER BY name;` returns the five seeded states. If no database is reachable, note that this end-to-end check was skipped and why — the build/vet/gofmt check is still mandatory and must be clean.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/contracts/repository/infrared_device_type.go \
        backend/internal/domain/contracts/repository/infrared_state.go \
        backend/internal/infrastructure/repository/infrared_device_type/ \
        backend/internal/infrastructure/repository/infrared_state/ \
        backend/internal/application/seeder/usecase.go \
        backend/internal/composition/seeder/infrastructure.go \
        backend/internal/composition/seeder/application.go
git commit -m "feat: add InfraredDeviceType/InfraredState repositories and wire seed persistence"
```

---

### Task 5: Device and state-definition repositories

**Files:**
- Create: `backend/internal/domain/contracts/repository/infrared_device.go`
- Create: `backend/internal/domain/contracts/repository/infrared_state_device_definition.go`
- Create: `backend/internal/infrastructure/repository/infrared_device/postgres.go`, `postgres_query.go`
- Create: `backend/internal/infrastructure/repository/infrared_state_device_definition/postgres.go`, `postgres_query.go`

**Interfaces:**
- Consumes: `domainmodels.InfraredDevice`, `domainmodels.InfraredStateDeviceDefinition` (Task 1).
- Produces: `domaincontractsrepository.InfraredDevice` with `Create(ctx, infraredDeviceTypeId uuid.UUID, brand string, model string) (id uuid.UUID, err error)` and `GetById(ctx, id uuid.UUID) (*domainmodels.InfraredDevice, error)`; `domaincontractsrepository.InfraredStateDeviceDefinition` with `CreateMany(ctx, infraredDeviceId uuid.UUID, definitions []domainmodels.InfraredStateDeviceDefinition) error` and `ListByDeviceId(ctx, infraredDeviceId uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error)`.

No dedicated test, same convention as Task 4.

- [ ] **Step 1: Write the domain contracts**

```go
// backend/internal/domain/contracts/repository/infrared_device.go
package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredDevice interface {
	Create(ctx context.Context, infraredDeviceTypeId uuid.UUID, brand string, model string) (id uuid.UUID, err error)
	GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredDevice, error)
}
```

```go
// backend/internal/domain/contracts/repository/infrared_state_device_definition.go
package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredStateDeviceDefinition interface {
	// CreateMany inserts one row per definition. All definitions must share
	// the given infraredDeviceId; this is validated by the usecase, not here.
	CreateMany(ctx context.Context, definitions []domainmodels.InfraredStateDeviceDefinition) error
	ListByDeviceId(ctx context.Context, infraredDeviceId uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error)
}
```

- [ ] **Step 2: Write both Postgres implementations**

Same `BasePostgres` pattern as Task 4. `InfraredDevice.Create` is a single-row `Insert(...).Suffix("RETURNING id")`, `.Scan(&id)`. `InfraredStateDeviceDefinition.CreateMany` loops the input slice, issuing one `Insert` per row inside a single call (no explicit transaction wrapper needed — if the codebase's `pgxdt.Pgxdt` doesn't expose a convenient multi-statement transaction helper, sequential inserts are acceptable here since a partial failure leaves an incomplete-but-not-corrupt state that the usecase can detect via row count, and this table has no uniqueness collision risk within one call). `ListByDeviceId` selects all columns including `options` (a Postgres `TEXT[]`, scan into `[]string` directly — `pgx` supports this natively) and the three nullable `DOUBLE PRECISION` columns into `*float64`.

- [ ] **Step 3: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/repository/ internal/infrastructure/repository/infrared_device/ internal/infrastructure/repository/infrared_state_device_definition/`
Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/domain/contracts/repository/infrared_device.go \
        backend/internal/domain/contracts/repository/infrared_state_device_definition.go \
        backend/internal/infrastructure/repository/infrared_device/ \
        backend/internal/infrastructure/repository/infrared_state_device_definition/
git commit -m "feat: add InfraredDevice/InfraredStateDeviceDefinition repositories"
```

---

### Task 6: Session repository

**Files:**
- Create: `backend/internal/domain/contracts/repository/infrared_record_session.go`
- Create: `backend/internal/infrastructure/repository/infrared_record_session/postgres.go`, `postgres_query.go`

**Interfaces:**
- Consumes: `domainmodels.InfraredRecordSession` (Task 1).
- Produces: `domaincontractsrepository.InfraredRecordSession` with `Create`, `GetById`, `GetActiveByNodeId`, `UpdateRecordingStateById` (the generic single-method status-transition pattern from `ActionLog.UpdateStatusByExecutionId` — see `internal/infrastructure/repository/action_log/postgres.go` for the exact reference), and `UpdateCurrentRecordCaseIdById`. `GetActiveByNodeId` is what Task 13's `CaptureIrRaw` will use to correlate an incoming MQTT capture (identified only by the node's device id) back to the session currently recording against it.

- [ ] **Step 1: Write the domain contract**

```go
package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredRecordSession interface {
	Create(ctx context.Context, nodeId uuid.UUID, infraredDeviceId uuid.UUID) (id uuid.UUID, err error)
	GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredRecordSession, error)
	// GetActiveByNodeId returns the session currently RECORDING against the
	// given node, or (nil, nil) if none is active — there is deliberately
	// no error for the common "nothing recording right now" case.
	GetActiveByNodeId(ctx context.Context, nodeId uuid.UUID) (*domainmodels.InfraredRecordSession, error)
	UpdateRecordingStateById(ctx context.Context, id uuid.UUID, recordingState string, isCompleted bool) error
	UpdateCurrentRecordCaseIdById(ctx context.Context, id uuid.UUID, currentRecordCaseId *uuid.UUID) error
}
```

- [ ] **Step 2: Write the Postgres implementation**

Read `internal/infrastructure/repository/action_log/postgres.go`'s `UpdateStatusByExecutionId` and its `postgres_query.go`'s `queryUpdateStatusByExecutionId` first — `UpdateRecordingStateById` and `UpdateCurrentRecordCaseIdById` are direct analogs (a squirrel `Update(...).Set(...).Where(squirrel.Eq{"id": id})`, checking rows-affected via the command tag and returning `infrastructurerepositoryshared.NotFound(...)` if zero rows matched — confirm this check exists in the `action_log` reference and copy it; if it doesn't, a zero-rows update silently succeeding is acceptable here too, matching that reference exactly either way). `Create` inserts with `recording_state` defaulting to `'DRAFT'` at the SQL level (per the migration's `DEFAULT`) — the Go insert only supplies `id` (generated), `node_id`, `infrared_device_id`, letting every other column take its default. `GetActiveByNodeId` is a `Select(...).Where(squirrel.Eq{"node_id": nodeId, "recording_state": domainmodels.InfraredRecordingStateRecording}).Limit(1)` — on `pgx.ErrNoRows`, return `(nil, nil)`, not an error (this is the normal "nothing recording" case, not a failure).

- [ ] **Step 3: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/repository/ internal/infrastructure/repository/infrared_record_session/`
Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/domain/contracts/repository/infrared_record_session.go \
        backend/internal/infrastructure/repository/infrared_record_session/
git commit -m "feat: add InfraredRecordSession repository"
```

---

### Task 7: Case, state, and raw-capture repositories

**Files:**
- Create: `backend/internal/domain/contracts/repository/infrared_state_device_record_case.go`
- Create: `backend/internal/infrastructure/repository/infrared_state_device_record_case/postgres.go`, `postgres_query.go`

**Interfaces:**
- Consumes: `domainmodels.InfraredStateDeviceRecordCase`, `InfraredStateDeviceRecordState`, `InfraredStateDeviceRecordRaw` (Task 1).
- Produces: `domaincontractsrepository.InfraredStateDeviceRecordCase` with:
  - `CreateWithStates(ctx, sessionId uuid.UUID, step int32, description string, states []domainmodels.InfraredStateDeviceRecordState) (caseId uuid.UUID, err error)`
  - `ListBySessionId(ctx, sessionId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordCase, error)`
  - `GetById(ctx, id uuid.UUID) (*domainmodels.InfraredStateDeviceRecordCase, error)`
  - `UpdateStatusById(ctx, id uuid.UUID, status domainmodels.InfraredRecordCaseStatus) error`
  - `ListStatesByCaseId(ctx, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordState, error)`
  - `CreateRaw(ctx, caseId uuid.UUID, rawData []byte) (rawId uuid.UUID, err error)`
  - `ListRawByCaseId(ctx, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordRaw, error)`
  - `UpdateRawStatusById(ctx, id uuid.UUID, status domainmodels.InfraredRecordRawStatus, discardedReason *string) error`
  - `CountAcceptedRawByCaseId(ctx, caseId uuid.UUID) (int, error)`

One contract/implementation pair covering all three record-pipeline tables — they're always read/written together in service of one case's lifecycle, and splitting them into three separate repositories would mean every call site touching a case also has to inject and coordinate two more repository interfaces for no isolation benefit.

- [ ] **Step 1: Write the domain contract**

```go
package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredStateDeviceRecordCase interface {
	CreateWithStates(ctx context.Context, sessionId uuid.UUID, step int32, description string, states []domainmodels.InfraredStateDeviceRecordState) (caseId uuid.UUID, err error)
	ListBySessionId(ctx context.Context, sessionId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordCase, error)
	GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredStateDeviceRecordCase, error)
	UpdateStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredRecordCaseStatus) error
	ListStatesByCaseId(ctx context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordState, error)
	CreateRaw(ctx context.Context, caseId uuid.UUID, rawData []byte) (rawId uuid.UUID, err error)
	ListRawByCaseId(ctx context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordRaw, error)
	UpdateRawStatusById(ctx context.Context, id uuid.UUID, status domainmodels.InfraredRecordRawStatus, discardedReason *string) error
	CountAcceptedRawByCaseId(ctx context.Context, caseId uuid.UUID) (int, error)
}
```

- [ ] **Step 2: Write the Postgres implementation**

Same `BasePostgres`/squirrel pattern as every prior repository task. `CreateWithStates` inserts the case row (`RETURNING id`), then loops `states` inserting one `infrared_state_device_record_state` row per entry with the returned case id — sequential inserts, same reasoning as Task 5's `CreateMany` (no transaction wrapper needed for this codebase's `pgxdt.Pgxdt`, per the pattern already accepted there). `CountAcceptedRawByCaseId` is a `SELECT COUNT(*) FROM infrared_state_device_record_raw WHERE infrared_state_device_record_case_id = ? AND status = 'ACCEPTED'`. `UpdateRawStatusById` sets both `status` and `discarded_reason` in one `Update(...).Set(...)` call (the latter to `NULL` when transitioning to `ACCEPTED`, to the given reason when transitioning to `DISCARDED`).

- [ ] **Step 3: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/repository/ internal/infrastructure/repository/infrared_state_device_record_case/`
Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/domain/contracts/repository/infrared_state_device_record_case.go \
        backend/internal/infrastructure/repository/infrared_state_device_record_case/
git commit -m "feat: add InfraredStateDeviceRecordCase repository (cases, states, raw captures)"
```

---

### Task 8: Deterministic OFAT case generator

**Files:**
- Create: `backend/internal/application/infrared/case_generation/generator.go`
- Create: `backend/internal/application/infrared/case_generation/generator_test.go`

**Interfaces:**
- Consumes: `domainmodels.InfraredState`, `domainmodels.InfraredStateDeviceDefinition` (Task 1).
- Produces: `applicationinfraredcasegeneration.GeneratedCase{Step int32, States map[uuid.UUID]string}` and `applicationinfraredcasegeneration.Generate(states []domainmodels.InfraredState, definitions []domainmodels.InfraredStateDeviceDefinition) []GeneratedCase`.

This is pure logic with no I/O — the strongest TDD candidate in this plan.

- [ ] **Step 1: Write the failing tests**

`backend/internal/application/infrared/case_generation/generator_test.go`:

```go
package applicationinfraredcasegeneration

import (
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

func TestGenerateProducesBaselinePlusOneCasePerStateValue(t *testing.T) {
	powerStateId := uuid.New()
	modeStateId := uuid.New()

	states := []domainmodels.InfraredState{
		{Id: powerStateId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum},
		{Id: modeStateId, Name: "MODE", Type: domainmodels.InfraredStateTypeEnum},
	}
	definitions := []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerStateId, Options: []string{"ON", "OFF"}},
		{InfraredStateId: modeStateId, Options: []string{"COOL", "HEAT", "FAN"}},
	}

	cases := Generate(states, definitions)

	// Baseline (first value of every state) + one case per NON-baseline value:
	// POWER has 1 non-baseline value (OFF), MODE has 2 (HEAT, FAN) => 1 + 1 + 2 = 4.
	if len(cases) != 4 {
		t.Fatalf("Generate() returned %d cases, want 4", len(cases))
	}

	baseline := cases[0]
	if baseline.States[powerStateId] != "ON" || baseline.States[modeStateId] != "COOL" {
		t.Fatalf("baseline states = %v, want POWER=ON MODE=COOL", baseline.States)
	}
	if baseline.Step != 1 {
		t.Fatalf("baseline step = %d, want 1", baseline.Step)
	}

	for i, c := range cases {
		if c.Step != int32(i+1) {
			t.Fatalf("case[%d].Step = %d, want %d (steps must be contiguous starting at 1)", i, c.Step, i+1)
		}
	}
}

func TestGenerateEachNonBaselineCaseDiffersFromBaselineInExactlyOneState(t *testing.T) {
	powerStateId := uuid.New()
	modeStateId := uuid.New()

	states := []domainmodels.InfraredState{
		{Id: powerStateId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum},
		{Id: modeStateId, Name: "MODE", Type: domainmodels.InfraredStateTypeEnum},
	}
	definitions := []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerStateId, Options: []string{"ON", "OFF"}},
		{InfraredStateId: modeStateId, Options: []string{"COOL", "HEAT"}},
	}

	cases := Generate(states, definitions)
	baseline := cases[0]

	for _, c := range cases[1:] {
		differences := 0
		for stateId, value := range c.States {
			if baseline.States[stateId] != value {
				differences++
			}
		}
		if differences != 1 {
			t.Errorf("case at step %d differs from baseline in %d states, want exactly 1 (states = %v)", c.Step, differences, c.States)
		}
	}
}

func TestGenerateRangeStateSweepsFromMinimumToMaximumByStep(t *testing.T) {
	temperatureStateId := uuid.New()

	states := []domainmodels.InfraredState{
		{Id: temperatureStateId, Name: "TEMPERATURE", Type: domainmodels.InfraredStateTypeRange},
	}
	minimum, maximum, step := 16.0, 18.0, 1.0
	definitions := []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: temperatureStateId, Minimum: &minimum, Maximum: &maximum, Step: &step},
	}

	cases := Generate(states, definitions)

	// Baseline = minimum (16); non-baseline sweep = 17, 18 => 3 cases total.
	if len(cases) != 3 {
		t.Fatalf("Generate() returned %d cases, want 3", len(cases))
	}
	want := []string{"16", "17", "18"}
	for i, c := range cases {
		if c.States[temperatureStateId] != want[i] {
			t.Errorf("case[%d] TEMPERATURE = %q, want %q", i, c.States[temperatureStateId], want[i])
		}
	}
}

func TestGenerateReturnsOnlyBaselineWhenNoStatesGiven(t *testing.T) {
	cases := Generate(nil, nil)
	if len(cases) != 1 {
		t.Fatalf("Generate() returned %d cases, want 1 (baseline only)", len(cases))
	}
	if len(cases[0].States) != 0 {
		t.Fatalf("baseline States = %v, want empty", cases[0].States)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/case_generation/... -v`
Expected: FAIL — `Generate` undefined.

- [ ] **Step 3: Write the implementation**

`backend/internal/application/infrared/case_generation/generator.go`:

```go
package applicationinfraredcasegeneration

import (
	"fmt"
	"sort"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type GeneratedCase struct {
	Step   int32
	States map[uuid.UUID]string
}

// Generate enumerates one baseline case (the first value of every state,
// enum options in declared order / range at its minimum) plus one case per
// non-baseline value of every state, changing exactly one state at a time
// from the baseline. This one-factor-at-a-time sweep is what makes every
// later bit-attribution step (Plan B2) sound: any bit that changes between
// a case and the baseline is attributable to that case's single changed
// state.
func Generate(states []domainmodels.InfraredState, definitions []domainmodels.InfraredStateDeviceDefinition) []GeneratedCase {
	definitionByStateId := make(map[uuid.UUID]domainmodels.InfraredStateDeviceDefinition, len(definitions))
	for _, definition := range definitions {
		definitionByStateId[definition.InfraredStateId] = definition
	}

	// Stable, deterministic ordering: sort states by name so re-running
	// Generate on the same input always produces the same case sequence.
	orderedStates := make([]domainmodels.InfraredState, len(states))
	copy(orderedStates, states)
	sort.Slice(orderedStates, func(i, j int) bool { return orderedStates[i].Name < orderedStates[j].Name })

	baseline := make(map[uuid.UUID]string, len(orderedStates))
	valuesByStateId := make(map[uuid.UUID][]string, len(orderedStates))
	for _, state := range orderedStates {
		values := allValues(definitionByStateId[state.Id])
		if len(values) == 0 {
			continue
		}
		valuesByStateId[state.Id] = values
		baseline[state.Id] = values[0]
	}

	cases := []GeneratedCase{{Step: 1, States: cloneMap(baseline)}}
	step := int32(2)
	for _, state := range orderedStates {
		values := valuesByStateId[state.Id]
		for _, value := range values[1:] {
			caseStates := cloneMap(baseline)
			caseStates[state.Id] = value
			cases = append(cases, GeneratedCase{Step: step, States: caseStates})
			step++
		}
	}
	return cases
}

func allValues(definition domainmodels.InfraredStateDeviceDefinition) []string {
	if len(definition.Options) > 0 {
		return definition.Options
	}
	if definition.Minimum == nil || definition.Maximum == nil || definition.Step == nil || *definition.Step <= 0 {
		return nil
	}
	var values []string
	for v := *definition.Minimum; v <= *definition.Maximum; v += *definition.Step {
		values = append(values, formatRangeValue(v))
	}
	return values
}

func formatRangeValue(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%g", v)
}

func cloneMap(m map[uuid.UUID]string) map[uuid.UUID]string {
	clone := make(map[uuid.UUID]string, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/case_generation/... -v`
Expected: PASS on all four tests.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/application/infrared/case_generation/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/infrared/case_generation/
git commit -m "feat: add deterministic OFAT case generator"
```

---

### Task 9: Websocket broadcaster for session status

**Files:**
- Create: `backend/internal/domain/contracts/broadcaster/infrared_record_session.go`
- Create: `backend/internal/infrastructure/broadcaster/infrared_record_session/gorilla.go`
- Create: `backend/internal/infrastructure/broadcaster/infrared_record_session/gorilla_client.go`
- Create: `backend/internal/presentation/http/handler/infrared/broadcast_handler.go`

**Interfaces:**
- Produces: `domaincontractsbroadcaster.InfraredRecordSession` with `Register(ctx, w http.ResponseWriter, r *http.Request, sessionId uuid.UUID) error` and `Send(ctx, event domainmodels.InfraredRecordSessionEvent) error`; a new domain model `domainmodels.InfraredRecordSessionEvent{SessionId uuid.UUID, RecordingState string, CurrentRecordCaseId *uuid.UUID}` (add this to `infrared.go` in this task, not Task 1, since it's specific to the broadcaster wire format); an HTTP handler method `InfraredRecordSessionBroadcastRegister`.

- [ ] **Step 1: Add the broadcaster event model**

Add to `backend/internal/domain/models/infrared.go`:

```go
type InfraredRecordSessionEvent struct {
	SessionId           uuid.UUID
	RecordingState      string
	CurrentRecordCaseId *uuid.UUID
}
```

- [ ] **Step 2: Write the domain contract**

```go
package domaincontractsbroadcaster

import (
	"context"
	"net/http"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type InfraredRecordSession interface {
	Register(ctx context.Context, w http.ResponseWriter, r *http.Request, sessionId uuid.UUID) (err error)
	Send(ctx context.Context, event domainmodels.InfraredRecordSessionEvent) (err error)
}
```

- [ ] **Step 3: Write the implementation**

Read `backend/internal/infrastructure/broadcaster/telemetry/gorilla.go` and `gorilla_client.go` in full first — this task is a direct structural copy with a narrower filter. `gorilla.go`:

```go
package infrastructurebroadcasterinfraredrecordsession

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	domaincontractsbroadcaster "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/broadcaster"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type gorillaImpl struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*client
	upgrader websocket.Upgrader
}

func NewGorillaImpl(allowedOrigins []string) *gorillaImpl {
	return &gorillaImpl{
		sessions: make(map[uuid.UUID]*client),
		upgrader: newUpgrader(allowedOrigins), // reuse the same origin-check helper telemetry's gorilla.go uses, or inline an equivalent
	}
}

func (h *gorillaImpl) Register(ctx context.Context, w http.ResponseWriter, r *http.Request, sessionId uuid.UUID) error {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return domainmodels.NewError("failed to upgrade websocket connection", domainmodels.ErrTypeFailure, err)
	}

	c := newClient(conn, sessionId)
	h.add(c)
	defer h.remove(c)

	go c.writePump()
	c.readPump(ctx)
	return nil
}

func (h *gorillaImpl) Send(ctx context.Context, event domainmodels.InfraredRecordSessionEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return domainmodels.NewError("failed to marshal infrared record session event", domainmodels.ErrTypeFailure, err)
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.sessions {
		if c.sessionId == event.SessionId {
			c.enqueue(payload)
		}
	}
	return nil
}

func (h *gorillaImpl) add(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[c.id] = c
}

func (h *gorillaImpl) remove(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions, c.id)
}

var _ domaincontractsbroadcaster.InfraredRecordSession = (*gorillaImpl)(nil)
```

`gorilla_client.go` — copy `telemetry`'s `client` struct and its `enqueue`/`writePump`/`readPump` methods verbatim, replacing its filter fields (`nodeDeviceId *string`, `metricName *string`, and whatever `matches(...)` does) with a single `sessionId uuid.UUID` field and dropping `matches` entirely (filtering happens inline in `Send` above, by direct equality, since there's only one filter dimension here versus telemetry's two).

**Find `newUpgrader`/the origin-check helper**: `telemetry/gorilla.go`'s `NewGorillaImpl` builds its `websocket.Upgrader` inline or via a shared helper — read it and either reuse that exact helper (if it's exported from a shared location) or copy its `CheckOrigin` logic verbatim. Don't invent new origin-checking logic.

- [ ] **Step 4: Write the HTTP upgrade handler**

`backend/internal/presentation/http/handler/infrared/broadcast_handler.go` — follow `internal/presentation/http/handler/telemetry/handler.go`'s `TelemetryBroadcastRegister` exactly: validate the access token from the `?token=` query param (browsers can't set an `Authorization` header on a WS handshake), check the `infrared_record_session:get` permission claim, parse `sessionId` from the path, then call `h.broadcaster.Register(r.Context(), c.Response(), r, sessionId)`. Include the same comment noting that past the upgrade point, failures can no longer be reported through `presentationhttputils.Error`.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/contracts/broadcaster/ internal/infrastructure/broadcaster/infrared_record_session/ internal/presentation/http/handler/infrared/ internal/domain/models/`
Expected: no output. (The handler file will not yet compile standalone if it references a `broadcaster` field the `handler` struct doesn't have yet — that's fine, Task 14 adds the struct; if `go build` fails here specifically because of that, note it as an expected, isolated failure in your report, matching how the prior plan's Task 10 handled the same situation.)

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/models/infrared.go \
        backend/internal/domain/contracts/broadcaster/infrared_record_session.go \
        backend/internal/infrastructure/broadcaster/infrared_record_session/ \
        backend/internal/presentation/http/handler/infrared/broadcast_handler.go
git commit -m "feat: add websocket broadcaster for infrared record session status"
```

---

### Task 10: Validation helpers

**Files:**
- Modify: `backend/internal/application/shared/validation.go`
- Create: `backend/internal/application/shared/validation_infrared_test.go`

**Interfaces:**
- Produces: `applicationshared.RequiredInfraredRecordingState(value string, field string) (string, error)` (closed membership check against the nine `InfraredRecordingState*` constants from Task 1); `applicationshared.RequiredInfraredBrand(value string, field string) (string, error)` and `RequiredInfraredModel(value string, field string) (string, error)` (simple non-empty checks, following `RequiredLlmModel`'s shape from the prior plan).

- [ ] **Step 1: Write the failing tests**

`backend/internal/application/shared/validation_infrared_test.go`:

```go
package applicationshared

import (
	"errors"
	"testing"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

func TestRequiredInfraredRecordingStateRejectsUnknownValue(t *testing.T) {
	_, err := RequiredInfraredRecordingState("SLEEPING", "recording_state")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RequiredInfraredRecordingState() error = %v, want validation error", err)
	}
}

func TestRequiredInfraredRecordingStateAcceptsKnownValues(t *testing.T) {
	values := []string{
		domainmodels.InfraredRecordingStateDraft,
		domainmodels.InfraredRecordingStateCasesGenerating,
		domainmodels.InfraredRecordingStateRecording,
		domainmodels.InfraredRecordingStateAnalyzing,
		domainmodels.InfraredRecordingStateFunctionGenerating,
		domainmodels.InfraredRecordingStateTestCasesGenerating,
		domainmodels.InfraredRecordingStateTesting,
		domainmodels.InfraredRecordingStateCompleted,
		domainmodels.InfraredRecordingStateFailed,
	}
	for _, value := range values {
		if _, err := RequiredInfraredRecordingState(value, "recording_state"); err != nil {
			t.Errorf("RequiredInfraredRecordingState(%q) error = %v, want nil", value, err)
		}
	}
}

func TestRequiredInfraredBrandRejectsEmpty(t *testing.T) {
	_, err := RequiredInfraredBrand("", "brand")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RequiredInfraredBrand(\"\") error = %v, want validation error", err)
	}
}

func TestRequiredInfraredModelRejectsEmpty(t *testing.T) {
	_, err := RequiredInfraredModel("", "model")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("RequiredInfraredModel(\"\") error = %v, want validation error", err)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/shared/... -run TestRequiredInfrared -v`
Expected: FAIL — undefined functions.

- [ ] **Step 3: Write the validators**

Add to `backend/internal/application/shared/validation.go`:

```go
var validInfraredRecordingStates = map[string]struct{}{
	domainmodels.InfraredRecordingStateDraft:               {},
	domainmodels.InfraredRecordingStateCasesGenerating:     {},
	domainmodels.InfraredRecordingStateRecording:           {},
	domainmodels.InfraredRecordingStateAnalyzing:           {},
	domainmodels.InfraredRecordingStateFunctionGenerating:  {},
	domainmodels.InfraredRecordingStateTestCasesGenerating: {},
	domainmodels.InfraredRecordingStateTesting:             {},
	domainmodels.InfraredRecordingStateCompleted:           {},
	domainmodels.InfraredRecordingStateFailed:              {},
}

func RequiredInfraredRecordingState(value string, field string) (string, error) {
	if _, ok := validInfraredRecordingStates[value]; !ok {
		return "", domainmodels.NewError(fmt.Sprintf("%s is not a recognized recording state", field), domainmodels.ErrTypeValidation, nil)
	}
	return value, nil
}

func RequiredInfraredBrand(value string, field string) (string, error) {
	if value == "" {
		return "", domainmodels.NewError(fmt.Sprintf("%s is required", field), domainmodels.ErrTypeValidation, nil)
	}
	return value, nil
}

func RequiredInfraredModel(value string, field string) (string, error) {
	if value == "" {
		return "", domainmodels.NewError(fmt.Sprintf("%s is required", field), domainmodels.ErrTypeValidation, nil)
	}
	return value, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/shared/... -run TestRequiredInfrared -v`
Expected: PASS on all four tests.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/application/shared/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/shared/validation.go \
        backend/internal/application/shared/validation_infrared_test.go
git commit -m "feat: add validators for infrared recording state, brand, model"
```

---

### Task 11: MQTT capture (subscribe, DTO, handler)

**Files:**
- Modify: `backend/internal/domain/contracts/node/subscriptions.go` (or wherever the file exposing the `Subscriptions` interface lives — confirm the exact path first)
- Modify: `backend/internal/infrastructure/node/subscriptions/mqtt.go`
- Create: `backend/internal/presentation/mqtt/dto/ir_capture.go`
- Create: `backend/internal/presentation/mqtt/handler/ir_capture.go`
- Modify: `backend/internal/presentation/mqtt/event/message.go`

**Interfaces:**
- Produces: `Subscriptions.IrCapture(ctx, nodeDeviceId string) error`; `presentationmqttdto.DecodeIrCapture(deviceId string, payload []byte) (domainusecasesinfrared.CaptureIrRawRequest, error)`; a new `Handler.IrCapture(ctx, msg mqtt.Message, deviceId string)` method; topic routing for the new topic.

- [ ] **Step 1: Add the subscription method**

Find the exact file defining the `domaincontractsnode.Subscriptions` interface (the prior research found the implementation at `internal/infrastructure/node/subscriptions/mqtt.go`; the interface itself lives under `internal/domain/contracts/node/`). Add:

```go
IrCapture(ctx context.Context, nodeDeviceId string) (err error)
```

Implement it in `mqtt.go` following the exact shape of the existing `Telemetry` method:

```go
func (m *mqttImpl) IrCapture(ctx context.Context, nodeDeviceId string) (err error) {
	return m.subscribe(ctx, infrastructurenodeshared.NodePubTopic(nodeDeviceId, "ir_capture"), irCaptureQos)
}
```

Define `irCaptureQos` as a `byte` constant alongside the file's existing `telemetryQos` (same value, unless the existing constants differ per topic for a documented reason — match whatever `telemetryQos` uses).

- [ ] **Step 2: Write the DTO decoder**

`backend/internal/presentation/mqtt/dto/ir_capture.go` — the ESP32 publishes `{"raw_data": [9000, 4500, 560, 560, ...]}` (an array of microsecond mark/space durations) on every IR signal it receives, unconditionally, whenever the node is subscribed. Follow `presentation/mqtt/dto/action_ack.go`'s exact decode-and-validate shape:

```go
package presentationmqttdto

import (
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
)

type IrCapture struct {
	RawData []int32 `json:"raw_data"`
}

func DecodeIrCapture(deviceId string, payload []byte) (domainusecasesinfrared.CaptureIrRawRequest, error) {
	var dto IrCapture
	if err := json.Unmarshal(payload, &dto); err != nil {
		return domainusecasesinfrared.CaptureIrRawRequest{}, domainmodels.NewError("invalid ir capture payload", domainmodels.ErrTypeValidation, err)
	}
	if len(dto.RawData) == 0 {
		return domainusecasesinfrared.CaptureIrRawRequest{}, domainmodels.NewError("ir capture payload has no raw_data", domainmodels.ErrTypeValidation, nil)
	}
	return domainusecasesinfrared.CaptureIrRawRequest{
		NodeDeviceId: deviceId,
		RawData:      dto.RawData,
	}, nil
}
```

`domainusecasesinfrared.CaptureIrRawRequest` is defined in Task 13 — this file will not build in isolation until that task lands; note that as an expected, isolated dependency in your report rather than inventing the type here.

- [ ] **Step 3: Write the MQTT handler method**

`backend/internal/presentation/mqtt/handler/ir_capture.go` — follow `handler/action_ack.go`'s exact shape (decode, call the usecase, log-and-return on error, no response published back):

```go
package presentationmqtthandler

import (
	"context"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationmqttdto "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/dto"
)

func (h *Handler) IrCapture(ctx context.Context, msg mqtt.Message, deviceId string) {
	const tag = path + "/IrCapture"

	request, err := presentationmqttdto.DecodeIrCapture(deviceId, msg.Payload())
	if err != nil {
		h.Logger.Warn(ctx, tag, "invalid ir capture payload", domainmodels.LoggerMeta{"err": err, "device_id": deviceId})
		return
	}

	if err := h.InfraredRecordSession.CaptureIrRaw(ctx, request); err != nil {
		h.Logger.Error(ctx, tag, "failed to handle ir capture", domainmodels.LoggerMeta{"err": err, "device_id": deviceId})
		return
	}
}
```

`h.InfraredRecordSession` is a new field on the `Handler` struct in `handler.go` (Task 13 adds the usecase it points at) — add the field to `Handler` and its constructor in this task, since the struct definition and this new method belong together; leave the constructor call site in `composition/main/` for Task 15 to fix (it will fail to compile until then — expected and isolated, same pattern as noted in prior tasks).

- [ ] **Step 4: Register the topic route**

In `backend/internal/presentation/mqtt/event/message.go`, find where incoming topics are pattern-matched to handler methods (e.g. a switch/if-chain on the topic suffix after the device id) and add a case for `"ir_capture"` calling `handler.IrCapture(ctx, msg, deviceId)`, following the exact shape of the existing `"action/ack"` (or whatever `ActionAck`'s registered topic suffix is) routing entry.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... 2>&1 | grep -v "InfraredRecordSession\|CaptureIrRawRequest"` (filtering out the two expected, forward-referenced symbols this task doesn't define)
Expected: no other build errors. `go vet` and `gofmt -l` on the files this task touched should be clean modulo the same two forward references.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/contracts/node/ \
        backend/internal/infrastructure/node/subscriptions/mqtt.go \
        backend/internal/presentation/mqtt/dto/ir_capture.go \
        backend/internal/presentation/mqtt/handler/ir_capture.go \
        backend/internal/presentation/mqtt/handler/handler.go \
        backend/internal/presentation/mqtt/event/message.go
git commit -m "feat: add MQTT IR capture subscription, decoder, and handler"
```

---

### Task 12: LLM case-script generation

**Files:**
- Create: `backend/internal/application/infrared/case_generation/llm_script.go`
- Create: `backend/internal/application/infrared/case_generation/llm_script_test.go`

**Interfaces:**
- Consumes: `domaincontractsllm.Client`/`GenerateTextRequest`/`GenerateTextResult` (merged prior plan); `applicationinfraredcasegeneration.GeneratedCase` (Task 8); `domainmodels.InfraredState`.
- Produces: `applicationinfraredcasegeneration.WriteScript(ctx context.Context, client domaincontractsllm.Client, deviceBrand string, deviceModel string, states []domainmodels.InfraredState, cases []GeneratedCase) ([]GeneratedCase, error)` — returns the same cases with `Description` populated and `Step` re-ordered per the LLM's chosen press-order.

This is this plan's first consumer of the `llm.Client` abstraction merged in the prior plan.

- [ ] **Step 1: Write the failing test**

`backend/internal/application/infrared/case_generation/llm_script_test.go` — uses a hand-written fake `domaincontractsllm.Client`, not a real provider call:

```go
package applicationinfraredcasegeneration

import (
	"context"
	"testing"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type fakeLlmClient struct {
	responseText string
	err          error
	lastRequest  domaincontractsllm.GenerateTextRequest
}

func (f *fakeLlmClient) GenerateText(_ context.Context, req domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	f.lastRequest = req
	return domaincontractsllm.GenerateTextResult{Text: f.responseText}, f.err
}

func TestWriteScriptAppliesDescriptionsAndReorderingFromLlmResponse(t *testing.T) {
	powerStateId := uuid.New()
	states := []domainmodels.InfraredState{{Id: powerStateId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}
	cases := []GeneratedCase{
		{Step: 1, States: map[uuid.UUID]string{powerStateId: "ON"}},
		{Step: 2, States: map[uuid.UUID]string{powerStateId: "OFF"}},
	}

	client := &fakeLlmClient{responseText: `[
		{"case_index": 0, "description": "Set the remote to POWER ON.", "order": 2},
		{"case_index": 1, "description": "Set the remote to POWER OFF.", "order": 1}
	]`}

	result, err := WriteScript(context.Background(), client, "Polytron", "PAC-09HDN", states, cases)
	if err != nil {
		t.Fatalf("WriteScript() error = %v, want nil", err)
	}
	if len(result) != 2 {
		t.Fatalf("WriteScript() returned %d cases, want 2", len(result))
	}

	// case_index 1 (order 1) comes first now, at Step 1; case_index 0 (order 2) is Step 2.
	if result[0].Description != "Set the remote to POWER OFF." || result[0].Step != 1 {
		t.Errorf("result[0] = %+v, want reordered OFF case at step 1", result[0])
	}
	if result[1].Description != "Set the remote to POWER ON." || result[1].Step != 2 {
		t.Errorf("result[1] = %+v, want reordered ON case at step 2", result[1])
	}

	if client.lastRequest.ResponseSchema == nil {
		t.Fatalf("GenerateText() request had no ResponseSchema — structured output was not requested")
	}
	if client.lastRequest.MaxOutputTokens <= 0 {
		t.Fatalf("GenerateText() request had MaxOutputTokens = %d, want > 0", client.lastRequest.MaxOutputTokens)
	}
}

func TestWriteScriptPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: context.DeadlineExceeded}
	_, err := WriteScript(context.Background(), client, "Polytron", "PAC-09HDN", nil, []GeneratedCase{{Step: 1, States: map[uuid.UUID]string{}}})
	if err == nil {
		t.Fatalf("WriteScript() error = nil, want propagated error")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test ./internal/application/infrared/case_generation/... -run TestWriteScript -v`
Expected: FAIL — `WriteScript` undefined.

- [ ] **Step 3: Write the implementation**

`backend/internal/application/infrared/case_generation/llm_script.go`:

```go
package applicationinfraredcasegeneration

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

const responseSchema = `{
	"type": "array",
	"items": {
		"type": "object",
		"properties": {
			"case_index": {"type": "integer"},
			"description": {"type": "string"},
			"order": {"type": "integer"}
		},
		"required": ["case_index", "description", "order"],
		"additionalProperties": false
	}
}`

type scriptEntry struct {
	CaseIndex   int    `json:"case_index"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

// WriteScript asks the LLM to write a human-facing instruction for every
// case and choose a press-order minimizing physical remote interaction. It
// never changes which state values a case targets — only Description and
// Step, per this plan's contract that case existence is deterministic and
// the LLM's role is language and ordering only.
func WriteScript(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	states []domainmodels.InfraredState,
	cases []GeneratedCase,
) ([]GeneratedCase, error) {
	stateNameById := make(map[string]string, len(states))
	for _, state := range states {
		stateNameById[state.Id.String()] = state.Name
	}

	var promptBuilder strings.Builder
	fmt.Fprintf(&promptBuilder, "Device: %s %s\n\n", deviceBrand, deviceModel)
	promptBuilder.WriteString("Below is a numbered list of remote-control states to physically set up, one per case. ")
	promptBuilder.WriteString("For each case, write a short, clear instruction telling a technician the FULL state to set the remote to " +
		"(every field, not just what changed from the previous case), then choose a press-order (\"order\", 1-based) that groups similar " +
		"button sequences together to minimize physical button presses across the whole list. Respond as a JSON array matching the given schema, " +
		"one entry per case_index below.\n\n")
	for i, c := range cases {
		fmt.Fprintf(&promptBuilder, "case_index %d: ", i)
		fields := make([]string, 0, len(c.States))
		for stateId, value := range c.States {
			fields = append(fields, fmt.Sprintf("%s=%s", stateNameById[stateId.String()], value))
		}
		sort.Strings(fields)
		promptBuilder.WriteString(strings.Join(fields, ", "))
		promptBuilder.WriteString("\n")
	}

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You write concise, technician-facing instructions for recording infrared remote-control signals.",
		Prompt:          promptBuilder.String(),
		MaxOutputTokens: 4096,
		ResponseSchema:  []byte(responseSchema),
	})
	if err != nil {
		return nil, domainmodels.NewError("failed to generate case script", domainmodels.ErrTypeFailure, err)
	}

	var entries []scriptEntry
	if err := json.Unmarshal([]byte(result.Text), &entries); err != nil {
		return nil, domainmodels.NewError("llm returned malformed case script", domainmodels.ErrTypeFailure, err)
	}
	if len(entries) != len(cases) {
		return nil, domainmodels.NewError(
			fmt.Sprintf("llm returned %d case script entries, want %d", len(entries), len(cases)),
			domainmodels.ErrTypeFailure, nil,
		)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Order < entries[j].Order })

	reordered := make([]GeneratedCase, len(cases))
	for newStep, entry := range entries {
		if entry.CaseIndex < 0 || entry.CaseIndex >= len(cases) {
			return nil, domainmodels.NewError(fmt.Sprintf("llm referenced out-of-range case_index %d", entry.CaseIndex), domainmodels.ErrTypeFailure, nil)
		}
		original := cases[entry.CaseIndex]
		reordered[newStep] = GeneratedCase{
			Step:        int32(newStep + 1),
			States:      original.States,
			Description: entry.Description,
		}
	}
	return reordered, nil
}
```

Add a `Description string` field to `GeneratedCase` in `generator.go` (Task 8) as part of this task's change — it wasn't needed until now.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/case_generation/... -v`
Expected: PASS on every test in the package, including Task 8's.

- [ ] **Step 5: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/application/infrared/case_generation/`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/application/infrared/case_generation/
git commit -m "feat: add LLM-driven case script writing and press-order"
```

---

### Task 13: Session management usecase

**Files:**
- Create: `backend/internal/domain/usecases/infrared/record_session_management.go`
- Create: `backend/internal/application/infrared/record_session_management/usecase.go`
- Create: `backend/internal/application/infrared/record_session_management/usecase_test.go`

**Interfaces:**
- Consumes: every repository from Tasks 4–7, `applicationinfraredcasegeneration.Generate`/`WriteScript` (Tasks 8, 12), `domaincontractsbroadcaster.InfraredRecordSession` (Task 9), an LLM `ClientFactory`-shaped dependency (the merged prior plan's `infrastructurellm.ClientFactory`, injected as an interface — see Step 4 for the exact shape this usecase needs from it), and the existing `domaincontractsrepository.Node` repository (find its device-lookup method — confirmed to exist, since `internal/domain/models/node.go`'s `DeviceId` field is used for MQTT addressing elsewhere).
- Produces: `domainusecasesinfrared.RecordSessionManagement` with `Start`, `GetById`, `ListCases`, `AcceptRaw`, `DiscardRaw`, `RetryCase`, `CaptureIrRaw` (the MQTT-driven entry point Task 11's handler calls, using Task 6's `GetActiveByNodeId` to correlate the capture to a session).

This is the task where the background-goroutine pattern is designed — there is no existing convention in this codebase to copy (confirmed by the earlier survey), so this task defines it. `CaptureIrRaw`'s real implementation (not a stub) is written here too, in the same task, since it needs nothing that only composition wiring can provide — Task 15 is pure wiring.

- [ ] **Step 1: Write the usecase interface and request/result DTOs**

`backend/internal/domain/usecases/infrared/record_session_management.go`:

```go
package domainusecasesinfrared

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type RecordSessionManagement interface {
	Start(ctx context.Context, request StartRecordSessionRequest) (sessionId uuid.UUID, err error)
	GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredRecordSession, error)
	ListCases(ctx context.Context, sessionId uuid.UUID) ([]CaseWithStatesAndRaw, error)
	AcceptRaw(ctx context.Context, rawId uuid.UUID) error
	DiscardRaw(ctx context.Context, rawId uuid.UUID, reason string) error
	RetryCase(ctx context.Context, caseId uuid.UUID) error
	CaptureIrRaw(ctx context.Context, request CaptureIrRawRequest) error
}

type StartRecordSessionRequest struct {
	NodeId               uuid.UUID
	InfraredDeviceTypeId uuid.UUID
	Brand                string
	Model                string
	Definitions          []StartRecordSessionDefinition
}

type StartRecordSessionDefinition struct {
	InfraredStateId uuid.UUID
	Options         []string
	Minimum         *float64
	Maximum         *float64
	Step            *float64
}

type CaseWithStatesAndRaw struct {
	Case  domainmodels.InfraredStateDeviceRecordCase
	States []domainmodels.InfraredStateDeviceRecordState
	Raw    []domainmodels.InfraredStateDeviceRecordRaw
}

type CaptureIrRawRequest struct {
	NodeDeviceId string
	RawData      []int32
}
```

- [ ] **Step 2: Write the failing tests**

`backend/internal/application/infrared/record_session_management/usecase_test.go` — this is the biggest test file in this plan; write it before the implementation, following this codebase's hand-written-fake convention (embed the real interface so unimplemented methods are compile-time-satisfied but panic if actually called):

```go
package applicationinfraredrecordsessionmanagement

import (
	"context"
	"errors"
	"testing"

	domaincontractsbroadcaster "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/broadcaster"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
	"github.com/google/uuid"
)

type fakeSessionRepository struct {
	domaincontractsrepository.InfraredRecordSession
	createdNodeId   uuid.UUID
	createdDeviceId uuid.UUID
	created         uuid.UUID
	statusUpdates   []string
	getResult       *domainmodels.InfraredRecordSession
}

func (f *fakeSessionRepository) Create(_ context.Context, nodeId uuid.UUID, deviceId uuid.UUID) (uuid.UUID, error) {
	f.createdNodeId, f.createdDeviceId = nodeId, deviceId
	f.created = uuid.New()
	return f.created, nil
}
func (f *fakeSessionRepository) GetById(_ context.Context, id uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
	return f.getResult, nil
}
func (f *fakeSessionRepository) UpdateRecordingStateById(_ context.Context, _ uuid.UUID, state string, _ bool) error {
	f.statusUpdates = append(f.statusUpdates, state)
	return nil
}
func (f *fakeSessionRepository) UpdateCurrentRecordCaseIdById(_ context.Context, _ uuid.UUID, _ *uuid.UUID) error {
	return nil
}
func (f *fakeSessionRepository) GetActiveByNodeId(_ context.Context, _ uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
	return f.getResult, nil
}

type fakeNodeRepository struct {
	domaincontractsrepository.Node
	result *domainmodels.Node
}

func (f *fakeNodeRepository) GetByDeviceId(_ context.Context, _ string) (*domainmodels.Node, error) {
	return f.result, nil
}

type fakeDeviceRepository struct {
	domaincontractsrepository.InfraredDevice
	created uuid.UUID
}

func (f *fakeDeviceRepository) Create(_ context.Context, _ uuid.UUID, _ string, _ string) (uuid.UUID, error) {
	f.created = uuid.New()
	return f.created, nil
}
func (f *fakeDeviceRepository) GetById(_ context.Context, id uuid.UUID) (*domainmodels.InfraredDevice, error) {
	return &domainmodels.InfraredDevice{Id: id}, nil
}

type fakeDefinitionRepository struct {
	domaincontractsrepository.InfraredStateDeviceDefinition
	createCalls int
}

func (f *fakeDefinitionRepository) CreateMany(_ context.Context, _ []domainmodels.InfraredStateDeviceDefinition) error {
	f.createCalls++
	return nil
}
func (f *fakeDefinitionRepository) ListByDeviceId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error) {
	return nil, nil
}

type fakeStateRepository struct {
	domaincontractsrepository.InfraredState
}

func (f *fakeStateRepository) ListByDeviceTypeId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredState, error) {
	return nil, nil
}

type fakeCaseRepository struct {
	domaincontractsrepository.InfraredStateDeviceRecordCase
	createRawCaseId uuid.UUID
	createRawData   []byte
	createRawCalls  int
}

func (f *fakeCaseRepository) CreateRaw(_ context.Context, caseId uuid.UUID, rawData []byte) (uuid.UUID, error) {
	f.createRawCaseId, f.createRawData = caseId, rawData
	f.createRawCalls++
	return uuid.New(), nil
}

type fakeBroadcaster struct {
	domaincontractsbroadcaster.InfraredRecordSession
	sentEvents []domainmodels.InfraredRecordSessionEvent
}

func (f *fakeBroadcaster) Send(_ context.Context, event domainmodels.InfraredRecordSessionEvent) error {
	f.sentEvents = append(f.sentEvents, event)
	return nil
}

type fakeSubscriptions struct {
	domaincontractsnode.Subscriptions
	subscribedDeviceIds []string
}

func (f *fakeSubscriptions) IrCapture(_ context.Context, nodeDeviceId string) error {
	f.subscribedDeviceIds = append(f.subscribedDeviceIds, nodeDeviceId)
	return nil
}

type noopLogger struct{ domaincontractslogger.Leveled }

func TestStartRejectsEmptyBrand(t *testing.T) {
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &noopLogger{},
	)

	_, err := usecase.Start(context.Background(), domainusecasesinfrared.StartRecordSessionRequest{
		Brand: "", Model: "PAC-09HDN",
	})
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("Start() error = %v, want validation error", err)
	}
}

func TestStartCreatesDeviceDefinitionsAndSessionThenKicksOffCaseGeneration(t *testing.T) {
	sessionRepo := &fakeSessionRepository{}
	deviceRepo := &fakeDeviceRepository{}
	definitionRepo := &fakeDefinitionRepository{}
	usecase := NewUsecaseImpl(
		sessionRepo, deviceRepo, definitionRepo,
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &noopLogger{},
	)

	nodeId := uuid.New()
	sessionId, err := usecase.Start(context.Background(), domainusecasesinfrared.StartRecordSessionRequest{
		NodeId: nodeId, Brand: "Polytron", Model: "PAC-09HDN",
	})
	if err != nil {
		t.Fatalf("Start() error = %v, want nil", err)
	}
	if sessionId != sessionRepo.created {
		t.Fatalf("Start() returned %v, want the repository's created id %v", sessionId, sessionRepo.created)
	}
	if sessionRepo.createdNodeId != nodeId {
		t.Fatalf("session created with node id %v, want %v", sessionRepo.createdNodeId, nodeId)
	}
	if definitionRepo.createCalls != 1 {
		t.Fatalf("definition repository CreateMany() calls = %d, want 1", definitionRepo.createCalls)
	}

	// The case-generation goroutine runs asynchronously; give it a moment,
	// then confirm the session transitioned into CASES_GENERATING at least.
	// (Task 13's real implementation must make this observable synchronously
	// for the *first* transition — see the "Step 3" note on making the DRAFT
	// -> CASES_GENERATING handoff itself synchronous up to the point the
	// goroutine is launched.)
	found := false
	for _, s := range sessionRepo.statusUpdates {
		if s == domainmodels.InfraredRecordingStateCasesGenerating {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include CASES_GENERATING", sessionRepo.statusUpdates)
	}
}

func TestDiscardRawRequiresNonEmptyReason(t *testing.T) {
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &noopLogger{},
	)

	err := usecase.DiscardRaw(context.Background(), uuid.New(), "")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("DiscardRaw() error = %v, want validation error", err)
	}
}

func TestCaptureIrRawPersistsRawAgainstCurrentCase(t *testing.T) {
	caseId := uuid.New()
	sessionId := uuid.New()
	nodeId := uuid.New()

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{
		Id: sessionId, RecordingState: domainmodels.InfraredRecordingStateRecording, CurrentRecordCaseId: &caseId,
	}}
	caseRepo := &fakeCaseRepository{}
	nodeRepo := &fakeNodeRepository{result: &domainmodels.Node{Id: nodeId}}
	broadcaster := &fakeBroadcaster{}

	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, broadcaster, &fakeSubscriptions{},
		nil, nodeRepo, &noopLogger{},
	)

	err := usecase.CaptureIrRaw(context.Background(), domainusecasesinfrared.CaptureIrRawRequest{
		NodeDeviceId: "AC276E5E030C", RawData: []int32{9000, 4500, 560, 560},
	})
	if err != nil {
		t.Fatalf("CaptureIrRaw() error = %v, want nil", err)
	}
	if caseRepo.createRawCalls != 1 {
		t.Fatalf("CreateRaw() calls = %d, want 1", caseRepo.createRawCalls)
	}
	if caseRepo.createRawCaseId != caseId {
		t.Fatalf("CreateRaw() case id = %v, want the session's CurrentRecordCaseId %v", caseRepo.createRawCaseId, caseId)
	}
	if len(broadcaster.sentEvents) != 1 || broadcaster.sentEvents[0].SessionId != sessionId {
		t.Fatalf("broadcaster sent events = %+v, want one event for session %v", broadcaster.sentEvents, sessionId)
	}
}

func TestCaptureIrRawDoesNothingWhenNodeUnknown(t *testing.T) {
	caseRepo := &fakeCaseRepository{}
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{result: nil}, &noopLogger{},
	)

	err := usecase.CaptureIrRaw(context.Background(), domainusecasesinfrared.CaptureIrRawRequest{
		NodeDeviceId: "unknown-device", RawData: []int32{9000},
	})
	if err != nil {
		t.Fatalf("CaptureIrRaw() error = %v, want nil (unknown node is a no-op, not a failure)", err)
	}
	if caseRepo.createRawCalls != 0 {
		t.Fatalf("CreateRaw() calls = %d, want 0", caseRepo.createRawCalls)
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -v`
Expected: FAIL — `NewUsecaseImpl` undefined.

- [ ] **Step 4: Write the usecase implementation**

`backend/internal/application/infrared/record_session_management/usecase.go` — this is the task's core design decision. **Background-job pattern (new to this codebase):** `Start` synchronously creates the device, definitions, and session (all fast, request-scoped work), synchronously flips `RecordingState` to `CASES_GENERATING` and broadcasts that transition (so the *first* observable state change happens within the HTTP request, making the pattern testable without a sleep/poll in the test above), then launches `go u.runCaseGeneration(sessionId, ...)` detached from the request context (`context.Background()` with its own `context.WithTimeout`, since the request that triggered it will already have returned) to do the actual LLM-backed case generation and persistence, updating `RecordingState` to `RECORDING` (success) or `FAILED` (error) when the goroutine finishes, broadcasting at each transition, and logging any error from inside the goroutine (nothing else can observe it — there's no request to return an error to).

```go
package applicationinfraredrecordsessionmanagement

import (
	"context"
	"encoding/json"
	"time"

	applicationinfraredcasegeneration "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/case_generation"
	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractsbroadcaster "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/broadcaster"
	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
	"github.com/google/uuid"
)

// LlmClientFactory is the subset of infrastructurellm.ClientFactory this
// usecase needs — declared as a domain-layer interface so the usecase
// doesn't import the infrastructure package directly.
type LlmClientFactory interface {
	Current(ctx context.Context) (domaincontractsllm.Client, error)
}

const caseGenerationTimeout = 2 * time.Minute

type usecase struct {
	session       domaincontractsrepository.InfraredRecordSession
	device        domaincontractsrepository.InfraredDevice
	definition    domaincontractsrepository.InfraredStateDeviceDefinition
	state         domaincontractsrepository.InfraredState
	recordCase    domaincontractsrepository.InfraredStateDeviceRecordCase
	broadcaster   domaincontractsbroadcaster.InfraredRecordSession
	subscriptions domaincontractsnode.Subscriptions
	llmFactory    LlmClientFactory
	node          domaincontractsrepository.Node
	logger        domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	session domaincontractsrepository.InfraredRecordSession,
	device domaincontractsrepository.InfraredDevice,
	definition domaincontractsrepository.InfraredStateDeviceDefinition,
	state domaincontractsrepository.InfraredState,
	recordCase domaincontractsrepository.InfraredStateDeviceRecordCase,
	broadcaster domaincontractsbroadcaster.InfraredRecordSession,
	subscriptions domaincontractsnode.Subscriptions,
	llmFactory LlmClientFactory,
	node domaincontractsrepository.Node,
	logger domaincontractslogger.Leveled,
) domainusecasesinfrared.RecordSessionManagement {
	return &usecase{
		session: session, device: device, definition: definition, state: state,
		recordCase: recordCase, broadcaster: broadcaster, subscriptions: subscriptions,
		llmFactory: llmFactory, node: node, logger: logger,
	}
}

func (u *usecase) Start(ctx context.Context, request domainusecasesinfrared.StartRecordSessionRequest) (uuid.UUID, error) {
	const tag = "infrared/record_session_management/Start"

	brand, err := applicationshared.RequiredInfraredBrand(request.Brand, "brand")
	if err != nil {
		return uuid.Nil, err
	}
	model, err := applicationshared.RequiredInfraredModel(request.Model, "model")
	if err != nil {
		return uuid.Nil, err
	}

	deviceId, err := u.device.Create(ctx, request.InfraredDeviceTypeId, brand, model)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create infrared device", domainmodels.LoggerMeta{"err": err})
		return uuid.Nil, err
	}

	definitions := make([]domainmodels.InfraredStateDeviceDefinition, 0, len(request.Definitions))
	for _, d := range request.Definitions {
		definitions = append(definitions, domainmodels.InfraredStateDeviceDefinition{
			InfraredDeviceId: deviceId, InfraredStateId: d.InfraredStateId,
			Options: d.Options, Minimum: d.Minimum, Maximum: d.Maximum, Step: d.Step,
		})
	}
	if err := u.definition.CreateMany(ctx, definitions); err != nil {
		u.logger.Error(ctx, tag, "failed to create infrared state device definitions", domainmodels.LoggerMeta{"err": err})
		return uuid.Nil, err
	}

	sessionId, err := u.session.Create(ctx, request.NodeId, deviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create infrared record session", domainmodels.LoggerMeta{"err": err})
		return uuid.Nil, err
	}

	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateCasesGenerating)

	go u.runCaseGeneration(sessionId, deviceId, request.InfraredDeviceTypeId, brand, model)

	return sessionId, nil
}

func (u *usecase) runCaseGeneration(sessionId uuid.UUID, deviceId uuid.UUID, deviceTypeId uuid.UUID, brand string, model string) {
	const tag = "infrared/record_session_management/runCaseGeneration"

	ctx, cancel := context.WithTimeout(context.Background(), caseGenerationTimeout)
	defer cancel()

	states, err := u.state.ListByDeviceTypeId(ctx, deviceTypeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list infrared states", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	definitions, err := u.definition.ListByDeviceId(ctx, deviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list infrared state device definitions", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	cases := applicationinfraredcasegeneration.Generate(states, definitions)

	client, err := u.llmFactory.Current(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve llm client", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	scriptedCases, err := applicationinfraredcasegeneration.WriteScript(ctx, client, brand, model, states, cases)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to write case script", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	for _, c := range scriptedCases {
		states := make([]domainmodels.InfraredStateDeviceRecordState, 0, len(c.States))
		for stateId, value := range c.States {
			states = append(states, domainmodels.InfraredStateDeviceRecordState{InfraredStateId: stateId, StateValue: value})
		}
		if _, err := u.recordCase.CreateWithStates(ctx, sessionId, c.Step, c.Description, states); err != nil {
			u.logger.Error(ctx, tag, "failed to persist record case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "step": c.Step})
			u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
			return
		}
	}

	session, err := u.session.GetById(ctx, sessionId)
	if err == nil && session != nil {
		// TODO(B-later): fetch the node's DeviceId and call
		// u.subscriptions.IrCapture(ctx, node.DeviceId) here once the node
		// repository lookup is wired in — deferred to Task 15's composition
		// wiring, which has the node repository available to inject.
		_ = session
	}

	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateRecording)
}

func (u *usecase) transition(ctx context.Context, tag string, sessionId uuid.UUID, state string) {
	isCompleted := state == domainmodels.InfraredRecordingStateCompleted || state == domainmodels.InfraredRecordingStateFailed
	if err := u.session.UpdateRecordingStateById(ctx, sessionId, state, isCompleted); err != nil {
		u.logger.Error(ctx, tag, "failed to update recording state", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "state": state})
		return
	}
	if err := u.broadcaster.Send(ctx, domainmodels.InfraredRecordSessionEvent{SessionId: sessionId, RecordingState: state}); err != nil {
		u.logger.Warn(ctx, tag, "failed to broadcast recording state", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
	}
}

func (u *usecase) GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
	return u.session.GetById(ctx, id)
}

func (u *usecase) ListCases(ctx context.Context, sessionId uuid.UUID) ([]domainusecasesinfrared.CaseWithStatesAndRaw, error) {
	cases, err := u.recordCase.ListBySessionId(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	result := make([]domainusecasesinfrared.CaseWithStatesAndRaw, 0, len(cases))
	for _, c := range cases {
		states, err := u.recordCase.ListStatesByCaseId(ctx, c.Id)
		if err != nil {
			return nil, err
		}
		raw, err := u.recordCase.ListRawByCaseId(ctx, c.Id)
		if err != nil {
			return nil, err
		}
		result = append(result, domainusecasesinfrared.CaseWithStatesAndRaw{Case: c, States: states, Raw: raw})
	}
	return result, nil
}

func (u *usecase) AcceptRaw(ctx context.Context, rawId uuid.UUID) error {
	return u.recordCase.UpdateRawStatusById(ctx, rawId, domainmodels.InfraredRecordRawStatusAccepted, nil)
}

func (u *usecase) DiscardRaw(ctx context.Context, rawId uuid.UUID, reason string) error {
	if reason == "" {
		return domainmodels.NewError("reason is required", domainmodels.ErrTypeValidation, nil)
	}
	return u.recordCase.UpdateRawStatusById(ctx, rawId, domainmodels.InfraredRecordRawStatusDiscarded, &reason)
}

func (u *usecase) RetryCase(ctx context.Context, caseId uuid.UUID) error {
	const tag = "infrared/record_session_management/RetryCase"

	raw, err := u.recordCase.ListRawByCaseId(ctx, caseId)
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
	return u.recordCase.UpdateStatusById(ctx, caseId, domainmodels.InfraredRecordCaseStatusActive)
}

func (u *usecase) CaptureIrRaw(ctx context.Context, request domainusecasesinfrared.CaptureIrRawRequest) error {
	const tag = "infrared/record_session_management/CaptureIrRaw"

	node, err := u.node.GetByDeviceId(ctx, request.NodeDeviceId)
	if err != nil || node == nil {
		u.logger.Warn(ctx, tag, "ir capture from unknown node", domainmodels.LoggerMeta{"device_id": request.NodeDeviceId})
		return nil
	}

	session, err := u.session.GetActiveByNodeId(ctx, node.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up active session for node", domainmodels.LoggerMeta{"err": err, "node_id": node.Id})
		return err
	}
	if session == nil || session.CurrentRecordCaseId == nil {
		u.logger.Warn(ctx, tag, "ir capture with no active recording case", domainmodels.LoggerMeta{"node_id": node.Id})
		return nil
	}

	rawBytes, err := json.Marshal(request.RawData)
	if err != nil {
		return domainmodels.NewError("failed to encode raw ir data", domainmodels.ErrTypeFailure, err)
	}
	if _, err := u.recordCase.CreateRaw(ctx, *session.CurrentRecordCaseId, rawBytes); err != nil {
		u.logger.Error(ctx, tag, "failed to persist raw capture", domainmodels.LoggerMeta{"err": err, "case_id": *session.CurrentRecordCaseId})
		return err
	}

	if err := u.broadcaster.Send(ctx, domainmodels.InfraredRecordSessionEvent{
		SessionId: session.Id, RecordingState: session.RecordingState, CurrentRecordCaseId: session.CurrentRecordCaseId,
	}); err != nil {
		u.logger.Warn(ctx, tag, "failed to broadcast raw capture", domainmodels.LoggerMeta{"err": err, "session_id": session.Id})
	}
	return nil
}
```

`u.node.GetByDeviceId` is a placeholder method name for this draft — before writing this method for real, grep the actual `domaincontractsrepository.Node` (or equivalent) interface for its real device-lookup method name and signature, and use that exact name instead (it is confirmed to exist, since `internal/domain/models/node.go`'s `DeviceId` field is used for MQTT addressing elsewhere in this codebase — this task is finding and reusing it, not inventing a new one). If the real repository's `Node` type or lookup method differs in shape from the sketch above (e.g. returns an error type this file needs to import differently), adapt accordingly — the logic (look up node, find its active session, persist the raw capture against the current case, broadcast) is what this task must preserve, not the exact placeholder names.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -v`
Expected: PASS on every test, including the new one for `CaptureIrRaw` you wrote in Step 4.

- [ ] **Step 6: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l internal/domain/usecases/infrared/ internal/application/infrared/record_session_management/`
Expected: no output.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/domain/usecases/infrared/ \
        backend/internal/application/infrared/record_session_management/
git commit -m "feat: add RecordSessionManagement usecase with async case generation"
```

---

### Task 14: HTTP presentation layer

**Files:**
- Create: `backend/internal/presentation/http/request/infrared.go`
- Create: `backend/internal/presentation/http/response/infrared.go`
- Create: `backend/internal/presentation/http/handler/infrared/handler.go`
- Modify: `backend/internal/presentation/http/handler/infrared/broadcast_handler.go` (Task 9 — add the `broadcaster` field this file references to the same `handler` struct this task defines)
- Modify: `backend/internal/presentation/http/route/route.go`
- Modify: `backend/database/seeder/permission.json`, `role.json` (or their real names, per the prior plan's Task 10 finding)

**Interfaces:**
- Consumes: `domainusecasesinfrared.RecordSessionManagement` (Task 13), `domaincontractsbroadcaster.InfraredRecordSession` (Task 9).
- Produces: `POST /v1/infrared/record-sessions`, `GET /v1/infrared/record-sessions/:id`, `GET /v1/infrared/record-sessions/:id/cases`, `POST /v1/infrared/record-cases/:caseId/raw/:rawId/accept`, `POST /v1/infrared/record-cases/:caseId/raw/:rawId/discard`, `POST /v1/infrared/record-cases/:caseId/retry`, `GET /v1/infrared/record-sessions/:id/broadcast` (the websocket upgrade), gated by new `infrared_record_session:get`/`infrared_record_session:add`/`infrared_record_session:set` permissions.

- [ ] **Step 1: Write request/response DTOs**

`backend/internal/presentation/http/request/infrared.go`:

```go
package presentationhttprequest

type StartRecordSessionRequest struct {
	NodeId               string                          `json:"node_id" example:"..."`
	InfraredDeviceTypeId string                          `json:"infrared_device_type_id" example:"..."`
	Brand                string                          `json:"brand" example:"Polytron"`
	Model                string                          `json:"model" example:"PAC-09HDN"`
	Definitions          []StartRecordSessionDefinition `json:"definitions"`
}

type StartRecordSessionDefinition struct {
	InfraredStateId string    `json:"infrared_state_id" example:"..."`
	Options         []string  `json:"options,omitempty" example:"ON,OFF"`
	Minimum         *float64  `json:"minimum,omitempty" example:"16"`
	Maximum         *float64  `json:"maximum,omitempty" example:"30"`
	Step            *float64  `json:"step,omitempty" example:"1"`
}

type DiscardRawRequest struct {
	Reason string `json:"reason" example:"pressed the wrong button"`
}
```

`backend/internal/presentation/http/response/infrared.go` — one `XResponse` struct + one mapper function per domain model, following `llm_config.go`'s exact shape from the prior plan: `InfraredRecordSessionResponse`, `InfraredStateDeviceRecordCaseResponse` (nested `States []InfraredStateDeviceRecordStateResponse` and `Raw []InfraredStateDeviceRecordRawResponse`), each with a plural `Xs(...)` mapper.

- [ ] **Step 2: Write the handler**

`backend/internal/presentation/http/handler/infrared/handler.go` — follow `handler/admin/handler.go`'s struct-of-usecases + swagger-annotated-methods shape exactly:

```go
package presentationhttphandlerinfrared

import (
	"net/http"

	domaincontractsbroadcaster "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/broadcaster"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
	presentationhttprequest "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

type handler struct {
	recordSessionUseCase domainusecasesinfrared.RecordSessionManagement
	broadcaster           domaincontractsbroadcaster.InfraredRecordSession
}

func NewHandler(
	recordSessionUseCase domainusecasesinfrared.RecordSessionManagement,
	broadcaster domaincontractsbroadcaster.InfraredRecordSession,
) *handler {
	return &handler{recordSessionUseCase: recordSessionUseCase, broadcaster: broadcaster}
}
```

Write `RecordSessionPost` (binds `StartRecordSessionRequest`, parses the two UUID fields and each definition's `InfraredStateId` via `presentationhttputils.RequiredUUID`, calls `Start`, returns `201` with `{"id": "..."}`), `RecordSessionGetById`, `RecordSessionCasesGetList`, `RecordCaseRawAccept`, `RecordCaseRawDiscard` (binds `DiscardRawRequest`), `RecordCaseRetry` — each following `handler/admin/handler.go`'s exact validate-bind-delegate-respond shape and swagger comment block.

`InfraredRecordSessionBroadcastRegister` (from Task 9) is completed here — it references `h.broadcaster`, which now exists on the struct.

- [ ] **Step 3: Register routes and permissions**

In `route.go`, add an `InfraredHandler` interface (listing every method above) and a `routeInfrared(v1 *echo.Group, handler InfraredHandler, permission PermissionMiddleware)` function, following `routeAdmin`'s exact shape:

```go
v1.POST("/infrared/record-sessions", handler.RecordSessionPost, permission("infrared_record_session:add"))
v1.GET("/infrared/record-sessions/:id", handler.RecordSessionGetById, permission("infrared_record_session:get"))
v1.GET("/infrared/record-sessions/:id/cases", handler.RecordSessionCasesGetList, permission("infrared_record_session:get"))
v1.GET("/infrared/record-sessions/:id/broadcast", handler.InfraredRecordSessionBroadcastRegister, permission("infrared_record_session:get"))
v1.POST("/infrared/record-cases/:caseId/raw/:rawId/accept", handler.RecordCaseRawAccept, permission("infrared_record_session:set"))
v1.POST("/infrared/record-cases/:caseId/raw/:rawId/discard", handler.RecordCaseRawDiscard, permission("infrared_record_session:set"))
v1.POST("/infrared/record-cases/:caseId/retry", handler.RecordCaseRetry, permission("infrared_record_session:set"))
```

Add `infrared_record_session:get`/`add`/`set` to the seed permission file found in the prior plan's Task 10, granted to whichever role already receives `node_config:*`-shaped operational permissions (following that task's precedent for where a new operational permission belongs).

- [ ] **Step 4: Verify**

Run: `cd backend && go build ./... 2>&1 | grep -v "composition/main"` (composition wiring is Task 15; an isolated failure there is expected)
Expected: no other build errors. `go vet` and `gofmt -l` on everything this task touched should be clean.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/presentation/http/request/infrared.go \
        backend/internal/presentation/http/response/infrared.go \
        backend/internal/presentation/http/handler/infrared/ \
        backend/internal/presentation/http/route/route.go \
        backend/database/seeder/permission.json backend/database/seeder/role.json
git commit -m "feat: add infrared record session HTTP endpoints"
```

---

### Task 15: Composition wiring

**Files:**
- Modify: `backend/internal/composition/main/infrastructure.go`
- Modify: `backend/internal/composition/main/application.go`
- Modify: `backend/internal/composition/main/presentation.go`

**Interfaces:**
- Consumes: every repository, the case generator, the broadcaster, and the handler from Tasks 4–14; `infrastructurellm.ClientFactory` (merged prior plan); the existing node repository (already a constructor dependency of Task 13's usecase).

Pure wiring — no business logic. `CaptureIrRaw` is already fully implemented and tested as of Task 13; this task only constructs and connects everything.

- [ ] **Step 1: Wire everything into the composition root**

In `infrastructure.go`, construct every new repository (`infraredDeviceTypeRepository`, `infraredStateRepository`, `infraredDeviceRepository`, `infraredStateDeviceDefinitionRepository`, `infraredRecordSessionRepository`, `infraredStateDeviceRecordCaseRepository`) and the broadcaster (`infraredRecordSessionBroadcaster := infrastructurebroadcasterinfraredrecordsession.NewGorillaImpl(config.HttpCorsAllowedOrigins)`), following the exact one-field-plus-one-constructor-line pattern used for every existing entity in that file.

In `application.go`, construct the usecase: `infraredRecordSessionManagement := applicationinfraredrecordsessionmanagement.NewUsecaseImpl(l.infra.infraredRecordSessionRepository, l.infra.infraredDeviceRepository, l.infra.infraredStateDeviceDefinitionRepository, l.infra.infraredStateRepository, l.infra.infraredStateDeviceRecordCaseRepository, l.infra.infraredRecordSessionBroadcaster, l.infra.nodeSubscriptions, l.infra.llmClientFactory, l.infra.nodeRepository, l.infra.logger)` — matching whatever exact parameter order Task 13/Step 1's final constructor signature ends up with.

In `presentation.go`, construct the handler: `infraredHandler := presentationhttphandlerinfrared.NewHandler(l.app.infraredRecordSessionManagement, l.infra.infraredRecordSessionBroadcaster)`, and pass it (plus the `mqttHandler` struct's new `InfraredRecordSession` field from Task 11) into wherever the MQTT `Handler` and HTTP route registration are constructed.

- [ ] **Step 2: Verify**

Run: `cd backend && go build ./... && go vet ./... && gofmt -l . && go test -count=1 ./...`
Expected: no build/vet/gofmt output; every test from Tasks 1–15 passes; the only failures anywhere are pre-existing ones already known to be unrelated to this plan (confirm via `git log` that any failing package's test file predates this plan's first commit).

- [ ] **Step 3: Manual smoke check (best-effort)**

If a local Postgres + this backend running is available: `POST /v1/infrared/record-sessions` with a body naming a real node and the seeded Air Conditioner device type plus five definitions, confirm a `202`-equivalent response (or whatever status Task 14 chose) with a session id, then `GET /v1/infrared/record-sessions/:id` a few seconds later and confirm `recording_state` has moved to `RECORDING` (or `FAILED`, if no LLM provider is configured — check `GET /v1/admin/llm-config` first, from the prior plan, and configure one if `api_key_set` is `false`). If no environment is available, skip and say so — same as the prior plan's Task 11.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/composition/main/
git commit -m "feat: wire IR recording session capture pipeline into composition root"
```

---

## Self-review notes

- **Spec coverage:** every element of the design spec's B1-scoped pipeline (form submission → device/definition creation → deterministic case enumeration → LLM script/ordering → MQTT capture → misclick retry via case/raw status → live status via websocket) has a task. Analysis, function generation, goja, test-case generation, and the TESTING loop are explicitly out of scope for B1 and are named as such in the Goal.
- **Placeholder scan:** no shipped stub or `panic` remains — `CaptureIrRaw` is fully implemented and tested in Task 13 itself (an earlier draft of this plan split it across Task 13 and Task 15, which was a real self-contradiction caught by the pre-execution conflict scan and fixed before dispatch: the method needs nothing composition-layer-specific, so it belongs in the same task as the rest of the usecase, and Task 15 is pure wiring).
- **Type consistency:** `domainmodels.InfraredRecordSession`/`InfraredStateDeviceRecordCase`/etc. (Task 1) are used with identical field names across every repository, usecase, and HTTP task. `applicationinfraredcasegeneration.GeneratedCase` (Task 8) gains its `Description` field in Task 12, and every later reference to it (Task 13) already expects that field. `domainusecasesinfrared.RecordSessionManagement`'s method set (Task 13) matches the handler methods built against it in Task 14 one-to-one. `domaincontractsrepository.InfraredRecordSession.GetActiveByNodeId` (Task 6) and `NewUsecaseImpl`'s final 10-parameter signature (Task 13) are used identically by Task 15's wiring line.
- **Seed-persistence gap (found during execution, fixed):** the original Task 4 specified only read-only `List`/`ListByDeviceTypeId` repository methods, while Task 2 only made the seed JSON parseable into `seederdata.Data` — nothing in the plan as originally written ever called a `Create` to actually persist the seeded "Air Conditioner" device type/states. Fixed by amending Task 4 to add `ReadByName`/`Create` to both repositories plus `seedInfraredDeviceTypes`/`seedInfraredStates` usecase functions and composition wiring, so the seed data lands in the database end-to-end.
