# Action ↔ Node Class many-to-many design

## Context

Today `actions.node_class_id` is a required single foreign key: every action
belongs to exactly one node class, and dispatch compatibility is a scalar
equality check (`node.NodeClassId == action.NodeClassId`,
`internal/application/action/execution/usecase.go:95`). The user wants
actions to be reusable across multiple node classes, with a node class's
detail page able to toggle which actions it accepts.

This is a direct structural mirror of the existing `role_permission` pivot
table already in this codebase (roles ↔ permissions, many-to-many). Every
layer below follows that precedent exactly — package names, method
shapes, cache-aside pattern, handler routes, and the frontend
checkbox-diff-and-save UX from Access Control's `RoleDetails.tsx` +
`PermissionGroups.tsx`.

Decisions confirmed with the user before writing this spec:
- The relationship is managed **only from the Node Class detail page**
  (toggle actions on/off for a class, "Save assignments" button). The
  Action page shows a **read-only** compatible-class count, no toggle
  there — asymmetric by design, matching how `PermissionCard.tsx` doesn't
  manage its roles.
- No data backfill: the database has already been reset, so the migration
  just drops the old column and creates the pivot table clean, with no
  concern for preserving existing action→node_class rows.
- The Actions list page's `node_class_id` filter and `ActionCard`'s
  node-class display both need to change since neither can rely on a
  scalar `action.node_class_id` anymore (raised by the user after the
  initial design pass).

## 1. Data model

Drop `actions.node_class_id` (column, FK, index) entirely. Add a new pivot
table structurally identical to `role_permission`:

```sql
CREATE TABLE node_class_action (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    node_class_id UUID NOT NULL REFERENCES node_classes (id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES actions (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    CONSTRAINT uq_node_class_action_node_class_id_action_id UNIQUE (node_class_id, action_id)
);

CREATE INDEX idx_node_class_action_node_class_id ON node_class_action (node_class_id);

CREATE INDEX idx_node_class_action_action_id ON node_class_action (action_id);
```

New domain error `ErrTypeNodeClassActionExists = errors.New("NODE_CLASS_ACTION_EXISTS")`
(`internal/domain/models/error.go`, alongside `ErrTypeRolePermissionExists`),
mapped from the unique-constraint violation via `ConflictMatch`.

The migration's down script drops `node_class_action` and re-adds
`actions.node_class_id` as a nullable column (no FK/index/NOT NULL —
structure can't be meaningfully restored without data, matching how other
irreversible-data migrations in this repo handle down scripts).

## 2. Backend layers — new `node_class_action` stack

One-to-one mirror of the `role_permission` stack, swapping
role→node_class, permission→action:

- `internal/domain/models/node_class_action.go` —
  `NodeClassAction{Id, NodeClassId, ActionId, CreatedAt, CreatedBy}`
- `internal/domain/contracts/repository/node_class_action.go` —
  `Create/ReadById/ReadByNodeClassIdAndActionId/ReadByPagination/DeleteById/DeleteByNodeClassIdAndActionId`,
  reads returning `(nodeClassAction, nodeClass, action, err)` triples
- `internal/domain/contracts/cache/node_class_action.go` — same shape as
  `contracts/cache/role_permission.go` (`GetById/SetById/DeleteById`,
  `Get/Set/DeleteByNodeClassIdAndActionId`, `Get/SetPagination`,
  `InvalidatePagination`, `InvalidateByNodeClassId`,
  `InvalidateByActionId`, `InvalidateAll`)
- `internal/infrastructure/cache/node_class_action/redis.go` — mirrors
  `infrastructure/cache/role_permission/redis.go`
- `internal/infrastructure/repository/node_class_action/{postgres.go,postgres_query.go}` —
  mirrors `infrastructure/repository/role_permission/*`: joins to
  `node_classes`/`actions`, same `baseReadQuery`/pagination/scan pattern,
  `Create` maps the unique-constraint conflict to
  `ErrTypeNodeClassActionExists`
- `internal/domain/usecases/repocache/node_class_action.go` — one-line
  interface embedding the repository contract (mirrors
  `usecases/repocache/role_permission.go`)
- `internal/application/repocache/node_class_action/usecase.go` —
  cache-aside decorator. On any write, invalidates itself **and** the
  owning node class's cached action list (`nodeClassCache.InvalidateActions`
  or a scoped `DeleteActions(nodeClassId)`). This fan-out is narrower than
  `role_permission`'s (which also invalidates a `userCache` for computed
  auth grants) — there's no downstream "computed compatibility" cache here
  beyond the node class's own action list.

## 3. `NodeClass` gains a direct joined read method

Exactly like `Role.ReadPermissions` (a method on the Role repository
itself, not routed through the `role_permission` package), `NodeClass`
gets:

- `internal/domain/contracts/repository/node_class.go`:
  `+ ReadActions(ctx, nodeClassId uuid.UUID) ([]domainmodels.Action, error)`
- `internal/domain/contracts/cache/node_class.go`:
  `+ GetActions/SetActions/DeleteActions/InvalidateActions`
- `internal/infrastructure/repository/node_class/postgres_query.go`:
  `+ queryReadActions` —
  `SELECT a.* FROM node_class_action nca JOIN actions a ON a.id = nca.action_id WHERE nca.node_class_id = ? AND a.deleted_at IS NULL ORDER BY a.created_at DESC, a.id ASC`,
  scanned with the existing `ScanPgxActions` shared helper (once
  `node_class_id` is dropped from the `actions` column list it scans)
- `internal/infrastructure/cache/node_class/redis.go` +
  `internal/application/repocache/node_class/usecase.go`: add the
  cache-aside `ReadActions` method mirroring Role's `ReadPermissions`

`internal/domain/usecases/node/class_management.go`'s `ClassManagement`
interface gains `AssignAction`/`RevokeAction`/`ReadActions`, backed by the
new `nodeClassActionRepoCache` + `nodeClass.ReadActions` — mirrors
`role_management`'s `AssignPermission`/`RevokePermission`/`ReadPermissions`
exactly.

## 4. Action loses `NodeClassId`, gains a list-view enrichment

`NodeClassId` is removed from: `domain/models/action.go`,
`domain/usecases/action/definition.go` (Create/ReadByPagination/Update
request structs), `domain/contracts/repository/action.go`,
`domain/contracts/cache/action.go`, `infrastructure/repository/action/*`,
`application/repocache/action/usecase.go`,
`presentation/http/request/action.go`,
`presentation/http/response/action.go`,
`presentation/http/handler/action/handler.go`.

For the Actions **list** view, a new enrichment mirrors the
`ActionLogListItem` pattern already established in this codebase (Action
History refactor): a joined/derived read model distinct from the base
`Action`.

- `internal/domain/models/action.go`: add
  `ActionListItem{Action, CompatibleNodeClassCount int}`
- `ReadByPagination` (repository contract, infra, repocache, usecase, all
  the way to the handler) returns `[]ActionListItem` instead of
  `[]Action`. The query adds a scalar subquery:
  `(SELECT COUNT(*) FROM node_class_action nca WHERE nca.action_id = a.id) AS compatible_node_class_count`
- `presentation/http/response/action.go`: `ActionResponse` gains
  `CompatibleNodeClassCount int`, populated only by the list mapper
  (`ActionListItem()`/`ActionListItems()`, mirroring
  `ActionLogListItem()`/`ActionLogListItems()`); the single-item
  `Action()` mapper leaves it at zero value, matching how
  `ActionLogResponse.ActionName` is only populated by the list path today.

## 5. `ActionGetList`'s `node_class_id` filter changes semantics

The existing `node_class_id` query param on `GET /actions` (and its
threading through `ReadActionsByPaginationRequest`,
`queryReadByPagination`) changes from `WHERE actions.node_class_id = ?` to:

```sql
WHERE EXISTS (
    SELECT 1 FROM node_class_action nca
    WHERE nca.action_id = a.id AND nca.node_class_id = ?
)
```

Same param name, same query-string shape, same frontend `Input` field on
`actions/page.tsx` — "show actions compatible with this node class" reads
identically to a user; only the backing implementation changes.

## 6. Dispatch compatibility check

`action/execution/usecase.go`'s `Dispatch` gets a new
`nodeClassAction domainusecasesrepocache.NodeClassAction` dependency,
replacing:

```go
if node.NodeClassId != action.NodeClassId {
    return u.createActionLog(..., "node class does not match action's node class")
}
```

with:

```go
if _, _, _, err := u.nodeClassAction.ReadByNodeClassIdAndActionId(ctx, node.NodeClassId, action.Id); err != nil {
    if errors.Is(err, domainmodels.ErrTypeNotFound) {
        return u.createActionLog(..., "action is not compatible with node's class")
    }
    ...
}
```

## 7. API surface

New permission strings: `node_class_action:get`, `node_class_action:add`,
`node_class_action:remove` (seeded into `permission.json`, granted to
`super` and to whichever roles currently hold both `action:set` and
`node_class:set` in `role.json`).

```
GET    /node-classes/:id/actions                        node_class_action:get    → []ActionResponse
POST   /node-classes/:node_class_id/actions/:action_id  node_class_action:add    → 201 IdResponse
DELETE /node-classes/:node_class_id/actions/:action_id  node_class_action:remove → 204
GET    /node-class-actions                              node_class_action:get    → PageDataResponse[NodeClassActionDetailResponse]
GET    /node-class-actions/by-pair                      node_class_action:get    → NodeClassActionDetailResponse
GET    /node-class-actions/:id                          node_class_action:get    → NodeClassActionDetailResponse
```

Exact mirror of `/admin/roles/:id/permissions` +
`/admin/role-permissions*`. New response types
`NodeClassActionResponse`/`NodeClassActionDetailResponse` in
`presentation/http/response/node_class.go`, mirroring
`RolePermissionResponse`/`RolePermissionDetailResponse` in
`response/role.go`.

## 8. Frontend

**Node Class detail page** (`node-classes/[id]/page.tsx`): new "Compatible
actions" section, structurally identical to Access Control's
`RoleDetails`/`PermissionGroups` — fetches all actions (paginated, same as
today's `listActions`) + this class's assigned action ids
(`GET /node-classes/:id/actions`), renders a checkbox list, single "Save
assignments" button that diffs current vs. desired and fires
assign/revoke calls via `Promise.allSettled` (new
`NodeClassActionChecklist.tsx` + `updateNodeClassActionsAction`, mirroring
`PermissionGroups.tsx` + `updateRoleAssignmentsAction`, minus the
resource-grouping logic which doesn't apply to a flat action list).

**Action form** (`ActionForm.tsx`): the `node_class_id` `<select>` is
removed entirely; the `nodeClasses` prop and its threading from
`actions/page.tsx`/`actions/[id]/page.tsx` are removed.

**`ActionCard.tsx`**: drops the `nodeClassName` prop entirely (used
inconsistently today — passed on the Actions list page, absent on the
Node detail page's Actions tab where it already fell back to a raw id).
Summary line becomes `"${count} compatible node class(es)"` for both call
sites, sourced from `ActionResponse.compatible_node_class_count`.

**`DispatchActionDialog.tsx`**: the node filter
`nodes.filter(node => node.node_class_id === action.node_class_id)`
becomes a filter against a `compatibleNodeClassIds: Set<string>` prop.
Since the dialog is only ever opened for a single node
(`nodes={[node]}`, per `NodeActionsWorkspace`) or a fleet-wide dispatch
context, the simplest source for this set is the same
`GET /node-class-actions?action_id=` list already needed for the Action
detail page's read-only compatible-class display — fetched once,
passed down.

**Actions list page** (`actions/page.tsx`):
- `node_class_id` filter `Input` stays as-is (see §5 for the backend
  semantics change).
- Drops `listAllNodeClasses()` and the `classNames` map — no longer used
  by `ActionForm` (field removed) or `ActionCard` (count instead of name).
- `canCreate` drops the `permissions.has("node_class:get")` requirement —
  creating an action no longer touches node classes at all.

**Action detail page** (`actions/[id]/page.tsx`): add a read-only
"Compatible node classes" list, sourced from
`GET /node-class-actions?action_id=<id>` (paginated detail rows, each
carrying the full `NodeClassResponse`), no toggle UI — matches the
decision that only the Node Class page manages the relationship.

## 9. Seeder

`database/seeder/action.json` drops `node_class_name`. A new seeding step
mirrors `seedRolePermissions`'s idempotent create-if-missing loop: each
action seed entry gains `node_class_names: string[]`, and
`internal/application/seeder/usecase.go` gets a `seedNodeClassActions`
step that loops action → node_class_names, resolves ids via the existing
lookup maps, checks `ReadByNodeClassIdAndActionId` for idempotency, and
creates if missing.

## 10. Out of scope

- Test case planning/writing (per explicit user instruction — this spec
  and its implementation plan skip test-case design; existing
  `go build`/`go vet`/`tsc`/`eslint` verification still applies).
- Any change to how `role_permission`, `Role`, or `Permission` themselves
  work — they're reference material only.
- Bulk "assign this action to N classes" from the Action side (explicitly
  out of scope per the "Node Class page only" decision).
