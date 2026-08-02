# Action History Page Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** refactor the Action History page — drop the manual Execution ID
filter (kept as a URL-only deep-link param), add real server-side
pagination with searchable Action/Node dropdowns and a status filter (a
full contract → infrastructure → application → presentation chain on the
backend), and replace the card grid with an expandable table (Timestamp /
Action / Node / Status).

**Architecture:** One backend chain: repository contract gains
`actionStatus`/`page`/`limit`, the Postgres query joins in
`actions.name`/`nodes.device_id`/`nodes.name` and adds `LIMIT`/`OFFSET`,
the usecase threads the new fields through unchanged, the HTTP handler
switches from an unbounded `CountDataResponse` to the same
`PageDataResponse` + `PageArgs` pattern every other paginated endpoint
already uses. One frontend rewrite: two new search comboboxes built on the
existing `RoleSearchCombobox` pattern, a dedicated (not shared) filter
parser so `telemetry`/`node-logs` are untouched, and a first-of-its-kind
`<table>` component with expandable rows.

**Tech Stack:** Go / pgx / squirrel (backend), Next.js App Router / React
19 / TypeScript (frontend).

## Global Constraints

- No automated tests are being planned for this feature (explicit scope
  decision) — verification is build/compile/lint plus a live check
  against the real running stack.
- `action_status` is a fixed 4-value enum
  (`UNEXECUTED|UNRESPONDED|FAILED|SUCCESS` — `domainmodels.ActionStatus`'s
  existing constants). Never introduce a second hardcoded list of these
  values — reuse the existing constants everywhere a Go-side check is
  needed, and the existing `_components/ActionLogCard.tsx`-derived label
  map (moved to `_lib/status.ts`, see Task F3) on the frontend.
- `telemetry` and `node-logs` share `lib/record-filters.ts`,
  `components/records/RecordWindow.tsx`, and
  `components/records/ScopedDeleteDialog.tsx` with today's Action History
  page. None of those three shared files are modified by this plan —
  Action History gets its own dedicated filter-parsing module
  (`_lib/filters.ts`) and drops `RecordWindow` entirely in favor of a new
  table + the existing `components/collection/Pagination` component,
  so the other two pages are provably unaffected.
- The dashboard's `?execution_id=...&start=...` deep link (`_components/UrgentAttention.tsx`
  in the dashboard feature) must keep working — `execution_id` stays a
  recognized, parsed query param; it's simply never rendered as an
  editable form field.
- No database migration — `action_status`, `actions.name`,
  `nodes.device_id`, `nodes.name` all already exist.

---

## Task B1 (backend): domain model + repository contract

**Repo:** `mate-things` (backend)

**Files:**
- Modify: `internal/domain/models/action_log.go`
- Modify: `internal/domain/contracts/repository/action_log.go`

**Interfaces:**
- Produces (consumed by B2, B3): `domainmodels.ActionLogListItem` (embeds
  `ActionLog`, adds `ActionName string`, `NodeDeviceId *string`,
  `NodeName *string`); `domaincontractsrepository.ActionLog`'s
  `ReadByFilter(ctx, executedAtStart, executedAtEnd, actionId, nodeId,
  actionStatus *domainmodels.ActionStatus, page int, limit int)
  ([]domainmodels.ActionLogListItem, int, error)` and `DeleteByFilter(ctx,
  executedAtStart, executedAtEnd, actionId, nodeId, actionStatus
  *domainmodels.ActionStatus) (int, error)`.

- [ ] **Step 1: Add the list read-model**

In `internal/domain/models/action_log.go`, add at the end of the file:

```go
// ActionLogListItem is the read model ReadByFilter returns - it embeds the
// table-mapped ActionLog (used as-is by Create and the single-dispatch
// response) plus the human-readable names joined in from actions/nodes,
// so the frontend's Action History table doesn't have to resolve UUIDs
// to names itself. NodeDeviceId/NodeName are pointers because node_id
// itself is nullable (an action log can predate/lack a resolved node);
// ActionName is a plain string because action_id is NOT NULL and actions
// are soft-deleted (never hard-deleted), so the join always matches.
type ActionLogListItem struct {
	ActionLog
	ActionName   string
	NodeDeviceId *string
	NodeName     *string
}
```

- [ ] **Step 2: Extend the repository contract**

In `internal/domain/contracts/repository/action_log.go`, replace:

```go
	ReadByFilter(
		ctx context.Context,
		executedAtStart *time.Time,
		executedAtEnd *time.Time,
		actionId *uuid.UUID,
		nodeId *uuid.UUID,
	) (actionLogs []domainmodels.ActionLog, total int, err error)

	DeleteByFilter(
		ctx context.Context,
		executedAtStart *time.Time,
		executedAtEnd *time.Time,
		actionId *uuid.UUID,
		nodeId *uuid.UUID,
	) (total int, err error)
```

with:

```go
	ReadByFilter(
		ctx context.Context,
		executedAtStart *time.Time,
		executedAtEnd *time.Time,
		actionId *uuid.UUID,
		nodeId *uuid.UUID,
		actionStatus *domainmodels.ActionStatus,
		page int,
		limit int,
	) (actionLogs []domainmodels.ActionLogListItem, total int, err error)

	DeleteByFilter(
		ctx context.Context,
		executedAtStart *time.Time,
		executedAtEnd *time.Time,
		actionId *uuid.UUID,
		nodeId *uuid.UUID,
		actionStatus *domainmodels.ActionStatus,
	) (total int, err error)
```

- [ ] **Step 3: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/backend
go build ./... 2>&1 | head -60
```

Expected: build errors in every file that implements/calls the old
signatures (the Postgres impl, the usecase, the HTTP handler) — that's
correct at this point, they're fixed in Tasks B2-B4. Confirm the errors
are exactly the expected call sites, nothing else.

- [ ] **Step 4: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add backend/internal/domain/models/action_log.go \
        backend/internal/domain/contracts/repository/action_log.go
git commit -m "feat(action-log): add ActionLogListItem and extend the repository contract for status + pagination"
```

(This commit intentionally leaves the build red — the plan's later tasks
fix every call site. If executing task-by-task with review gates, note
this explicitly rather than treating the red build as a surprise.)

---

## Task B2 (backend): Postgres infrastructure — join, status filter, pagination

**Repo:** `mate-things` (backend)

**Files:**
- Modify: `internal/infrastructure/repository/action_log/postgres_query.go`
- Modify: `internal/infrastructure/repository/action_log/postgres.go`
- Modify: `internal/infrastructure/repository/shared/scan_pgx.go`

**Interfaces:**
- Consumes: `domainmodels.ActionLogListItem` (B1).
- Produces (consumed by B3 indirectly via the contract): the
  `postgresImpl` methods now satisfy B1's extended contract exactly;
  `infrastructurerepositoryshared.ScanPgxActionLogListItem(row pgx.Row)
  (domainmodels.ActionLogListItem, error)` and
  `ScanPgxActionLogListItems(rows pgx.Rows)
  ([]domainmodels.ActionLogListItem, error)`.

- [ ] **Step 1: Add the two new scan functions**

In `internal/infrastructure/repository/shared/scan_pgx.go`, add right
after the existing `ScanPgxActionLogs` function:

```go
func ScanPgxActionLogListItem(row pgx.Row) (domainmodels.ActionLogListItem, error) {
	var item domainmodels.ActionLogListItem
	err := row.Scan(
		&item.Id,
		&item.ExecutionId,
		&item.ActionId,
		&item.NodeId,
		&item.ActionStatus,
		&item.ActionMessage,
		&item.Payload,
		&item.ExecutedAt,
		&item.CreatedAt,
		&item.ActionName,
		&item.NodeDeviceId,
		&item.NodeName,
	)
	return item, err
}

func ScanPgxActionLogListItems(rows pgx.Rows) ([]domainmodels.ActionLogListItem, error) {
	items := make([]domainmodels.ActionLogListItem, 0)
	for rows.Next() {
		item, err := ScanPgxActionLogListItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
```

(The 9-field order matches `actionLogColumns` exactly, then the 3 joined
columns in the order Step 3 below selects them — `action_name`,
`node_device_id`, `node_name`.)

- [ ] **Step 2: Add the `domainmodels` import to `postgres_query.go`**

```go
import (
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)
```
→
```go
import (
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)
```

- [ ] **Step 3: Rewrite `queryReadByFilter` — join, status, pagination**

Replace the whole function:

```go
func (p *postgresImpl) queryReadByFilter(
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.SqrD.Select(actionLogColumns...).
		From("action_logs")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("action_logs")

	baseQ, totalQ = applyActionLogFilters(baseQ, totalQ, executedAtStart, executedAtEnd, actionId, nodeId)

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("executed_at DESC", "id ASC").
		ToSql()
	return
}
```

with:

```go
func (p *postgresImpl) queryReadByFilter(
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
	actionStatus *domainmodels.ActionStatus,
	page int,
	limit int,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	qualifiedColumns := make([]string, len(actionLogColumns))
	for i, col := range actionLogColumns {
		qualifiedColumns[i] = "action_logs." + col
	}

	// LEFT JOIN (not INNER) on both: node_id is nullable, and defensively,
	// a row referencing a since-deleted action/node should still surface
	// rather than silently vanish from the list. The count query stays
	// join-free - it never needs the joined name columns, only the same
	// WHERE filters, none of which reference an ambiguous column name.
	baseQ := p.SqrD.Select(qualifiedColumns...).
		Column("actions.name AS action_name").
		Column("nodes.device_id AS node_device_id").
		Column("nodes.name AS node_name").
		From("action_logs").
		LeftJoin("actions ON actions.id = action_logs.action_id").
		LeftJoin("nodes ON nodes.id = action_logs.node_id")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("action_logs")

	baseQ, totalQ = applyActionLogFilters(baseQ, totalQ, executedAtStart, executedAtEnd, actionId, nodeId, actionStatus)

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("action_logs.executed_at DESC", "action_logs.id ASC").
		Limit(uint64(limit)).
		Offset(uint64((page - 1) * limit)).
		ToSql()
	return
}
```

- [ ] **Step 4: Add the status parameter to `queryDeleteByFilter`**

Replace:

```go
func (p *postgresImpl) queryDeleteByFilter(
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Delete("action_logs")

	if executedAtStart != nil {
		q = q.Where(squirrel.GtOrEq{"executed_at": *executedAtStart})
	}
	if executedAtEnd != nil {
		q = q.Where(squirrel.LtOrEq{"executed_at": *executedAtEnd})
	}
	if actionId != nil {
		q = q.Where(squirrel.Eq{"action_id": *actionId})
	}
	if nodeId != nil {
		q = q.Where(squirrel.Eq{"node_id": *nodeId})
	}

	return q.ToSql()
}
```

with:

```go
func (p *postgresImpl) queryDeleteByFilter(
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
	actionStatus *domainmodels.ActionStatus,
) (query string, args []any, err error) {
	q := p.SqrD.Delete("action_logs")

	if executedAtStart != nil {
		q = q.Where(squirrel.GtOrEq{"executed_at": *executedAtStart})
	}
	if executedAtEnd != nil {
		q = q.Where(squirrel.LtOrEq{"executed_at": *executedAtEnd})
	}
	if actionId != nil {
		q = q.Where(squirrel.Eq{"action_id": *actionId})
	}
	if nodeId != nil {
		q = q.Where(squirrel.Eq{"node_id": *nodeId})
	}
	if actionStatus != nil {
		q = q.Where(squirrel.Eq{"action_status": string(*actionStatus)})
	}

	return q.ToSql()
}
```

- [ ] **Step 5: Add the status parameter to `applyActionLogFilters`**

Replace:

```go
func applyActionLogFilters(
	baseQ squirrel.SelectBuilder,
	totalQ squirrel.SelectBuilder,
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	if executedAtStart != nil {
		condition := squirrel.GtOrEq{"executed_at": *executedAtStart}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if executedAtEnd != nil {
		condition := squirrel.LtOrEq{"executed_at": *executedAtEnd}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if actionId != nil {
		condition := squirrel.Eq{"action_id": *actionId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if nodeId != nil {
		condition := squirrel.Eq{"node_id": *nodeId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	return baseQ, totalQ
}
```

with:

```go
func applyActionLogFilters(
	baseQ squirrel.SelectBuilder,
	totalQ squirrel.SelectBuilder,
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
	actionStatus *domainmodels.ActionStatus,
) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	if executedAtStart != nil {
		condition := squirrel.GtOrEq{"executed_at": *executedAtStart}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if executedAtEnd != nil {
		condition := squirrel.LtOrEq{"executed_at": *executedAtEnd}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if actionId != nil {
		condition := squirrel.Eq{"action_id": *actionId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if nodeId != nil {
		condition := squirrel.Eq{"node_id": *nodeId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if actionStatus != nil {
		condition := squirrel.Eq{"action_status": string(*actionStatus)}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	return baseQ, totalQ
}
```

(None of `executed_at`, `action_id`, `node_id`, `action_status` exist as
column names on `actions` or `nodes`, so these bare, unqualified `WHERE`
conditions remain unambiguous even once the data query joins those two
tables in — only the `SELECT`/`ORDER BY` column lists needed explicit
`action_logs.` qualification, since `id` collides across all three
tables.)

- [ ] **Step 6: Update `postgres.go`'s `ReadByFilter`/`DeleteByFilter`**

Replace:

```go
func (p *postgresImpl) ReadByFilter(
	ctx context.Context,
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
) (actionLogs []domainmodels.ActionLog, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByFilter(executedAtStart, executedAtEnd, actionId, nodeId)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read action logs query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count action logs", err)
	}
	if total == 0 {
		return []domainmodels.ActionLog{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read action logs", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxActionLogs(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan action logs", err)
	}

	return items, total, nil
}
```

with:

```go
func (p *postgresImpl) ReadByFilter(
	ctx context.Context,
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
	actionStatus *domainmodels.ActionStatus,
	page int,
	limit int,
) (actionLogs []domainmodels.ActionLogListItem, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByFilter(executedAtStart, executedAtEnd, actionId, nodeId, actionStatus, page, limit)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read action logs query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count action logs", err)
	}
	if total == 0 {
		return []domainmodels.ActionLogListItem{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read action logs", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxActionLogListItems(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan action logs", err)
	}

	return items, total, nil
}
```

Replace:

```go
func (p *postgresImpl) DeleteByFilter(
	ctx context.Context,
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
) (total int, err error) {
	query, args, err := p.queryDeleteByFilter(executedAtStart, executedAtEnd, actionId, nodeId)
```

with:

```go
func (p *postgresImpl) DeleteByFilter(
	ctx context.Context,
	executedAtStart *time.Time,
	executedAtEnd *time.Time,
	actionId *uuid.UUID,
	nodeId *uuid.UUID,
	actionStatus *domainmodels.ActionStatus,
) (total int, err error) {
	query, args, err := p.queryDeleteByFilter(executedAtStart, executedAtEnd, actionId, nodeId, actionStatus)
```

(the rest of `DeleteByFilter`'s body is unchanged).

- [ ] **Step 7: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/backend
go build ./... 2>&1 | head -60
```

Expected: remaining errors only in the usecase and HTTP handler layers
(fixed in B3/B4) — the repository package itself must compile cleanly
now.

- [ ] **Step 8: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add backend/internal/infrastructure/repository/action_log/postgres_query.go \
        backend/internal/infrastructure/repository/action_log/postgres.go \
        backend/internal/infrastructure/repository/shared/scan_pgx.go
git commit -m "feat(action-log): join action/node names, add status filter and pagination to the read query"
```

---

## Task B3 (backend): application usecase + domain usecase types

**Repo:** `mate-things` (backend)

**Files:**
- Modify: `internal/domain/usecases/action/history.go`
- Modify: `internal/application/action/history/usecase.go`

**Interfaces:**
- Consumes: `domaincontractsrepository.ActionLog`'s extended methods (B2).
- Produces (consumed by B4): `domainusecasesaction.ReadActionLogsByFilterRequest`
  gains `ActionStatus *domainmodels.ActionStatus`, `Page int`, `Limit
  int`; `DeleteActionLogsByFilterRequest` gains `ActionStatus
  *domainmodels.ActionStatus`; `History.ReadByFilter` returns
  `([]domainmodels.ActionLogListItem, int, error)`.

- [ ] **Step 1: Extend the domain usecase types**

In `internal/domain/usecases/action/history.go`, replace:

```go
type History interface {
	ReadByFilter(ctx context.Context, request ReadActionLogsByFilterRequest) ([]domainmodels.ActionLog, int, error)
	DeleteByFilter(ctx context.Context, request DeleteActionLogsByFilterRequest) (int, error)
}

type ReadActionLogsByFilterRequest struct {
	ExecutedAtStart *time.Time
	ExecutedAtEnd   *time.Time
	ActionId        *uuid.UUID
	NodeId          *uuid.UUID
}

type DeleteActionLogsByFilterRequest struct {
	ExecutedAtStart *time.Time
	ExecutedAtEnd   *time.Time
	ActionId        *uuid.UUID
	NodeId          *uuid.UUID
}
```

with:

```go
type History interface {
	ReadByFilter(ctx context.Context, request ReadActionLogsByFilterRequest) ([]domainmodels.ActionLogListItem, int, error)
	DeleteByFilter(ctx context.Context, request DeleteActionLogsByFilterRequest) (int, error)
}

type ReadActionLogsByFilterRequest struct {
	ExecutedAtStart *time.Time
	ExecutedAtEnd   *time.Time
	ActionId        *uuid.UUID
	NodeId          *uuid.UUID
	ActionStatus    *domainmodels.ActionStatus
	Page            int
	Limit           int
}

type DeleteActionLogsByFilterRequest struct {
	ExecutedAtStart *time.Time
	ExecutedAtEnd   *time.Time
	ActionId        *uuid.UUID
	NodeId          *uuid.UUID
	ActionStatus    *domainmodels.ActionStatus
}
```

- [ ] **Step 2: Thread the new fields through the usecase implementation**

In `internal/application/action/history/usecase.go`, replace:

```go
func (u *usecase) ReadByFilter(
	ctx context.Context,
	request domainusecasesaction.ReadActionLogsByFilterRequest,
) ([]domainmodels.ActionLog, int, error) {
	const tag = "action/history/ReadByFilter"

	actionLogs, total, err := u.actionLog.ReadByFilter(
		ctx,
		request.ExecutedAtStart,
		request.ExecutedAtEnd,
		request.ActionId,
		request.NodeId,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read action logs", domainmodels.LoggerMeta{
			"err":       err,
			"action_id": request.ActionId,
			"node_id":   request.NodeId,
		})
		return nil, 0, err
	}

	return actionLogs, total, nil
}
```

with:

```go
func (u *usecase) ReadByFilter(
	ctx context.Context,
	request domainusecasesaction.ReadActionLogsByFilterRequest,
) ([]domainmodels.ActionLogListItem, int, error) {
	const tag = "action/history/ReadByFilter"

	actionLogs, total, err := u.actionLog.ReadByFilter(
		ctx,
		request.ExecutedAtStart,
		request.ExecutedAtEnd,
		request.ActionId,
		request.NodeId,
		request.ActionStatus,
		request.Page,
		request.Limit,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read action logs", domainmodels.LoggerMeta{
			"err":       err,
			"action_id": request.ActionId,
			"node_id":   request.NodeId,
		})
		return nil, 0, err
	}

	return actionLogs, total, nil
}
```

Replace:

```go
	total, err := u.actionLog.DeleteByFilter(
		ctx,
		request.ExecutedAtStart,
		request.ExecutedAtEnd,
		request.ActionId,
		request.NodeId,
	)
```

with:

```go
	total, err := u.actionLog.DeleteByFilter(
		ctx,
		request.ExecutedAtStart,
		request.ExecutedAtEnd,
		request.ActionId,
		request.NodeId,
		request.ActionStatus,
	)
```

(inside `DeleteByFilter` — the rest of that function body is unchanged).

- [ ] **Step 3: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/backend
go build ./... 2>&1 | head -60
```

Expected: remaining errors only in the presentation layer (handler.go and
response/action_log.go) — fixed in B4.

- [ ] **Step 4: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add backend/internal/domain/usecases/action/history.go \
        backend/internal/application/action/history/usecase.go
git commit -m "feat(action-log): thread status/pagination through the history usecase"
```

---

## Task B4 (backend): HTTP presentation — status query param, pagination, response shape

**Repo:** `mate-things` (backend)

**Files:**
- Modify: `internal/presentation/http/response/action_log.go`
- Modify: `internal/presentation/http/handler/action/handler.go`

**Interfaces:**
- Consumes: `domainusecasesaction.History` (B3).
- Produces: `GET /v1/action-logs` returns
  `presentationhttpresponse.PageDataResponse[ActionLogResponse]` (was
  `CountDataResponse[...]`) with a new `status` query param;
  `presentationhttpresponse.ActionLogResponse` gains `action_name`,
  `node_device_id`, `node_name`.

- [ ] **Step 1: Extend the response type and add the list mapper**

In `internal/presentation/http/response/action_log.go`, replace the whole
file:

```go
package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ActionLogResponse struct {
	Id            int64                     `json:"id" example:"4821"`
	ExecutionId   string                    `json:"execution_id" example:"a4d7f1c9-3e6b-4a8d-9c2f-5b1e7d4a6c02"`
	ActionId      string                    `json:"action_id" example:"6d9e2f5a-8b1c-4d3e-9f6a-2c5d8e1f4b07"`
	NodeId        *string                   `json:"node_id,omitempty" example:"5e8a1c3f-2b7d-4f6a-9c1e-3a8b6d2f4e09"`
	ActionStatus  domainmodels.ActionStatus `json:"action_status" example:"SUCCESS"`
	ActionMessage *string                   `json:"action_message,omitempty" example:"Shot pulled: 93.5C for 28s. Crema looked great."`
	Payload       json.RawMessage           `json:"payload" swaggertype:"object"`
	ExecutedAt    time.Time                 `json:"executed_at" example:"2026-07-28T08:15:00Z"`
	CreatedAt     time.Time                 `json:"created_at" example:"2026-07-28T08:15:02Z"`
}

func ActionLog(actionLog domainmodels.ActionLog) ActionLogResponse {
	return ActionLogResponse{
		Id:            actionLog.Id,
		ExecutionId:   UUIDString(actionLog.ExecutionId),
		ActionId:      UUIDString(actionLog.ActionId),
		NodeId:        UUIDPtrString(actionLog.NodeId),
		ActionStatus:  actionLog.ActionStatus,
		ActionMessage: actionLog.ActionMessage,
		Payload:       NormalizeJSON(actionLog.Payload),
		ExecutedAt:    actionLog.ExecutedAt,
		CreatedAt:     actionLog.CreatedAt,
	}
}

func ActionLogs(actionLogs []domainmodels.ActionLog) []ActionLogResponse {
	result := make([]ActionLogResponse, 0, len(actionLogs))
	for _, actionLog := range actionLogs {
		result = append(result, ActionLog(actionLog))
	}
	return result
}
```

with:

```go
package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ActionLogResponse struct {
	Id            int64                     `json:"id" example:"4821"`
	ExecutionId   string                    `json:"execution_id" example:"a4d7f1c9-3e6b-4a8d-9c2f-5b1e7d4a6c02"`
	ActionId      string                    `json:"action_id" example:"6d9e2f5a-8b1c-4d3e-9f6a-2c5d8e1f4b07"`
	ActionName    string                    `json:"action_name" example:"pull_espresso_shot"`
	NodeId        *string                   `json:"node_id,omitempty" example:"5e8a1c3f-2b7d-4f6a-9c1e-3a8b6d2f4e09"`
	NodeDeviceId  *string                   `json:"node_device_id,omitempty" example:"AC276E5E030C"`
	NodeName      *string                   `json:"node_name,omitempty" example:"Kitchen Espresso Machine"`
	ActionStatus  domainmodels.ActionStatus `json:"action_status" example:"SUCCESS"`
	ActionMessage *string                   `json:"action_message,omitempty" example:"Shot pulled: 93.5C for 28s. Crema looked great."`
	Payload       json.RawMessage           `json:"payload" swaggertype:"object"`
	ExecutedAt    time.Time                 `json:"executed_at" example:"2026-07-28T08:15:00Z"`
	CreatedAt     time.Time                 `json:"created_at" example:"2026-07-28T08:15:02Z"`
}

func ActionLog(actionLog domainmodels.ActionLog) ActionLogResponse {
	return ActionLogResponse{
		Id:            actionLog.Id,
		ExecutionId:   UUIDString(actionLog.ExecutionId),
		ActionId:      UUIDString(actionLog.ActionId),
		NodeId:        UUIDPtrString(actionLog.NodeId),
		ActionStatus:  actionLog.ActionStatus,
		ActionMessage: actionLog.ActionMessage,
		Payload:       NormalizeJSON(actionLog.Payload),
		ExecutedAt:    actionLog.ExecutedAt,
		CreatedAt:     actionLog.CreatedAt,
	}
}

// ActionLogListItem maps the joined list read-model - the single-item
// ActionLog() mapper above is unchanged and still used by the dispatch
// response, which never needed name enrichment.
func ActionLogListItem(item domainmodels.ActionLogListItem) ActionLogResponse {
	resp := ActionLog(item.ActionLog)
	resp.ActionName = item.ActionName
	resp.NodeDeviceId = item.NodeDeviceId
	resp.NodeName = item.NodeName
	return resp
}

func ActionLogListItems(items []domainmodels.ActionLogListItem) []ActionLogResponse {
	result := make([]ActionLogResponse, 0, len(items))
	for _, item := range items {
		result = append(result, ActionLogListItem(item))
	}
	return result
}
```

- [ ] **Step 2: Add the `domainmodels` import to the handler**

In `internal/presentation/http/handler/action/handler.go`:

```go
import (
	"net/http"
	"time"

	domainusecasesaction "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/action"
	presentationhttprequest "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)
```
→
```go
import (
	"net/http"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesaction "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/action"
	presentationhttprequest "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)
```

- [ ] **Step 3: Rewrite `ActionLogGetList`**

Replace:

```go
// ActionLogGetList godoc
//
// @Summary Action Log List
// @Tags Action Logs
// @Produce json
// @Security BearerAuth
// @Param executed_at_start query string false "RFC3339 timestamp"
// @Param executed_at_end query string false "RFC3339 timestamp"
// @Param action_id query string false "action id"
// @Param node_id query string false "node id"
// @Success 200 {object} presentationhttpresponse.CountDataResponse[presentationhttpresponse.ActionLogResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/action-logs [get]
func (h *handler) ActionLogGetList(c *echo.Context) error {
	filter, err := h.actionLogFilter(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	actionLogs, total, err := h.historyUseCase.ReadByFilter(c.Request().Context(), filter)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.CountDataResponse[presentationhttpresponse.ActionLogResponse]{
		Data:       presentationhttpresponse.ActionLogs(actionLogs),
		TotalItems: total,
	})
}
```

with:

```go
// ActionLogGetList godoc
//
// @Summary Action Log List
// @Tags Action Logs
// @Produce json
// @Security BearerAuth
// @Param page query int false "page number"
// @Param limit query int false "page size"
// @Param executed_at_start query string false "RFC3339 timestamp"
// @Param executed_at_end query string false "RFC3339 timestamp"
// @Param action_id query string false "action id"
// @Param node_id query string false "node id"
// @Param status query string false "action status (UNEXECUTED, UNRESPONDED, FAILED, SUCCESS)"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.ActionLogResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/action-logs [get]
func (h *handler) ActionLogGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	filter, err := h.actionLogFilter(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	filter.Page = page.Page
	filter.Limit = page.Limit

	actionLogs, total, err := h.historyUseCase.ReadByFilter(c.Request().Context(), filter)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.ActionLogResponse]{
		Data: presentationhttpresponse.ActionLogListItems(actionLogs),
		Page: presentationhttputils.PageResponse(page, total),
	})
}
```

- [ ] **Step 4: Add the `status` param to the delete endpoint's Swagger doc**

Replace:

```go
// ActionLogDelete godoc
//
// @Summary Action Log Delete
// @Tags Action Logs
// @Produce json
// @Security BearerAuth
// @Param executed_at_start query string false "RFC3339 timestamp"
// @Param executed_at_end query string false "RFC3339 timestamp"
// @Param action_id query string false "action id"
// @Param node_id query string false "node id"
// @Success 200 {object} presentationhttpresponse.CountResponse
```

with:

```go
// ActionLogDelete godoc
//
// @Summary Action Log Delete
// @Tags Action Logs
// @Produce json
// @Security BearerAuth
// @Param executed_at_start query string false "RFC3339 timestamp"
// @Param executed_at_end query string false "RFC3339 timestamp"
// @Param action_id query string false "action id"
// @Param node_id query string false "node id"
// @Param status query string false "action status (UNEXECUTED, UNRESPONDED, FAILED, SUCCESS)"
// @Success 200 {object} presentationhttpresponse.CountResponse
```

(`ActionLogDelete`'s function body itself needs no change — it already
delegates to `actionLogFilter`/`actionLogDeleteFilter`, updated next.)

- [ ] **Step 5: Parse `status` in `actionLogFilter`, add the validation helper, update `actionLogDeleteFilter`**

Replace:

```go
func (h *handler) actionLogFilter(c *echo.Context) (domainusecasesaction.ReadActionLogsByFilterRequest, error) {
	executedAtStart, err := presentationhttputils.QueryTime(c, "executed_at_start")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}
	executedAtEnd, err := presentationhttputils.QueryTime(c, "executed_at_end")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}
	actionId, err := presentationhttputils.QueryUUID(c, "action_id")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}
	nodeId, err := presentationhttputils.QueryUUID(c, "node_id")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}

	return domainusecasesaction.ReadActionLogsByFilterRequest{
		ExecutedAtStart: executedAtStart,
		ExecutedAtEnd:   executedAtEnd,
		ActionId:        actionId,
		NodeId:          nodeId,
	}, nil
}

func (h *handler) actionLogDeleteFilter(c *echo.Context) (domainusecasesaction.DeleteActionLogsByFilterRequest, error) {
	filter, err := h.actionLogFilter(c)
	if err != nil {
		return domainusecasesaction.DeleteActionLogsByFilterRequest{}, err
	}

	return domainusecasesaction.DeleteActionLogsByFilterRequest{
		ExecutedAtStart: filter.ExecutedAtStart,
		ExecutedAtEnd:   filter.ExecutedAtEnd,
		ActionId:        filter.ActionId,
		NodeId:          filter.NodeId,
	}, nil
}
```

with:

```go
func (h *handler) actionLogFilter(c *echo.Context) (domainusecasesaction.ReadActionLogsByFilterRequest, error) {
	executedAtStart, err := presentationhttputils.QueryTime(c, "executed_at_start")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}
	executedAtEnd, err := presentationhttputils.QueryTime(c, "executed_at_end")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}
	actionId, err := presentationhttputils.QueryUUID(c, "action_id")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}
	nodeId, err := presentationhttputils.QueryUUID(c, "node_id")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}
	actionStatus, err := actionLogStatus(presentationhttputils.QueryString(c, "status"))
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}

	return domainusecasesaction.ReadActionLogsByFilterRequest{
		ExecutedAtStart: executedAtStart,
		ExecutedAtEnd:   executedAtEnd,
		ActionId:        actionId,
		NodeId:          nodeId,
		ActionStatus:    actionStatus,
	}, nil
}

func (h *handler) actionLogDeleteFilter(c *echo.Context) (domainusecasesaction.DeleteActionLogsByFilterRequest, error) {
	filter, err := h.actionLogFilter(c)
	if err != nil {
		return domainusecasesaction.DeleteActionLogsByFilterRequest{}, err
	}

	return domainusecasesaction.DeleteActionLogsByFilterRequest{
		ExecutedAtStart: filter.ExecutedAtStart,
		ExecutedAtEnd:   filter.ExecutedAtEnd,
		ActionId:        filter.ActionId,
		NodeId:          filter.NodeId,
		ActionStatus:    filter.ActionStatus,
	}, nil
}

// actionLogStatus mirrors node_log/handler.go's nodeLogLevel helper - a
// package-local enum-string validator, not a shared presentationhttputils
// addition, matching this codebase's existing precedent for this kind of
// query param.
func actionLogStatus(value *string) (*domainmodels.ActionStatus, error) {
	if value == nil {
		return nil, nil
	}

	status := domainmodels.ActionStatus(*value)
	switch status {
	case domainmodels.ActionStatusUnexecuted,
		domainmodels.ActionStatusUnresponded,
		domainmodels.ActionStatusFailed,
		domainmodels.ActionStatusSuccess:
		return &status, nil
	default:
		return nil, domainmodels.NewError(
			"status must be one of UNEXECUTED, UNRESPONDED, FAILED, SUCCESS",
			domainmodels.ErrTypeValidation,
			nil,
		)
	}
}
```

- [ ] **Step 6: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/backend
go build ./...
go vet ./...
go test ./...
```

Expected: clean build and vet; full test suite passes (the one
pre-existing, unrelated `node_log` handler test failure documented
earlier this session is not a regression from this work — confirm no
*new* failures beyond that known one).

- [ ] **Step 7: Regenerate Swagger docs**

```bash
cd /home/dodol/Repositories/mate/mate-things/backend
swag init -g cmd/main/main.go -o docs/swagger --parseInternal --parseDependency
```

(Per this repo's documented convention — `docs.go`/`swagger.json`/`swagger.yaml`
are generated, never hand-edited. If the `swag` CLI isn't available in
this environment, note that plainly rather than skipping silently.)

- [ ] **Step 8: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add backend/internal/presentation/http/response/action_log.go \
        backend/internal/presentation/http/handler/action/handler.go \
        backend/docs/swagger
git commit -m "feat(action-log): add status filter, pagination, and joined names to the action-logs API"
```

---

## Task F1 (frontend): API client types + query shape

**Repo:** `mate-things` (frontend)

**Files:**
- Modify: `frontend/src/lib/api/action-logs.ts`

**Interfaces:**
- Produces (consumed by F2-F5): `ActionLogResponse` gains `action_name:
  string`, `node_device_id?: string`, `node_name?: string`;
  `ActionLogFilterQuery extends PageQuery` gains `status?: ActionStatus`;
  `listActionLogs` returns `Promise<PageDataResponse<ActionLogResponse>>`
  (was `CountDataResponse`).

- [ ] **Step 1: Rewrite the file**

Replace the whole file:

```ts
"use server";

import { apiFetch, buildQuery } from "@/lib/api/client";
import type { CountDataResponse, CountResponse } from "@/lib/api/types";

export type ActionStatus = "UNEXECUTED" | "UNRESPONDED" | "FAILED" | "SUCCESS";

export interface ActionLogResponse {
  id: number;
  execution_id: string;
  action_id: string;
  node_id?: string;
  action_status: ActionStatus;
  action_message?: string;
  payload: Record<string, unknown>;
  executed_at: string;
  created_at: string;
}

export interface ActionLogFilterQuery {
  /** ISO 8601 timestamps. */
  executed_at_start?: string;
  executed_at_end?: string;
  action_id?: string;
  node_id?: string;
}

/** Filter-only, no pagination - see AGENTS.md's data-fetching note on the two list-response shapes. */
export async function listActionLogs(
  query: ActionLogFilterQuery = {},
): Promise<CountDataResponse<ActionLogResponse>> {
  return apiFetch(`/action-logs${buildQuery(query)}`);
}

export async function deleteActionLogs(
  query: ActionLogFilterQuery = {},
): Promise<CountResponse> {
  return apiFetch(`/action-logs${buildQuery(query)}`, { method: "DELETE" });
}
```

with:

```ts
"use server";

import { apiFetch, buildQuery } from "@/lib/api/client";
import type { CountResponse, PageDataResponse, PageQuery } from "@/lib/api/types";

export type ActionStatus = "UNEXECUTED" | "UNRESPONDED" | "FAILED" | "SUCCESS";

export interface ActionLogResponse {
  id: number;
  execution_id: string;
  action_id: string;
  action_name: string;
  node_id?: string;
  node_device_id?: string;
  node_name?: string;
  action_status: ActionStatus;
  action_message?: string;
  payload: Record<string, unknown>;
  executed_at: string;
  created_at: string;
}

export interface ActionLogFilterQuery extends PageQuery {
  /** ISO 8601 timestamps. */
  executed_at_start?: string;
  executed_at_end?: string;
  action_id?: string;
  node_id?: string;
  status?: ActionStatus;
}

export async function listActionLogs(
  query: ActionLogFilterQuery = {},
): Promise<PageDataResponse<ActionLogResponse>> {
  return apiFetch(`/action-logs${buildQuery(query)}`);
}

export async function deleteActionLogs(
  query: ActionLogFilterQuery = {},
): Promise<CountResponse> {
  return apiFetch(`/action-logs${buildQuery(query)}`, { method: "DELETE" });
}
```

- [ ] **Step 2: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
export PATH=/tmp/mate-node-v24.18.1/bin:$PATH
node_modules/.bin/tsc --noEmit 2>&1 | grep -v "TS2307\|e2e/\|tsbuildinfo\|actions.test.ts\|status-badge.test.tsx"
```

Expected: errors in `action-history/page.tsx` and
`action-history/_components/*` (still using the old shape) — fixed in
F3-F5. No errors outside the `action-history` directory (confirms
`telemetry`/`node-logs` don't import this file's now-changed exports).

- [ ] **Step 3: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add frontend/src/lib/api/action-logs.ts
git commit -m "feat(action-log): update API client types for status filter and real pagination"
```

---

## Task F2 (frontend): Action and Node search comboboxes

**Repo:** `mate-things` (frontend)

**Files:**
- Create: `frontend/src/components/actions/ActionSearchCombobox.tsx`
- Create: `frontend/src/components/nodes/NodeSearchCombobox.tsx`

**Interfaces:**
- Consumes: `listActions` (`@/lib/api/actions`), `listNodes`
  (`@/lib/api/nodes`) — both already support `{search, page, limit}`.
- Produces (consumed by F3): `<ActionSearchCombobox name defaultActionId?
  defaultActionName? label? placeholder? />`; `<NodeSearchCombobox name
  defaultNodeId? defaultNodeName? label? placeholder? />` — both render a
  hidden `<input name=... value={selectedId}>` for a surrounding
  `<form>`, exactly like `components/roles/RoleSearchCombobox.tsx`.

- [ ] **Step 1: Create `ActionSearchCombobox`**

Create `frontend/src/components/actions/ActionSearchCombobox.tsx`:

```tsx
"use client";

import { ChevronLeft, ChevronRight, Search, X } from "lucide-react";
import { useEffect, useId, useRef, useState } from "react";

import Input from "@/components/ui/input";
import { listActions, type ActionResponse } from "@/lib/api/actions";

const RESULTS_PER_PAGE = 6;
const DEBOUNCE_MS = 300;

interface ActionSearchComboboxProps {
  name: string;
  defaultActionId?: string;
  defaultActionName?: string;
  label?: string;
  placeholder?: string;
}

export default function ActionSearchCombobox({
  name,
  defaultActionId,
  defaultActionName,
  label = "Action",
  placeholder = "Any action",
}: ActionSearchComboboxProps) {
  const id = useId();
  const containerRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(1);
  const [actions, setActions] = useState<ActionResponse[]>([]);
  const [totalItems, setTotalItems] = useState(0);
  const [loading, setLoading] = useState(false);
  const [selectedId, setSelectedId] = useState(defaultActionId ?? "");
  const [selectedName, setSelectedName] = useState(defaultActionName ?? "");

  useEffect(() => {
    if (!open) return undefined;

    const timeout = setTimeout(() => {
      let cancelled = false;
      setLoading(true);
      listActions({ search: query || undefined, page, limit: RESULTS_PER_PAGE })
        .then((result) => {
          if (cancelled) return;
          setActions(result.data);
          setTotalItems(result.page.total_items);
        })
        .catch(() => {
          if (cancelled) return;
          setActions([]);
          setTotalItems(0);
        })
        .finally(() => {
          if (!cancelled) setLoading(false);
        });
      return () => {
        cancelled = true;
      };
    }, DEBOUNCE_MS);

    return () => clearTimeout(timeout);
  }, [open, query, page]);

  useEffect(() => {
    if (!open) return undefined;

    function handlePointerDown(event: MouseEvent) {
      if (!containerRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    return () => document.removeEventListener("mousedown", handlePointerDown);
  }, [open]);

  const totalPages = Math.max(1, Math.ceil(totalItems / RESULTS_PER_PAGE));

  function select(action: ActionResponse | null) {
    setSelectedId(action?.id ?? "");
    setSelectedName(action?.name ?? "");
    setOpen(false);
  }

  return (
    <div ref={containerRef} className="relative">
      <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
        {label}
      </span>
      <input type="hidden" name={name} value={selectedId} />
      <button
        type="button"
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={() => {
          setOpen((current) => !current);
          setPage(1);
        }}
        className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus flex min-h-11 w-full items-center justify-between gap-2 rounded-xl border px-3.5 py-2.5 text-left text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none"
      >
        <span className={selectedName ? "" : "text-foreground/40"}>
          {selectedName || placeholder}
        </span>
        <ChevronRight
          aria-hidden="true"
          className={`size-4 shrink-0 transition-transform ${open ? "rotate-90" : ""}`}
        />
      </button>

      {open ? (
        <div className="border-border bg-surface absolute z-20 mt-1.5 w-full min-w-64 rounded-xl border p-2 shadow-lg">
          <div className="relative">
            <Search
              aria-hidden="true"
              className="text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2"
            />
            <Input
              autoFocus
              className="pl-9"
              placeholder="Search actions"
              value={query}
              onChange={(event) => {
                setQuery(event.target.value);
                setPage(1);
              }}
              aria-label="Search actions"
            />
          </div>

          <ul className="mt-2 max-h-56 space-y-0.5 overflow-y-auto" role="listbox">
            <li>
              <button
                type="button"
                onClick={() => select(null)}
                className="hover:bg-highlight/40 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm transition-colors"
              >
                <X aria-hidden="true" className="text-muted-foreground size-3.5" />
                Any action
              </button>
            </li>
            {loading ? (
              <li className="text-muted-foreground px-2.5 py-2 text-sm">
                Searching…
              </li>
            ) : actions.length ? (
              actions.map((action) => (
                <li key={action.id}>
                  <button
                    type="button"
                    role="option"
                    aria-selected={action.id === selectedId}
                    onClick={() => select(action)}
                    className={`hover:bg-highlight/40 w-full rounded-lg px-2.5 py-2 text-left text-sm transition-colors ${
                      action.id === selectedId ? "bg-highlight/40 font-semibold" : ""
                    }`}
                  >
                    {action.name}
                  </button>
                </li>
              ))
            ) : (
              <li className="text-muted-foreground px-2.5 py-2 text-sm">
                No actions match.
              </li>
            )}
          </ul>

          <div className="border-border mt-2 flex items-center justify-between border-t pt-2">
            <span id={`${id}-page-info`} className="text-muted-foreground text-xs">
              Page {page} of {totalPages}
            </span>
            <div className="flex gap-1">
              <button
                type="button"
                aria-label="Previous actions"
                disabled={page <= 1 || loading}
                onClick={() => setPage((current) => Math.max(1, current - 1))}
                className="hover:bg-highlight/40 flex size-7 items-center justify-center rounded-lg transition-colors disabled:opacity-40"
              >
                <ChevronLeft aria-hidden="true" className="size-4" />
              </button>
              <button
                type="button"
                aria-label="Next actions"
                disabled={page >= totalPages || loading}
                onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
                className="hover:bg-highlight/40 flex size-7 items-center justify-center rounded-lg transition-colors disabled:opacity-40"
              >
                <ChevronRight aria-hidden="true" className="size-4" />
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
```

- [ ] **Step 2: Create `NodeSearchCombobox`**

Create `frontend/src/components/nodes/NodeSearchCombobox.tsx` — identical
structure, swapping the data source and labels:

```tsx
"use client";

import { ChevronLeft, ChevronRight, Search, X } from "lucide-react";
import { useEffect, useId, useRef, useState } from "react";

import Input from "@/components/ui/input";
import { listNodes, type NodeResponse } from "@/lib/api/nodes";

const RESULTS_PER_PAGE = 6;
const DEBOUNCE_MS = 300;

interface NodeSearchComboboxProps {
  name: string;
  defaultNodeId?: string;
  defaultNodeName?: string;
  label?: string;
  placeholder?: string;
}

export default function NodeSearchCombobox({
  name,
  defaultNodeId,
  defaultNodeName,
  label = "Node",
  placeholder = "Any node",
}: NodeSearchComboboxProps) {
  const id = useId();
  const containerRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(1);
  const [nodes, setNodes] = useState<NodeResponse[]>([]);
  const [totalItems, setTotalItems] = useState(0);
  const [loading, setLoading] = useState(false);
  const [selectedId, setSelectedId] = useState(defaultNodeId ?? "");
  const [selectedName, setSelectedName] = useState(defaultNodeName ?? "");

  useEffect(() => {
    if (!open) return undefined;

    const timeout = setTimeout(() => {
      let cancelled = false;
      setLoading(true);
      listNodes({ search: query || undefined, page, limit: RESULTS_PER_PAGE })
        .then((result) => {
          if (cancelled) return;
          setNodes(result.data);
          setTotalItems(result.page.total_items);
        })
        .catch(() => {
          if (cancelled) return;
          setNodes([]);
          setTotalItems(0);
        })
        .finally(() => {
          if (!cancelled) setLoading(false);
        });
      return () => {
        cancelled = true;
      };
    }, DEBOUNCE_MS);

    return () => clearTimeout(timeout);
  }, [open, query, page]);

  useEffect(() => {
    if (!open) return undefined;

    function handlePointerDown(event: MouseEvent) {
      if (!containerRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    return () => document.removeEventListener("mousedown", handlePointerDown);
  }, [open]);

  const totalPages = Math.max(1, Math.ceil(totalItems / RESULTS_PER_PAGE));

  function select(node: NodeResponse | null) {
    setSelectedId(node?.id ?? "");
    setSelectedName(node?.name || node?.device_id || "");
    setOpen(false);
  }

  return (
    <div ref={containerRef} className="relative">
      <span className="text-foreground/70 mb-1.5 block text-xs font-semibold">
        {label}
      </span>
      <input type="hidden" name={name} value={selectedId} />
      <button
        type="button"
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={() => {
          setOpen((current) => !current);
          setPage(1);
        }}
        className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus flex min-h-11 w-full items-center justify-between gap-2 rounded-xl border px-3.5 py-2.5 text-left text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none"
      >
        <span className={selectedName ? "" : "text-foreground/40"}>
          {selectedName || placeholder}
        </span>
        <ChevronRight
          aria-hidden="true"
          className={`size-4 shrink-0 transition-transform ${open ? "rotate-90" : ""}`}
        />
      </button>

      {open ? (
        <div className="border-border bg-surface absolute z-20 mt-1.5 w-full min-w-64 rounded-xl border p-2 shadow-lg">
          <div className="relative">
            <Search
              aria-hidden="true"
              className="text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2"
            />
            <Input
              autoFocus
              className="pl-9"
              placeholder="Search nodes"
              value={query}
              onChange={(event) => {
                setQuery(event.target.value);
                setPage(1);
              }}
              aria-label="Search nodes"
            />
          </div>

          <ul className="mt-2 max-h-56 space-y-0.5 overflow-y-auto" role="listbox">
            <li>
              <button
                type="button"
                onClick={() => select(null)}
                className="hover:bg-highlight/40 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm transition-colors"
              >
                <X aria-hidden="true" className="text-muted-foreground size-3.5" />
                Any node
              </button>
            </li>
            {loading ? (
              <li className="text-muted-foreground px-2.5 py-2 text-sm">
                Searching…
              </li>
            ) : nodes.length ? (
              nodes.map((node) => (
                <li key={node.id}>
                  <button
                    type="button"
                    role="option"
                    aria-selected={node.id === selectedId}
                    onClick={() => select(node)}
                    className={`hover:bg-highlight/40 w-full rounded-lg px-2.5 py-2 text-left text-sm transition-colors ${
                      node.id === selectedId ? "bg-highlight/40 font-semibold" : ""
                    }`}
                  >
                    {node.name || node.device_id}
                  </button>
                </li>
              ))
            ) : (
              <li className="text-muted-foreground px-2.5 py-2 text-sm">
                No nodes match.
              </li>
            )}
          </ul>

          <div className="border-border mt-2 flex items-center justify-between border-t pt-2">
            <span id={`${id}-page-info`} className="text-muted-foreground text-xs">
              Page {page} of {totalPages}
            </span>
            <div className="flex gap-1">
              <button
                type="button"
                aria-label="Previous nodes"
                disabled={page <= 1 || loading}
                onClick={() => setPage((current) => Math.max(1, current - 1))}
                className="hover:bg-highlight/40 flex size-7 items-center justify-center rounded-lg transition-colors disabled:opacity-40"
              >
                <ChevronLeft aria-hidden="true" className="size-4" />
              </button>
              <button
                type="button"
                aria-label="Next nodes"
                disabled={page >= totalPages || loading}
                onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
                className="hover:bg-highlight/40 flex size-7 items-center justify-center rounded-lg transition-colors disabled:opacity-40"
              >
                <ChevronRight aria-hidden="true" className="size-4" />
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
```

- [ ] **Step 3: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
export PATH=/tmp/mate-node-v24.18.1/bin:$PATH
node_modules/.bin/tsc --noEmit 2>&1 | grep -v "TS2307\|e2e/\|tsbuildinfo\|actions.test.ts\|status-badge.test.tsx"
node_modules/.bin/eslint \
  src/components/actions/ActionSearchCombobox.tsx \
  src/components/nodes/NodeSearchCombobox.tsx
```

- [ ] **Step 4: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add frontend/src/components/actions/ActionSearchCombobox.tsx \
        frontend/src/components/nodes/NodeSearchCombobox.tsx
git commit -m "feat(action-history): add Action and Node search comboboxes"
```

---

## Task F3 (frontend): status label module, filters, dedicated filter parser

**Repo:** `mate-things` (frontend)

**Files:**
- Create: `frontend/src/app/(authenticated)/action-history/_lib/status.ts`
- Create: `frontend/src/app/(authenticated)/action-history/_lib/filters.ts`
- Modify: `frontend/src/app/(authenticated)/action-history/_components/ActionLogFilters.tsx`
- Delete: `frontend/src/app/(authenticated)/action-history/_components/ActionLogCard.tsx`

**Interfaces:**
- Consumes: `ActionSearchCombobox`, `NodeSearchCombobox` (F2);
  `ActionStatus` (F1).
- Produces (consumed by F4, F5): `ACTION_STATUS_LABELS: Record<ActionStatus,
  {label: string; variant: StatusVariant}>` (`_lib/status.ts`);
  `parseActionHistoryFilters(raw): {filters: {start?, end?, actionId?,
  nodeId?, status?, executionId?}, error?: string}` (`_lib/filters.ts`);
  `<ActionLogFilters start? end? actionId? actionName? nodeId? nodeName?
  status? />` (no more `executionId` prop — removed).

- [ ] **Step 1: Create the shared status label module**

Create `frontend/src/app/(authenticated)/action-history/_lib/status.ts`:

```ts
import type { StatusVariant } from "@/components/ui/status-badge";
import type { ActionStatus } from "@/lib/api/action-logs";

export const ACTION_STATUS_LABELS: Record<
  ActionStatus,
  { label: string; variant: StatusVariant }
> = {
  UNEXECUTED: { label: "Unexecuted", variant: "neutral" },
  UNRESPONDED: { label: "Unresponded", variant: "warning" },
  FAILED: { label: "Failed", variant: "critical" },
  SUCCESS: { label: "Success", variant: "success" },
};
```

- [ ] **Step 2: Delete `ActionLogCard.tsx`**

Its card-grid rendering is fully replaced by the new table (Task F4); its
only reusable piece (the status label map) was just moved to
`_lib/status.ts` above. Confirmed via grep that nothing outside this
directory imports it.

```bash
rm "/home/dodol/Repositories/mate/mate-things/frontend/src/app/(authenticated)/action-history/_components/ActionLogCard.tsx"
```

- [ ] **Step 3: Create the dedicated filter parser**

Create `frontend/src/app/(authenticated)/action-history/_lib/filters.ts`
(deliberately separate from `@/lib/record-filters.ts` — `telemetry` and
`node-logs` keep using that shared module unmodified; this page's filter
shape has diverged enough — status, no free-text execution ID field — to
warrant its own parser rather than growing the shared one with
page-specific concerns):

```ts
import type { ActionStatus } from "@/lib/api/action-logs";

export interface ActionHistoryFilters {
  start?: string;
  end?: string;
  actionId?: string;
  nodeId?: string;
  status?: ActionStatus;
  executionId?: string;
}

export interface ParsedActionHistoryFilters {
  filters: ActionHistoryFilters;
  error?: string;
}

type RawActionHistoryParams = Record<string, string | string[] | undefined>;

const VALID_STATUSES: readonly ActionStatus[] = [
  "UNEXECUTED",
  "UNRESPONDED",
  "FAILED",
  "SUCCESS",
];

function first(value: string | string[] | undefined): string {
  return (Array.isArray(value) ? value[0] : (value ?? "")).trim();
}

function iso(
  value: string,
  field: string,
): { value?: string; error?: string } {
  if (!value) return {};
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return { error: `${field} must be a valid date and time.` };
  }
  return { value: date.toISOString() };
}

export function parseActionHistoryFilters(
  raw: RawActionHistoryParams,
): ParsedActionHistoryFilters {
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

  const statusRaw = first(raw.status).toUpperCase();
  const status = VALID_STATUSES.includes(statusRaw as ActionStatus)
    ? (statusRaw as ActionStatus)
    : undefined;
  if (!error && statusRaw && !status) {
    error = "Select a valid status.";
  }

  return {
    filters: {
      start: start.value,
      end: end.value,
      actionId: first(raw.action_id) || undefined,
      nodeId: first(raw.node_id) || undefined,
      status,
      executionId: first(raw.execution_id) || undefined,
    },
    error,
  };
}
```

- [ ] **Step 4: Rewrite `ActionLogFilters.tsx`**

Replace the whole file:

```tsx
import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import Input from "@/components/ui/input";

interface ActionLogFiltersProps {
  start?: string;
  end?: string;
  actionId?: string;
  nodeId?: string;
  executionId?: string;
}

export default function ActionLogFilters(props: ActionLogFiltersProps) {
  return (
    <form className="border-border bg-muted grid gap-3 rounded-2xl border p-4 sm:grid-cols-2 lg:grid-cols-3">
      <TimeRangeFilter start={props.start} end={props.end} />
      <Input
        name="action_id"
        defaultValue={props.actionId}
        placeholder="Action ID"
        aria-label="Action ID"
      />
      <Input
        name="node_id"
        defaultValue={props.nodeId}
        placeholder="Node ID"
        aria-label="Node ID"
      />
      <Input
        name="execution_id"
        defaultValue={props.executionId}
        placeholder="Execution ID"
        aria-label="Execution ID"
      />
      <div className="flex gap-2">
        <button className="bg-primary text-surface min-h-11 flex-1 rounded-xl px-4 text-sm font-semibold">
          Apply filters
        </button>
        <a
          href="/action-history"
          className="border-border text-primary inline-flex min-h-11 items-center rounded-xl border px-4 text-sm font-semibold"
        >
          Clear
        </a>
      </div>
    </form>
  );
}
```

with:

```tsx
import ActionSearchCombobox from "@/components/actions/ActionSearchCombobox";
import NodeSearchCombobox from "@/components/nodes/NodeSearchCombobox";
import TimeRangeFilter from "@/components/records/TimeRangeFilter";
import type { ActionStatus } from "@/lib/api/action-logs";

import { ACTION_STATUS_LABELS } from "../_lib/status";

interface ActionLogFiltersProps {
  start?: string;
  end?: string;
  actionId?: string;
  actionName?: string;
  nodeId?: string;
  nodeName?: string;
  status?: ActionStatus;
}

const STATUS_KEYS = Object.keys(ACTION_STATUS_LABELS) as ActionStatus[];

export default function ActionLogFilters(props: ActionLogFiltersProps) {
  return (
    <form className="border-border bg-muted grid gap-3 rounded-2xl border p-4 sm:grid-cols-2 lg:grid-cols-3">
      <TimeRangeFilter start={props.start} end={props.end} />
      <ActionSearchCombobox
        name="action_id"
        defaultActionId={props.actionId}
        defaultActionName={props.actionName}
      />
      <NodeSearchCombobox
        name="node_id"
        defaultNodeId={props.nodeId}
        defaultNodeName={props.nodeName}
      />
      <div>
        <label
          className="text-foreground/70 mb-1.5 block text-xs font-semibold"
          htmlFor="action-log-status"
        >
          Status
        </label>
        <select
          id="action-log-status"
          name="status"
          defaultValue={props.status ?? ""}
          className="border-control-border bg-background text-foreground focus-visible:border-focus focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 py-2.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
        >
          <option value="">Any status</option>
          {STATUS_KEYS.map((status) => (
            <option key={status} value={status}>
              {ACTION_STATUS_LABELS[status].label}
            </option>
          ))}
        </select>
      </div>
      <div className="flex gap-2">
        <button className="bg-primary text-surface min-h-11 flex-1 rounded-xl px-4 text-sm font-semibold">
          Apply filters
        </button>
        <a
          href="/action-history"
          className="border-border text-primary inline-flex min-h-11 items-center rounded-xl border px-4 text-sm font-semibold"
        >
          Clear
        </a>
      </div>
    </form>
  );
}
```

(No Execution ID field anywhere in this form — `execution_id` is parsed
from the URL by `_lib/filters.ts` independent of what this form renders,
so the dashboard's existing deep link keeps working with zero markup here
referencing it.)

- [ ] **Step 5: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
export PATH=/tmp/mate-node-v24.18.1/bin:$PATH
node_modules/.bin/tsc --noEmit 2>&1 | grep -v "TS2307\|e2e/\|tsbuildinfo\|actions.test.ts\|status-badge.test.tsx"
node_modules/.bin/eslint \
  "src/app/(authenticated)/action-history/_lib/status.ts" \
  "src/app/(authenticated)/action-history/_lib/filters.ts" \
  "src/app/(authenticated)/action-history/_components/ActionLogFilters.tsx"
```

Expected: remaining `tsc` errors only in `page.tsx` and `_lib/actions.ts`
(fixed in F5) — nothing outside `action-history`.

- [ ] **Step 6: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add "frontend/src/app/(authenticated)/action-history/_lib/status.ts" \
        "frontend/src/app/(authenticated)/action-history/_lib/filters.ts" \
        "frontend/src/app/(authenticated)/action-history/_components/ActionLogFilters.tsx"
git rm "frontend/src/app/(authenticated)/action-history/_components/ActionLogCard.tsx"
git commit -m "feat(action-history): searchable action/node filters, status select, drop Execution ID input"
```

---

## Task F4 (frontend): the expandable table

**Repo:** `mate-things` (frontend)

**Files:**
- Create: `frontend/src/app/(authenticated)/action-history/_components/ActionHistoryTable.tsx`

**Interfaces:**
- Consumes: `ActionLogResponse` (F1), `ACTION_STATUS_LABELS` (F3),
  `JsonPayload` (`@/components/records/JsonPayload`, existing/unchanged),
  `StatusBadge` (`@/components/ui/status-badge`, existing/unchanged).
- Produces (consumed by F5): `<ActionHistoryTable records={ActionLogResponse[]} />`.

- [ ] **Step 1: Create the component**

Create `frontend/src/app/(authenticated)/action-history/_components/ActionHistoryTable.tsx`:

```tsx
"use client";

import { ChevronRight } from "lucide-react";
import { Fragment, useState } from "react";

import JsonPayload from "@/components/records/JsonPayload";
import StatusBadge from "@/components/ui/status-badge";
import type { ActionLogResponse } from "@/lib/api/action-logs";

import { ACTION_STATUS_LABELS } from "../_lib/status";

const DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "UTC",
});

export default function ActionHistoryTable({
  records,
}: {
  records: readonly ActionLogResponse[];
}) {
  const [expandedId, setExpandedId] = useState<number | null>(null);

  return (
    <div className="border-border overflow-x-auto rounded-2xl border">
      <table className="w-full text-left text-sm">
        <thead className="bg-muted text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          <tr>
            <th scope="col" className="px-4 py-3">
              Timestamp
            </th>
            <th scope="col" className="px-4 py-3">
              Action
            </th>
            <th scope="col" className="px-4 py-3">
              Node
            </th>
            <th scope="col" className="px-4 py-3">
              Status
            </th>
          </tr>
        </thead>
        <tbody className="divide-border divide-y">
          {records.map((record) => {
            const expanded = expandedId === record.id;
            const status = ACTION_STATUS_LABELS[record.action_status];
            const nodeLabel =
              record.node_name || record.node_device_id || "Not assigned";

            return (
              <Fragment key={record.id}>
                <tr
                  onClick={() => setExpandedId(expanded ? null : record.id)}
                  aria-expanded={expanded}
                  className="hover:bg-highlight/20 cursor-pointer transition-colors"
                >
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <ChevronRight
                        aria-hidden="true"
                        className={`size-3.5 shrink-0 transition-transform ${expanded ? "rotate-90" : ""}`}
                      />
                      <time dateTime={record.executed_at}>
                        {DATE_FORMATTER.format(new Date(record.executed_at))}
                      </time>
                    </div>
                  </td>
                  <td className="px-4 py-3" title={record.action_id}>
                    {record.action_name}
                  </td>
                  <td className="px-4 py-3" title={record.node_id ?? undefined}>
                    {nodeLabel}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge variant={status.variant}>
                      {status.label}
                    </StatusBadge>
                  </td>
                </tr>
                {expanded ? (
                  <tr>
                    <td colSpan={4} className="bg-muted/50 px-4 py-4">
                      {record.action_message ? (
                        <p className="mb-3 text-sm whitespace-pre-wrap">
                          {record.action_message}
                        </p>
                      ) : null}
                      <JsonPayload value={record.payload} />
                    </td>
                  </tr>
                ) : null}
              </Fragment>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
```

- [ ] **Step 2: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
export PATH=/tmp/mate-node-v24.18.1/bin:$PATH
node_modules/.bin/tsc --noEmit 2>&1 | grep -v "TS2307\|e2e/\|tsbuildinfo\|actions.test.ts\|status-badge.test.tsx"
node_modules/.bin/eslint "src/app/(authenticated)/action-history/_components/ActionHistoryTable.tsx"
```

- [ ] **Step 3: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add "frontend/src/app/(authenticated)/action-history/_components/ActionHistoryTable.tsx"
git commit -m "feat(action-history): add the expandable Action History table"
```

---

## Task F5 (frontend): page rewrite + delete action

**Repo:** `mate-things` (frontend)

**Files:**
- Modify: `frontend/src/app/(authenticated)/action-history/page.tsx`
- Modify: `frontend/src/app/(authenticated)/action-history/_lib/actions.ts`

**Interfaces:**
- Consumes: everything from F1-F4.
- Produces: the finished page. `deleteActionHistoryAction` gains
  `status` in the delete filter it builds.

- [ ] **Step 1: Rewrite `page.tsx`**

Replace the whole file:

```tsx
import type { Metadata } from "next";

import ScopedDeleteDialog from "@/components/records/ScopedDeleteDialog";
import RecordWindow from "@/components/records/RecordWindow";
import RefreshBoundary from "@/components/refresh/RefreshBoundary";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listActionLogs } from "@/lib/api/action-logs";
import { activeFilterEntries, parseRecordFilters } from "@/lib/record-filters";
import { requirePermission } from "@/lib/session";

import ActionLogCard from "./_components/ActionLogCard";
import ActionLogFilters from "./_components/ActionLogFilters";
import { deleteActionHistoryAction } from "./_lib/actions";

export const metadata: Metadata = { title: "Action History — Mate Things" };
type RawSearchParams = Record<string, string | string[] | undefined>;

export default async function ActionHistoryPage({
  searchParams,
}: {
  searchParams: Promise<RawSearchParams>;
}) {
  const [raw, { permissions }] = await Promise.all([
    searchParams,
    requirePermission("action_log:get"),
  ]);
  const parsed = parseRecordFilters(raw);
  const query = {
    executed_at_start: parsed.filters.start,
    executed_at_end: parsed.filters.end,
    action_id: parsed.filters.actionId,
    node_id: parsed.filters.nodeId,
  };
  const result = parsed.error
    ? { data: [], total_items: 0 }
    : await listActionLogs(query);
  const records = parsed.filters.executionId
    ? result.data.filter(
        (log) => log.execution_id === parsed.filters.executionId,
      )
    : result.data;
  const deletionFilters = activeFilterEntries({
    start: parsed.filters.start,
    end: parsed.filters.end,
    action_id: parsed.filters.actionId,
    node_id: parsed.filters.nodeId,
  });
  const latest = records[0]?.created_at;

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Action History"
        description="Inspect execution outcomes and payloads across the fleet."
        actions={
          permissions.has("action_log:remove") &&
          !parsed.filters.executionId ? (
            <ScopedDeleteDialog
              action={deleteActionHistoryAction}
              filters={deletionFilters}
              label="action history"
            />
          ) : undefined
        }
      />
      <ActionLogFilters
        start={parsed.filters.start}
        end={parsed.filters.end}
        actionId={parsed.filters.actionId}
        nodeId={parsed.filters.nodeId}
        executionId={parsed.filters.executionId}
      />
      {parsed.error ? (
        <EmptyState title="Check the time range" description={parsed.error} />
      ) : (
        <>
          <div className="flex flex-wrap justify-between gap-3 text-sm">
            <p>
              <strong>
                {parsed.filters.executionId
                  ? records.length
                  : result.total_items}
              </strong>{" "}
              records
            </p>
            <p className="text-muted-foreground">
              Use a bounded time range for faster operational review.
            </p>
          </div>
          <RefreshBoundary updatedAt={latest}>
            {records.length ? (
              <RecordWindow
                records={records}
                getKey={(record) => record.id}
                renderRecord={(record) => <ActionLogCard log={record} />}
              />
            ) : (
              <EmptyState
                title="No action history found"
                description="Change the active filters or wait for a new action execution."
              />
            )}
          </RefreshBoundary>
        </>
      )}
    </main>
  );
}
```

with:

```tsx
import type { Metadata } from "next";
import { redirect } from "next/navigation";

import Pagination from "@/components/collection/Pagination";
import ScopedDeleteDialog from "@/components/records/ScopedDeleteDialog";
import RefreshBoundary from "@/components/refresh/RefreshBoundary";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { getActionById } from "@/lib/api/actions";
import { listActionLogs } from "@/lib/api/action-logs";
import { getNodeById } from "@/lib/api/nodes";
import {
  getOutOfRangePageRedirect,
  parsePageQuery,
} from "@/lib/collection-query";
import { activeFilterEntries } from "@/lib/record-filters";
import { requirePermission } from "@/lib/session";

import ActionHistoryTable from "./_components/ActionHistoryTable";
import ActionLogFilters from "./_components/ActionLogFilters";
import { deleteActionHistoryAction } from "./_lib/actions";
import { parseActionHistoryFilters } from "./_lib/filters";

export const metadata: Metadata = { title: "Action History — Mate Things" };
type RawSearchParams = Record<string, string | string[] | undefined>;

export default async function ActionHistoryPage({
  searchParams,
}: {
  searchParams: Promise<RawSearchParams>;
}) {
  const [raw, { permissions }] = await Promise.all([
    searchParams,
    requirePermission("action_log:get"),
  ]);
  const pageQuery = parsePageQuery(raw);
  const parsed = parseActionHistoryFilters(raw);

  const [actionDefault, nodeDefault] = await Promise.all([
    parsed.filters.actionId
      ? getActionById(parsed.filters.actionId).catch(() => null)
      : Promise.resolve(null),
    parsed.filters.nodeId
      ? getNodeById(parsed.filters.nodeId).catch(() => null)
      : Promise.resolve(null),
  ]);

  const result = parsed.error
    ? {
        data: [],
        page: { page: pageQuery.page, limit: pageQuery.limit, total_items: 0 },
      }
    : await listActionLogs({
        page: pageQuery.page,
        limit: pageQuery.limit,
        executed_at_start: parsed.filters.start,
        executed_at_end: parsed.filters.end,
        action_id: parsed.filters.actionId,
        node_id: parsed.filters.nodeId,
        status: parsed.filters.status,
      });

  if (!parsed.error) {
    const redirectTarget = getOutOfRangePageRedirect(
      "/action-history",
      raw,
      result.page,
    );
    if (redirectTarget) {
      redirect(redirectTarget);
    }
  }

  const records = parsed.filters.executionId
    ? result.data.filter(
        (log) => log.execution_id === parsed.filters.executionId,
      )
    : result.data;
  const deletionFilters = activeFilterEntries({
    start: parsed.filters.start,
    end: parsed.filters.end,
    action_id: parsed.filters.actionId,
    node_id: parsed.filters.nodeId,
    status: parsed.filters.status,
  });
  const latest = records[0]?.created_at;

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Action History"
        description="Inspect execution outcomes and payloads across the fleet."
        actions={
          permissions.has("action_log:remove") &&
          !parsed.filters.executionId ? (
            <ScopedDeleteDialog
              action={deleteActionHistoryAction}
              filters={deletionFilters}
              label="action history"
            />
          ) : undefined
        }
      />
      <ActionLogFilters
        start={parsed.filters.start}
        end={parsed.filters.end}
        actionId={parsed.filters.actionId}
        actionName={actionDefault?.name}
        nodeId={parsed.filters.nodeId}
        nodeName={nodeDefault?.name}
        status={parsed.filters.status}
      />
      {parsed.error ? (
        <EmptyState title="Check the time range" description={parsed.error} />
      ) : (
        <>
          <div className="flex flex-wrap justify-between gap-3 text-sm">
            <p>
              <strong>
                {parsed.filters.executionId
                  ? records.length
                  : result.page.total_items}
              </strong>{" "}
              records
            </p>
            <p className="text-muted-foreground">
              Use a bounded time range for faster operational review.
            </p>
          </div>
          <RefreshBoundary updatedAt={latest}>
            {records.length ? (
              <ActionHistoryTable records={records} />
            ) : (
              <EmptyState
                title="No action history found"
                description="Change the active filters or wait for a new action execution."
              />
            )}
          </RefreshBoundary>
          {!parsed.filters.executionId ? (
            <div className="border-border border-t pt-5">
              <Pagination
                page={result.page}
                pathname="/action-history"
                searchParams={raw}
              />
            </div>
          ) : null}
        </>
      )}
    </main>
  );
}
```

(`getActionById`/`getNodeById` already exist in `lib/api/actions.ts`/
`lib/api/nodes.ts`; `.catch(() => null)` handles a filter referencing a
since-deleted action/node gracefully — the combobox just shows its
placeholder instead of a resolved name, the filter itself still applies.
Pagination is suppressed entirely when viewing a single
`execution_id`-scoped result, since that's a fixed 0-or-1-row view, not a
real paginated list.)

- [ ] **Step 2: Update the delete server action**

In `frontend/src/app/(authenticated)/action-history/_lib/actions.ts`,
replace:

```ts
"use server";

import type { ScopedDeleteState } from "@/components/records/ScopedDeleteDialog";
import {
  deleteActionLogs,
  type ActionLogFilterQuery,
} from "@/lib/api/action-logs";
import { ApiError } from "@/lib/api/client";
import { parseRecordFilters } from "@/lib/record-filters";
import { requireSessionContext } from "@/lib/session";

export async function deleteActionHistoryAction(
  _state: ScopedDeleteState,
  formData: FormData,
): Promise<ScopedDeleteState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("action_log:remove")) {
    return {
      status: "error",
      message: "You do not have permission to delete action history.",
    };
  }
  const raw = Object.fromEntries(formData.entries()) as Record<string, string>;
  const parsed = parseRecordFilters(raw);
  if (parsed.error) return { status: "error", message: parsed.error };
  const filters: ActionLogFilterQuery = {
    executed_at_start: parsed.filters.start,
    executed_at_end: parsed.filters.end,
    action_id: parsed.filters.actionId,
    node_id: parsed.filters.nodeId,
  };
```

with:

```ts
"use server";

import type { ScopedDeleteState } from "@/components/records/ScopedDeleteDialog";
import {
  deleteActionLogs,
  type ActionLogFilterQuery,
} from "@/lib/api/action-logs";
import { ApiError } from "@/lib/api/client";
import { requireSessionContext } from "@/lib/session";

import { parseActionHistoryFilters } from "./filters";

export async function deleteActionHistoryAction(
  _state: ScopedDeleteState,
  formData: FormData,
): Promise<ScopedDeleteState> {
  const session = await requireSessionContext();
  if (!session.permissions.has("action_log:remove")) {
    return {
      status: "error",
      message: "You do not have permission to delete action history.",
    };
  }
  const raw = Object.fromEntries(formData.entries()) as Record<string, string>;
  const parsed = parseActionHistoryFilters(raw);
  if (parsed.error) return { status: "error", message: parsed.error };
  const filters: ActionLogFilterQuery = {
    executed_at_start: parsed.filters.start,
    executed_at_end: parsed.filters.end,
    action_id: parsed.filters.actionId,
    node_id: parsed.filters.nodeId,
    status: parsed.filters.status,
  };
```

(the rest of the file — the confirmation-phrase check and the
try/catch around `deleteActionLogs` — is unchanged).

- [ ] **Step 3: Verify**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
export PATH=/tmp/mate-node-v24.18.1/bin:$PATH
node_modules/.bin/tsc --noEmit 2>&1 | grep -v "TS2307\|e2e/\|tsbuildinfo\|actions.test.ts\|status-badge.test.tsx"
node_modules/.bin/eslint \
  "src/app/(authenticated)/action-history/page.tsx" \
  "src/app/(authenticated)/action-history/_lib/actions.ts"
```

Expected: fully clean — this is the last frontend file touched by this
plan.

- [ ] **Step 4: Commit**

```bash
cd /home/dodol/Repositories/mate/mate-things
git add "frontend/src/app/(authenticated)/action-history/page.tsx" \
        "frontend/src/app/(authenticated)/action-history/_lib/actions.ts"
git commit -m "feat(action-history): rewrite the page with real pagination and the new table"
```

---

## Task T6: full build + live verification

**Repo:** `mate-things` (both backend and frontend).

**Files:** none (verification only).

- [ ] **Step 1: Backend**

```bash
cd /home/dodol/Repositories/mate/mate-things/backend
go build ./...
go vet ./...
go test ./...
```

Expected: clean, aside from the one already-known pre-existing
`node_log` handler test failure (not a regression — confirmed unrelated
earlier this session via `git stash`).

- [ ] **Step 2: Frontend**

```bash
cd /home/dodol/Repositories/mate/mate-things/frontend
export PATH=/tmp/mate-node-v24.18.1/bin:$PATH
node_modules/.bin/tsc --noEmit 2>&1 | grep -v "TS2307\|e2e/\|tsbuildinfo\|actions.test.ts\|status-badge.test.tsx"
node_modules/.bin/eslint "src/app/(authenticated)/action-history/**/*.{ts,tsx}" \
  src/components/actions/ActionSearchCombobox.tsx \
  src/components/nodes/NodeSearchCombobox.tsx
```

Expected: fully clean.

- [ ] **Step 3: Rebuild and run the real stack**

```bash
cd /home/dodol/Repositories/mate/mate-things
docker compose build --no-cache app
docker compose up -d
```

(`--no-cache` — this environment's `docker compose build` has repeatedly
shown stale-cache false positives for genuine source changes this
session.)

- [ ] **Step 4: Live check with a real browser**

Log in and exercise `/action-history` end-to-end (a real browser session
— this session has a working recipe using `playwright-core` against the
already-cached Chromium binary at
`~/.cache/ms-playwright/chromium-1234/chrome-linux64/chrome`, reused from
the last live-verification pass):

- No Execution ID input visible anywhere on the page.
- The Action and Node fields are searchable dropdowns, not text inputs;
  typing filters results live (debounced), selecting one applies it.
- The Status dropdown filters results to only that status.
- Pagination controls appear, move between pages, and match the returned
  `page.total_items`/`page.limit`.
- The table shows Timestamp / Action / Node / Status columns with
  resolved names (not raw UUIDs) as the primary text; clicking a row
  expands it to show the message/payload, clicking again collapses it.
- Navigate directly to `/action-history?execution_id=<a real
  execution_id>&start=<its executed_at>` (simulating the dashboard's
  "Urgent attention" link) and confirm it narrows to exactly that one
  row, with no Execution ID input rendered.
- With a status filter active, open the scoped delete dialog and confirm
  its "deletion scope" summary lists the active status.
- Check the container logs for the duration of this check — zero new
  errors (matches the standard this session has held for every prior
  live-verification pass).
