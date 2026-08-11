# Infrared Record Sessions: List + Detail (Spec A) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the infrared Record list page (filters + paginated table) and a full detail page for an existing record session (overview, cases with raws, coder, test cases, delete controls), backed by one new paginated list endpoint and the record-session REST surface that already exists.

**Architecture:** Backend gains one new repository method + usecase method + route (`GET /v1/infrared/record-sessions`), mirroring `ActionLog.ReadByFilter`'s exact shape (joined list query, `PageDataResponse` envelope). Frontend gains a server-rendered list page (`record/page.tsx`, filters + table + pagination, same skeleton as `action-history/page.tsx`) and a server-rendered detail page (`record/[id]/page.tsx`, same skeleton as `nodes/[id]/page.tsx`) built from small, focused client components for each mutable section (raw accept/discard, case retry, deletes).

**Tech Stack:** Go/Echo/squirrel/pgx (existing backend stack), Next.js 16 App Router Server Components + Server Actions (existing frontend stack), Tailwind (existing styling).

## Global Constraints

- New list endpoint permission: `infrared_record_session:get` (already exists, reused — no new permission).
- Mutation permission for accept/discard/retry: `infrared_record_session:set` (already exists, reused).
- Delete permission: `infrared_record_session:delete` (already exists, reused).
- List filters: `recording_state` (exact match against one of the 9 `InfraredRecordingState*` constants), `infrared_device_type_id` (UUID), `created_at_start`/`created_at_end` (RFC3339, via the existing `presentationhttputils.QueryTime`).
- List response envelope: `presentationhttpresponse.PageDataResponse[InfraredRecordSessionListItemResponse]` — same generic type every other paginated endpoint in this codebase uses.
- No schema/migration changes — this plan only adds a read path over existing tables.
- Frontend: every new file follows this codebase's existing conventions exactly — `"use server"` action files can only export async functions (constants/types go in a sibling `_lib/state.ts`), server components fetch data directly via the `server-only` API client in `src/lib/api/`, client components are the smallest possible unit that needs interactivity.

---

### Task 1: Backend — `InfraredRecordSessionListItem` domain model + repository `ReadByFilter`

**Files:**
- Modify: `backend/internal/domain/models/infrared.go` (add `InfraredRecordSessionListItem` struct, near the existing `InfraredRecordSession` struct at line 45)
- Modify: `backend/internal/domain/contracts/repository/infrared_record_session.go` (add `ReadByFilter` to the interface)
- Modify: `backend/internal/infrastructure/repository/infrared_record_session/postgres_query.go` (add `queryReadByFilter` + `applyInfraredRecordSessionFilters`)
- Modify: `backend/internal/infrastructure/repository/infrared_record_session/postgres.go` (add `ReadByFilter` implementation + `scanInfraredRecordSessionListItem`)
- Modify: `backend/internal/infrastructure/repository/shared/scan_pgx.go` (add `ScanPgxInfraredRecordSessionListItem`/`ScanPgxInfraredRecordSessionListItems`, following the exact pattern of `ScanPgxActionLogListItem`/`ScanPgxActionLogListItems` at line 328)
- Test: `backend/internal/infrastructure/repository/infrared_record_session/postgres_test.go` (extend — this file already exists from the earlier repair-loop plan's Task 5)

**Interfaces:**
- Consumes: nothing new — `infrared_record_session`/`infrared_device`/`infrared_device_type` tables already exist (created in `20260810090001_infrared_recording.up.sql` and `20260810090000_infrared_reference.up.sql`).
- Produces (for Task 2 to consume):
  ```go
  type InfraredRecordSessionListItem struct {
      Id                   uuid.UUID
      RecordingState       string
      IsCompleted          bool
      InfraredDeviceId     uuid.UUID
      Brand                string
      Model                string
      InfraredDeviceTypeId uuid.UUID
      DeviceTypeName       string
      CreatedAt            time.Time
  }
  ```
  ```go
  ReadByFilter(
      ctx context.Context,
      recordingState *string,
      infraredDeviceTypeId *uuid.UUID,
      createdAtStart *time.Time,
      createdAtEnd *time.Time,
      page int,
      limit int,
  ) (items []domainmodels.InfraredRecordSessionListItem, total int, err error)
  ```

- [ ] **Step 1: Add the domain model**

In `backend/internal/domain/models/infrared.go`, add right after the existing `InfraredRecordSession` struct (currently ending at line 56):

```go
type InfraredRecordSessionListItem struct {
	Id                   uuid.UUID
	RecordingState       string
	IsCompleted          bool
	InfraredDeviceId     uuid.UUID
	Brand                string
	Model                string
	InfraredDeviceTypeId uuid.UUID
	DeviceTypeName       string
	CreatedAt            time.Time
}
```

- [ ] **Step 2: Add the repository interface method**

In `backend/internal/domain/contracts/repository/infrared_record_session.go`, add to the `InfraredRecordSession` interface, right after `ReadActiveByNodeId`:

```go
	// ReadByFilter lists sessions joined against their device and device
	// type for display purposes (brand/model/device-type name), filtered
	// and paginated — the list-page counterpart to ReadById's single-item
	// lookup.
	ReadByFilter(
		ctx context.Context,
		recordingState *string,
		infraredDeviceTypeId *uuid.UUID,
		createdAtStart *time.Time,
		createdAtEnd *time.Time,
		page int,
		limit int,
	) (items []domainmodels.InfraredRecordSessionListItem, total int, err error)
```

Add `"time"` to this file's imports if not already present — check first.

- [ ] **Step 3: Add the scan helper**

In `backend/internal/infrastructure/repository/shared/scan_pgx.go`, add right after `ScanPgxActionLogListItems` (currently ending at line 357):

```go
func ScanPgxInfraredRecordSessionListItem(row pgx.Row) (domainmodels.InfraredRecordSessionListItem, error) {
	var item domainmodels.InfraredRecordSessionListItem
	err := row.Scan(
		&item.Id,
		&item.RecordingState,
		&item.IsCompleted,
		&item.InfraredDeviceId,
		&item.Brand,
		&item.Model,
		&item.InfraredDeviceTypeId,
		&item.DeviceTypeName,
		&item.CreatedAt,
	)
	return item, err
}

func ScanPgxInfraredRecordSessionListItems(rows pgx.Rows) ([]domainmodels.InfraredRecordSessionListItem, error) {
	items := make([]domainmodels.InfraredRecordSessionListItem, 0)
	for rows.Next() {
		item, err := ScanPgxInfraredRecordSessionListItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
```

The scan order MUST exactly match the `SELECT` column order built in Step 4 below.

- [ ] **Step 4: Add the query builder**

In `backend/internal/infrastructure/repository/infrared_record_session/postgres_query.go`, add:

```go
func (p *postgresImpl) queryReadByFilter(
	recordingState *string,
	infraredDeviceTypeId *uuid.UUID,
	createdAtStart *time.Time,
	createdAtEnd *time.Time,
	page int,
	limit int,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(
		"infrared_record_session.id",
		"infrared_record_session.recording_state",
		"infrared_record_session.is_completed",
		"infrared_device.id",
		"infrared_device.brand",
		"infrared_device.model",
		"infrared_device_type.id",
		"infrared_device_type.name",
		"infrared_record_session.created_at",
	).
		From("infrared_record_session").
		Join("infrared_device ON infrared_device.id = infrared_record_session.infrared_device_id").
		Join("infrared_device_type ON infrared_device_type.id = infrared_device.infrared_device_type_id")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("infrared_record_session").
		Join("infrared_device ON infrared_device.id = infrared_record_session.infrared_device_id")

	baseQ = baseQ.Where("infrared_record_session.deleted_at IS NULL")
	totalQ = totalQ.Where("infrared_record_session.deleted_at IS NULL")

	baseQ, totalQ = applyInfraredRecordSessionFilters(baseQ, totalQ, recordingState, infraredDeviceTypeId, createdAtStart, createdAtEnd)

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("infrared_record_session.created_at DESC", "infrared_record_session.id ASC").
		Limit(uint64(limit)).
		Offset(uint64((page - 1) * limit)).
		ToSql()
	return
}

func applyInfraredRecordSessionFilters(
	baseQ squirrel.SelectBuilder,
	totalQ squirrel.SelectBuilder,
	recordingState *string,
	infraredDeviceTypeId *uuid.UUID,
	createdAtStart *time.Time,
	createdAtEnd *time.Time,
) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	if recordingState != nil {
		condition := squirrel.Eq{"infrared_record_session.recording_state": *recordingState}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if infraredDeviceTypeId != nil {
		condition := squirrel.Eq{"infrared_device_type.id": *infraredDeviceTypeId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if createdAtStart != nil {
		condition := squirrel.GtOrEq{"infrared_record_session.created_at": *createdAtStart}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if createdAtEnd != nil {
		condition := squirrel.LtOrEq{"infrared_record_session.created_at": *createdAtEnd}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	return baseQ, totalQ
}
```

Note `totalQ` needs the `infrared_device_type` join too if filtering by
`infraredDeviceTypeId` — but since that filter condition references
`infrared_device_type.id`, and `totalQ` above only joins
`infrared_device` (not `infrared_device_type`), add the missing join.
Fix `totalQ`'s construction to also include:
```go
	totalQ := p.SqrD.Select("COUNT(*)").
		From("infrared_record_session").
		Join("infrared_device ON infrared_device.id = infrared_record_session.infrared_device_id").
		Join("infrared_device_type ON infrared_device_type.id = infrared_device.infrared_device_type_id")
```
(Both joins are always applied, matching `baseQ` — simpler and still
correct even though `totalQ` doesn't select any device/device-type
columns, since squirrel doesn't complain about an unused join.)

Add `"time"` and `"github.com/Masterminds/squirrel"` to this file's
imports if not already present — check first (squirrel is likely already
imported for the existing query builders in this file).

- [ ] **Step 5: Add the repository method**

In `backend/internal/infrastructure/repository/infrared_record_session/postgres.go`, add right after `ReadActiveByNodeId`:

```go
func (p *postgresImpl) ReadByFilter(
	ctx context.Context,
	recordingState *string,
	infraredDeviceTypeId *uuid.UUID,
	createdAtStart *time.Time,
	createdAtEnd *time.Time,
	page int,
	limit int,
) (items []domainmodels.InfraredRecordSessionListItem, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByFilter(recordingState, infraredDeviceTypeId, createdAtStart, createdAtEnd, page, limit)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read infrared_record_session query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count infrared_record_session", err)
	}
	if total == 0 {
		return []domainmodels.InfraredRecordSessionListItem{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read infrared_record_session", err)
	}
	defer rows.Close()

	items, err = infrastructurerepositoryshared.ScanPgxInfraredRecordSessionListItems(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan infrared_record_session", err)
	}

	return items, total, nil
}
```

Add `"time"` to this file's imports if not already present — check first.

- [ ] **Step 6: Write a query-builder test**

Check `backend/internal/infrastructure/repository/infrared_record_session/postgres_test.go`'s existing style first (it was created in an earlier plan for `queryMarkChecksumClarificationUsedById` — match that exact assertion pattern: build the query, assert the generated SQL string contains expected fragments). Add:

```go
func TestQueryReadByFilterJoinsDeviceAndDeviceType(t *testing.T) {
	sqrDollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	p := &postgresImpl{}
	p.SqrD = &sqrDollar

	_, _, query, _, err := p.queryReadByFilter(nil, nil, nil, nil, 1, 10)
	if err != nil {
		t.Fatalf("queryReadByFilter() error = %v, want nil", err)
	}
	if !strings.Contains(query, "infrared_device_type.name") {
		t.Fatalf("query = %q, want it to select the device type name", query)
	}
	if !strings.Contains(query, "deleted_at IS NULL") {
		t.Fatalf("query = %q, want it to exclude soft-deleted sessions", query)
	}
}

func TestQueryReadByFilterAppliesRecordingStateFilter(t *testing.T) {
	sqrDollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	p := &postgresImpl{}
	p.SqrD = &sqrDollar

	state := "RECORDING"
	_, _, query, args, err := p.queryReadByFilter(&state, nil, nil, nil, 1, 10)
	if err != nil {
		t.Fatalf("queryReadByFilter() error = %v, want nil", err)
	}
	if !strings.Contains(query, "recording_state") {
		t.Fatalf("query = %q, want a recording_state condition", query)
	}
	found := false
	for _, a := range args {
		if a == state {
			found = true
		}
	}
	if !found {
		t.Fatalf("args = %v, want %q among them", args, state)
	}
}
```

- [ ] **Step 7: Run tests and build**

Run: `cd backend && go build ./... && go vet ./... && go test ./internal/infrastructure/repository/infrared_record_session/... -v`
Expected: build/vet clean, all tests `PASS`.

- [ ] **Step 8: Commit**

```bash
cd backend
git add internal/domain/models/infrared.go \
        internal/domain/contracts/repository/infrared_record_session.go \
        internal/infrastructure/repository/infrared_record_session/postgres_query.go \
        internal/infrastructure/repository/infrared_record_session/postgres.go \
        internal/infrastructure/repository/infrared_record_session/postgres_test.go \
        internal/infrastructure/repository/shared/scan_pgx.go
git commit -m "infrared: add ReadByFilter for paginated record session listing"
```

---

### Task 2: Backend — usecase method, route, response type

**Files:**
- Modify: `backend/internal/domain/usecases/infrared/record_session_management.go` (add `ListByFilter` + `ListRecordSessionsRequest`)
- Modify: `backend/internal/application/infrared/record_session_management/usecase.go` (implement `ListByFilter`)
- Modify: `backend/internal/application/infrared/record_session_management/usecase_test.go` (add tests; update `fakeSessionRepository` for interface conformance)
- Modify: `backend/internal/presentation/http/response/infrared.go` (add `InfraredRecordSessionListItemResponse` + mapper)
- Modify: `backend/internal/presentation/http/handler/infrared/handler.go` (add `RecordSessionGetList`)
- Modify: `backend/internal/presentation/http/route/route.go` (register the route + interface method)

**Interfaces:**
- Consumes: Task 1's `domainmodels.InfraredRecordSessionListItem` and repository `ReadByFilter`.
- Produces (for the frontend to consume): `GET /v1/infrared/record-sessions?recording_state=&infrared_device_type_id=&created_at_start=&created_at_end=&page=&limit=` → `presentationhttpresponse.PageDataResponse[InfraredRecordSessionListItemResponse]`.

- [ ] **Step 1: Add the usecase interface method + request type**

In `backend/internal/domain/usecases/infrared/record_session_management.go`, add to the `RecordSessionManagement` interface, right after `GetById`:

```go
	ListByFilter(ctx context.Context, request ListRecordSessionsRequest) (items []domainmodels.InfraredRecordSessionListItem, total int, err error)
```

Add the request type near `StartRecordSessionRequest`:

```go
type ListRecordSessionsRequest struct {
	RecordingState       *string
	InfraredDeviceTypeId *uuid.UUID
	CreatedAtStart       *time.Time
	CreatedAtEnd         *time.Time
	Page                 int
	Limit                int
}
```

Add `"time"` to this file's imports if not already present — check first.

- [ ] **Step 2: Write the failing usecase test**

In `backend/internal/application/infrared/record_session_management/usecase_test.go`, first extend `fakeSessionRepository` (defined near line 22) with a `ReadByFilter` stub — every existing repository interface implementer needs this method now for the package to compile:

```go
func (f *fakeSessionRepository) ReadByFilter(_ context.Context, _ *string, _ *uuid.UUID, _ *time.Time, _ *time.Time, page int, limit int) ([]domainmodels.InfraredRecordSessionListItem, int, error) {
	return f.listByFilterResult, f.listByFilterTotal, f.listByFilterErr
}
```

Add the backing fields (`listByFilterResult []domainmodels.InfraredRecordSessionListItem`, `listByFilterTotal int`, `listByFilterErr error`) to the `fakeSessionRepository` struct definition.

Then add the test:

```go
func TestListByFilterDelegatesToRepository(t *testing.T) {
	sessionRepo := &fakeSessionRepository{
		listByFilterResult: []domainmodels.InfraredRecordSessionListItem{
			{Id: uuid.New(), RecordingState: domainmodels.InfraredRecordingStateCompleted, Brand: "Daikin", Model: "FTWX35AXV1"},
		},
		listByFilterTotal: 1,
	}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	)

	items, total, err := impl.ListByFilter(context.Background(), domainusecasesinfrared.ListRecordSessionsRequest{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListByFilter() error = %v, want nil", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("ListByFilter() = %v, %d, want 1 item, total 1", items, total)
	}
	if items[0].Brand != "Daikin" {
		t.Fatalf("items[0].Brand = %q, want Daikin", items[0].Brand)
	}
}
```

Add `domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"` to this test file's imports if not already present — check first.

- [ ] **Step 3: Run test to verify it fails**

Run: `cd backend && go test ./internal/application/infrared/record_session_management/... -run TestListByFilter -v`
Expected: compile error — `ListByFilter` doesn't exist on `*usecase` yet.

- [ ] **Step 4: Implement `ListByFilter`**

In `backend/internal/application/infrared/record_session_management/usecase.go`, add (near `GetById`, currently at line 215):

```go
func (u *usecase) ListByFilter(ctx context.Context, request domainusecasesinfrared.ListRecordSessionsRequest) ([]domainmodels.InfraredRecordSessionListItem, int, error) {
	return u.session.ReadByFilter(ctx, request.RecordingState, request.InfraredDeviceTypeId, request.CreatedAtStart, request.CreatedAtEnd, request.Page, request.Limit)
}
```

No validation needed beyond what the HTTP layer already does (Task 3's
`RecordSessionGetList` validates `recording_state` against the known
constants before this is ever called) — this method is a thin pass-through,
consistent with `GetById`/`GetCoderBySessionId` immediately above/below it.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && go build ./... && go test ./internal/application/infrared/record_session_management/... -v`
Expected: `PASS` across the whole package (confirms the `fakeSessionRepository` conformance fix didn't break any existing test).

- [ ] **Step 6: Add the response type**

In `backend/internal/presentation/http/response/infrared.go`, add right after `InfraredRecordSessions` (currently ending at line 76):

```go
type InfraredRecordSessionListItemResponse struct {
	Id             string    `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	RecordingState string    `json:"recording_state" example:"RECORDING"`
	IsCompleted    bool      `json:"is_completed" example:"false"`
	Brand          string    `json:"brand" example:"Daikin"`
	Model          string    `json:"model" example:"FTWX35AXV1"`
	DeviceTypeName string    `json:"device_type_name" example:"Air Conditioner"`
	CreatedAt      time.Time `json:"created_at" example:"2026-06-15T09:30:00Z"`
}

func InfraredRecordSessionListItems(models []domainmodels.InfraredRecordSessionListItem) []InfraredRecordSessionListItemResponse {
	responses := make([]InfraredRecordSessionListItemResponse, len(models))
	for i, model := range models {
		responses[i] = InfraredRecordSessionListItemResponse{
			Id:             UUIDString(model.Id),
			RecordingState: model.RecordingState,
			IsCompleted:    model.IsCompleted,
			Brand:          model.Brand,
			Model:          model.Model,
			DeviceTypeName: model.DeviceTypeName,
			CreatedAt:      model.CreatedAt,
		}
	}
	return responses
}
```

- [ ] **Step 7: Add the handler**

Check `backend/internal/presentation/http/utils/query.go`'s exact
existing helper names before writing this (already confirmed present:
`PageArgs`, `QueryUUID`, `QueryString`, `QueryTime`). In
`backend/internal/presentation/http/handler/infrared/handler.go`, add
right after `RecordSessionPost` (before `RecordSessionGetById`):

```go
// RecordSessionGetList godoc
//
// @Summary Infrared Record Session List
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param recording_state query string false "recording state"
// @Param infrared_device_type_id query string false "device type id"
// @Param created_at_start query string false "created at start (RFC3339)"
// @Param created_at_end query string false "created at end (RFC3339)"
// @Param page query int false "page"
// @Param limit query int false "limit"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.InfraredRecordSessionListItemResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-sessions [get]
func (h *handler) RecordSessionGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	recordingState := presentationhttputils.QueryString(c, "recording_state")
	if recordingState != nil && !domainmodels.IsValidInfraredRecordingState(*recordingState) {
		return presentationhttputils.Error(c, domainmodels.NewError("recording_state must be a known recording state", domainmodels.ErrTypeValidation, nil))
	}
	infraredDeviceTypeId, err := presentationhttputils.QueryUUID(c, "infrared_device_type_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	createdAtStart, err := presentationhttputils.QueryTime(c, "created_at_start")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	createdAtEnd, err := presentationhttputils.QueryTime(c, "created_at_end")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	items, total, err := h.recordSessionUseCase.ListByFilter(c.Request().Context(), domainusecasesinfrared.ListRecordSessionsRequest{
		RecordingState:       recordingState,
		InfraredDeviceTypeId: infraredDeviceTypeId,
		CreatedAtStart:       createdAtStart,
		CreatedAtEnd:         createdAtEnd,
		Page:                 page.Page,
		Limit:                page.Limit,
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.InfraredRecordSessionListItemResponse]{
		Data: presentationhttpresponse.InfraredRecordSessionListItems(items),
		Page: presentationhttputils.PageResponse(page, total),
	})
}
```

This references a new helper `domainmodels.IsValidInfraredRecordingState`
— add it to `backend/internal/domain/models/infrared.go` right after the
`InfraredRecordingState*` constants (currently ending at line 42):

```go
var validInfraredRecordingStates = map[string]struct{}{
	InfraredRecordingStateDraft:               {},
	InfraredRecordingStateCasesGenerating:     {},
	InfraredRecordingStateRecording:           {},
	InfraredRecordingStateAnalyzing:           {},
	InfraredRecordingStateFunctionGenerating:  {},
	InfraredRecordingStateTestCasesGenerating: {},
	InfraredRecordingStateTesting:             {},
	InfraredRecordingStateCompleted:           {},
	InfraredRecordingStateFailed:              {},
}

func IsValidInfraredRecordingState(state string) bool {
	_, ok := validInfraredRecordingStates[state]
	return ok
}
```

- [ ] **Step 8: Register the route**

In `backend/internal/presentation/http/route/route.go`:
- Add `RecordSessionGetList(c *echo.Context) error` to the route interface (find where `RecordSessionPost`/`RecordSessionGetById` etc. are declared in that interface, add alongside them).
- Register, right after the existing `v1.POST("/infrared/record-sessions", ...)` line (currently line 379):
  ```go
  v1.GET("/infrared/record-sessions", handler.RecordSessionGetList, permission("infrared_record_session:get"))
  ```

- [ ] **Step 8b: Add `GetDeviceById` — the detail page needs to resolve a device's brand/model/device-type name, and no existing endpoint exposes a single device by id today**

`InfraredRecordSessionResponse` (from `GetById`) only carries
`infrared_device_id`, not brand/model/device-type-name — those live on
`InfraredDevice`, and no endpoint currently exposes a single device by
id (only `DELETE /infrared/devices/{id}` exists). `ReferenceManagement`
already has the `device` repository injected (used by `DeleteDeviceById`)
and `InfraredDevice.ReadById` already exists on that repository — this
is a small, additive read endpoint reusing both.

In `backend/internal/domain/usecases/infrared/reference_management.go`,
add to the `ReferenceManagement` interface, right after
`DeleteDeviceById`:

```go
	GetDeviceById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredDevice, error)
```

In `backend/internal/application/infrared/reference_management/usecase.go`,
add right after `DeleteDeviceById`:

```go
func (u *usecase) GetDeviceById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredDevice, error) {
	return u.device.ReadById(ctx, id)
}
```

In `backend/internal/presentation/http/response/infrared.go`, add a
response type (no existing single-device response type — only
`InfraredDeviceType` has one today):

```go
type InfraredDeviceResponse struct {
	Id                   string `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	InfraredDeviceTypeId string `json:"infrared_device_type_id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	Brand                string `json:"brand" example:"Daikin"`
	Model                string `json:"model" example:"FTWX35AXV1"`
	AuditResponse
}

func InfraredDevice(model domainmodels.InfraredDevice) InfraredDeviceResponse {
	return InfraredDeviceResponse{
		Id:                   UUIDString(model.Id),
		InfraredDeviceTypeId: UUIDString(model.InfraredDeviceTypeId),
		Brand:                model.Brand,
		Model:                model.Model,
		AuditResponse:        Audit(model.CreatedAt, model.UpdatedAt, model.DeletedAt, model.CreatedBy, model.UpdatedBy, model.DeletedBy),
	}
}
```

In `backend/internal/presentation/http/handler/infrared/handler.go`, add
right after `DeviceDelete`:

```go
// DeviceGetById godoc
//
// @Summary Infrared Device Get By ID
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.InfraredDeviceResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/devices/{id} [get]
func (h *handler) DeviceGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	device, err := h.referenceUseCase.GetDeviceById(c.Request().Context(), id)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if device == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("device"))
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.InfraredDevice(*device))
}
```

In `backend/internal/presentation/http/route/route.go`, add
`DeviceGetById(c *echo.Context) error` to the route interface, and
register right after the existing
`v1.DELETE("/infrared/devices/:id", ...)` line:

```go
v1.GET("/infrared/devices/:id", handler.DeviceGetById, permission("infrared_reference:get"))
```

(Reuses the existing `infrared_reference:get` permission — the same one
that already gates `DeviceTypeGetList`/`StateGetList`, since a device
read is reference-data access, not session-lifecycle access.)

Extend `backend/internal/application/infrared/reference_management/usecase_test.go`
with a test confirming the pass-through:

```go
func TestGetDeviceByIdDelegatesToRepository(t *testing.T) {
	deviceId := uuid.New()
	deviceRepo := &fakeDeviceRepository{getByIdResult: &domainmodels.InfraredDevice{Id: deviceId, Brand: "Daikin", Model: "FTWX35AXV1"}}
	impl := NewUsecaseImpl(&fakeDeviceTypeRepository{}, deviceRepo, &fakeStateRepository{}, &fakeDefinitionRepository{})

	device, err := impl.GetDeviceById(context.Background(), deviceId)
	if err != nil {
		t.Fatalf("GetDeviceById() error = %v, want nil", err)
	}
	if device == nil || device.Brand != "Daikin" {
		t.Fatalf("GetDeviceById() = %v, want Brand=Daikin", device)
	}
}
```

Check `reference_management/usecase_test.go`'s existing
`fakeDeviceRepository` fake first — if it doesn't already have a
`getByIdResult` field/`ReadById` method, add both (following this
package's existing fake conventions); if a `ReadById` stub already
exists for a different reason, reuse it rather than adding a duplicate.

- [ ] **Step 9: Build, vet, test the whole backend**

Run: `cd backend && go build ./... && go vet ./... && go test ./...`
Expected: clean build/vet, every package `ok`.

- [ ] **Step 10: Regenerate swagger docs**

Check this repo's swagger generation command (likely `swag init` or a
Makefile target — check `backend/Makefile` or `backend/docs/swagger` for
the existing generation convention before running). Run it if one
exists; if none is discoverable, skip this step and note it in the
task's completion report rather than guessing at a command.

- [ ] **Step 11: Commit**

```bash
cd backend
git add internal/domain/usecases/infrared/record_session_management.go \
        internal/domain/models/infrared.go \
        internal/application/infrared/record_session_management/usecase.go \
        internal/application/infrared/record_session_management/usecase_test.go \
        internal/presentation/http/response/infrared.go \
        internal/presentation/http/handler/infrared/handler.go \
        internal/presentation/http/route/route.go \
        docs/swagger/ 2>/dev/null || true
git commit -m "infrared: add GET /infrared/record-sessions list endpoint"
```

---

### Task 3: Frontend — API client for record session read/mutate/delete endpoints

**Files:**
- Modify: `frontend/src/lib/api/infrared.ts`

**Interfaces:**
- Consumes: Task 2's new endpoint, plus every already-shipped record-session endpoint (Start, GetById, ListCases, coder, test cases, accept/discard/retry, deletes).
- Produces (for Tasks 4-5 to consume): the full typed client surface listed below.

- [ ] **Step 1: Read the real handler/response shapes for every endpoint before typing**

Before writing types, re-confirm exact JSON field names against
`backend/internal/presentation/http/response/infrared.go` (already read
in full during planning — types below mirror it exactly) and the route
list in `backend/internal/presentation/http/route/route.go`. Do not
guess field names.

- [ ] **Step 2: Add types and functions to `infrared.ts`**

Append to the existing file (after the existing `deleteInfraredState`):

```ts
export type InfraredRecordingState =
  | "DRAFT"
  | "CASES_GENERATING"
  | "RECORDING"
  | "ANALYZING"
  | "FUNCTION_GENERATING"
  | "TEST_CASES_GENERATING"
  | "TESTING"
  | "COMPLETED"
  | "FAILED";

export interface InfraredRecordSessionListItemResponse {
  id: string;
  recording_state: InfraredRecordingState;
  is_completed: boolean;
  brand: string;
  model: string;
  device_type_name: string;
  created_at: string;
}

export interface ListInfraredRecordSessionsQuery extends PageQuery {
  recording_state?: InfraredRecordingState;
  infrared_device_type_id?: string;
  created_at_start?: string;
  created_at_end?: string;
}

export async function listInfraredRecordSessions(
  query: ListInfraredRecordSessionsQuery = {},
): Promise<PageDataResponse<InfraredRecordSessionListItemResponse>> {
  return apiFetch(`/infrared/record-sessions${buildQuery(query)}`);
}

export interface InfraredRecordSessionResponse extends AuditFields {
  id: string;
  node_id: string;
  infrared_device_id: string;
  recording_state: InfraredRecordingState;
  current_record_case_id?: string;
  is_completed: boolean;
}

export async function getInfraredRecordSession(
  id: string,
): Promise<InfraredRecordSessionResponse> {
  return apiFetch(`/infrared/record-sessions/${id}`);
}

export interface InfraredDeviceResponse extends AuditFields {
  id: string;
  infrared_device_type_id: string;
  brand: string;
  model: string;
}

export async function getInfraredDevice(
  id: string,
): Promise<InfraredDeviceResponse> {
  return apiFetch(`/infrared/devices/${id}`);
}

export interface StartRecordSessionDefinition {
  infrared_state_id: string;
  options?: string[];
  minimum?: number;
  maximum?: number;
  step?: number;
}

export interface StartRecordSessionRequest {
  node_id: string;
  infrared_device_type_id: string;
  brand: string;
  model: string;
  definitions: StartRecordSessionDefinition[];
}

export async function startInfraredRecordSession(
  request: StartRecordSessionRequest,
): Promise<IdResponse> {
  return apiFetch("/infrared/record-sessions", {
    method: "POST",
    body: request,
  });
}

export interface InfraredStateDeviceRecordStateResponse {
  id: string;
  infrared_state_id: string;
  state_value: string;
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export interface InfraredStateDeviceRecordRawResponse {
  id: string;
  status: "CAPTURED" | "ACCEPTED" | "DISCARDED";
  discarded_reason?: string;
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export interface InfraredStateDeviceRecordCaseResponse {
  id: string;
  step: number;
  description: string;
  status: "PENDING" | "ACTIVE" | "ACCEPTED";
  states: InfraredStateDeviceRecordStateResponse[];
  raw: InfraredStateDeviceRecordRawResponse[];
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export async function listInfraredRecordSessionCases(
  sessionId: string,
): Promise<InfraredStateDeviceRecordCaseResponse[]> {
  return apiFetch(`/infrared/record-sessions/${sessionId}/cases`);
}

export async function acceptInfraredRecordRaw(
  caseId: string,
  rawId: string,
): Promise<void> {
  return apiFetch(`/infrared/record-cases/${caseId}/raw/${rawId}/accept`, {
    method: "POST",
  });
}

export async function discardInfraredRecordRaw(
  caseId: string,
  rawId: string,
  reason: string,
): Promise<void> {
  return apiFetch(`/infrared/record-cases/${caseId}/raw/${rawId}/discard`, {
    method: "POST",
    body: { reason },
  });
}

export async function retryInfraredRecordCase(caseId: string): Promise<void> {
  return apiFetch(`/infrared/record-cases/${caseId}/retry`, {
    method: "POST",
  });
}

export async function setInfraredRecordSessionCurrentCase(
  sessionId: string,
  caseId: string,
): Promise<void> {
  return apiFetch(
    `/infrared/record-sessions/${sessionId}/cases/${caseId}/current`,
    { method: "POST" },
  );
}

export interface InfraredStateCoderResponse {
  id: string;
  encoder_source: string;
  decoder_source: string;
  summary_readme: string;
  detail_readme: string;
  status: "UNVERIFIED" | "ACTIVE" | "SUPERSEDED";
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export async function getInfraredRecordSessionCoder(
  sessionId: string,
): Promise<InfraredStateCoderResponse> {
  return apiFetch(`/infrared/record-sessions/${sessionId}/coder`);
}

export interface InfraredTestCaseStateResponse {
  id: string;
  infrared_state_id: string;
  state_value: string;
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export interface InfraredTestCaseResponse {
  id: string;
  step: number;
  description: string;
  status: "PENDING" | "PASSED" | "FAILED";
  states: InfraredTestCaseStateResponse[];
  created_at: string;
  deleted_at?: string;
  deleted_by?: string;
}

export async function listInfraredRecordSessionTestCases(
  sessionId: string,
): Promise<InfraredTestCaseResponse[]> {
  return apiFetch(`/infrared/record-sessions/${sessionId}/test-cases`);
}

export async function transmitInfraredTestCase(
  testCaseId: string,
): Promise<void> {
  return apiFetch(`/infrared/test-cases/${testCaseId}/transmit`, {
    method: "POST",
  });
}

export async function recordInfraredTestCaseResult(
  testCaseId: string,
  passed: boolean,
): Promise<void> {
  return apiFetch(`/infrared/test-cases/${testCaseId}/result`, {
    method: "POST",
    body: { passed },
  });
}

export async function deleteInfraredRecordSession(id: string): Promise<void> {
  return apiFetch(`/infrared/record-sessions/${id}`, { method: "DELETE" });
}

export async function deleteInfraredRecordCase(caseId: string): Promise<void> {
  return apiFetch(`/infrared/record-cases/${caseId}`, { method: "DELETE" });
}

export async function deleteInfraredRecordState(id: string): Promise<void> {
  return apiFetch(`/infrared/record-states/${id}`, { method: "DELETE" });
}

export async function deleteInfraredRecordRaw(
  caseId: string,
  rawId: string,
): Promise<void> {
  return apiFetch(`/infrared/record-cases/${caseId}/raw/${rawId}`, {
    method: "DELETE",
  });
}

export async function deleteInfraredStateCoder(id: string): Promise<void> {
  return apiFetch(`/infrared/state-coders/${id}`, { method: "DELETE" });
}

export async function deleteInfraredTestCase(id: string): Promise<void> {
  return apiFetch(`/infrared/test-cases/${id}`, { method: "DELETE" });
}

export async function deleteInfraredTestCaseState(
  testCaseId: string,
  stateId: string,
): Promise<void> {
  return apiFetch(`/infrared/test-cases/${testCaseId}/states/${stateId}`, {
    method: "DELETE",
  });
}
```

Add `PageDataResponse, PageQuery, IdResponse, AuditFields` to this
file's existing import from `@/lib/api/types` if not already all
present — check the current import line first (`AuditFields` and
`IdResponse` are already imported per the file's existing content; add
`PageDataResponse`/`PageQuery` alongside them). Add `buildQuery` to the
existing `@/lib/api/client` import alongside `apiFetch` if not already
present.

Note: `RecordCaseRawDiscard`'s request body field is `reason` per
`presentationhttprequest.DiscardRawRequest` — verify this exact field
name against `backend/internal/presentation/http/request/infrared.go`
before finalizing (read that file if not already confirmed in an
earlier exploration pass).

- [ ] **Step 3: Verify the frontend builds and lints**

Since `frontend/node_modules`/`.next` are root-owned in the main
checkout, run this verification in the `.worktrees/mate-things-frontend-consistency`
worktree per this project's established workflow: copy the changed
file in, then run `npm run lint` and `npx tsc --noEmit` (or this
project's equivalent type-check command — check `package.json` scripts
first) from that worktree. Restore the worktree afterward
(`git checkout -- . && git clean -fd`).

- [ ] **Step 4: Commit**

```bash
cd frontend
git add src/lib/api/infrared.ts
git commit -m "infrared: add record session list/detail/mutation API client"
```

---

### Task 4: Frontend — list page (`record/page.tsx`)

**Files:**
- Modify: `frontend/src/app/(authenticated)/apps/infrared/record/page.tsx` (replace the stub)
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/_components/RecordSessionFilters.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/_components/RecordSessionsTable.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/_lib/filters.ts`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/_lib/status.ts`

**Interfaces:**
- Consumes: Task 3's `listInfraredRecordSessions`, existing `listInfraredDeviceTypes` (already in `infrared.ts`).
- Produces: nothing consumed by later tasks in this plan (Task 5 is a separate route).

- [ ] **Step 1: Write `_lib/status.ts`**

```ts
import type { StatusVariant } from "@/components/ui/status-badge";
import type { InfraredRecordingState } from "@/lib/api/infrared";

export const RECORDING_STATE_LABELS: Record<
  InfraredRecordingState,
  { label: string; variant: StatusVariant }
> = {
  DRAFT: { label: "Draft", variant: "neutral" },
  CASES_GENERATING: { label: "Building cases", variant: "info" },
  RECORDING: { label: "Recording", variant: "info" },
  ANALYZING: { label: "Analyzing", variant: "info" },
  FUNCTION_GENERATING: { label: "Generating encoder", variant: "info" },
  TEST_CASES_GENERATING: { label: "Building tests", variant: "info" },
  TESTING: { label: "Testing", variant: "info" },
  COMPLETED: { label: "Completed", variant: "success" },
  FAILED: { label: "Failed", variant: "critical" },
};
```

- [ ] **Step 2: Write `_lib/filters.ts`**

Mirror `action-history/_lib/filters.ts`'s exact shape:

```ts
import type { InfraredRecordingState } from "@/lib/api/infrared";

import { RECORDING_STATE_LABELS } from "./status";

export interface RecordSessionFilters {
  recordingState?: InfraredRecordingState;
  deviceTypeId?: string;
  start?: string;
  end?: string;
}

export interface ParsedRecordSessionFilters {
  filters: RecordSessionFilters;
  error?: string;
}

type RawRecordSessionParams = Record<string, string | string[] | undefined>;

const VALID_STATES = Object.keys(
  RECORDING_STATE_LABELS,
) as InfraredRecordingState[];

function first(value: string | string[] | undefined): string {
  return (Array.isArray(value) ? value[0] : (value ?? "")).trim();
}

function iso(value: string, field: string): { value?: string; error?: string } {
  if (!value) return {};
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return { error: `${field} must be a valid date and time.` };
  }
  return { value: date.toISOString() };
}

export function parseRecordSessionFilters(
  raw: RawRecordSessionParams,
): ParsedRecordSessionFilters {
  const start = iso(first(raw.start), "Start");
  const end = iso(first(raw.end), "End");

  let error = start.error ?? end.error;
  if (
    !error &&
    start.value &&
    end.value &&
    new Date(start.value).getTime() > new Date(end.value).getTime()
  ) {
    error = "End must be after start.";
  }

  const stateRaw = first(raw.recording_state).toUpperCase();
  const recordingState = VALID_STATES.includes(
    stateRaw as InfraredRecordingState,
  )
    ? (stateRaw as InfraredRecordingState)
    : undefined;
  if (!error && stateRaw && !recordingState) {
    error = "Select a valid status.";
  }

  return {
    filters: {
      start: start.value,
      end: end.value,
      recordingState,
      deviceTypeId: first(raw.infrared_device_type_id) || undefined,
    },
    error,
  };
}
```

- [ ] **Step 3: Write `RecordSessionFilters.tsx`**

Mirror `ActionLogFilters.tsx`'s structure (`FilterBar` + `TimeRangeFilter`
+ selects), but both selects here are plain (no combobox needed — device
type list is small and fully enumerable, unlike actions/nodes):

```tsx
import FilterBar from "@/components/collection/FilterBar";
import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import Select from "@/components/ui/select";
import type {
  InfraredDeviceTypeResponse,
  InfraredRecordingState,
} from "@/lib/api/infrared";

import { RECORDING_STATE_LABELS } from "../_lib/status";

const STATE_KEYS = Object.keys(
  RECORDING_STATE_LABELS,
) as InfraredRecordingState[];

export default function RecordSessionFilters({
  deviceTypes,
  start,
  end,
  recordingState,
  deviceTypeId,
}: {
  deviceTypes: readonly InfraredDeviceTypeResponse[];
  start?: string;
  end?: string;
  recordingState?: InfraredRecordingState;
  deviceTypeId?: string;
}) {
  return (
    <FilterBar clearHref="/apps/infrared/record">
      <TimeRangeFilter start={start} end={end} />
      <div>
        <label
          className="text-foreground/70 mb-1.5 block text-xs font-semibold"
          htmlFor="record-session-status"
        >
          Status
        </label>
        <Select
          id="record-session-status"
          name="recording_state"
          defaultValue={recordingState ?? ""}
        >
          <option value="">Any status</option>
          {STATE_KEYS.map((state) => (
            <option key={state} value={state}>
              {RECORDING_STATE_LABELS[state].label}
            </option>
          ))}
        </Select>
      </div>
      <div>
        <label
          className="text-foreground/70 mb-1.5 block text-xs font-semibold"
          htmlFor="record-session-device-type"
        >
          Device type
        </label>
        <Select
          id="record-session-device-type"
          name="infrared_device_type_id"
          defaultValue={deviceTypeId ?? ""}
        >
          <option value="">Any device type</option>
          {deviceTypes.map((dt) => (
            <option key={dt.id} value={dt.id}>
              {dt.name}
            </option>
          ))}
        </Select>
      </div>
    </FilterBar>
  );
}
```

- [ ] **Step 4: Write `RecordSessionsTable.tsx`**

```tsx
import Link from "next/link";

import LocalDateTime from "@/components/ui/local-date-time";
import StatusBadge from "@/components/ui/status-badge";
import type { InfraredRecordSessionListItemResponse } from "@/lib/api/infrared";

import { RECORDING_STATE_LABELS } from "../_lib/status";

export default function RecordSessionsTable({
  records,
}: {
  records: readonly InfraredRecordSessionListItemResponse[];
}) {
  return (
    <div className="border-border overflow-x-auto rounded-2xl border">
      <table className="w-full text-left text-sm">
        <thead className="bg-muted text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          <tr>
            <th scope="col" className="px-4 py-3">
              Device
            </th>
            <th scope="col" className="px-4 py-3">
              Device type
            </th>
            <th scope="col" className="px-4 py-3">
              Status
            </th>
            <th scope="col" className="px-4 py-3">
              Created
            </th>
          </tr>
        </thead>
        <tbody className="divide-border divide-y">
          {records.map((record) => {
            const status = RECORDING_STATE_LABELS[record.recording_state];
            return (
              <tr key={record.id} className="hover:bg-highlight/20 transition-colors">
                <td className="px-4 py-3">
                  <Link
                    href={`/apps/infrared/record/${record.id}`}
                    className="text-primary font-semibold"
                  >
                    {record.brand} {record.model}
                  </Link>
                </td>
                <td className="px-4 py-3">{record.device_type_name}</td>
                <td className="px-4 py-3">
                  <StatusBadge variant={status.variant}>
                    {status.label}
                  </StatusBadge>
                </td>
                <td className="px-4 py-3">
                  <LocalDateTime value={record.created_at} />
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
```

- [ ] **Step 5: Rewrite `page.tsx`**

Mirror `action-history/page.tsx`'s structure exactly:

```tsx
import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import Pagination from "@/components/collection/Pagination";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import {
  listInfraredDeviceTypes,
  listInfraredRecordSessions,
} from "@/lib/api/infrared";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

import RecordSessionFilters from "./_components/RecordSessionFilters";
import RecordSessionsTable from "./_components/RecordSessionsTable";
import { parseRecordSessionFilters } from "./_lib/filters";

export const metadata: Metadata = { title: "Infrared Record — Mate Things" };
type RawSearchParams = Record<string, string | string[] | undefined>;

export default async function InfraredRecordPage({
  searchParams,
}: {
  searchParams: Promise<RawSearchParams>;
}) {
  const [raw, { permissions }] = await Promise.all([
    searchParams,
    requirePermission("infrared_record_session:get"),
  ]);
  const pageQuery = parsePageQuery(raw);
  const parsed = parseRecordSessionFilters(raw);
  const canStart = permissions.has("infrared_record_session:add");

  const deviceTypes = await listInfraredDeviceTypes();

  const result = parsed.error
    ? { data: [], page: { page: pageQuery.page, limit: pageQuery.limit, total_items: 0 } }
    : await listInfraredRecordSessions({
        page: pageQuery.page,
        limit: pageQuery.limit,
        recording_state: parsed.filters.recordingState,
        infrared_device_type_id: parsed.filters.deviceTypeId,
        created_at_start: parsed.filters.start,
        created_at_end: parsed.filters.end,
      });

  if (!parsed.error) {
    const redirectTarget = getOutOfRangePageRedirect(
      "/apps/infrared/record",
      raw,
      result.page,
    );
    if (redirectTarget) {
      redirect(redirectTarget);
    }
  }

  const records = result.data;

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Record"
        description="Identify the device control by capturing its remote's infrared signals."
        actions={
          canStart ? (
            <Link
              href="/apps/infrared/record/new"
              className="bg-primary text-surface inline-flex min-h-11 items-center rounded-xl px-6 text-sm font-semibold"
            >
              Record New Device
            </Link>
          ) : undefined
        }
      />
      <RecordSessionFilters
        deviceTypes={deviceTypes}
        start={parsed.filters.start}
        end={parsed.filters.end}
        recordingState={parsed.filters.recordingState}
        deviceTypeId={parsed.filters.deviceTypeId}
      />
      {parsed.error ? (
        <EmptyState title="Check the active filters" description={parsed.error} />
      ) : (
        <>
          <div className="flex flex-wrap justify-between gap-3 text-sm">
            <p>
              <strong>{result.page.total_items}</strong> sessions
            </p>
          </div>
          {records.length ? (
            <RecordSessionsTable records={records} />
          ) : (
            <EmptyState
              title="No record sessions yet"
              description="Start recording a device to teach the Infrared app how to control it."
            />
          )}
          <div className="border-border border-t pt-5">
            <Pagination page={result.page} pathname="/apps/infrared/record" searchParams={raw} />
          </div>
        </>
      )}
    </main>
  );
}
```

Note: the `Link` in `PageHeader`'s `actions` points at
`/apps/infrared/record/new`, which does not exist until Spec B ships —
per the design doc, this is intentional (Spec B's own plan is
responsible for the route existing; committing this now with the link
present is fine since Spec B is the very next planned unit of work, not
speculative future work).

Check `PageHeader`'s exact `actions` prop type before using a raw
`<Link>` as its child — if it expects a specific button-shaped
component, use that instead of a bare anchor-styled `Link` (read
`frontend/src/components/ui/page-header.tsx` first to confirm).

- [ ] **Step 6: Verify — build, lint, manual check**

In the `.worktrees/mate-things-frontend-consistency` worktree: copy in
the new/changed files, run `npm run lint`, `npx prettier --check .` (or
this project's exact commands per `package.json`), and `npm run build`.
Then start the dev server and visually confirm the list page renders
(with at least the empty state, since no sessions exist yet in a fresh
environment) at `/apps/infrared/record` — per this project's established
"test the golden path in a browser" requirement for UI changes. Use
`localhost`, not `127.0.0.1`, per this project's known
frontend-dev-server hydration quirk.

- [ ] **Step 7: Commit**

```bash
cd frontend
git add "src/app/(authenticated)/apps/infrared/record/page.tsx" \
        "src/app/(authenticated)/apps/infrared/record/_components/RecordSessionFilters.tsx" \
        "src/app/(authenticated)/apps/infrared/record/_components/RecordSessionsTable.tsx" \
        "src/app/(authenticated)/apps/infrared/record/_lib/filters.ts" \
        "src/app/(authenticated)/apps/infrared/record/_lib/status.ts"
git commit -m "infrared: build the record sessions list page"
```

---

### Task 5: Frontend — detail page (`record/[id]/page.tsx`) with cases, coder, test cases

**Files:**
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/page.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_components/RecordSessionOverview.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_components/RecordCaseList.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_components/RecordCaseRawControls.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_components/RecordCoderPanel.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_components/RecordTestCaseList.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_components/DeleteRecordSessionDialog.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_components/CodeBlock.tsx`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_lib/actions.ts`
- Create: `frontend/src/app/(authenticated)/apps/infrared/record/[id]/_lib/state.ts`

**Interfaces:**
- Consumes: Task 3's full API client surface.
- Produces (for Spec B's plan to consume, once written): `RecordCaseList` and `RecordCaseRawControls` are the reusable building blocks Spec B's wizard reuses for its own in-flight recording/re-record steps — designed as standalone components taking `sessionId`/`cases`/permission booleans as props, no page-specific coupling.

- [ ] **Step 1: Write `_lib/state.ts`**

```ts
import type { ActionState } from "@/lib/forms/action-state";

export type RecordSessionActionState = ActionState<string>;

export const EMPTY_RECORD_SESSION_ACTION_STATE: RecordSessionActionState = {
  status: "idle",
};
```

- [ ] **Step 2: Write `_lib/actions.ts`**

Mirror the settings `_lib/actions.ts` pattern exactly (permission check,
try/catch → `ApiError` mapping, success message). Confirmed by reading
`use-refresh-after-action.ts`: this codebase's sole post-mutation-refresh
convention is the client-side `useRefreshAfterAction(state)` hook
(`router.refresh()` triggered by the caller watching `useActionState`'s
returned state) — `revalidatePath` is not used anywhere in this
codebase. Every calling component (`RecordCaseRawControls.tsx`,
`RecordCaseList.tsx`, `RecordTestCaseList.tsx`, the delete dialogs)
already calls `useRefreshAfterAction(state)` per their own steps in this
task, so these actions need no refresh mechanism of their own — no
`sessionId` extraction needed either, since it existed only to build a
`revalidatePath` target. One action per mutation, each taking whatever
ids it needs as hidden form fields:

```ts
"use server";

import { ApiError } from "@/lib/api/client";
import {
  acceptInfraredRecordRaw,
  deleteInfraredRecordCase,
  deleteInfraredRecordRaw,
  deleteInfraredRecordSession,
  deleteInfraredRecordState,
  deleteInfraredStateCoder,
  deleteInfraredTestCase,
  deleteInfraredTestCaseState,
  discardInfraredRecordRaw,
  recordInfraredTestCaseResult,
  retryInfraredRecordCase,
  transmitInfraredTestCase,
} from "@/lib/api/infrared";
import { requireSessionContext } from "@/lib/session";

import type { RecordSessionActionState } from "./state";

function permissionDenied(): RecordSessionActionState {
  return {
    status: "error",
    title: "Permission denied",
    message: "You do not have permission to make this change.",
  };
}

function actionError(error: unknown): RecordSessionActionState {
  if (error instanceof ApiError) {
    return { status: "error", title: error.title, message: error.message };
  }
  return {
    status: "error",
    title: "Something went wrong",
    message: "Please try again.",
  };
}

export async function acceptRawAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:set")) {
    return permissionDenied();
  }
  const caseId = String(formData.get("case_id") ?? "").trim();
  const rawId = String(formData.get("raw_id") ?? "").trim();
  try {
    await acceptInfraredRecordRaw(caseId, rawId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Capture accepted", message: "" };
}

export async function discardRawAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:set")) {
    return permissionDenied();
  }
  const caseId = String(formData.get("case_id") ?? "").trim();
  const rawId = String(formData.get("raw_id") ?? "").trim();
  const reason = String(formData.get("reason") ?? "").trim() || "user discarded";
  try {
    await discardInfraredRecordRaw(caseId, rawId, reason);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Capture discarded", message: "" };
}

export async function retryCaseAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:set")) {
    return permissionDenied();
  }
  const caseId = String(formData.get("case_id") ?? "").trim();
  try {
    await retryInfraredRecordCase(caseId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Case reset for retake", message: "" };
}

export async function transmitTestCaseAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:set")) {
    return permissionDenied();
  }
  const testCaseId = String(formData.get("test_case_id") ?? "").trim();
  try {
    await transmitInfraredTestCase(testCaseId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Transmitted", message: "" };
}

export async function recordTestCaseResultAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:set")) {
    return permissionDenied();
  }
  const testCaseId = String(formData.get("test_case_id") ?? "").trim();
  const passed = formData.get("passed") === "true";
  try {
    await recordInfraredTestCaseResult(testCaseId, passed);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Result recorded", message: "" };
}

export async function deleteRecordSessionAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const id = String(formData.get("id") ?? "").trim();
  try {
    await deleteInfraredRecordSession(id);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Session deleted", message: "" };
}

export async function deleteRecordCaseAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const caseId = String(formData.get("case_id") ?? "").trim();
  try {
    await deleteInfraredRecordCase(caseId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Case deleted", message: "" };
}

export async function deleteRecordStateAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const id = String(formData.get("id") ?? "").trim();
  try {
    await deleteInfraredRecordState(id);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "State deleted", message: "" };
}

export async function deleteRecordRawAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const caseId = String(formData.get("case_id") ?? "").trim();
  const rawId = String(formData.get("raw_id") ?? "").trim();
  try {
    await deleteInfraredRecordRaw(caseId, rawId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Capture deleted", message: "" };
}

export async function deleteStateCoderAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const id = String(formData.get("id") ?? "").trim();
  try {
    await deleteInfraredStateCoder(id);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Coder deleted", message: "" };
}

export async function deleteTestCaseAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const id = String(formData.get("id") ?? "").trim();
  try {
    await deleteInfraredTestCase(id);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Test case deleted", message: "" };
}

export async function deleteTestCaseStateAction(
  _previousState: RecordSessionActionState,
  formData: FormData,
): Promise<RecordSessionActionState> {
  const { permissions } = await requireSessionContext();
  if (!permissions.has("infrared_record_session:delete")) {
    return permissionDenied();
  }
  const testCaseId = String(formData.get("test_case_id") ?? "").trim();
  const stateId = String(formData.get("state_id") ?? "").trim();
  try {
    await deleteInfraredTestCaseState(testCaseId, stateId);
  } catch (error) {
    return actionError(error);
  }
  return { status: "success", title: "Test case state deleted", message: "" };
}
```

- [ ] **Step 3: Write `CodeBlock.tsx`**

A minimal read-only code viewer with copy button, for the encoder/
decoder source (no existing generic component for plain-text code in
this codebase — `JsonPayload` is JSON-specific). Model its copy-button
UI after `JsonPayload.tsx`'s existing copy button:

```tsx
"use client";

import { useState } from "react";
import { Check, Copy } from "lucide-react";

export default function CodeBlock({
  value,
  label,
}: {
  value: string;
  label: string;
}) {
  const [copied, setCopied] = useState(false);

  return (
    <div className="relative">
      <button
        type="button"
        aria-label={`Copy ${label}`}
        className="border-border bg-background/80 focus-visible:ring-focus absolute top-2 right-2 inline-flex min-h-9 items-center gap-2 rounded-xl border px-3 text-xs font-semibold backdrop-blur focus-visible:ring-2 focus-visible:outline-none"
        onClick={async () => {
          await navigator.clipboard.writeText(value);
          setCopied(true);
          window.setTimeout(() => setCopied(false), 1500);
        }}
      >
        {copied ? (
          <Check aria-hidden="true" className="h-4 w-4" />
        ) : (
          <Copy aria-hidden="true" className="h-4 w-4" />
        )}
        {copied ? "Copied" : "Copy"}
      </button>
      <pre className="border-border bg-muted overflow-x-auto rounded-2xl border p-4 text-xs">
        <code>{value}</code>
      </pre>
    </div>
  );
}
```

- [ ] **Step 4: Write `RecordCaseRawControls.tsx`**

The reusable accept/discard/retry unit for one case's raws — this is
the exact component Spec B's live wizard will also render for the active
recording step, so keep it free of any page-specific assumptions
(everything it needs comes in as props).

```tsx
"use client";

import { useActionState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { InfraredStateDeviceRecordRawResponse } from "@/lib/api/infrared";

import {
  acceptRawAction,
  discardRawAction,
} from "../_lib/actions";
import { EMPTY_RECORD_SESSION_ACTION_STATE } from "../_lib/state";

export default function RecordCaseRawControls({
  caseId,
  raw,
  canMutate,
}: {
  caseId: string;
  raw: InfraredStateDeviceRecordRawResponse;
  canMutate: boolean;
}) {
  const [acceptState, acceptAction, acceptPending] = useActionState(
    acceptRawAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  const [discardState, discardAction, discardPending] = useActionState(
    discardRawAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  useRefreshAfterAction(acceptState);
  useRefreshAfterAction(discardState);

  if (raw.status !== "CAPTURED" || !canMutate) {
    return null;
  }

  return (
    <div className="flex items-center gap-2">
      <form action={acceptAction}>
        <input type="hidden" name="case_id" value={caseId} />
        <input type="hidden" name="raw_id" value={raw.id} />
        <Button type="submit" disabled={acceptPending || discardPending}>
          {acceptPending ? "Accepting…" : "Accept"}
        </Button>
      </form>
      <form action={discardAction}>
        <input type="hidden" name="case_id" value={caseId} />
        <input type="hidden" name="raw_id" value={raw.id} />
        <Button
          type="submit"
          variant="secondary"
          disabled={acceptPending || discardPending}
        >
          {discardPending ? "Discarding…" : "Discard"}
        </Button>
      </form>
      <ActionMessage state={acceptState} />
      <ActionMessage state={discardState} />
    </div>
  );
}
```

Check `Button`'s exact variant prop names
(`frontend/src/components/ui/button.tsx`) before using `"secondary"` —
confirm it's a real variant, not guessed.

- [ ] **Step 5: Write `RecordCaseList.tsx`**

The reusable case list — pulse-count/duration summary per raw (no
waveform), each case's states, and a per-case "Retry" action for
already-accepted cases. Also the exact reusable unit Spec B's wizard
will render for its recording step.

```tsx
"use client";

import { useActionState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import StatusBadge from "@/components/ui/status-badge";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { InfraredStateDeviceRecordCaseResponse } from "@/lib/api/infrared";

import { retryCaseAction } from "../_lib/actions";
import { EMPTY_RECORD_SESSION_ACTION_STATE } from "../_lib/state";
import RecordCaseRawControls from "./RecordCaseRawControls";

const CASE_STATUS_VARIANT = {
  PENDING: "neutral",
  ACTIVE: "info",
  ACCEPTED: "success",
} as const;

function summarizeRaw(rawCount: number): string {
  return rawCount === 1 ? "1 capture" : `${rawCount} captures`;
}

export default function RecordCaseList({
  cases,
  canMutate,
}: {
  cases: readonly InfraredStateDeviceRecordCaseResponse[];
  canMutate: boolean;
}) {
  const [retryState, retryAction, retryPending] = useActionState(
    retryCaseAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  useRefreshAfterAction(retryState);

  return (
    <ul className="space-y-3">
      {cases.map((c) => (
        <li
          key={c.id}
          className="border-border rounded-2xl border p-4 space-y-3"
        >
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div>
              <p className="font-semibold">{c.description || `Step ${c.step}`}</p>
              <p className="text-muted-foreground text-xs">
                {c.states.map((s) => s.state_value).join(", ")} ·{" "}
                {summarizeRaw(c.raw.length)}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <StatusBadge variant={CASE_STATUS_VARIANT[c.status]}>
                {c.status}
              </StatusBadge>
              {canMutate && c.status === "ACCEPTED" ? (
                <form action={retryAction}>
                  <input type="hidden" name="case_id" value={c.id} />
                  <Button type="submit" variant="secondary" disabled={retryPending}>
                    {retryPending ? "Resetting…" : "Retry"}
                  </Button>
                </form>
              ) : null}
            </div>
          </div>
          {c.raw.map((raw) => (
            <RecordCaseRawControls
              key={raw.id}
              caseId={c.id}
              raw={raw}
              canMutate={canMutate}
            />
          ))}
        </li>
      ))}
      <ActionMessage state={retryState} />
    </ul>
  );
}
```

Verify `StatusBadge`'s `variant` prop accepts `"neutral"`/`"info"`/
`"success"` (already confirmed in exploration — `StatusVariant` type
includes exactly `success | warning | critical | info | neutral`).

- [ ] **Step 6: Write `RecordCoderPanel.tsx`**

Server component (no interactivity beyond the delete dialog, which is
its own small client island):

```tsx
import type { InfraredStateCoderResponse } from "@/lib/api/infrared";

import CodeBlock from "./CodeBlock";
import DeleteStateCoderDialog from "./DeleteStateCoderDialog";

export default function RecordCoderPanel({
  coder,
  canDelete,
}: {
  coder: InfraredStateCoderResponse;
  canDelete: boolean;
}) {
  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="font-display text-lg">{coder.summary_readme}</h3>
          <p className="text-muted-foreground text-sm">{coder.detail_readme}</p>
        </div>
        {canDelete ? (
          <DeleteStateCoderDialog coderId={coder.id} />
        ) : null}
      </div>
      <div>
        <h4 className="mb-2 text-sm font-semibold">Encoder</h4>
        <CodeBlock value={coder.encoder_source} label="encoder source" />
      </div>
      <div>
        <h4 className="mb-2 text-sm font-semibold">Decoder</h4>
        <CodeBlock value={coder.decoder_source} label="decoder source" />
      </div>
    </div>
  );
}
```

This references a new `DeleteStateCoderDialog.tsx` — write it following
`DeleteDeviceTypeDialog.tsx`'s exact structure (dialog + confirm form
via `deleteStateCoderAction`), taking a `coderId` prop and adding a
single hidden `id` field.

- [ ] **Step 7: Write `RecordTestCaseList.tsx`**

```tsx
"use client";

import { useActionState } from "react";

import ActionMessage from "@/components/forms/ActionMessage";
import Button from "@/components/ui/button";
import StatusBadge from "@/components/ui/status-badge";
import { useRefreshAfterAction } from "@/hooks/use-refresh-after-action";
import type { InfraredTestCaseResponse } from "@/lib/api/infrared";

import {
  recordTestCaseResultAction,
  transmitTestCaseAction,
} from "../_lib/actions";
import { EMPTY_RECORD_SESSION_ACTION_STATE } from "../_lib/state";

const TEST_STATUS_VARIANT = {
  PENDING: "neutral",
  PASSED: "success",
  FAILED: "critical",
} as const;

export default function RecordTestCaseList({
  testCases,
  canMutate,
}: {
  testCases: readonly InfraredTestCaseResponse[];
  canMutate: boolean;
}) {
  const [transmitState, transmitAction, transmitPending] = useActionState(
    transmitTestCaseAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  const [resultState, resultAction, resultPending] = useActionState(
    recordTestCaseResultAction,
    EMPTY_RECORD_SESSION_ACTION_STATE,
  );
  useRefreshAfterAction(resultState);

  return (
    <ul className="space-y-3">
      {testCases.map((tc) => (
        <li key={tc.id} className="border-border rounded-2xl border p-4">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div>
              <p className="font-semibold">{tc.description}</p>
              <p className="text-muted-foreground text-xs">
                {tc.states.map((s) => s.state_value).join(", ")}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <StatusBadge variant={TEST_STATUS_VARIANT[tc.status]}>
                {tc.status}
              </StatusBadge>
              {canMutate && tc.status === "PENDING" ? (
                <>
                  <form action={transmitAction}>
                    <input type="hidden" name="test_case_id" value={tc.id} />
                    <Button type="submit" disabled={transmitPending}>
                      {transmitPending ? "Transmitting…" : "Transmit"}
                    </Button>
                  </form>
                  <form action={resultAction}>
                    <input type="hidden" name="test_case_id" value={tc.id} />
                    <input type="hidden" name="passed" value="true" />
                    <Button type="submit" disabled={resultPending}>
                      Passed
                    </Button>
                  </form>
                  <form action={resultAction}>
                    <input type="hidden" name="test_case_id" value={tc.id} />
                    <input type="hidden" name="passed" value="false" />
                    <Button type="submit" variant="secondary" disabled={resultPending}>
                      Failed
                    </Button>
                  </form>
                </>
              ) : null}
            </div>
          </div>
        </li>
      ))}
      <ActionMessage state={transmitState} />
      <ActionMessage state={resultState} />
    </ul>
  );
}
```

- [ ] **Step 8: Write `RecordSessionOverview.tsx`**

```tsx
import LocalDateTime from "@/components/ui/local-date-time";
import StatusBadge from "@/components/ui/status-badge";
import type { InfraredRecordSessionResponse } from "@/lib/api/infrared";

import { RECORDING_STATE_LABELS } from "../../_lib/status";
import DeleteRecordSessionDialog from "./DeleteRecordSessionDialog";

export default function RecordSessionOverview({
  session,
  brand,
  model,
  deviceTypeName,
  canDelete,
}: {
  session: InfraredRecordSessionResponse;
  brand: string;
  model: string;
  deviceTypeName: string;
  canDelete: boolean;
}) {
  const status = RECORDING_STATE_LABELS[session.recording_state];
  return (
    <div className="flex flex-wrap items-start justify-between gap-4">
      <dl className="grid grid-cols-2 gap-4 text-sm sm:grid-cols-4">
        <div>
          <dt className="text-muted-foreground">Device</dt>
          <dd className="mt-1 font-medium">
            {brand} {model}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Device type</dt>
          <dd className="mt-1 font-medium">{deviceTypeName}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Status</dt>
          <dd className="mt-1">
            <StatusBadge variant={status.variant}>{status.label}</StatusBadge>
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Created</dt>
          <dd className="mt-1 font-medium">
            <LocalDateTime value={session.created_at} />
          </dd>
        </div>
      </dl>
      {canDelete ? <DeleteRecordSessionDialog sessionId={session.id} /> : null}
    </div>
  );
}
```

Note this imports `RECORDING_STATE_LABELS` from
`../../_lib/status` (Task 4's list-page `_lib/status.ts`) rather than
duplicating it — the `[id]` directory is one level deeper than
`record/`, so the relative import climbs two levels. If Next.js's App
Router `_lib` convention in this codebase doesn't allow reaching into a
sibling route segment's `_lib` this way (check by looking at whether any
existing nested route does this — e.g.
`nodes/[id]/_lib` vs `nodes/_lib`), instead move `status.ts` up to a
location both routes can import from cleanly, such as
`record/_lib/status.ts` imported via a relative path from `record/[id]/`
(one level up, not two) — reconcile the exact relative path against the
real directory structure at implementation time rather than trusting
the path written here.

- [ ] **Step 9: Write `DeleteRecordSessionDialog.tsx`**

Same structure as `DeleteDeviceTypeDialog.tsx`, using
`deleteRecordSessionAction`, redirecting to `/apps/infrared/record` on
success (check `useActionDialog`'s options for a post-success redirect
hook, or handle it via `useEffect` watching `state.status === "success"`
and calling `useRouter().push(...)`).

- [ ] **Step 10: Write `page.tsx`**

Resolves brand/model/device-type-name via two calls: `getInfraredDevice`
(Task 2 Step 8b's new endpoint, gives brand/model +
`infrared_device_type_id`) cross-referenced against `listInfraredDeviceTypes()`
(already needed elsewhere, gives id → name) — no new join, no new
session-level endpoint, both calls run in parallel with everything else.

```tsx
import type { Metadata } from "next";
import { notFound } from "next/navigation";

import PageHeader from "@/components/ui/page-header";
import {
  getInfraredDevice,
  getInfraredRecordSession,
  getInfraredRecordSessionCoder,
  listInfraredDeviceTypes,
  listInfraredRecordSessionCases,
  listInfraredRecordSessionTestCases,
} from "@/lib/api/infrared";
import { getOptionalById } from "@/lib/api/optional";
import { requirePermission } from "@/lib/session";

import RecordCaseList from "./_components/RecordCaseList";
import RecordCoderPanel from "./_components/RecordCoderPanel";
import RecordSessionOverview from "./_components/RecordSessionOverview";
import RecordTestCaseList from "./_components/RecordTestCaseList";

export const metadata: Metadata = { title: "Record Session — Mate Things" };

export default async function RecordSessionDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const { permissions } = await requirePermission("infrared_record_session:get");

  const session = await getOptionalById(() => getInfraredRecordSession(id));
  if (!session) {
    notFound();
  }

  const [cases, coder, testCases, device, deviceTypes] = await Promise.all([
    listInfraredRecordSessionCases(id),
    getOptionalById(() => getInfraredRecordSessionCoder(id)),
    listInfraredRecordSessionTestCases(id),
    getInfraredDevice(session.infrared_device_id),
    listInfraredDeviceTypes(),
  ]);

  const deviceTypeName =
    deviceTypes.find((dt) => dt.id === device.infrared_device_type_id)
      ?.name ?? "Unknown";
  const canSet = permissions.has("infrared_record_session:set");
  const canDelete = permissions.has("infrared_record_session:delete");

  return (
    <main className="mx-auto w-full max-w-5xl space-y-8 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader title="Record Session" description={`Session ${session.id}`} />
      <RecordSessionOverview
        session={session}
        brand={device.brand}
        model={device.model}
        deviceTypeName={deviceTypeName}
        canDelete={canDelete}
      />
      <section className="space-y-3">
        <h2 className="font-display text-primary text-xl tracking-wide">Cases</h2>
        <RecordCaseList cases={cases} canMutate={canSet} />
      </section>
      {coder ? (
        <section className="space-y-3">
          <h2 className="font-display text-primary text-xl tracking-wide">Coder</h2>
          <RecordCoderPanel coder={coder} canDelete={canDelete} />
        </section>
      ) : null}
      {testCases.length ? (
        <section className="space-y-3">
          <h2 className="font-display text-primary text-xl tracking-wide">Test cases</h2>
          <RecordTestCaseList testCases={testCases} canMutate={canSet} />
        </section>
      ) : null}
    </main>
  );
}
```

- [ ] **Step 11: Verify — build, lint, manual check**

Same worktree-based verification as Task 4 Step 6. Additionally, since a
real record session requires backend state that doesn't exist without
actually running the recording flow, verify what CAN be verified without
one: the page's `notFound()` path (visit `/apps/infrared/record/nonexistent-uuid`),
and — if any session exists from prior manual backend testing this
session's earlier work may have left in the dev database — the real
data-bearing path. If no real session exists to test against, note this
explicitly in the task completion report as a real, acknowledged gap
rather than claiming full verification.

- [ ] **Step 12: Commit**

```bash
cd frontend
git add "src/app/(authenticated)/apps/infrared/record/[id]/"
git commit -m "infrared: build the record session detail page"
```

---

### Task 6: Full-repo verification

**Files:** none (verification only).

- [ ] **Step 1: Backend — build, vet, test, format**

Run: `cd backend && go build ./... && go vet ./... && go test ./... && gofmt -l .`
Expected: clean build/vet, all tests pass, empty gofmt output.

- [ ] **Step 2: Backend — module tidiness**

Run: `cd backend && go mod tidy && git diff --stat go.mod go.sum`
Expected: no diff (no new dependencies added).

- [ ] **Step 3: Frontend — lint, type-check, format, build**

In the `.worktrees/mate-things-frontend-consistency` worktree (per this
project's established verification workflow): copy in every changed
frontend file from this plan, run `npm run lint`, `npx prettier --check .`,
and `npm run build`. Restore the worktree afterward.

- [ ] **Step 4: Confirm no stray files**

Run: `cd /home/dodol/Repositories/mate/mate-things && git status --short`
Expected: clean, or showing only files intentionally modified across
this plan's tasks (all already committed by their own step — this
should show nothing).
