# Action History page refactor — design

## Context

The Action History page (`/action-history`) currently shows a card grid,
one card per `action_logs` row, with a manual filter form (time range,
free-text Action ID / Node ID / Execution ID inputs). The backend's
`/v1/action-logs` list endpoint has **no server-side pagination at all** —
it returns every row matching the filter, unbounded, and the frontend
relies on client-side "show more" windowing (`RecordWindow`) plus a
required time range to keep result sets bounded.

This refactor:

1. Drops the manual Execution ID filter field (kept working as a
   URL-only, hidden filter — see below).
2. Adds real server-side pagination with searchable dropdowns for the
   Action and Node filters, replacing free-text UUID entry.
3. Adds a status filter (`action_status`), requiring changes through the
   backend's repository contract, infrastructure (Postgres query),
   application (usecase), and presentation (HTTP handler) layers.
4. Replaces the card grid with an actual `<table>` — the first real
   HTML table in this frontend — with columns Timestamp, Action, Node,
   Status, each row expandable in place to reveal the action message and
   JSON payload.

No test cases are being planned for this work (explicit scope decision) —
verification is build/compile/lint plus a live check against the running
stack, matching this session's established bar for backend+frontend work.

## Decisions confirmed with the user before writing this spec

- The dashboard's "Urgent attention" widget deep-links a failed action to
  `/action-history?execution_id=...&start=...`. Removing the *manual*
  Execution ID input must not break that link — `execution_id` stays a
  recognized (but no longer user-facing) query param that narrows results
  to one execution when present.
- The new Action/Node table columns show the resolved **name** as primary
  text (action's `name`, node's `name`/`device_id`), with the raw UUID
  available as a secondary detail (e.g. `title` attribute), not shown as
  the primary literal UUID text the request's wording implied.
- Payload/message detail (shown inline on today's cards) moves to an
  **expandable table row** — click a row to reveal the message and JSON
  payload beneath it, click again to collapse. No separate detail page or
  dialog.

## Backend changes

### 1. Repository contract (`internal/domain/contracts/repository/action_log.go`)

`ReadByFilter` and `DeleteByFilter` both gain an `actionStatus
*domainmodels.ActionStatus` parameter (nil = no status filter, matching
the existing nil-means-unfiltered convention for the other filter params).

`ReadByFilter` additionally gains `page int, limit int` and its return
type changes from `[]domainmodels.ActionLog` to a new
`[]domainmodels.ActionLogListItem`:

```go
type ActionLogListItem struct {
	ActionLog
	ActionName   string
	NodeDeviceId *string
	NodeName     *string
}
```

(`domain/models/action_log.go`). This is a read-model specific to the list
query — `ActionLog` itself (used by `Create` and the single-dispatch
response) is untouched, so nothing about action dispatch or the existing
`ActionLog()` response mapper for a single log needs to change.
`NodeDeviceId`/`NodeName` are pointers because `node_id` itself is
nullable (an action log can predate/lack a resolved node).

`DeleteByFilter` does **not** get `page`/`limit` — "delete everything
matching this filter" is intentionally unbounded, same as today.

### 2. Infrastructure (`internal/infrastructure/repository/action_log/postgres_query.go`)

- `applyActionLogFilters` gains an `actionStatus *domainmodels.ActionStatus`
  parameter; when set, adds `WHERE action_status = ?` to both the data
  and count queries (mirrors every other filter field already there).
- `queryReadByFilter` joins in the two name columns:
  ```sql
  SELECT action_logs.*, actions.name AS action_name,
         nodes.device_id AS node_device_id, nodes.name AS node_name
  FROM action_logs
  LEFT JOIN actions ON actions.id = action_logs.action_id
  LEFT JOIN nodes ON nodes.id = action_logs.node_id
  ```
  (expressed via squirrel's `.Join(...)`/`.LeftJoin(...)` builder, matching
  the query-builder style already used throughout this file — not raw SQL
  string concatenation.) `LEFT JOIN` on both, since `node_id` is nullable
  and — defensively — `action_id` referencing a soft-deleted action
  should still show *a* name via the join rather than silently dropping
  the row.
- Adds `.Limit(uint64(limit)).Offset(uint64((page-1)*limit))` to the data
  query only (never the count query).
- `queryDeleteByFilter` gains the same status `WHERE` clause as the read
  query, no pagination (unchanged shape otherwise).
- `internal/infrastructure/repository/shared/scan_pgx.go` already has
  `ScanPgxActionLog(row pgx.Row) (domainmodels.ActionLog, error)` (scans
  the 9 `actionLogColumns` in a fixed order) and `ScanPgxActionLogs(rows
  pgx.Rows) ([]domainmodels.ActionLog, error)`. Add
  `ScanPgxActionLogListItem`/`ScanPgxActionLogListItems` alongside them,
  scanning the same 9 columns into the embedded `ActionLog` plus the 3
  joined columns (in the same order they're selected in) into
  `ActionName`/`NodeDeviceId`/`NodeName`.

### 3. Application (`internal/application/action/history/usecase.go`,
   `internal/domain/usecases/action/history.go`)

- `ReadActionLogsByFilterRequest` gains `ActionStatus
  *domainmodels.ActionStatus`, `Page int`, `Limit int`.
- `DeleteActionLogsByFilterRequest` gains `ActionStatus
  *domainmodels.ActionStatus`.
- `History.ReadByFilter`'s return type changes to
  `([]domainmodels.ActionLogListItem, int, error)`.
- Both usecase methods thread the new field(s) straight through to the
  repository call, same as every existing field.

### 4. Presentation

- `internal/presentation/http/handler/action/handler.go`'s
  `actionLogFilter` (shared by both the GET and DELETE handlers) parses a
  new `status` query param via a small local helper function,
  `actionLogStatus(value *string) (*domainmodels.ActionStatus, error)`,
  mirroring `internal/presentation/http/handler/node_log/handler.go`'s
  existing `nodeLogLevel` helper exactly (a package-local function, not a
  shared `presentationhttputils` addition — that's the established
  precedent for this kind of enum-string query param in this codebase):
  absent/empty → `nil`; present but not one of the 4
  `domainmodels.ActionStatus` constants → a `domainmodels.ErrTypeValidation`
  error ("action_status must be one of UNEXECUTED, UNRESPONDED, FAILED,
  SUCCESS"), which the existing `presentationhttputils.Error(c, err)` call
  already turns into a 400.
- `ActionLogGetList` switches from manually building
  `ReadActionLogsByFilterRequest` + `CountDataResponse` to also calling
  `presentationhttputils.PageArgs(c)` for `page`/`limit` (same helper
  `nodes`/`actions`/`roles` list endpoints already use — `page`/`limit`
  are **not** part of `actionLogFilter`, they're orthogonal to the
  time/action/node/status filter fields) and returns
  `presentationhttpresponse.PageDataResponse[ActionLogResponse]` instead
  of `CountDataResponse[...]`, using `presentationhttputils.PageResponse(page,
  total)` to build the `page` envelope field — bringing this endpoint in
  line with every other paginated list endpoint in this API instead of
  its current one-off unbounded shape.
- `ActionLogDelete` gains the `status` field on its filter (already
  covered by reusing `actionLogFilter`/`actionLogDeleteFilter`) — no
  envelope shape change, `DeleteByFilter` isn't paginated.
- `internal/presentation/http/response/action_log.go`: `ActionLogResponse`
  gains `action_name string`, `node_device_id *string`, `node_name
  *string`. `ActionLog()` (the single-item mapper, still used for the
  dispatch-response path) leaves these three fields as their zero values —
  a single freshly-dispatched action log's response never needed name
  enrichment before and doesn't need it now; only the new
  `ActionLogListItem`-based list mapper (a new `ActionLogListItems([]domainmodels.ActionLogListItem)
  []ActionLogResponse` function, or equivalent) populates them.
- Swagger comments on `ActionLogGetList`/`ActionLogDelete` updated: new
  `status` query param documented; `ActionLogGetList`'s `@Success`
  annotation changes from `CountDataResponse[...]` to
  `PageDataResponse[...]`.

### Global backend constraints for this work

- `action_status` is a Postgres enum (`action_status` type,
  `UNEXECUTED|UNRESPONDED|FAILED|SUCCESS`) — the Go-side validation set
  must match exactly, reusing `domainmodels.ActionStatus`'s existing 4
  constants, not a new hardcoded list.
- No migration needed — `action_status`, `actions.name`, `nodes.device_id`,
  `nodes.name` all already exist; this is a query/response shape change
  only.
- Every existing filter field (time range, action_id, node_id) keeps its
  current behavior byte-for-byte; only status and pagination are new.

## Frontend changes

### 1. `_lib` — new dedicated query parsing (does not touch the shared
   `lib/record-filters.ts`, which `telemetry` and `node-logs` still use
   unmodified)

Action History gets its own param parsing (new file, e.g.
`_lib/filters.ts`, or inline in `page.tsx` if small enough once written)
covering: `page`, `limit` (via the existing `parsePageQuery`, reused
as-is), `start`/`end` (reusing the same ISO-parse-and-validate logic
`record-filters.ts`'s `iso()` helper already encodes — duplicate the small
helper rather than coupling the two pages through a shared module, since
their overall shapes now diverge), `action_id`, `node_id`, `status`, and
`execution_id` (parsed but never rendered as an editable field — see next
section).

### 2. `_components/ActionLogFilters.tsx` — rewritten

- Time range: unchanged (`TimeRangeFilter`, already generic).
- Action ID / Node ID free-text inputs → two new combobox components,
  built on the **exact existing pattern** in
  `components/roles/RoleSearchCombobox.tsx` (debounced 300ms search,
  6-per-page results, hidden input carrying the selected UUID for the
  surrounding `<form>`):
  - `components/actions/ActionSearchCombobox.tsx` — wraps
    `listActions({ search, page, limit })` (already supports `search`,
    confirmed via the existing `/v1/actions` handler already calling
    `presentationhttputils.PageArgs`), displays each option as the
    action's `name`.
  - `components/nodes/NodeSearchCombobox.tsx` — wraps
    `listNodes({ search, page, limit })` (same — `/v1/nodes` already
    supports `search`), displays each option as the node's `name`
    (falling back to `device_id` if a node has no name set — check
    whether `name` can be empty for self-registered nodes during
    implementation and handle accordingly).
- New Status `<select>`: `Any status | Unexecuted | Unresponded | Failed |
  Success` — reuses the same label set already defined in
  `_components/ActionLogCard.tsx`'s `STATUS` map (export and import it
  rather than redefining the labels a second time).
- Execution ID input is **removed** from the rendered form entirely. No
  hidden input is needed either — since the filter parser (previous
  section) reads `execution_id` straight from the URL search params
  independent of what the form renders, the dashboard's existing link
  still works with zero markup changes here.

### 3. `page.tsx` — rewritten

- Calls `listActionLogs` with the new paginated query shape (`page`,
  `limit`, `start`, `end`, `action_id`, `node_id`, `status`) instead of
  today's filter-only call; drops the `execution_id`-based client-side
  `.filter(...)` post-processing (now that `execution_id` is a first-class
  concern, decide during implementation whether it stays a client-only
  post-filter over an otherwise-normal paginated query, or whether it's
  worth a dedicated backend query param — **default to keeping it a
  client-side post-filter over the current page's results**, same as
  today, since it's a narrow "jump to one specific execution from a
  dashboard link" case, not a general filter worth new backend surface).
- Drops `RecordWindow` entirely (no longer needed once real pagination
  exists) in favor of the new `ActionHistoryTable` component (below) +
  the existing `components/collection/Pagination` component (same
  `page`/`searchParams`/`pathname` props the Nodes/Actions pages already
  pass).
- `ScopedDeleteDialog`'s `deletionFilters` gains `status` when present
  (alongside the existing start/end/action_id/node_id), so the delete
  confirmation dialog's "deletion scope" summary reflects it; the delete
  server action (`_lib/actions.ts`'s `deleteActionHistoryAction`) parses
  `status` the same way and passes it through to `deleteActionLogs`.

### 4. New `_components/ActionHistoryTable.tsx` (replaces
   `ActionLogCard.tsx`'s grid usage; the labels/status-badge logic in
   `ActionLogCard.tsx` is reused, not duplicated)

- A real `<table>`: `<thead>` with Timestamp / Action / Node / Status;
  one `<tr>` per log row (client component, since it needs expand/collapse
  state), a second, conditionally-rendered `<tr>` immediately below it
  holding the expanded detail (action message + `JsonPayload`) spanning
  all 4 columns, toggled on row click (mirrors the chevron-rotate
  affordance already established in `RoleSearchCombobox`).
- Action/Node cells: primary text is the resolved name
  (`log.action_name`, `log.node_name ?? log.node_device_id ?? "Not
  assigned"`), the raw UUID goes on a `title` attribute on the cell for
  anyone who needs to copy/inspect it.
- Status cell: reuses `StatusBadge`/the existing `STATUS` label map
  as-is.

### 5. `lib/api/action-logs.ts` — updated types

- `ActionLogResponse` gains `action_name: string`, `node_device_id?:
  string`, `node_name?: string`.
- `ActionLogFilterQuery` extends `PageQuery` (gains `page`/`limit`/drops
  its implicit "unbounded" framing) and gains `status?: ActionStatus`.
- `listActionLogs`'s return type changes from
  `Promise<CountDataResponse<ActionLogResponse>>` to
  `Promise<PageDataResponse<ActionLogResponse>>`, matching the backend
  envelope change; the file's existing "filter-only, no pagination" doc
  comment is corrected/removed since it's no longer true.
- `deleteActionLogs`'s `ActionLogFilterQuery` param picks up `status` for
  free (same type), no signature change needed there beyond the type
  extension.

## Verification

No new automated tests are being planned for this feature (explicit scope
decision). Verification is:

- **Backend**: `go build ./...`, `go vet ./...`, `go test ./...` (existing
  suite must stay green — note the one already-known, pre-existing,
  unrelated `node_log` handler test failure documented earlier this
  session; don't treat it as a regression caused by this work, but also
  don't let *new* failures hide behind that expectation).
- **Frontend**: `tsc --noEmit`, `eslint` on every touched/created file.
- **Live check**: rebuild and run the actual Docker stack (`docker compose
  build --no-cache app && docker compose up -d`), then exercise the real
  page end-to-end (a real browser session, not just curl) — apply each
  filter (time range, action combobox, node combobox, status), confirm
  pagination controls work and match the returned `page` envelope,
  confirm the dashboard's existing `?execution_id=...` deep link still
  narrows to one row with no visible Execution ID input anywhere, expand
  a table row and confirm the message/payload render, and confirm the
  scoped delete dialog's summary reflects an active status filter when
  set.
