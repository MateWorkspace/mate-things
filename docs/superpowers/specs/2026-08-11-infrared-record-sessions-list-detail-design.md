# Infrared Record Sessions: List + Detail (Spec A)

## Context

The Infrared app's Record page (`frontend/src/app/(authenticated)/apps/infrared/record/page.tsx`)
is currently a stub — "Coming soon." The backend, however, already has the
full record-session lifecycle built and wired to HTTP: starting a session,
listing its cases, accepting/discarding/retrying raw captures, reading the
generated coder, listing/transmitting test cases, and deleting any of the
above — plus a real-time WebSocket broadcast endpoint
(`GET /infrared/record-sessions/{id}/broadcast`) for live status updates.
The one missing piece is a **paginated list** of sessions — today only
`GET /infrared/record-sessions/{id}` (single, by id) exists.

This spec covers building the list + filter + detail pages that let a user
browse and manage existing sessions, following this codebase's established
list/detail conventions (`action-history` for list+filters+table,
`nodes/[id]` for a rich per-entity detail page). It deliberately does not
cover starting a new session or the WebSocket live view — that's Spec B
(the "Record New Device" wizard), designed separately.

## 1. Backend: list endpoint

**New repository method**, `internal/domain/contracts/repository/infrared_record_session.go`:

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

Same shape as `ActionLog.ReadByFilter` (`internal/domain/contracts/repository/action_log.go`).
The Postgres implementation joins `infrared_record_session` →
`infrared_device` → `infrared_device_type` so brand/model/type-name come
back without N+1 lookups — same pattern as `ActionLogListItem`'s join
against `nodes`.

**New domain read-model**, `internal/domain/models/infrared.go`:

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

**New usecase method** on `RecordSessionManagement`
(`internal/domain/usecases/infrared/record_session_management.go`):

```go
ListByFilter(ctx context.Context, request ListRecordSessionsRequest) (items []domainmodels.InfraredRecordSessionListItem, total int, err error)

type ListRecordSessionsRequest struct {
    RecordingState       *string
    InfraredDeviceTypeId *uuid.UUID
    CreatedAtStart       *time.Time
    CreatedAtEnd         *time.Time
    Page                 int
    Limit                int
}
```

Validates `RecordingState` against the known `InfraredRecordingState*`
constants when present (same pattern as other enum-filter validation in
this codebase), otherwise passes through to the repository unchanged — no
new business logic beyond filtering.

**New route**: `GET /v1/infrared/record-sessions`, permission
`infrared_record_session:get` (already exists, reused). Handler follows
`ActionLogGetList`'s exact shape: `presentationhttputils.PageArgs(c)` for
page/limit, `presentationhttputils.QueryUUID`/`QueryString`/query-time
helpers for the filters, response wrapped in the existing
`presentationhttpresponse.PageDataResponse[InfraredRecordSessionListItemResponse]`.

**New response type**, `internal/presentation/http/response/infrared.go`:

```go
type InfraredRecordSessionListItemResponse struct {
    Id             string    `json:"id"`
    RecordingState string    `json:"recording_state"`
    IsCompleted    bool      `json:"is_completed"`
    Brand          string    `json:"brand"`
    Model          string    `json:"model"`
    DeviceTypeName string    `json:"device_type_name"`
    CreatedAt      time.Time `json:"created_at"`
}
```

## 2. Frontend: API client

`frontend/src/lib/api/infrared.ts` gains, all following this file's
existing `apiFetch`/`server-only` conventions:

- `listInfraredRecordSessions(query): Promise<PageDataResponse<InfraredRecordSessionListItemResponse>>`
- `getInfraredRecordSession(id): Promise<InfraredRecordSessionResponse>`
- `listInfraredRecordSessionCases(id): Promise<InfraredRecordCaseResponse[]>`
- `getInfraredRecordSessionCoder(id): Promise<InfraredStateCoderResponse | null>`
- `listInfraredRecordSessionTestCases(id): Promise<InfraredTestCaseResponse[]>`
- `acceptRecordRaw`, `discardRecordRaw`, `retryRecordCase` (mutations, existing endpoints)
- `deleteRecordSession`, `deleteRecordCase`, `deleteRecordState`, `deleteRecordRaw`, `deleteStateCoder`, `deleteTestCase`, `deleteTestCaseState`

Response TypeScript types mirror whatever shape the existing (already
shipped) handlers already return — read the real handler/response Go code
for each endpoint before typing the client, rather than guessing field
names.

## 3. Frontend: list page

`record/page.tsx`, rewritten (server component, same skeleton as
`action-history/page.tsx`):

- `PageHeader` — title "Record", a "Record New Device" button in `actions`
  (routes to `/apps/infrared/record/new`; until Spec B ships, this route
  doesn't exist yet, so the button is omitted entirely rather than linking
  to a 404 — added back when Spec B lands).
- `RecordSessionFilters` (`_components/`, client component): status
  select (9 `InfraredRecordingState` values), device type select
  (populated via `listInfraredDeviceTypes`), created-date range — same
  structure as `ActionLogFilters`.
- `RecordSessionsTable` (`_components/`, server-rendered): columns
  device (brand/model), device type, status badge, created date; each row
  a `Link` to `/apps/infrared/record/[id]`.
- `Pagination`, `EmptyState` for zero results — same as every other list
  page in this codebase.
- `_lib/filters.ts` (parse/validate query params, same shape as
  `action-history/_lib/filters.ts`) and `_lib/status.ts`
  (`RECORDING_STATE_LABELS: Record<InfraredRecordingState, string>` +
  `StatusBadge` variant per state, same pattern as `ACTION_STATUS_LABELS`).

## 4. Frontend: detail page

`record/[id]/page.tsx` (new, server component, same skeleton as
`nodes/[id]/page.tsx`):

- **Overview card**: brand/model, device type, status badge, created
  timestamp, a delete-session control (client component, existing
  `infrared_record_session:delete` permission).
- **Cases section**: every recorded case, its target states, and its raw
  captures with accept/discard/retry controls (client components wrapping
  the existing endpoints). This block is written as a standalone,
  reusable component (`RecordCaseList` or similar) — Spec B's wizard will
  reuse it for its own in-progress recording/re-record steps rather than
  duplicating the same case/raw UI.
- **Coder section** (only rendered if `getInfraredRecordSessionCoder`
  returns non-null): encoder/decoder source in read-only code blocks,
  summary/detail README text, delete-coder control.
- **Test cases section** (only rendered if any exist): target state per
  test case, pass/fail status, a transmit control.
- All destructive controls gated on `infrared_record_session:delete`;
  mutation controls (accept/discard/retry/transmit) gated on
  `infrared_record_session:set`.

## Out of scope

- Starting a new recording session (Spec B).
- The WebSocket broadcast channel / any live-updating view (Spec B).
- Any change to backend session-lifecycle logic — this spec only adds a
  list/filter read path around what already exists.
