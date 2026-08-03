# Action ↔ Node Class many-to-many Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `actions.node_class_id` (single required FK) with a `node_class_action` pivot table, so one action can be reusable across multiple node classes, managed from the Node Class detail page.

**Architecture:** Full mirror of the existing `role_permission` pivot stack (contracts → postgres repo → redis cache → repocache decorator → usecase → handler → routes), swapping role→node_class, permission→action. `NodeClass` gains a direct `ReadActions` joined-read method mirroring `Role.ReadPermissions`. The Actions list view gains a `compatible_node_class_count` enrichment mirroring the existing `ActionLogListItem` pattern.

**Tech Stack:** Go (Echo v5, squirrel, pgx, go-redis), Next.js App Router / React 19, Tailwind.

## Global Constraints

- Skip all test-case planning and writing (explicit user instruction). Verification is `go build`/`go vet`/`tsc`/`eslint`/`prettier` plus one live Docker+Playwright pass at the end (Task T6) — no new `_test.go` or `.test.tsx` files.
- No data backfill in the migration — the database has already been reset (user's explicit instruction). Drop `actions.node_class_id` clean; create `node_class_action` empty.
- The relationship is managed **only** from the Node Class detail page (toggle checklist, "Save assignments"). The Action page shows a **read-only** compatible-class count/list — never add a toggle there.
- Every new backend package must mirror its `role_permission`/`Role.ReadPermissions` reference exactly in structure, naming convention, and error handling — this is a deliberate architectural mirror, not a redesign.
- New permission strings: `node_class_action:get`, `node_class_action:add`, `node_class_action:remove`.
- New error type: `ErrTypeNodeClassActionExists = errors.New("NODE_CLASS_ACTION_EXISTS")`.
- Migration timestamps: `20260803090000` (drop column + create pivot) — later than the last existing migration `20260730140000`.

---

## Task B1: Migration — drop `actions.node_class_id`, create `node_class_action`

**Files:**
- Create: `backend/database/migrations/20260803090000_node_class_action.up.sql`
- Create: `backend/database/migrations/20260803090000_node_class_action.down.sql`
- Modify: `backend/internal/domain/models/error.go`

- [ ] **Step 1: Write the up migration**

```sql
-- backend/database/migrations/20260803090000_node_class_action.up.sql
ALTER TABLE actions DROP CONSTRAINT actions_node_class_id_fkey;

DROP INDEX IF EXISTS idx_actions_node_class_id;

ALTER TABLE actions DROP COLUMN node_class_id;

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

- [ ] **Step 2: Write the down migration**

```sql
-- backend/database/migrations/20260803090000_node_class_action.down.sql
DROP TABLE IF EXISTS node_class_action;

ALTER TABLE actions ADD COLUMN node_class_id UUID;
```

- [ ] **Step 3: Add the new error type**

In `backend/internal/domain/models/error.go`, add to the `ErrType*Exists` block (after `ErrTypeRolePermissionExists`):

```go
	ErrTypeRolePermissionExists       = errors.New("ROLE_PERMISSION_EXISTS")
	ErrTypeNodeClassActionExists      = errors.New("NODE_CLASS_ACTION_EXISTS")
	ErrTypeFirmwareConfigKeyExists    = errors.New("FIRMWARE_CONFIG_KEY_EXISTS")
```

- [ ] **Step 4: Run the migration**

Run: `cd backend && docker compose exec app /app/migrate up` (or whatever this repo's documented migration command is — check `backend/AGENTS.md`/Makefile if unsure) against the dev database, or rely on the app container's entrypoint auto-migrating on next `docker compose up --build`.

- [ ] **Step 5: Commit**

```bash
git add backend/database/migrations/20260803090000_node_class_action.up.sql backend/database/migrations/20260803090000_node_class_action.down.sql backend/internal/domain/models/error.go
git commit -m "db: replace actions.node_class_id with node_class_action pivot table"
```

---

## Task B2: Domain models and contracts

**Files:**
- Create: `backend/internal/domain/models/node_class_action.go`
- Modify: `backend/internal/domain/models/action.go`
- Create: `backend/internal/domain/contracts/repository/node_class_action.go`
- Modify: `backend/internal/domain/contracts/repository/action.go`
- Modify: `backend/internal/domain/contracts/repository/node_class.go`
- Create: `backend/internal/domain/contracts/cache/node_class_action.go`
- Modify: `backend/internal/domain/contracts/cache/action.go`
- Modify: `backend/internal/domain/contracts/cache/node_class.go`
- Modify: `backend/internal/domain/contracts/cache/types.go`
- Create: `backend/internal/domain/usecases/repocache/node_class_action.go`

**Interfaces:**
- Produces: `domainmodels.NodeClassAction`, `domainmodels.ActionListItem`, `domaincontractsrepository.NodeClassAction`, `domaincontractscache.NodeClassAction`, `domaincontractscache.NodeClassActionItem`, `domainusecasesrepocache.NodeClassAction` — all consumed by Tasks B3-B7.

- [ ] **Step 1: New domain model `NodeClassAction`**

```go
// backend/internal/domain/models/node_class_action.go
package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type NodeClassAction struct {
	Id          uuid.UUID  `db:"id" json:"id"`
	NodeClassId uuid.UUID  `db:"node_class_id" json:"node_class_id"`
	ActionId    uuid.UUID  `db:"action_id" json:"action_id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	CreatedBy   *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
}
```

- [ ] **Step 2: Remove `NodeClassId` from `Action`, add `ActionListItem`**

In `backend/internal/domain/models/action.go`, remove the `NodeClassId` field:

```go
type Action struct {
	Id                   uuid.UUID       `db:"id" json:"id"`
	Name                 string          `db:"name" json:"name"`
	Description          string          `db:"description" json:"description"`
	PayloadSchemaName    string          `db:"payload_schema_name" json:"payload_schema_name"`
	PayloadSchemaVersion int32           `db:"payload_schema_version" json:"payload_schema_version"`
	Preferences          json.RawMessage `db:"preferences" json:"preferences"`
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt            *time.Time      `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt            *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedBy            *uuid.UUID      `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy            *uuid.UUID      `db:"updated_by" json:"updated_by,omitempty"`
	DeletedBy            *uuid.UUID      `db:"deleted_by" json:"deleted_by,omitempty"`
}

// ActionListItem enriches Action with data only needed by the list view -
// mirrors ActionLogListItem's embedding pattern in action_log.go.
type ActionListItem struct {
	Action
	CompatibleNodeClassCount int `db:"compatible_node_class_count" json:"compatible_node_class_count"`
}
```

- [ ] **Step 3: New repository contract `NodeClassAction`**

```go
// backend/internal/domain/contracts/repository/node_class_action.go
package domaincontractsrepository

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type NodeClassAction interface {
	Create(
		ctx context.Context,
		nodeClassId uuid.UUID,
		actionId uuid.UUID,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (nodeClassAction *domainmodels.NodeClassAction, nodeClass *domainmodels.NodeClass, action *domainmodels.Action, err error)

	ReadByNodeClassIdAndActionId(
		ctx context.Context,
		nodeClassId uuid.UUID,
		actionId uuid.UUID,
	) (nodeClassAction *domainmodels.NodeClassAction, nodeClass *domainmodels.NodeClass, action *domainmodels.Action, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		nodeClassId *uuid.UUID,
		actionId *uuid.UUID,
	) (nodeClassActions []domainmodels.NodeClassAction, nodeClasses []domainmodels.NodeClass, actions []domainmodels.Action, total int, err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
	) (err error)

	DeleteByNodeClassIdAndActionId(
		ctx context.Context,
		nodeClassId *uuid.UUID,
		actionId *uuid.UUID,
	) (err error)
}
```

- [ ] **Step 4: Update repository contract `Action`**

In `backend/internal/domain/contracts/repository/action.go`, remove `nodeClassId` from `Create`, `ReadByPagination` (also change return type), and `UpdateById`:

```go
package domaincontractsrepository

import (
	"context"
	"encoding/json"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Action interface {
	Create(
		ctx context.Context,
		name string,
		description *string,
		payloadSchemaName string,
		payloadSchemaVersion int32,
		createdBy *uuid.UUID,
	) (id uuid.UUID, err error)

	ReadById(
		ctx context.Context,
		id uuid.UUID,
	) (action *domainmodels.Action, err error)

	ReadByName(
		ctx context.Context,
		name string,
	) (action *domainmodels.Action, err error)

	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		search *string,
		nodeClassId *uuid.UUID,
		payloadSchemaName *string,
		payloadSchemaVersion *int32,
	) (actions []domainmodels.ActionListItem, total int, err error)

	UpdateById(
		ctx context.Context,
		id uuid.UUID,
		name *string,
		description *string,
		payloadSchemaName *string,
		payloadSchemaVersion *int32,
		preferences *json.RawMessage,
		updatedBy *uuid.UUID,
	) (err error)

	DeleteById(
		ctx context.Context,
		id uuid.UUID,
		deletedBy *uuid.UUID,
	) (err error)
}
```

Note `nodeClassId *uuid.UUID` stays as a `ReadByPagination` filter param — its SQL meaning changes to an `EXISTS` subquery in Task B3, not its Go signature.

- [ ] **Step 5: Update repository contract `NodeClass`**

In `backend/internal/domain/contracts/repository/node_class.go`, add `ReadActions` (mirrors `Role.ReadPermissions`):

```go
	ReadByPagination(
		ctx context.Context,
		page int,
		limit int,
		search *string,
	) (nodeClasses []domainmodels.NodeClass, total int, err error)

	ReadActions(
		ctx context.Context,
		nodeClassId uuid.UUID,
	) (actions []domainmodels.Action, err error)

	UpdateById(
```

(inserted between `ReadByPagination` and `UpdateById`, matching where `Role.ReadPermissions` sits relative to its neighbors)

- [ ] **Step 6: New cache contract `NodeClassAction`**

```go
// backend/internal/domain/contracts/cache/node_class_action.go
package domaincontractscache

import (
	"context"

	"github.com/google/uuid"
)

type NodeClassAction interface {
	GetById(ctx context.Context, id uuid.UUID) (item *NodeClassActionItem, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, item *NodeClassActionItem) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID) (item *NodeClassActionItem, hit bool, err error)
	SetByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID, item *NodeClassActionItem) error
	DeleteByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID) error
	GetPagination(ctx context.Context, page int, limit int, nodeClassId *uuid.UUID, actionId *uuid.UUID) (pagination Pagination[NodeClassActionItem], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, nodeClassId *uuid.UUID, actionId *uuid.UUID, pagination Pagination[NodeClassActionItem]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateByNodeClassId(ctx context.Context, nodeClassId uuid.UUID) error
	InvalidateByActionId(ctx context.Context, actionId uuid.UUID) error
	InvalidateAll(ctx context.Context) error
}
```

- [ ] **Step 7: Add `NodeClassActionItem` to `types.go`**

In `backend/internal/domain/contracts/cache/types.go`:

```go
package domaincontractscache

import domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"

type Pagination[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

type RolePermissionItem struct {
	RolePermission domainmodels.RolePermission `json:"role_permission"`
	Role           domainmodels.Role           `json:"role"`
	Permission     domainmodels.Permission     `json:"permission"`
}

type NodeClassActionItem struct {
	NodeClassAction domainmodels.NodeClassAction `json:"node_class_action"`
	NodeClass       domainmodels.NodeClass       `json:"node_class"`
	Action          domainmodels.Action          `json:"action"`
}
```

- [ ] **Step 8: Update cache contract `Action`**

In `backend/internal/domain/contracts/cache/action.go`, `GetPagination`/`SetPagination` return `Pagination[domainmodels.ActionListItem]` instead of `Pagination[domainmodels.Action]` (the `nodeClassId`/other params are unchanged):

```go
package domaincontractscache

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Action interface {
	GetById(ctx context.Context, id uuid.UUID) (action *domainmodels.Action, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, action *domainmodels.Action) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByName(ctx context.Context, name string) (action *domainmodels.Action, hit bool, err error)
	SetByName(ctx context.Context, name string, action *domainmodels.Action) error
	DeleteByName(ctx context.Context, name string) error
	GetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, payloadSchemaName *string, payloadSchemaVersion *int32) (pagination Pagination[domainmodels.ActionListItem], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, payloadSchemaName *string, payloadSchemaVersion *int32, pagination Pagination[domainmodels.ActionListItem]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateAll(ctx context.Context) error
}
```

- [ ] **Step 9: Update cache contract `NodeClass`**

In `backend/internal/domain/contracts/cache/node_class.go`, add the `Actions` family mirroring `Role`'s `Permissions` family:

```go
package domaincontractscache

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type NodeClass interface {
	GetById(ctx context.Context, id uuid.UUID) (nodeClass *domainmodels.NodeClass, hit bool, err error)
	SetById(ctx context.Context, id uuid.UUID, nodeClass *domainmodels.NodeClass) error
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetByName(ctx context.Context, name string) (nodeClass *domainmodels.NodeClass, hit bool, err error)
	SetByName(ctx context.Context, name string, nodeClass *domainmodels.NodeClass) error
	DeleteByName(ctx context.Context, name string) error
	GetActions(ctx context.Context, nodeClassId uuid.UUID) (actions []domainmodels.Action, hit bool, err error)
	SetActions(ctx context.Context, nodeClassId uuid.UUID, actions []domainmodels.Action) error
	DeleteActions(ctx context.Context, nodeClassId uuid.UUID) error
	InvalidateActions(ctx context.Context) error
	GetPagination(ctx context.Context, page int, limit int, search *string) (pagination Pagination[domainmodels.NodeClass], hit bool, err error)
	SetPagination(ctx context.Context, page int, limit int, search *string, pagination Pagination[domainmodels.NodeClass]) error
	InvalidatePagination(ctx context.Context) error
	InvalidateAll(ctx context.Context) error
}
```

- [ ] **Step 10: New `domainusecasesrepocache.NodeClassAction`**

```go
// backend/internal/domain/usecases/repocache/node_class_action.go
package domainusecasesrepocache

import domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"

type NodeClassAction interface {
	domaincontractsrepository.NodeClassAction
}
```

- [ ] **Step 11: Build to confirm the contracts compile in isolation**

Run: `cd backend && go build ./internal/domain/...`
Expected: fails only in packages that implement these interfaces (repository/cache implementations, not yet updated) — confirm the failures are exactly `infrastructure/repository/action`, `infrastructure/cache/action`, `infrastructure/repository/node_class`, `infrastructure/cache/node_class`, and any `application/*` packages consuming them. This confirms the contract changes are complete and correctly propagate; Tasks B3+ fix the implementations.

- [ ] **Step 12: Commit**

```bash
git add backend/internal/domain
git commit -m "domain: add node_class_action contracts, drop Action.NodeClassId, add ActionListItem"
```

---

## Task B3: Infrastructure — Postgres repositories

**Files:**
- Create: `backend/internal/infrastructure/repository/node_class_action/postgres.go`
- Create: `backend/internal/infrastructure/repository/node_class_action/postgres_query.go`
- Modify: `backend/internal/infrastructure/repository/action/postgres.go`
- Modify: `backend/internal/infrastructure/repository/action/postgres_query.go`
- Modify: `backend/internal/infrastructure/repository/node_class/postgres.go`
- Modify: `backend/internal/infrastructure/repository/node_class/postgres_query.go`
- Modify: `backend/internal/infrastructure/repository/shared/scan_pgx.go`

**Interfaces:**
- Consumes: contracts from Task B2.
- Produces: `infrastructurerepositorynodeclassaction.NewPostgresImpl(...) domaincontractsrepository.NodeClassAction`, updated `infrastructurerepositoryaction.NewPostgresImpl` and `infrastructurerepositorynodeclass.NewPostgresImpl` (same constructor signatures as before — only the returned interface's method bodies change).

- [ ] **Step 1: New `node_class_action` postgres repository — full mirror of `role_permission`**

```go
// backend/internal/infrastructure/repository/node_class_action/postgres.go
package infrastructurerepositorynodeclassaction

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/MateWorkspace/mate-things/backend/pkg/pgxdt"
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
) domaincontractsrepository.NodeClassAction {
	return &postgresImpl{
		BasePostgres: infrastructurerepositoryshared.BasePostgres{
			Dt:   dt,
			SqrQ: sqrQuestion,
			SqrD: sqrDollar,
		},
	}
}

func (p *postgresImpl) Create(
	ctx context.Context,
	nodeClassId uuid.UUID,
	actionId uuid.UUID,
	createdBy *uuid.UUID,
) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(nodeClassId, actionId, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create node class action query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create node class action", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "node_class_id_action_id", Type: domainmodels.ErrTypeNodeClassActionExists},
		)
	}

	return id, nil
}

func (p *postgresImpl) ReadById(
	ctx context.Context,
	id uuid.UUID,
) (nodeClassAction *domainmodels.NodeClassAction, nodeClass *domainmodels.NodeClass, action *domainmodels.Action, err error) {
	query, args, err := p.queryReadById(id)
	if err != nil {
		return nil, nil, nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node class action query", err)
	}

	nca, nc, a, err := scanPgxNodeClassActionJoined(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil, infrastructurerepositoryshared.NotFound("node class action not found", err)
		}
		return nil, nil, nil, infrastructurerepositoryshared.MapPgxError("failed to read node class action", err)
	}

	return &nca, &nc, &a, nil
}

func (p *postgresImpl) ReadByNodeClassIdAndActionId(
	ctx context.Context,
	nodeClassId uuid.UUID,
	actionId uuid.UUID,
) (nodeClassAction *domainmodels.NodeClassAction, nodeClass *domainmodels.NodeClass, action *domainmodels.Action, err error) {
	query, args, err := p.queryReadByNodeClassIdAndActionId(nodeClassId, actionId)
	if err != nil {
		return nil, nil, nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node class action query", err)
	}

	nca, nc, a, err := scanPgxNodeClassActionJoined(p.Dt.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, nil, infrastructurerepositoryshared.NotFound("node class action not found", err)
		}
		return nil, nil, nil, infrastructurerepositoryshared.MapPgxError("failed to read node class action", err)
	}

	return &nca, &nc, &a, nil
}

func (p *postgresImpl) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) (nodeClassActions []domainmodels.NodeClassAction, nodeClasses []domainmodels.NodeClass, actions []domainmodels.Action, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByPagination(page, limit, nodeClassId, actionId)
	if err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read node class actions query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count node class actions", err)
	}
	if total == 0 {
		return []domainmodels.NodeClassAction{}, []domainmodels.NodeClass{}, []domainmodels.Action{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read node class actions", err)
	}
	defer rows.Close()

	nodeClassActions = make([]domainmodels.NodeClassAction, 0, total)
	nodeClasses = make([]domainmodels.NodeClass, 0, total)
	actions = make([]domainmodels.Action, 0, total)
	for rows.Next() {
		nca, nc, a, err := scanPgxNodeClassActionJoined(rows)
		if err != nil {
			return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan node class action", err)
		}

		nodeClassActions = append(nodeClassActions, nca)
		nodeClasses = append(nodeClasses, nc)
		actions = append(actions, a)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, 0, infrastructurerepositoryshared.MapPgxError("failed to iterate node class actions", err)
	}

	return nodeClassActions, nodeClasses, actions, total, nil
}

func (p *postgresImpl) DeleteById(ctx context.Context, id uuid.UUID) (err error) {
	query, args, err := p.queryDeleteById(id)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete node class action query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete node class action", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("node class action not found", nil)
	}

	return nil
}

func (p *postgresImpl) DeleteByNodeClassIdAndActionId(
	ctx context.Context,
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) (err error) {
	query, args, err := p.queryDeleteByNodeClassIdAndActionId(nodeClassId, actionId)
	if err != nil {
		return infrastructurerepositoryshared.QueryBuildError("failed to build delete node class action query", err)
	}

	commandTag, err := p.Dt.Exec(ctx, query, args...)
	if err != nil {
		return infrastructurerepositoryshared.MapPgxError("failed to delete node class action", err)
	}
	if commandTag.RowsAffected() == 0 {
		return infrastructurerepositoryshared.NotFound("node class action not found", nil)
	}

	return nil
}

func scanPgxNodeClassActionJoined(row pgx.Row) (
	nodeClassAction domainmodels.NodeClassAction,
	nodeClass domainmodels.NodeClass,
	action domainmodels.Action,
	err error,
) {
	err = row.Scan(
		&nodeClassAction.Id,
		&nodeClassAction.NodeClassId,
		&nodeClassAction.ActionId,
		&nodeClassAction.CreatedAt,
		&nodeClassAction.CreatedBy,
		&nodeClass.Id,
		&nodeClass.Name,
		&nodeClass.Description,
		&nodeClass.Preferences,
		&nodeClass.CreatedAt,
		&nodeClass.UpdatedAt,
		&nodeClass.DeletedAt,
		&nodeClass.CreatedBy,
		&nodeClass.UpdatedBy,
		&nodeClass.DeletedBy,
		&action.Id,
		&action.Name,
		&action.Description,
		&action.PayloadSchemaName,
		&action.PayloadSchemaVersion,
		&action.Preferences,
		&action.CreatedAt,
		&action.UpdatedAt,
		&action.DeletedAt,
		&action.CreatedBy,
		&action.UpdatedBy,
		&action.DeletedBy,
	)
	return
}
```

- [ ] **Step 2: Query builder for `node_class_action`**

```go
// backend/internal/infrastructure/repository/node_class_action/postgres_query.go
package infrastructurerepositorynodeclassaction

import (
	"github.com/Masterminds/squirrel"
	infrastructurerepositoryshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/shared"
	"github.com/google/uuid"
)

func (p *postgresImpl) queryCreate(
	nodeClassId uuid.UUID,
	actionId uuid.UUID,
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	return p.SqrD.Insert("node_class_action").
		Columns("node_class_id", "action_id", "created_by").
		Values(nodeClassId, actionId, createdBy).
		Suffix("RETURNING id").
		ToSql()
}

func (p *postgresImpl) queryReadById(id uuid.UUID) (query string, args []any, err error) {
	return p.baseReadQuery().
		Where(squirrel.Eq{"nca.id": id}).
		ToSql()
}

func (p *postgresImpl) queryReadByNodeClassIdAndActionId(
	nodeClassId uuid.UUID,
	actionId uuid.UUID,
) (query string, args []any, err error) {
	return p.baseReadQuery().
		Where(squirrel.Eq{"nca.node_class_id": nodeClassId}).
		Where(squirrel.Eq{"nca.action_id": actionId}).
		Limit(1).
		ToSql()
}

func (p *postgresImpl) queryReadByPagination(
	page int,
	limit int,
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	baseQ := p.baseReadQuery()
	totalQ := p.SqrD.Select("COUNT(*)").
		From("node_class_action nca").
		Join("node_classes nc ON nca.node_class_id = nc.id").
		Join("actions a ON nca.action_id = a.id").
		Where("nc.deleted_at IS NULL").
		Where("a.deleted_at IS NULL")

	if nodeClassId != nil {
		condition := squirrel.Eq{"nca.node_class_id": *nodeClassId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if actionId != nil {
		condition := squirrel.Eq{"nca.action_id": *actionId}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("nca.created_at DESC", "nca.id ASC").
		Limit(infrastructurerepositoryshared.NormalizeLimit(limit)).
		Offset(infrastructurerepositoryshared.NormalizeOffset(page, limit)).
		ToSql()
	return
}

func (p *postgresImpl) queryDeleteById(id uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Delete("node_class_action").
		Where(squirrel.Eq{"id": id}).
		ToSql()
}

func (p *postgresImpl) queryDeleteByNodeClassIdAndActionId(
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Delete("node_class_action")

	if nodeClassId != nil {
		q = q.Where(squirrel.Eq{"node_class_id": *nodeClassId})
	}
	if actionId != nil {
		q = q.Where(squirrel.Eq{"action_id": *actionId})
	}

	return q.ToSql()
}

func (p *postgresImpl) baseReadQuery() squirrel.SelectBuilder {
	return p.SqrD.Select(
		"nca.id",
		"nca.node_class_id",
		"nca.action_id",
		"nca.created_at",
		"nca.created_by",
		"nc.id",
		"nc.name",
		"nc.description",
		"nc.preferences",
		"nc.created_at",
		"nc.updated_at",
		"nc.deleted_at",
		"nc.created_by",
		"nc.updated_by",
		"nc.deleted_by",
		"a.id",
		"a.name",
		"a.description",
		"a.payload_schema_name",
		"a.payload_schema_version",
		"a.preferences",
		"a.created_at",
		"a.updated_at",
		"a.deleted_at",
		"a.created_by",
		"a.updated_by",
		"a.deleted_by",
	).
		From("node_class_action nca").
		Join("node_classes nc ON nca.node_class_id = nc.id").
		Join("actions a ON nca.action_id = a.id").
		Where("nc.deleted_at IS NULL").
		Where("a.deleted_at IS NULL")
}
```

- [ ] **Step 3: Update shared scanners — drop `node_class_id`, add `ActionListItem` scanners**

In `backend/internal/infrastructure/repository/shared/scan_pgx.go`, change `ScanPgxAction` (remove `&item.NodeClassId`) and add new `ScanPgxActionListItem`/`ScanPgxActionListItems` next to it:

```go
func ScanPgxAction(row pgx.Row) (domainmodels.Action, error) {
	var item domainmodels.Action
	err := row.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
		&item.PayloadSchemaName,
		&item.PayloadSchemaVersion,
		&item.Preferences,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.DeletedBy,
	)
	return item, err
}

func ScanPgxActions(rows pgx.Rows) ([]domainmodels.Action, error) {
	items := make([]domainmodels.Action, 0)
	for rows.Next() {
		item, err := ScanPgxAction(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ScanPgxActionListItem(row pgx.Row) (domainmodels.ActionListItem, error) {
	var item domainmodels.ActionListItem
	err := row.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
		&item.PayloadSchemaName,
		&item.PayloadSchemaVersion,
		&item.Preferences,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.DeletedBy,
		&item.CompatibleNodeClassCount,
	)
	return item, err
}

func ScanPgxActionListItems(rows pgx.Rows) ([]domainmodels.ActionListItem, error) {
	items := make([]domainmodels.ActionListItem, 0)
	for rows.Next() {
		item, err := ScanPgxActionListItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
```

- [ ] **Step 4: Update `action` postgres repository — drop `nodeClassId` param, use `ActionListItem` for pagination**

In `backend/internal/infrastructure/repository/action/postgres.go`:

```go
func (p *postgresImpl) Create(
	ctx context.Context,
	name string,
	description *string,
	payloadSchemaName string,
	payloadSchemaVersion int32,
	createdBy *uuid.UUID,
) (id uuid.UUID, err error) {
	query, args, err := p.queryCreate(name, description, payloadSchemaName, payloadSchemaVersion, createdBy)
	if err != nil {
		return uuid.Nil, infrastructurerepositoryshared.QueryBuildError("failed to build create action query", err)
	}

	if err := p.Dt.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create action", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypeActionNameExists},
		)
	}

	return id, nil
}
```

(Only `Create`'s signature drops `nodeClassId uuid.UUID` and the call to `p.queryCreate` drops the arg — `ReadById`/`ReadByName`/`DeleteById` bodies are unchanged.)

```go
func (p *postgresImpl) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
	nodeClassId *uuid.UUID,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
) (actions []domainmodels.ActionListItem, total int, err error) {
	totalQuery, totalArgs, query, queryArgs, err := p.queryReadByPagination(page, limit, search, nodeClassId, payloadSchemaName, payloadSchemaVersion)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.QueryBuildError("failed to build read actions query", err)
	}

	if err := p.Dt.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to count actions", err)
	}
	if total == 0 {
		return []domainmodels.ActionListItem{}, 0, nil
	}

	rows, err := p.Dt.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to read actions", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxActionListItems(rows)
	if err != nil {
		return nil, 0, infrastructurerepositoryshared.MapPgxError("failed to scan actions", err)
	}

	return items, total, nil
}
```

```go
func (p *postgresImpl) UpdateById(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	description *string,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (err error) {
	query, args, err := p.queryUpdateById(id, name, description, payloadSchemaName, payloadSchemaVersion, preferences, updatedBy)
	// ... body unchanged below this line, just drops nodeClassId arg from the queryUpdateById call
```

- [ ] **Step 5: Update `action` query builder**

In `backend/internal/infrastructure/repository/action/postgres_query.go`:

```go
var actionColumns = []string{
	"id",
	"name",
	"description",
	"payload_schema_name",
	"payload_schema_version",
	"preferences",
	"created_at",
	"updated_at",
	"deleted_at",
	"created_by",
	"updated_by",
	"deleted_by",
}

func (p *postgresImpl) queryCreate(
	name string,
	description *string,
	payloadSchemaName string,
	payloadSchemaVersion int32,
	createdBy *uuid.UUID,
) (query string, args []any, err error) {
	columns := []string{"name", "payload_schema_name", "payload_schema_version", "created_by"}
	values := []any{name, payloadSchemaName, payloadSchemaVersion, createdBy}

	if description != nil {
		columns = append(columns, "description")
		values = append(values, *description)
	}

	return p.SqrD.Insert("actions").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING id").
		ToSql()
}
```

`queryReadById`/`queryReadByName` are unchanged (they already select via `actionColumns...`, which no longer contains `node_class_id`).

```go
func (p *postgresImpl) queryReadByPagination(
	page int,
	limit int,
	search *string,
	nodeClassId *uuid.UUID,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
) (totalQuery string, totalArgs []any, query string, queryArgs []any, err error) {
	selectColumns := append([]string{}, actionColumns...)
	selectColumns = append(selectColumns, "(SELECT COUNT(*) FROM node_class_action nca WHERE nca.action_id = actions.id) AS compatible_node_class_count")

	baseQ := p.SqrD.Select(selectColumns...).
		From("actions").
		Where("deleted_at IS NULL")
	totalQ := p.SqrD.Select("COUNT(*)").
		From("actions").
		Where("deleted_at IS NULL")

	if pattern, ok := infrastructurerepositoryshared.SearchPattern(search); ok {
		condition := squirrel.Expr("name ILIKE ?", pattern)
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if nodeClassId != nil {
		condition := squirrel.Expr(
			"EXISTS (SELECT 1 FROM node_class_action nca WHERE nca.action_id = actions.id AND nca.node_class_id = ?)",
			*nodeClassId,
		)
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if payloadSchemaName != nil {
		condition := squirrel.Eq{"payload_schema_name": *payloadSchemaName}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}
	if payloadSchemaVersion != nil {
		condition := squirrel.Eq{"payload_schema_version": *payloadSchemaVersion}
		baseQ = baseQ.Where(condition)
		totalQ = totalQ.Where(condition)
	}

	totalQuery, totalArgs, err = totalQ.ToSql()
	if err != nil {
		return
	}

	query, queryArgs, err = baseQ.
		OrderBy("created_at DESC", "id ASC").
		Limit(infrastructurerepositoryshared.NormalizeLimit(limit)).
		Offset(infrastructurerepositoryshared.NormalizeOffset(page, limit)).
		ToSql()
	return
}
```

```go
func (p *postgresImpl) queryUpdateById(
	id uuid.UUID,
	name *string,
	description *string,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
	preferences *json.RawMessage,
	updatedBy *uuid.UUID,
) (query string, args []any, err error) {
	q := p.SqrD.Update("actions").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL")

	if name != nil {
		q = q.Set("name", *name)
	}
	if description != nil {
		q = q.Set("description", *description)
	}
	if payloadSchemaName != nil {
		q = q.Set("payload_schema_name", *payloadSchemaName)
	}
	if payloadSchemaVersion != nil {
		q = q.Set("payload_schema_version", *payloadSchemaVersion)
	}
	if preferences != nil {
		q = q.Set("preferences", *preferences)
	}

	return q.
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_by", updatedBy).
		ToSql()
}
```

`queryDeleteById` is unchanged.

- [ ] **Step 6: Add `ReadActions` to `node_class` postgres repository**

In `backend/internal/infrastructure/repository/node_class/postgres.go`, add (after `ReadByPagination`, before `UpdateById`):

```go
func (p *postgresImpl) ReadActions(ctx context.Context, nodeClassId uuid.UUID) (actions []domainmodels.Action, err error) {
	query, args, err := p.queryReadActions(nodeClassId)
	if err != nil {
		return nil, infrastructurerepositoryshared.QueryBuildError("failed to build read node class actions query", err)
	}

	rows, err := p.Dt.Query(ctx, query, args...)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read node class actions", err)
	}
	defer rows.Close()

	items, err := infrastructurerepositoryshared.ScanPgxActions(rows)
	if err != nil {
		return nil, infrastructurerepositoryshared.MapPgxError("failed to scan node class actions", err)
	}

	return items, nil
}
```

- [ ] **Step 7: Add `queryReadActions` to `node_class` query builder**

In `backend/internal/infrastructure/repository/node_class/postgres_query.go`, add (mirrors `queryReadPermissions` in `role/postgres_query.go`):

```go
func (p *postgresImpl) queryReadActions(nodeClassId uuid.UUID) (query string, args []any, err error) {
	return p.SqrD.Select(
		"a.id", "a.name", "a.description", "a.payload_schema_name",
		"a.payload_schema_version", "a.preferences", "a.created_at",
		"a.updated_at", "a.deleted_at", "a.created_by", "a.updated_by", "a.deleted_by",
	).
		From("node_class_action nca").
		Join("actions a ON a.id = nca.action_id").
		Where(squirrel.Eq{"nca.node_class_id": nodeClassId}).
		Where("a.deleted_at IS NULL").
		OrderBy("a.created_at DESC", "a.id ASC").
		ToSql()
}
```

(needs `"github.com/google/uuid"` import already present in this file for `queryReadById`/etc.)

- [ ] **Step 8: Build**

Run: `cd backend && go build ./internal/infrastructure/...`
Expected: clean build. Fix any column-order mismatches against `actionColumns`/`nodeClassColumns` if the compiler or a quick manual review catches one.

- [ ] **Step 9: Commit**

```bash
git add backend/internal/infrastructure/repository
git commit -m "infra: add node_class_action postgres repo, wire compatible_node_class_count into actions list"
```

---

## Task B4: Infrastructure — Redis caches

**Files:**
- Create: `backend/internal/infrastructure/cache/node_class_action/redis.go`
- Modify: `backend/internal/infrastructure/cache/action/redis.go`
- Modify: `backend/internal/infrastructure/cache/node_class/redis.go`

**Interfaces:**
- Consumes: cache contracts from Task B2.
- Produces: `infrastructurecachenodeclassaction.NewRedisImpl(...) domaincontractscache.NodeClassAction`.

- [ ] **Step 1: New `node_class_action` redis cache — full mirror of `role_permission`**

```go
// backend/internal/infrastructure/cache/node_class_action/redis.go
package infrastructurecachenodeclassaction

import (
	"context"

	domaincontractscache "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/cache"
	infrastructurecacheshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/shared"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisImpl struct {
	infrastructurecacheshared.BaseRedis
}

func NewRedisImpl(client redis.UniversalClient, namespace string, ttl infrastructurecacheshared.TtlConfig) domaincontractscache.NodeClassAction {
	return &redisImpl{BaseRedis: infrastructurecacheshared.NewRedis(client, namespace, "node_class_action", ttl)}
}

func (r *redisImpl) GetById(ctx context.Context, id uuid.UUID) (*domaincontractscache.NodeClassActionItem, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return nil, false, err
	}
	var item domaincontractscache.NodeClassActionItem
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetById(ctx context.Context, id uuid.UUID, item *domaincontractscache.NodeClassActionItem) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, item, r.IdentityTtl())
}

func (r *redisImpl) DeleteById(ctx context.Context, id uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "identity", "id", id.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID) (*domaincontractscache.NodeClassActionItem, bool, error) {
	key, err := r.ScopedKey(ctx, "identity", "node_class_action", nodeClassId.String(), actionId.String())
	if err != nil {
		return nil, false, err
	}
	var item domaincontractscache.NodeClassActionItem
	hit, err := r.Get(ctx, key, &item)
	return &item, hit, err
}

func (r *redisImpl) SetByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID, item *domaincontractscache.NodeClassActionItem) error {
	key, err := r.ScopedKey(ctx, "identity", "node_class_action", nodeClassId.String(), actionId.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, item, r.IdentityTtl())
}

func (r *redisImpl) DeleteByNodeClassIdAndActionId(ctx context.Context, nodeClassId uuid.UUID, actionId uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "identity", "node_class_action", nodeClassId.String(), actionId.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) GetPagination(ctx context.Context, page int, limit int, nodeClassId *uuid.UUID, actionId *uuid.UUID) (domaincontractscache.Pagination[domaincontractscache.NodeClassActionItem], bool, error) {
	key, err := r.paginationKey(ctx, page, limit, nodeClassId, actionId)
	if err != nil {
		return domaincontractscache.Pagination[domaincontractscache.NodeClassActionItem]{}, false, err
	}
	var pagination domaincontractscache.Pagination[domaincontractscache.NodeClassActionItem]
	hit, err := r.Get(ctx, key, &pagination)
	return pagination, hit, err
}

func (r *redisImpl) SetPagination(ctx context.Context, page int, limit int, nodeClassId *uuid.UUID, actionId *uuid.UUID, pagination domaincontractscache.Pagination[domaincontractscache.NodeClassActionItem]) error {
	key, err := r.paginationKey(ctx, page, limit, nodeClassId, actionId)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, pagination, r.PaginationTtl())
}

func (r *redisImpl) InvalidatePagination(ctx context.Context) error {
	return r.Invalidate(ctx, "pagination")
}

func (r *redisImpl) InvalidateByNodeClassId(ctx context.Context, nodeClassId uuid.UUID) error {
	return r.InvalidateAll(ctx)
}

func (r *redisImpl) InvalidateByActionId(ctx context.Context, actionId uuid.UUID) error {
	return r.InvalidateAll(ctx)
}

func (r *redisImpl) InvalidateAll(ctx context.Context) error {
	return r.Invalidate(ctx, "identity", "pagination")
}

func (r *redisImpl) paginationKey(ctx context.Context, page int, limit int, nodeClassId *uuid.UUID, actionId *uuid.UUID) (string, error) {
	hash, err := infrastructurecacheshared.HashPart(page, limit, nodeClassId, actionId)
	if err != nil {
		return "", err
	}
	return r.ScopedKey(ctx, "pagination", hash)
}
```

- [ ] **Step 2: Update `action` redis cache — `Pagination[domainmodels.ActionListItem]`**

In `backend/internal/infrastructure/cache/action/redis.go`, change `GetPagination`/`SetPagination`'s generic type param from `domainmodels.Action` to `domainmodels.ActionListItem` (three occurrences: the return type, the `var pagination` declaration, and the parameter type) — everything else in the file is unchanged:

```go
func (r *redisImpl) GetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, payloadSchemaName *string, payloadSchemaVersion *int32) (domaincontractscache.Pagination[domainmodels.ActionListItem], bool, error) {
	key, err := r.paginationKey(ctx, page, limit, search, nodeClassId, payloadSchemaName, payloadSchemaVersion)
	if err != nil {
		return domaincontractscache.Pagination[domainmodels.ActionListItem]{}, false, err
	}
	var pagination domaincontractscache.Pagination[domainmodels.ActionListItem]
	hit, err := r.Get(ctx, key, &pagination)
	return pagination, hit, err
}

func (r *redisImpl) SetPagination(ctx context.Context, page int, limit int, search *string, nodeClassId *uuid.UUID, payloadSchemaName *string, payloadSchemaVersion *int32, pagination domaincontractscache.Pagination[domainmodels.ActionListItem]) error {
	key, err := r.paginationKey(ctx, page, limit, search, nodeClassId, payloadSchemaName, payloadSchemaVersion)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, pagination, r.PaginationTtl())
}
```

- [ ] **Step 3: Add `Actions` family to `node_class` redis cache**

In `backend/internal/infrastructure/cache/node_class/redis.go`, add (mirrors `GetPermissions`/`SetPermissions`/`DeletePermissions`/`InvalidatePermissions` in `role/redis.go`, and update `InvalidateAll`):

```go
func (r *redisImpl) GetActions(ctx context.Context, nodeClassId uuid.UUID) ([]domainmodels.Action, bool, error) {
	key, err := r.ScopedKey(ctx, "actions", "node_class", nodeClassId.String())
	if err != nil {
		return nil, false, err
	}
	var items []domainmodels.Action
	hit, err := r.Get(ctx, key, &items)
	return items, hit, err
}

func (r *redisImpl) SetActions(ctx context.Context, nodeClassId uuid.UUID, actions []domainmodels.Action) error {
	key, err := r.ScopedKey(ctx, "actions", "node_class", nodeClassId.String())
	if err != nil {
		return err
	}
	return r.Set(ctx, key, actions, r.RelationTtl())
}

func (r *redisImpl) DeleteActions(ctx context.Context, nodeClassId uuid.UUID) error {
	key, err := r.ScopedKey(ctx, "actions", "node_class", nodeClassId.String())
	if err != nil {
		return err
	}
	return r.Delete(ctx, key)
}

func (r *redisImpl) InvalidateActions(ctx context.Context) error {
	return r.Invalidate(ctx, "actions")
}
```

And update `InvalidateAll`:

```go
func (r *redisImpl) InvalidateAll(ctx context.Context) error {
	return r.Invalidate(ctx, "identity", "actions", "pagination")
}
```

- [ ] **Step 4: Build**

Run: `cd backend && go build ./internal/infrastructure/... ./internal/domain/...`
Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/infrastructure/cache
git commit -m "infra: add node_class_action redis cache, wire ActionListItem pagination cache, node_class actions cache"
```

---

## Task B5: Application repocache decorators

**Files:**
- Create: `backend/internal/application/repocache/node_class_action/usecase.go`
- Modify: `backend/internal/application/repocache/action/usecase.go`
- Modify: `backend/internal/application/repocache/node_class/usecase.go`

**Interfaces:**
- Consumes: `domaincontractsrepository.NodeClassAction`, `domaincontractscache.NodeClassAction`, `domaincontractscache.NodeClass`, `domaincontractscache.Action`.
- Produces: `applicationrepocachenodeclassaction.NewRepoCacheImpl(repository, cache, nodeClassCache, actionCache) domainusecasesrepocache.NodeClassAction`.

- [ ] **Step 1: New `node_class_action` repocache decorator**

```go
// backend/internal/application/repocache/node_class_action/usecase.go
package applicationrepocachenodeclassaction

import (
	"context"

	domaincontractscache "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/cache"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	repository     domaincontractsrepository.NodeClassAction
	cache          domaincontractscache.NodeClassAction
	nodeClassCache domaincontractscache.NodeClass
	actionCache    domaincontractscache.Action
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.NodeClassAction,
	cache domaincontractscache.NodeClassAction,
	nodeClassCache domaincontractscache.NodeClass,
	actionCache domaincontractscache.Action,
) domainusecasesrepocache.NodeClassAction {
	return &usecase{
		repository:     repository,
		cache:          cache,
		nodeClassCache: nodeClassCache,
		actionCache:    actionCache,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	nodeClassId uuid.UUID,
	actionId uuid.UUID,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, nodeClassId, actionId, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.invalidateRelations(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(
	ctx context.Context,
	id uuid.UUID,
) (*domainmodels.NodeClassAction, *domainmodels.NodeClass, *domainmodels.Action, error) {
	if item, hit, err := u.cache.GetById(ctx, id); err != nil {
		return nil, nil, nil, err
	} else if hit {
		return &item.NodeClassAction, &item.NodeClass, &item.Action, nil
	}

	nodeClassAction, nodeClass, action, err := u.repository.ReadById(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := u.cache.SetById(ctx, id, &domaincontractscache.NodeClassActionItem{
		NodeClassAction: *nodeClassAction,
		NodeClass:       *nodeClass,
		Action:          *action,
	}); err != nil {
		return nil, nil, nil, err
	}

	return nodeClassAction, nodeClass, action, nil
}

func (u *usecase) ReadByNodeClassIdAndActionId(
	ctx context.Context,
	nodeClassId uuid.UUID,
	actionId uuid.UUID,
) (*domainmodels.NodeClassAction, *domainmodels.NodeClass, *domainmodels.Action, error) {
	if item, hit, err := u.cache.GetByNodeClassIdAndActionId(ctx, nodeClassId, actionId); err != nil {
		return nil, nil, nil, err
	} else if hit {
		return &item.NodeClassAction, &item.NodeClass, &item.Action, nil
	}

	nodeClassAction, nodeClass, action, err := u.repository.ReadByNodeClassIdAndActionId(ctx, nodeClassId, actionId)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := u.cache.SetByNodeClassIdAndActionId(ctx, nodeClassId, actionId, &domaincontractscache.NodeClassActionItem{
		NodeClassAction: *nodeClassAction,
		NodeClass:       *nodeClass,
		Action:          *action,
	}); err != nil {
		return nil, nil, nil, err
	}

	return nodeClassAction, nodeClass, action, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) ([]domainmodels.NodeClassAction, []domainmodels.NodeClass, []domainmodels.Action, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, nodeClassId, actionId); err != nil {
		return nil, nil, nil, 0, err
	} else if hit {
		nodeClassActions, nodeClasses, actions := splitNodeClassActionItems(pagination.Items)
		return nodeClassActions, nodeClasses, actions, pagination.Total, nil
	}

	nodeClassActions, nodeClasses, actions, total, err := u.repository.ReadByPagination(ctx, page, limit, nodeClassId, actionId)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, nodeClassId, actionId, domaincontractscache.Pagination[domaincontractscache.NodeClassActionItem]{
		Items: joinNodeClassActionItems(nodeClassActions, nodeClasses, actions),
		Total: total,
	}); err != nil {
		return nil, nil, nil, 0, err
	}

	return nodeClassActions, nodeClasses, actions, total, nil
}

func (u *usecase) DeleteById(ctx context.Context, id uuid.UUID) error {
	if err := u.repository.DeleteById(ctx, id); err != nil {
		return err
	}

	return u.invalidateRelations(ctx)
}

func (u *usecase) DeleteByNodeClassIdAndActionId(
	ctx context.Context,
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) error {
	if err := u.repository.DeleteByNodeClassIdAndActionId(ctx, nodeClassId, actionId); err != nil {
		return err
	}

	return u.invalidateRelations(ctx)
}

// invalidateRelations clears this pivot's own cache, the owning node
// class's cached action list (NodeClass.ReadActions), and the actions
// list pagination cache (compatible_node_class_count is embedded there).
// Narrower than role_permission's fan-out: there's no downstream
// "computed grant" cache analog here.
func (u *usecase) invalidateRelations(ctx context.Context) error {
	if err := u.cache.InvalidateAll(ctx); err != nil {
		return err
	}
	if err := u.nodeClassCache.InvalidateActions(ctx); err != nil {
		return err
	}
	if err := u.actionCache.InvalidatePagination(ctx); err != nil {
		return err
	}
	return nil
}

func joinNodeClassActionItems(
	nodeClassActions []domainmodels.NodeClassAction,
	nodeClasses []domainmodels.NodeClass,
	actions []domainmodels.Action,
) []domaincontractscache.NodeClassActionItem {
	items := make([]domaincontractscache.NodeClassActionItem, len(nodeClassActions))
	for i := range nodeClassActions {
		items[i].NodeClassAction = nodeClassActions[i]
		if i < len(nodeClasses) {
			items[i].NodeClass = nodeClasses[i]
		}
		if i < len(actions) {
			items[i].Action = actions[i]
		}
	}
	return items
}

func splitNodeClassActionItems(
	items []domaincontractscache.NodeClassActionItem,
) ([]domainmodels.NodeClassAction, []domainmodels.NodeClass, []domainmodels.Action) {
	nodeClassActions := make([]domainmodels.NodeClassAction, len(items))
	nodeClasses := make([]domainmodels.NodeClass, len(items))
	actions := make([]domainmodels.Action, len(items))

	for i, item := range items {
		nodeClassActions[i] = item.NodeClassAction
		nodeClasses[i] = item.NodeClass
		actions[i] = item.Action
	}

	return nodeClassActions, nodeClasses, actions
}
```

- [ ] **Step 2: Update `action` repocache decorator**

In `backend/internal/application/repocache/action/usecase.go`, `Create` drops `nodeClassId`, `ReadByPagination` returns `[]domainmodels.ActionListItem`:

```go
func (u *usecase) Create(
	ctx context.Context,
	name string,
	description *string,
	payloadSchemaName string,
	payloadSchemaVersion int32,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, name, description, payloadSchemaName, payloadSchemaVersion, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.cache.InvalidateAll(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
```

```go
func (u *usecase) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	search *string,
	nodeClassId *uuid.UUID,
	payloadSchemaName *string,
	payloadSchemaVersion *int32,
) ([]domainmodels.ActionListItem, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, search, nodeClassId, payloadSchemaName, payloadSchemaVersion); err != nil {
		return nil, 0, err
	} else if hit {
		return pagination.Items, pagination.Total, nil
	}

	items, total, err := u.repository.ReadByPagination(ctx, page, limit, search, nodeClassId, payloadSchemaName, payloadSchemaVersion)
	if err != nil {
		return nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, search, nodeClassId, payloadSchemaName, payloadSchemaVersion, domaincontractscache.Pagination[domainmodels.ActionListItem]{
		Items: items,
		Total: total,
	}); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
```

`ReadById`/`ReadByName`/`UpdateById`(minus `nodeClassId` param, mirrors Task B3 Step 4's `UpdateById` signature)/`DeleteById` are otherwise unchanged.

- [ ] **Step 3: Add `ReadActions` to `node_class` repocache decorator**

In `backend/internal/application/repocache/node_class/usecase.go`, add (mirrors `Role`'s `ReadPermissions` in `application/repocache/role/usecase.go`):

```go
func (u *usecase) ReadActions(ctx context.Context, nodeClassId uuid.UUID) ([]domainmodels.Action, error) {
	if items, hit, err := u.cache.GetActions(ctx, nodeClassId); err != nil {
		return nil, err
	} else if hit {
		return items, nil
	}

	items, err := u.repository.ReadActions(ctx, nodeClassId)
	if err != nil {
		return nil, err
	}

	if err := u.cache.SetActions(ctx, nodeClassId, items); err != nil {
		return nil, err
	}

	return items, nil
}
```

- [ ] **Step 4: Build**

Run: `cd backend && go build ./internal/application/repocache/... ./internal/domain/...`
Expected: clean (downstream `application/action/*`, `application/node/*` will still fail until Tasks B6-B7 — that's expected, not a regression here).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/application/repocache
git commit -m "app: add node_class_action repocache, wire ActionListItem and NodeClass.ReadActions caching"
```

---

## Task B6: Domain usecase interfaces

**Files:**
- Modify: `backend/internal/domain/usecases/action/definition.go`
- Modify: `backend/internal/domain/usecases/node/class_management.go`

**Interfaces:**
- Produces: updated `Definition`/`ClassManagement` interfaces consumed by Task B7 (application layer) and Task B9 (handlers).

- [ ] **Step 1: Update `action` definition usecase interface**

In `backend/internal/domain/usecases/action/definition.go`, remove `NodeClassId` from `CreateActionRequest`/`UpdateActionRequest`, change `ReadByPagination` return type:

```go
package domainusecasesaction

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type Definition interface {
	Create(ctx context.Context, request CreateActionRequest) (uuid.UUID, error)
	ReadById(ctx context.Context, request ReadActionByIdRequest) (*domainmodels.Action, error)
	ReadByName(ctx context.Context, request ReadActionByNameRequest) (*domainmodels.Action, error)
	ReadByPagination(ctx context.Context, request ReadActionsByPaginationRequest) ([]domainmodels.ActionListItem, int, error)
	UpdateById(ctx context.Context, request UpdateActionRequest) error
	DeleteById(ctx context.Context, request DeleteActionRequest) error
}

type CreateActionRequest struct {
	Name                 string
	Description          *string
	PayloadSchemaName    string
	PayloadSchemaVersion int32
	CreatedBy            *uuid.UUID
}

type ReadActionByIdRequest struct {
	Id uuid.UUID
}

type ReadActionByNameRequest struct {
	Name string
}

type ReadActionsByPaginationRequest struct {
	Page                 int
	Limit                int
	Search               *string
	NodeClassId          *uuid.UUID
	PayloadSchemaName    *string
	PayloadSchemaVersion *int32
}

type UpdateActionRequest struct {
	Id                   uuid.UUID
	Name                 *string
	Description          *string
	PayloadSchemaName    *string
	PayloadSchemaVersion *int32
	UpdatedBy            *uuid.UUID
}

type DeleteActionRequest struct {
	Id        uuid.UUID
	DeletedBy *uuid.UUID
}
```

(`ReadActionsByPaginationRequest.NodeClassId` stays — it's the list filter, unrelated to the removed create/update field.)

- [ ] **Step 2: Update `node` class management usecase interface**

In `backend/internal/domain/usecases/node/class_management.go`, add the action-assignment methods and their request/result types (mirrors `domainusecasesadmin.RoleManagement`'s permission methods):

```go
package domainusecasesnode

import (
	"context"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
)

type ClassManagement interface {
	Create(ctx context.Context, request CreateNodeClassRequest) (uuid.UUID, error)
	ReadById(ctx context.Context, request ReadNodeClassByIdRequest) (*domainmodels.NodeClass, error)
	ReadByName(ctx context.Context, request ReadNodeClassByNameRequest) (*domainmodels.NodeClass, error)
	ReadByPagination(ctx context.Context, request ReadNodeClassesByPaginationRequest) ([]domainmodels.NodeClass, int, error)
	UpdateById(ctx context.Context, request UpdateNodeClassRequest) error
	DeleteById(ctx context.Context, request DeleteNodeClassRequest) error
	AssignAction(ctx context.Context, request AssignNodeClassActionRequest) (uuid.UUID, error)
	RevokeAction(ctx context.Context, request RevokeNodeClassActionRequest) error
	ReadActions(ctx context.Context, request ReadNodeClassActionsRequest) ([]domainmodels.Action, error)
	ReadNodeClassActionById(ctx context.Context, request ReadNodeClassActionByIdRequest) (NodeClassActionResult, error)
	ReadNodeClassActionByNodeClassIdAndActionId(ctx context.Context, request ReadNodeClassActionByNodeClassIdAndActionIdRequest) (NodeClassActionResult, error)
	ReadNodeClassActionsByPagination(ctx context.Context, request ReadNodeClassActionsByPaginationRequest) ([]domainmodels.NodeClassAction, []domainmodels.NodeClass, []domainmodels.Action, int, error)
}

type CreateNodeClassRequest struct {
	Name        string
	Description *string
	CreatedBy   *uuid.UUID
}

type ReadNodeClassByIdRequest struct {
	Id uuid.UUID
}

type ReadNodeClassByNameRequest struct {
	Name string
}

type ReadNodeClassesByPaginationRequest struct {
	Page   int
	Limit  int
	Search *string
}

type UpdateNodeClassRequest struct {
	Id          uuid.UUID
	Name        *string
	Description *string
	UpdatedBy   *uuid.UUID
}

type DeleteNodeClassRequest struct {
	Id        uuid.UUID
	DeletedBy *uuid.UUID
}

type AssignNodeClassActionRequest struct {
	NodeClassId uuid.UUID
	ActionId    uuid.UUID
	CreatedBy   *uuid.UUID
}

type RevokeNodeClassActionRequest struct {
	NodeClassId uuid.UUID
	ActionId    uuid.UUID
}

type ReadNodeClassActionsRequest struct {
	NodeClassId uuid.UUID
}

type ReadNodeClassActionByIdRequest struct {
	Id uuid.UUID
}

type ReadNodeClassActionByNodeClassIdAndActionIdRequest struct {
	NodeClassId uuid.UUID
	ActionId    uuid.UUID
}

type ReadNodeClassActionsByPaginationRequest struct {
	Page        int
	Limit       int
	NodeClassId *uuid.UUID
	ActionId    *uuid.UUID
}

type NodeClassActionResult struct {
	NodeClassAction domainmodels.NodeClassAction
	NodeClass       domainmodels.NodeClass
	Action          domainmodels.Action
}
```

- [ ] **Step 3: Build**

Run: `cd backend && go build ./internal/domain/...`
Expected: clean (this package has no implementation, only interfaces — implementations are Task B7).

- [ ] **Step 4: Commit**

```bash
git add backend/internal/domain/usecases
git commit -m "domain: update action/node_class usecase interfaces for many-to-many"
```

---

## Task B7: Application usecases — action definition/execution, node class management

**Files:**
- Modify: `backend/internal/application/action/definition/usecase.go`
- Modify: `backend/internal/application/action/execution/usecase.go`
- Modify: `backend/internal/application/node/class_management/usecase.go`

**Interfaces:**
- Consumes: `domainusecasesrepocache.NodeClassAction` (new constructor param on both `action/execution` and `node/class_management`).
- Produces: updated `NewUsecaseImpl` constructor signatures — Task B11 (composition) must update both call sites.

- [ ] **Step 1: Update `action/definition` usecase — drop `NodeClassId`**

In `backend/internal/application/action/definition/usecase.go`:

```go
func (u *usecase) Create(ctx context.Context, request domainusecasesaction.CreateActionRequest) (uuid.UUID, error) {
	const tag = "action/definition/Create"

	name, err := applicationshared.RequiredActionName(request.Name, "name")
	if err != nil {
		return uuid.Nil, err
	}
	payloadSchemaName, err := applicationshared.RequiredSnakeCaseName(request.PayloadSchemaName, "payload_schema_name")
	if err != nil {
		return uuid.Nil, err
	}
	payloadSchemaVersion, err := applicationshared.RequiredPositiveVersion(request.PayloadSchemaVersion, "payload_schema_version")
	if err != nil {
		return uuid.Nil, err
	}

	id, err := u.action.Create(
		ctx,
		name,
		request.Description,
		payloadSchemaName,
		payloadSchemaVersion,
		request.CreatedBy,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create action", domainmodels.LoggerMeta{
			"err":        err,
			"created_by": request.CreatedBy,
		})
		return uuid.Nil, err
	}

	return id, nil
}
```

```go
func (u *usecase) ReadByPagination(
	ctx context.Context,
	request domainusecasesaction.ReadActionsByPaginationRequest,
) ([]domainmodels.ActionListItem, int, error) {
	const tag = "action/definition/ReadByPagination"

	actions, total, err := u.action.ReadByPagination(
		ctx,
		request.Page,
		request.Limit,
		request.Search,
		request.NodeClassId,
		request.PayloadSchemaName,
		request.PayloadSchemaVersion,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read actions", domainmodels.LoggerMeta{
			"err":   err,
			"page":  request.Page,
			"limit": request.Limit,
		})
		return nil, 0, err
	}

	return actions, total, nil
}
```

```go
func (u *usecase) UpdateById(ctx context.Context, request domainusecasesaction.UpdateActionRequest) error {
	const tag = "action/definition/UpdateById"

	name, err := applicationshared.OptionalActionName(request.Name, "name")
	if err != nil {
		return err
	}
	payloadSchemaName, err := applicationshared.OptionalSnakeCaseName(request.PayloadSchemaName, "payload_schema_name")
	if err != nil {
		return err
	}
	payloadSchemaVersion, err := applicationshared.OptionalPositiveVersion(request.PayloadSchemaVersion, "payload_schema_version")
	if err != nil {
		return err
	}

	if err := u.action.UpdateById(
		ctx,
		request.Id,
		name,
		request.Description,
		payloadSchemaName,
		payloadSchemaVersion,
		nil,
		request.UpdatedBy,
	); err != nil {
		u.logger.Error(ctx, tag, "failed to update action", domainmodels.LoggerMeta{
			"err":        err,
			"id":         request.Id,
			"updated_by": request.UpdatedBy,
		})
		return err
	}

	return nil
}
```

(`ReadById`/`ReadByName`/`DeleteById`/`NewUsecaseImpl`/struct fields are unchanged — only the three functions above lose `request.NodeClassId` threading.)

- [ ] **Step 2: Update `action/execution` usecase — swap compatibility check to pivot lookup**

In `backend/internal/application/action/execution/usecase.go`:

```go
type usecase struct {
	action          domainusecasesrepocache.Action
	node            domainusecasesrepocache.Node
	nodeClassAction domainusecasesrepocache.NodeClassAction
	payloadSchema   domainusecasesrepocache.PayloadSchema
	actionLog       domaincontractsrepository.ActionLog
	publisher       domaincontractsnode.Publish
	validator       domaincontractsutility.PayloadSchemaValidator
	logger          domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	action domainusecasesrepocache.Action,
	node domainusecasesrepocache.Node,
	nodeClassAction domainusecasesrepocache.NodeClassAction,
	payloadSchema domainusecasesrepocache.PayloadSchema,
	actionLog domaincontractsrepository.ActionLog,
	publisher domaincontractsnode.Publish,
	validator domaincontractsutility.PayloadSchemaValidator,
	logger domaincontractslogger.Leveled,
) domainusecasesaction.Execution {
	return &usecase{
		action:          action,
		node:            node,
		nodeClassAction: nodeClassAction,
		payloadSchema:   payloadSchema,
		actionLog:       actionLog,
		publisher:       publisher,
		validator:       validator,
		logger:          logger,
	}
}
```

And in `Dispatch`, replace the equality check:

```go
	if _, _, _, err := u.nodeClassAction.ReadByNodeClassIdAndActionId(ctx, node.NodeClassId, action.Id); err != nil {
		if errors.Is(err, domainmodels.ErrTypeNotFound) {
			return u.createActionLog(
				ctx,
				tag,
				executionId,
				request,
				&node.Id,
				domainmodels.ActionStatusUnexecuted,
				new("action is not compatible with node's class"),
			)
		}
		u.logger.Error(ctx, tag, "failed to check node class action compatibility", domainmodels.LoggerMeta{
			"err":           err,
			"node_class_id": node.NodeClassId,
			"action_id":     action.Id,
			"actor_id":      request.ActorId,
		})
		return nil, err
	}
```

(replaces the `if node.NodeClassId != action.NodeClassId { ... }` block in place — everything before and after it in `Dispatch` is unchanged.)

- [ ] **Step 3: Add `AssignAction`/`RevokeAction`/`ReadActions`/read-methods to `node/class_management` usecase**

In `backend/internal/application/node/class_management/usecase.go`, update the struct/constructor and add the new methods (mirrors `application/admin/role_management/usecase.go`'s permission methods exactly):

```go
type usecase struct {
	nodeClass       domainusecasesrepocache.NodeClass
	nodeClassAction domainusecasesrepocache.NodeClassAction
	logger          domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	nodeClass domainusecasesrepocache.NodeClass,
	nodeClassAction domainusecasesrepocache.NodeClassAction,
	logger domaincontractslogger.Leveled,
) domainusecasesnode.ClassManagement {
	return &usecase{
		nodeClass:       nodeClass,
		nodeClassAction: nodeClassAction,
		logger:          logger,
	}
}
```

```go
func (u *usecase) AssignAction(
	ctx context.Context,
	request domainusecasesnode.AssignNodeClassActionRequest,
) (uuid.UUID, error) {
	const tag = "node/class_management/AssignAction"

	id, err := u.nodeClassAction.Create(ctx, request.NodeClassId, request.ActionId, request.CreatedBy)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to assign node class action", domainmodels.LoggerMeta{
			"err":           err,
			"node_class_id": request.NodeClassId,
			"action_id":     request.ActionId,
			"created_by":    request.CreatedBy,
		})
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) RevokeAction(
	ctx context.Context,
	request domainusecasesnode.RevokeNodeClassActionRequest,
) error {
	const tag = "node/class_management/RevokeAction"

	if err := u.nodeClassAction.DeleteByNodeClassIdAndActionId(ctx, &request.NodeClassId, &request.ActionId); err != nil {
		u.logger.Error(ctx, tag, "failed to revoke node class action", domainmodels.LoggerMeta{
			"err":           err,
			"node_class_id": request.NodeClassId,
			"action_id":     request.ActionId,
		})
		return err
	}

	return nil
}

func (u *usecase) ReadActions(
	ctx context.Context,
	request domainusecasesnode.ReadNodeClassActionsRequest,
) ([]domainmodels.Action, error) {
	const tag = "node/class_management/ReadActions"

	actions, err := u.nodeClass.ReadActions(ctx, request.NodeClassId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node class actions", domainmodels.LoggerMeta{
			"err":           err,
			"node_class_id": request.NodeClassId,
		})
		return nil, err
	}

	return actions, nil
}

func (u *usecase) ReadNodeClassActionById(
	ctx context.Context,
	request domainusecasesnode.ReadNodeClassActionByIdRequest,
) (domainusecasesnode.NodeClassActionResult, error) {
	const tag = "node/class_management/ReadNodeClassActionById"

	nodeClassAction, nodeClass, action, err := u.nodeClassAction.ReadById(ctx, request.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node class action", domainmodels.LoggerMeta{
			"err": err,
			"id":  request.Id,
		})
		return domainusecasesnode.NodeClassActionResult{}, err
	}

	return domainusecasesnode.NodeClassActionResult{
		NodeClassAction: *nodeClassAction,
		NodeClass:       *nodeClass,
		Action:          *action,
	}, nil
}

func (u *usecase) ReadNodeClassActionByNodeClassIdAndActionId(
	ctx context.Context,
	request domainusecasesnode.ReadNodeClassActionByNodeClassIdAndActionIdRequest,
) (domainusecasesnode.NodeClassActionResult, error) {
	const tag = "node/class_management/ReadNodeClassActionByNodeClassIdAndActionId"

	nodeClassAction, nodeClass, action, err := u.nodeClassAction.ReadByNodeClassIdAndActionId(
		ctx,
		request.NodeClassId,
		request.ActionId,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node class action", domainmodels.LoggerMeta{
			"err":           err,
			"node_class_id": request.NodeClassId,
			"action_id":     request.ActionId,
		})
		return domainusecasesnode.NodeClassActionResult{}, err
	}

	return domainusecasesnode.NodeClassActionResult{
		NodeClassAction: *nodeClassAction,
		NodeClass:       *nodeClass,
		Action:          *action,
	}, nil
}

func (u *usecase) ReadNodeClassActionsByPagination(
	ctx context.Context,
	request domainusecasesnode.ReadNodeClassActionsByPaginationRequest,
) ([]domainmodels.NodeClassAction, []domainmodels.NodeClass, []domainmodels.Action, int, error) {
	const tag = "node/class_management/ReadNodeClassActionsByPagination"

	nodeClassActions, nodeClasses, actions, total, err := u.nodeClassAction.ReadByPagination(
		ctx,
		request.Page,
		request.Limit,
		request.NodeClassId,
		request.ActionId,
	)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to read node class actions", domainmodels.LoggerMeta{
			"err":           err,
			"page":          request.Page,
			"limit":         request.Limit,
			"node_class_id": request.NodeClassId,
			"action_id":     request.ActionId,
		})
		return nil, nil, nil, 0, err
	}

	return nodeClassActions, nodeClasses, actions, total, nil
}
```

(`Create`/`ReadById`/`ReadByName`/`ReadByPagination`/`UpdateById`/`DeleteById` are unchanged.)

- [ ] **Step 4: Build**

Run: `cd backend && go build ./internal/application/... ./internal/domain/...`
Expected: clean except `internal/composition/...` (constructor call sites — Task B11) and `internal/presentation/...` (handlers/DTOs — Tasks B8-B9).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/application/action backend/internal/application/node
git commit -m "app: wire node_class_action into action dispatch and node class management usecases"
```

---

## Task B8: Presentation DTOs — request/response

**Files:**
- Modify: `backend/internal/presentation/http/request/action.go`
- Modify: `backend/internal/presentation/http/response/action.go`
- Modify: `backend/internal/presentation/http/response/node_class.go`

**Interfaces:**
- Produces: `presentationhttpresponse.ActionListItem()`/`ActionListItems()`, `presentationhttpresponse.NodeClassAction()`/`NodeClassActionDetail()`/`NodeClassActionDetails()` — consumed by Task B9.

- [ ] **Step 1: Update `ActionPostRequest`/`ActionPatchRequest`**

In `backend/internal/presentation/http/request/action.go`, remove the `NodeClassId` field from both structs (they currently mirror `CreateActionRequest`/`UpdateActionRequest`'s shape — drop the same field there).

- [ ] **Step 2: Update `ActionResponse` and add list-item mapper**

In `backend/internal/presentation/http/response/action.go`:

```go
package presentationhttpresponse

import (
	"encoding/json"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

type ActionResponse struct {
	Id                        string          `json:"id" example:"6d9e2f5a-8b1c-4d3e-9f6a-2c5d8e1f4b07"`
	Name                      string          `json:"name" example:"pull_espresso_shot"`
	Description               string          `json:"description" example:"Runs a timed espresso extraction on the target node."`
	PayloadSchemaName         string          `json:"payload_schema_name" example:"pull_espresso_shot"`
	PayloadSchemaVersion      int32           `json:"payload_schema_version" example:"1"`
	Preferences               json.RawMessage `json:"preferences" swaggertype:"object"`
	CompatibleNodeClassCount  int             `json:"compatible_node_class_count,omitempty" example:"3"`
	AuditResponse
}

func Action(action domainmodels.Action) ActionResponse {
	return ActionResponse{
		Id:                   UUIDString(action.Id),
		Name:                 action.Name,
		Description:          action.Description,
		PayloadSchemaName:    action.PayloadSchemaName,
		PayloadSchemaVersion: action.PayloadSchemaVersion,
		Preferences:          NormalizeJSON(action.Preferences),
		AuditResponse: Audit(
			action.CreatedAt,
			action.UpdatedAt,
			action.DeletedAt,
			action.CreatedBy,
			action.UpdatedBy,
			action.DeletedBy,
		),
	}
}

func Actions(actions []domainmodels.Action) []ActionResponse {
	result := make([]ActionResponse, 0, len(actions))
	for _, action := range actions {
		result = append(result, Action(action))
	}
	return result
}

// ActionListItem maps the list read-model (base Action fields plus the
// compatible-class count) - mirrors ActionLogListItem() in action_log.go.
// The single-item Action() mapper above is unchanged and leaves
// CompatibleNodeClassCount at its zero value.
func ActionListItem(item domainmodels.ActionListItem) ActionResponse {
	resp := Action(item.Action)
	resp.CompatibleNodeClassCount = item.CompatibleNodeClassCount
	return resp
}

func ActionListItems(items []domainmodels.ActionListItem) []ActionResponse {
	result := make([]ActionResponse, 0, len(items))
	for _, item := range items {
		result = append(result, ActionListItem(item))
	}
	return result
}
```

(field name changed from having no `NodeClassId string` field — it's simply removed; `omitempty` on the count matches the precedent set by `ActionLogResponse.ActionName,omitempty` for the same "list-view-only" reason.)

- [ ] **Step 3: Add `NodeClassAction`/`NodeClassActionDetail` response types**

In `backend/internal/presentation/http/response/node_class.go`, add (mirrors `RolePermissionResponse`/`RolePermissionDetailResponse` in `response/role.go`):

```go
type NodeClassActionResponse struct {
	Id          string    `json:"id" example:"a4d7f1c9-3e6b-4a8d-9c2f-5b1e7d4a6c02"`
	NodeClassId string    `json:"node_class_id" example:"3f1c9a2e-6d4b-4e7a-8c2f-1a9b3d5e7f01"`
	ActionId    string    `json:"action_id" example:"6d9e2f5a-8b1c-4d3e-9f6a-2c5d8e1f4b07"`
	CreatedAt   time.Time `json:"created_at" example:"2026-08-03T09:00:00Z"`
	CreatedBy   *string   `json:"created_by,omitempty" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
}

type NodeClassActionDetailResponse struct {
	NodeClassAction NodeClassActionResponse `json:"node_class_action"`
	NodeClass       NodeClassResponse       `json:"node_class"`
	Action          ActionResponse          `json:"action"`
}

func NodeClassAction(nodeClassAction domainmodels.NodeClassAction) NodeClassActionResponse {
	return NodeClassActionResponse{
		Id:          UUIDString(nodeClassAction.Id),
		NodeClassId: UUIDString(nodeClassAction.NodeClassId),
		ActionId:    UUIDString(nodeClassAction.ActionId),
		CreatedAt:   nodeClassAction.CreatedAt,
		CreatedBy:   UUIDPtrString(nodeClassAction.CreatedBy),
	}
}

func NodeClassActionDetail(nodeClassAction domainmodels.NodeClassAction, nodeClass domainmodels.NodeClass, action domainmodels.Action) NodeClassActionDetailResponse {
	return NodeClassActionDetailResponse{
		NodeClassAction: NodeClassAction(nodeClassAction),
		NodeClass:       NodeClass(nodeClass),
		Action:          Action(action),
	}
}

func NodeClassActionDetails(
	nodeClassActions []domainmodels.NodeClassAction,
	nodeClasses []domainmodels.NodeClass,
	actions []domainmodels.Action,
) []NodeClassActionDetailResponse {
	result := make([]NodeClassActionDetailResponse, 0, len(nodeClassActions))
	for i, nodeClassAction := range nodeClassActions {
		var nodeClass domainmodels.NodeClass
		if i < len(nodeClasses) {
			nodeClass = nodeClasses[i]
		}

		var action domainmodels.Action
		if i < len(actions) {
			action = actions[i]
		}

		result = append(result, NodeClassActionDetail(nodeClassAction, nodeClass, action))
	}
	return result
}
```

Add `"time"` to this file's imports.

- [ ] **Step 4: Build**

Run: `cd backend && go build ./internal/presentation/http/request/... ./internal/presentation/http/response/...`
Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/presentation/http/request backend/internal/presentation/http/response
git commit -m "presentation: update action DTOs, add NodeClassAction response types"
```

---

## Task B9: Presentation handlers

**Files:**
- Modify: `backend/internal/presentation/http/handler/action/handler.go`
- Modify: `backend/internal/presentation/http/handler/node/handler.go`

**Interfaces:**
- Produces: `NodeClassActionsGet`, `NodeClassActionPost`, `NodeClassActionDeleteByPair`, `NodeClassActionGetList`, `NodeClassActionGetById`, `NodeClassActionGetByPair` on the node handler — consumed by Task B10 (routes).

- [ ] **Step 1: Update `ActionPost` — drop `node_class_id`**

In `backend/internal/presentation/http/handler/action/handler.go`:

```go
func (h *handler) ActionPost(c *echo.Context) error {
	var req presentationhttprequest.ActionPostRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	id, err := h.definitionUseCase.Create(c.Request().Context(), domainusecasesaction.CreateActionRequest{
		Name:                 req.Name,
		Description:          req.Description,
		PayloadSchemaName:    req.PayloadSchemaName,
		PayloadSchemaVersion: req.PayloadSchemaVersion,
		CreatedBy:            presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}
```

Also update the `@Param request body ...` swagger comment block if it references `node_class_id` (check; the reference dump didn't show one explicitly listing body fields beyond the struct type, so likely no doc line to edit besides the struct itself in Task B8).

- [ ] **Step 2: Update `ActionGetList` — use `ActionListItems`**

```go
	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.ActionResponse]{
		Data: presentationhttpresponse.ActionListItems(actions),
		Page: presentationhttputils.PageResponse(page, total),
	})
```

(only the `Data:` line changes — `actions, total, err := h.definitionUseCase.ReadByPagination(...)` above it is unchanged since `ReadByPagination` now returns `[]domainmodels.ActionListItem` already matching the new `actions` type.)

- [ ] **Step 3: Update `ActionPatch` — drop `node_class_id`**

```go
func (h *handler) ActionPatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.ActionPatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	if err := h.definitionUseCase.UpdateById(c.Request().Context(), domainusecasesaction.UpdateActionRequest{
		Id:                   id,
		Name:                 req.Name,
		Description:          req.Description,
		PayloadSchemaName:    req.PayloadSchemaName,
		PayloadSchemaVersion: req.PayloadSchemaVersion,
		UpdatedBy:            presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
```

(`ActionGetByName`/`ActionGetById`/`ActionDelete`/`ActionDispatchPost`/`ActionLogGetList`/`ActionLogDelete`/`actionLogFilter`/`actionLogDeleteFilter`/`actionLogStatus` are all unchanged.)

- [ ] **Step 4: Add `NodeClassAction*` handlers to `node/handler.go`**

In `backend/internal/presentation/http/handler/node/handler.go`, add after `NodeClassDelete` (mirrors `RolePermissionsGet`/`RolePermissionPost`/`RolePermissionDeleteByPair`/`RolePermissionGetList`/`RolePermissionGetById`/`RolePermissionGetByPair`/`rolePermissionPath` in `admin/handler.go`):

```go
// NodeClassActionsGet godoc
//
// @Summary Node Class Actions Get
// @Tags Node Classes
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {array} presentationhttpresponse.ActionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-classes/{id}/actions [get]
func (h *handler) NodeClassActionsGet(c *echo.Context) error {
	nodeClassId, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	actions, err := h.classUseCase.ReadActions(c.Request().Context(), domainusecasesnode.ReadNodeClassActionsRequest{NodeClassId: nodeClassId})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Actions(actions))
}

// NodeClassActionPost godoc
//
// @Summary Node Class Action
// @Tags Node Classes
// @Produce json
// @Security BearerAuth
// @Param node_class_id path string true "node_class_id"
// @Param action_id path string true "action_id"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-classes/{node_class_id}/actions/{action_id} [post]
func (h *handler) NodeClassActionPost(c *echo.Context) error {
	nodeClassId, actionId, err := h.nodeClassActionPath(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	id, err := h.classUseCase.AssignAction(c.Request().Context(), domainusecasesnode.AssignNodeClassActionRequest{
		NodeClassId: nodeClassId,
		ActionId:    actionId,
		CreatedBy:   presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// NodeClassActionDeleteByPair godoc
//
// @Summary Node Class Action Delete By Pair
// @Tags Node Classes
// @Produce json
// @Security BearerAuth
// @Param node_class_id path string true "node_class_id"
// @Param action_id path string true "action_id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-classes/{node_class_id}/actions/{action_id} [delete]
func (h *handler) NodeClassActionDeleteByPair(c *echo.Context) error {
	nodeClassId, actionId, err := h.nodeClassActionPath(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.classUseCase.RevokeAction(c.Request().Context(), domainusecasesnode.RevokeNodeClassActionRequest{
		NodeClassId: nodeClassId,
		ActionId:    actionId,
	}); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// NodeClassActionGetList godoc
//
// @Summary Node Class Action List
// @Tags Node Class Actions
// @Produce json
// @Security BearerAuth
// @Param page query int false "page number"
// @Param limit query int false "page size"
// @Param node_class_id query string false "node class id"
// @Param action_id query string false "action id"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.NodeClassActionDetailResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-class-actions [get]
func (h *handler) NodeClassActionGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	nodeClassId, err := presentationhttputils.QueryUUID(c, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	actionId, err := presentationhttputils.QueryUUID(c, "action_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	nodeClassActions, nodeClasses, actions, total, err := h.classUseCase.ReadNodeClassActionsByPagination(
		c.Request().Context(),
		domainusecasesnode.ReadNodeClassActionsByPaginationRequest{
			Page:        page.Page,
			Limit:       page.Limit,
			NodeClassId: nodeClassId,
			ActionId:    actionId,
		},
	)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.NodeClassActionDetailResponse]{
		Data: presentationhttpresponse.NodeClassActionDetails(nodeClassActions, nodeClasses, actions),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// NodeClassActionGetById godoc
//
// @Summary Node Class Action Get By ID
// @Tags Node Class Actions
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.NodeClassActionDetailResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-class-actions/{id} [get]
func (h *handler) NodeClassActionGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	result, err := h.classUseCase.ReadNodeClassActionById(c.Request().Context(), domainusecasesnode.ReadNodeClassActionByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.NodeClassActionDetail(result.NodeClassAction, result.NodeClass, result.Action))
}

// NodeClassActionGetByPair godoc
//
// @Summary Node Class Action Get By Pair
// @Tags Node Class Actions
// @Produce json
// @Security BearerAuth
// @Param node_class_id query string true "node class id"
// @Param action_id query string true "action id"
// @Success 200 {object} presentationhttpresponse.NodeClassActionDetailResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-class-actions/by-pair [get]
func (h *handler) NodeClassActionGetByPair(c *echo.Context) error {
	nodeClassId, err := presentationhttputils.QueryUUID(c, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	actionId, err := presentationhttputils.QueryUUID(c, "action_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if nodeClassId == nil || actionId == nil {
		return presentationhttputils.Error(c, presentationhttputils.BadPairQuery("node_class_id", "action_id"))
	}

	result, err := h.classUseCase.ReadNodeClassActionByNodeClassIdAndActionId(
		c.Request().Context(),
		domainusecasesnode.ReadNodeClassActionByNodeClassIdAndActionIdRequest{NodeClassId: *nodeClassId, ActionId: *actionId},
	)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.NodeClassActionDetail(result.NodeClassAction, result.NodeClass, result.Action))
}

func (h *handler) nodeClassActionPath(c *echo.Context) (uuid.UUID, uuid.UUID, error) {
	nodeClassId, err := presentationhttputils.RequiredUUID(c.Param("node_class_id"), "node_class_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	actionId, err := presentationhttputils.RequiredUUID(c.Param("action_id"), "action_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return nodeClassId, actionId, nil
}
```

(`domainusecasesnode` is already imported in this file; `uuid` is already imported per the struct/constructor dump.)

- [ ] **Step 5: Build**

Run: `cd backend && go build ./internal/presentation/...`
Expected: fails only in `internal/presentation/http/route` (interface not yet updated — Task B10) and `internal/composition` (constructors not yet updated — Task B11).

- [ ] **Step 6: Commit**

```bash
git add backend/internal/presentation/http/handler
git commit -m "presentation: add node class action handlers, update action handlers for many-to-many"
```

---

## Task B10: Routes

**Files:**
- Modify: `backend/internal/presentation/http/route/route.go`

- [ ] **Step 1: Add methods to `NodeHandler` interface**

In `backend/internal/presentation/http/route/route.go`, add to the `NodeHandler` interface (after `NodeClassDelete`):

```go
	NodeClassPatch(c *echo.Context) error
	NodeClassDelete(c *echo.Context) error
	NodeClassActionsGet(c *echo.Context) error
	NodeClassActionPost(c *echo.Context) error
	NodeClassActionDeleteByPair(c *echo.Context) error
	NodeClassActionGetList(c *echo.Context) error
	NodeClassActionGetById(c *echo.Context) error
	NodeClassActionGetByPair(c *echo.Context) error
```

- [ ] **Step 2: Add route registrations**

In `routeNode`, add after the existing `/node-classes/:node_class_id/firmwares` line and before `/node-classes/:id` (same "specific-path-before-param" convention already used in this function):

```go
	v1.GET("/node-classes/:node_class_id/firmwares", handler.FirmwareGetByNodeClassId, permission("firmware:get"))
	v1.GET("/node-classes/:id/actions", handler.NodeClassActionsGet, permission("node_class_action:get"))
	v1.POST("/node-classes/:node_class_id/actions/:action_id", handler.NodeClassActionPost, permission("node_class_action:add"))
	v1.DELETE("/node-classes/:node_class_id/actions/:action_id", handler.NodeClassActionDeleteByPair, permission("node_class_action:remove"))
	v1.GET("/node-class-actions", handler.NodeClassActionGetList, permission("node_class_action:get"))
	v1.GET("/node-class-actions/by-pair", handler.NodeClassActionGetByPair, permission("node_class_action:get"))
	v1.GET("/node-class-actions/:id", handler.NodeClassActionGetById, permission("node_class_action:get"))
	v1.GET("/firmwares/:id/config-parameters", handler.FirmwareConfigParametersGet, permission("firmware:get"))
	v1.GET("/node-classes/:id", handler.NodeClassGetById, permission("node_class:get"))
```

- [ ] **Step 3: Build**

Run: `cd backend && go build ./internal/presentation/...`
Expected: fails only in `internal/composition` now (Task B11).

- [ ] **Step 4: Commit**

```bash
git add backend/internal/presentation/http/route
git commit -m "presentation: register node class action routes"
```

---

## Task B11: Composition wiring

**Files:**
- Modify: `backend/internal/composition/main/infrastructure.go`
- Modify: `backend/internal/composition/main/application.go`
- Modify: `backend/internal/composition/seeder/infrastructure.go`

- [ ] **Step 1: Add `node_class_action` repo/cache to `main/infrastructure.go`**

Add imports:

```go
	infrastructurecachenodeclassaction "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/cache/node_class_action"
	infrastructurerepositorynodeclassaction "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/node_class_action"
```

Add struct fields (in the repository/cache blocks, alongside their `nodeClass*` neighbors):

```go
	nodeClassRepository       domaincontractsrepository.NodeClass
	nodeClassActionRepository domaincontractsrepository.NodeClassAction
```

```go
	nodeClassCache       domaincontractscache.NodeClass
	nodeClassActionCache domaincontractscache.NodeClassAction
```

Add construction in `newInfrastructure` (alongside the other repository/cache constructions):

```go
	nodeClassRepository := infrastructurerepositorynodeclass.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	nodeClassActionRepository := infrastructurerepositorynodeclassaction.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
```

```go
	nodeClassCache := infrastructurecachenodeclass.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
	nodeClassActionCache := infrastructurecachenodeclassaction.NewRedisImpl(l.drv.redisClient, config.RedisCacheNamespace, cacheTtl)
```

And add both to the `l.infra = &infrastructure{...}` literal (`nodeClassActionRepository: nodeClassActionRepository,` and `nodeClassActionCache: nodeClassActionCache,`, alongside their `nodeClass*` neighbors).

- [ ] **Step 2: Wire `node_class_action` repocache and update dependent constructors in `main/application.go`**

Add import:

```go
	applicationrepocachenodeclassaction "github.com/MateWorkspace/mate-things/backend/internal/application/repocache/node_class_action"
```

Add struct field:

```go
	nodeClassRepoCache       domainusecasesrepocache.NodeClass
	nodeClassActionRepoCache domainusecasesrepocache.NodeClassAction
```

In `newApplication`, construct it after `nodeClassRepoCache` and before it's needed by the two updated constructors:

```go
	nodeClassRepoCache := applicationrepocachenodeclass.NewRepoCacheImpl(l.infra.nodeClassRepository, l.infra.nodeClassCache)
	nodeClassActionRepoCache := applicationrepocachenodeclassaction.NewRepoCacheImpl(
		l.infra.nodeClassActionRepository,
		l.infra.nodeClassActionCache,
		l.infra.nodeClassCache,
		l.infra.actionCache,
	)
```

Update the two call sites that now take an extra param:

```go
	actionExecution := applicationactionexecution.NewUsecaseImpl(
		actionRepoCache,
		nodeRepoCache,
		nodeClassActionRepoCache,
		payloadSchemaRepoCache,
		l.infra.actionLogRepository,
		l.infra.nodePublisher,
		l.infra.payloadSchemaValidator,
		l.infra.logger,
	)
```

```go
	nodeClassManagement := applicationnodeclassmanagement.NewUsecaseImpl(nodeClassRepoCache, nodeClassActionRepoCache, l.infra.logger)
```

And add `nodeClassActionRepoCache: nodeClassActionRepoCache,` to the `l.app = &application{...}` literal.

- [ ] **Step 3: Wire `node_class_action` repository into the seeder composition**

In `backend/internal/composition/seeder/infrastructure.go`, add import:

```go
	infrastructurerepositorynodeclassaction "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/repository/node_class_action"
```

Add struct field:

```go
	nodeClassRepository       domaincontractsrepository.NodeClass
	nodeClassActionRepository domaincontractsrepository.NodeClassAction
```

Add construction and literal entry:

```go
	nodeClassRepository := infrastructurerepositorynodeclass.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
	nodeClassActionRepository := infrastructurerepositorynodeclassaction.NewPostgresImpl(l.drv.dt, &sqrQuestion, &sqrDollar)
```

(`nodeClassActionRepository: nodeClassActionRepository,` in the `l.infra = &infrastructure{...}` literal.)

- [ ] **Step 4: Build the whole backend**

Run: `cd backend && go build ./...`
Expected: fails only in `internal/application/seeder` (constructor signature will need the new dependency — Task B12) and `internal/composition/seeder/application.go` if it exists and calls that constructor (check for it; wire it there too if so, passing `l.infra.nodeClassActionRepository`).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/composition
git commit -m "composition: wire node_class_action repository/cache/repocache through main and seeder"
```

---

## Task B12: Seeder — data shape, seeding step, permissions/roles

**Files:**
- Modify: `backend/database/seeder/seeder.go`
- Modify: `backend/database/seeder/action.json`
- Modify: `backend/database/seeder/permission.json`
- Modify: `backend/database/seeder/role.json`
- Modify: `backend/internal/application/seeder/usecase.go`

- [ ] **Step 1: Update the `Action` seed struct**

In `backend/database/seeder/seeder.go`:

```go
type Action struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	NodeClassNames       []string `json:"node_class_names"`
	PayloadSchemaName    string   `json:"payload_schema_name"`
	PayloadSchemaVersion int32    `json:"payload_schema_version"`
}
```

- [ ] **Step 2: Update `action.json`**

```json
[
  {
    "name": "restart",
    "description": "Restart or reset the device.",
    "node_class_names": ["base_node"],
    "payload_schema_name": "restart",
    "payload_schema_version": 1
  }
]
```

- [ ] **Step 3: Add the three new permissions**

In `backend/database/seeder/permission.json`, add after `"role_permission:remove"`:

```json
  { "name": "role_permission:remove", "description": "Remove a permission from a role." },

  { "name": "node_class_action:get", "description": "View node-class-to-action compatibility assignments." },
  { "name": "node_class_action:add", "description": "Mark an action as compatible with a node class." },
  { "name": "node_class_action:remove", "description": "Remove an action's compatibility with a node class." },
```

- [ ] **Step 4: Grant the new permissions to `super` and `admin`**

In `backend/database/seeder/role.json`, add `"node_class_action:get", "node_class_action:add", "node_class_action:remove"` to both the `super` and `admin` permission arrays, next to the existing `"node_class:*"` entries (both roles already hold the full `node_class:*` and `action:*` sets per the dump in Task research — this mirrors that existing grouping). Leave `user` unchanged (it has no `node_class:*`/`action:add`/`action:set` grants today either).

- [ ] **Step 5: Update `seedActions` to return an id map, matching the other seed steps**

In `backend/internal/application/seeder/usecase.go`:

```go
func (u *usecase) seedActions(ctx context.Context, nodeClassIds map[string]uuid.UUID) (map[string]uuid.UUID, error) {
	const tag = "seeder/seedActions"

	ids := make(map[string]uuid.UUID, len(u.data.Actions))
	for _, action := range u.data.Actions {
		for _, nodeClassName := range action.NodeClassNames {
			if _, ok := nodeClassIds[nodeClassName]; !ok {
				err := domainmodels.NewError("node class not found for action seeding", domainmodels.ErrTypeNotFound, nil)
				u.logger.Error(ctx, tag, "unknown node class", domainmodels.LoggerMeta{
					"err":        err,
					"action":     action.Name,
					"node_class": nodeClassName,
				})
				return nil, err
			}
		}

		existing, err := u.action.ReadByName(ctx, action.Name)
		if err == nil {
			u.logger.Debug(ctx, tag, "action already exists, skipping", domainmodels.LoggerMeta{"name": action.Name})
			ids[action.Name] = existing.Id
			continue
		}
		if !errors.Is(err, domainmodels.ErrTypeNotFound) {
			u.logger.Error(ctx, tag, "failed to read action", domainmodels.LoggerMeta{
				"err":  err,
				"name": action.Name,
			})
			return nil, err
		}

		description := action.Description
		id, err := u.action.Create(
			ctx,
			action.Name,
			&description,
			action.PayloadSchemaName,
			action.PayloadSchemaVersion,
			nil,
		)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to create action", domainmodels.LoggerMeta{
				"err":  err,
				"name": action.Name,
			})
			return nil, err
		}

		u.logger.Info(ctx, tag, "action created", domainmodels.LoggerMeta{"name": action.Name})
		ids[action.Name] = id
	}

	return ids, nil
}
```

- [ ] **Step 6: Add `seedNodeClassActions`, mirroring `seedRolePermissions`'s idempotent loop**

```go
func (u *usecase) seedNodeClassActions(
	ctx context.Context,
	nodeClassIds map[string]uuid.UUID,
	actionIds map[string]uuid.UUID,
) error {
	const tag = "seeder/seedNodeClassActions"

	for _, action := range u.data.Actions {
		actionId, ok := actionIds[action.Name]
		if !ok {
			err := domainmodels.NewError("action not found for node_class_action seeding", domainmodels.ErrTypeNotFound, nil)
			u.logger.Error(ctx, tag, "unknown action", domainmodels.LoggerMeta{"err": err, "action": action.Name})
			return err
		}

		for _, nodeClassName := range action.NodeClassNames {
			nodeClassId, ok := nodeClassIds[nodeClassName]
			if !ok {
				err := domainmodels.NewError("node class not found for node_class_action seeding", domainmodels.ErrTypeNotFound, nil)
				u.logger.Error(ctx, tag, "unknown node class", domainmodels.LoggerMeta{
					"err":        err,
					"action":     action.Name,
					"node_class": nodeClassName,
				})
				return err
			}

			_, _, _, err := u.nodeClassAction.ReadByNodeClassIdAndActionId(ctx, nodeClassId, actionId)
			if err == nil {
				u.logger.Debug(ctx, tag, "node class action already assigned, skipping", domainmodels.LoggerMeta{
					"action":     action.Name,
					"node_class": nodeClassName,
				})
				continue
			}
			if !errors.Is(err, domainmodels.ErrTypeNotFound) {
				u.logger.Error(ctx, tag, "failed to read node class action", domainmodels.LoggerMeta{
					"err":        err,
					"action":     action.Name,
					"node_class": nodeClassName,
				})
				return err
			}

			if _, err := u.nodeClassAction.Create(ctx, nodeClassId, actionId, nil); err != nil {
				u.logger.Error(ctx, tag, "failed to create node class action", domainmodels.LoggerMeta{
					"err":        err,
					"action":     action.Name,
					"node_class": nodeClassName,
				})
				return err
			}

			u.logger.Info(ctx, tag, "node class action assigned", domainmodels.LoggerMeta{
				"action":     action.Name,
				"node_class": nodeClassName,
			})
		}
	}

	return nil
}
```

- [ ] **Step 7: Wire the new field/dependency and update `Run`**

Add `nodeClassAction domaincontractsrepository.NodeClassAction` to the `usecase` struct and `NewUsecaseImpl` params (insert alongside `nodeClass`), and update `Run`:

```go
	nodeClassIds, err := u.seedNodeClasses(ctx)
	if err != nil {
		return err
	}

	if err := u.seedPayloadSchemas(ctx); err != nil {
		return err
	}

	actionIds, err := u.seedActions(ctx, nodeClassIds)
	if err != nil {
		return err
	}

	if err := u.seedNodeClassActions(ctx, nodeClassIds, actionIds); err != nil {
		return err
	}

	if err := u.seedUsers(ctx, roleIds); err != nil {
		return err
	}
```

- [ ] **Step 8: Update the seeder composition call site**

Find wherever `applicationseeder.NewUsecaseImpl(...)` is called (likely `backend/internal/composition/seeder/application.go`, not yet read — locate it with `grep -rn "applicationseeder.NewUsecaseImpl" backend/internal/composition`) and add `l.infra.nodeClassActionRepository` to the call, matching the new constructor param position.

- [ ] **Step 9: Build the whole backend and vet**

Run: `cd backend && go build ./... && go vet ./...`
Expected: clean.

- [ ] **Step 10: Commit**

```bash
git add backend/database/seeder backend/internal/application/seeder backend/internal/composition
git commit -m "seeder: seed node_class_action pivot rows, add node_class_action permissions"
```

---

## Task F1: Frontend API clients

**Files:**
- Modify: `frontend/src/lib/api/actions.ts`
- Modify: `frontend/src/lib/api/node-classes.ts`

- [ ] **Step 1: Update `ActionResponse`/`CreateActionRequest`/`UpdateActionRequest`**

In `frontend/src/lib/api/actions.ts`:

```ts
export interface ActionResponse extends AuditFields {
  id: string;
  name: string;
  description: string;
  payload_schema_name: string;
  payload_schema_version: number;
  preferences: Record<string, unknown>;
  compatible_node_class_count?: number;
}

export interface CreateActionRequest {
  name: string;
  description?: string;
  payload_schema_name: string;
  payload_schema_version: number;
}

export interface UpdateActionRequest {
  name?: string;
  description?: string;
  payload_schema_name?: string;
  payload_schema_version?: number;
}
```

(`DispatchActionRequest`, `ListActionsQuery` — `node_class_id` stays as a filter param, unchanged — and every function are otherwise unchanged.)

- [ ] **Step 2: Add `NodeClassActionResponse` type and pivot functions to `node-classes.ts`**

In `frontend/src/lib/api/node-classes.ts`, add (mirrors `getRolePermissions`/`assignRolePermission`/`revokeRolePermission` in `lib/api/roles.ts`):

```ts
import type { ActionResponse } from "@/lib/api/actions";

export interface NodeClassActionResponse {
  id: string;
  node_class_id: string;
  action_id: string;
  created_at: string;
  created_by?: string;
}

export interface NodeClassActionDetailResponse {
  node_class_action: NodeClassActionResponse;
  node_class: NodeClassResponse;
  action: ActionResponse;
}

export interface ListNodeClassActionsQuery extends PageQuery {
  node_class_id?: string;
  action_id?: string;
}

export async function getNodeClassActions(
  nodeClassId: string,
): Promise<ActionResponse[]> {
  return apiFetch(`/node-classes/${nodeClassId}/actions`);
}

export async function assignNodeClassAction(
  nodeClassId: string,
  actionId: string,
): Promise<IdResponse> {
  return apiFetch(`/node-classes/${nodeClassId}/actions/${actionId}`, {
    method: "POST",
  });
}

export async function revokeNodeClassAction(
  nodeClassId: string,
  actionId: string,
): Promise<void> {
  return apiFetch(`/node-classes/${nodeClassId}/actions/${actionId}`, {
    method: "DELETE",
  });
}

export async function listNodeClassActions(
  query: ListNodeClassActionsQuery = {},
): Promise<PageDataResponse<NodeClassActionDetailResponse>> {
  return apiFetch(`/node-class-actions${buildQuery(query)}`);
}
```

(add `IdResponse` to the existing `import type { AuditFields, IdResponse, PageDataResponse, PageQuery }` line at the top if not already present — check the current import list first, per the earlier full-file dump it currently imports `AuditFields, IdResponse, PageDataResponse, PageQuery` already, so no import change needed beyond adding the `ActionResponse` type-only import shown above.)

- [ ] **Step 3: Typecheck**

Run: `cd frontend && export PATH=/tmp/mate-node-v24.18.1/bin:$PATH && node_modules/.bin/tsc --noEmit 2>&1 | grep -v "\.test\.\|vitest\|@playwright\|e2e/"`
Expected: errors only in files not yet updated (`ActionForm.tsx`, `ActionCard.tsx`, `DispatchActionDialog.tsx`, `actions/page.tsx`, `actions/[id]/page.tsx`, `actions/_lib/actions.ts` — all fixed in Tasks F2-F4).

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/api/actions.ts frontend/src/lib/api/node-classes.ts
git commit -m "frontend: update action/node-class API clients for many-to-many"
```

---

## Task F2: Action form — remove node-class field

**Files:**
- Modify: `frontend/src/app/(authenticated)/actions/_components/ActionForm.tsx`
- Modify: `frontend/src/app/(authenticated)/actions/_lib/actions.ts`

- [ ] **Step 1: Remove the node-class select and `nodeClasses` prop from `ActionForm.tsx`**

Remove this whole block from `EditorDialog`:

```tsx
        <div>
          <Label htmlFor={`${id}-class`}>Node class</Label>
          <select
            id={`${id}-class`}
            name="node_class_id"
            defaultValue={action?.node_class_id ?? ""}
            required
            className="border-control-border bg-background focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
          >
            <option value="">Select node class</option>
            {nodeClasses.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
        </div>
```

Remove `nodeClasses: readonly NodeClassResponse[];` from `ActionFormProps`, the `import type { NodeClassResponse } from "@/lib/api/node-classes";` import, and the `nodeClasses` prop/param from `ActionForm`'s destructured props, its passthrough to `EditorDialog`, and `EditorDialog`'s own props type (`Omit<ActionFormProps, "canEdit" | "canDelete">` stays correct automatically once the field is gone from `ActionFormProps`).

- [ ] **Step 2: Remove `node_class_id` from the create/update server actions**

In `frontend/src/app/(authenticated)/actions/_lib/actions.ts`:

```ts
function values(formData: FormData) {
  return {
    id: String(formData.get("action_id") ?? "").trim(),
    name: String(formData.get("name") ?? "").trim(),
    description: String(formData.get("description") ?? "").trim(),
    schemaName: String(formData.get("payload_schema_name") ?? "").trim(),
    schemaVersion: Number(formData.get("payload_schema_version")),
  };
}

function validateAction(
  input: ReturnType<typeof values>,
): ActionFormState | null {
  const fieldErrors: Record<string, string> = {};
  if (!input.name) fieldErrors.name = "Name is required.";
  if (!input.schemaName)
    fieldErrors.payload_schema_name = "Payload schema is required.";
  if (!Number.isInteger(input.schemaVersion) || input.schemaVersion <= 0)
    fieldErrors.payload_schema_version = "Select a valid schema version.";
  return Object.keys(fieldErrors).length
    ? {
        status: "error",
        title: "Check the action",
        message: "Complete the required fields.",
        fieldErrors,
      }
    : null;
}
```

In `createActionFormAction`:

```ts
    await createAction({
      name: input.name,
      description: input.description,
      payload_schema_name: input.schemaName,
      payload_schema_version: input.schemaVersion,
    });
```

In `updateActionFormAction`:

```ts
    await updateAction(input.id, {
      name: input.name,
      description: input.description,
      payload_schema_name: input.schemaName,
      payload_schema_version: input.schemaVersion,
    });
```

(`dispatchActionFormAction`/`deleteActionFormAction` are unchanged.)

- [ ] **Step 3: Typecheck this slice**

Run: `cd frontend && export PATH=/tmp/mate-node-v24.18.1/bin:$PATH && node_modules/.bin/tsc --noEmit 2>&1 | grep -i "ActionForm\|actions/_lib"`
Expected: no matches from these two files (remaining errors are in files not yet touched).

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(authenticated)/actions/_components/ActionForm.tsx" "frontend/src/app/(authenticated)/actions/_lib/actions.ts"
git commit -m "frontend: remove node-class field from action create/edit form"
```

---

## Task F3: Action card and list page

**Files:**
- Modify: `frontend/src/app/(authenticated)/actions/_components/ActionCard.tsx`
- Modify: `frontend/src/app/(authenticated)/actions/page.tsx`

- [ ] **Step 1: Update `ActionCard.tsx` — compatible-class count instead of name**

```tsx
// frontend/src/app/(authenticated)/actions/_components/ActionCard.tsx
import Link from "next/link";

import ResourceCard from "@/components/collection/ResourceCard";
import type { ActionResponse } from "@/lib/api/actions";

interface ActionCardProps {
  action: ActionResponse;
}

export default function ActionCard({ action }: ActionCardProps) {
  const count = action.compatible_node_class_count ?? 0;
  return (
    <ResourceCard
      title={action.name}
      summary={`${count} compatible node class${count === 1 ? "" : "es"}`}
    >
      <p className="text-foreground/75 min-h-12">
        {action.description || "No description provided."}
      </p>
      <dl className="mt-4 space-y-2">
        <div className="flex justify-between gap-3">
          <dt className="text-muted-foreground">Schema</dt>
          <dd>
            {action.payload_schema_name} · v{action.payload_schema_version}
          </dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-muted-foreground">Updated</dt>
          <dd>
            <time dateTime={action.updated_at ?? action.created_at}>
              {new Date(
                action.updated_at ?? action.created_at,
              ).toLocaleString()}
            </time>
          </dd>
        </div>
      </dl>
      <Link
        href={`/actions/${action.id}`}
        className="bg-primary text-surface mt-5 inline-flex min-h-11 w-full items-center justify-center rounded-xl px-4 text-sm font-semibold"
      >
        View and dispatch
      </Link>
    </ResourceCard>
  );
}
```

(callers `actions/page.tsx:87` and `NodeOperationsWorkspace.tsx:43` both already only pass `action={...}` — the `nodeClassName` prop drop needs no caller edits beyond `actions/page.tsx`'s own cleanup in the next step, since `NodeOperationsWorkspace.tsx` never passed it.)

- [ ] **Step 2: Simplify `actions/page.tsx` — drop `listAllNodeClasses`/`classNames`**

```tsx
// frontend/src/app/(authenticated)/actions/page.tsx
import type { Metadata } from "next";

import ActionCard from "@/app/(authenticated)/actions/_components/ActionCard";
import ActionForm from "@/app/(authenticated)/actions/_components/ActionForm";
import Pagination from "@/components/collection/Pagination";
import Input from "@/components/ui/input";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { listActions } from "@/lib/api/actions";
import { listAllPayloadSchemas } from "@/lib/api/payload-schemas";
import { parsePageQuery } from "@/lib/collection-query";
import { requirePermission } from "@/lib/session";

export const metadata: Metadata = { title: "Actions — Mate Things" };
type RawSearchParams = Record<string, string | string[] | undefined>;

export default async function ActionsPage({
  searchParams,
}: {
  searchParams: Promise<RawSearchParams>;
}) {
  const [raw, { permissions }] = await Promise.all([
    searchParams,
    requirePermission("action:get"),
  ]);
  const pageQuery = parsePageQuery(raw);
  const nodeClassId =
    typeof raw.node_class_id === "string" ? raw.node_class_id.trim() : "";
  const schemaName =
    typeof raw.payload_schema_name === "string"
      ? raw.payload_schema_name.trim()
      : "";
  const [result, schemas] = await Promise.all([
    listActions({
      ...pageQuery,
      node_class_id: nodeClassId || undefined,
      payload_schema_name: schemaName || undefined,
    }),
    permissions.has("payload_schema:get")
      ? listAllPayloadSchemas()
      : Promise.resolve([]),
  ]);
  const canCreate =
    permissions.has("action:add") && permissions.has("payload_schema:get");

  return (
    <main className="mx-auto w-full max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title="Actions"
        description="Define fleet commands, match them to compatible node classes, and dispatch them safely."
        actions={canCreate ? <ActionForm schemas={schemas} /> : undefined}
      />
      <form className="border-border bg-muted grid gap-3 rounded-2xl border p-4 sm:grid-cols-[1fr_1fr_auto]">
        <Input
          name="node_class_id"
          defaultValue={nodeClassId}
          placeholder="Node class ID"
          aria-label="Filter by node class ID"
        />
        <Input
          name="payload_schema_name"
          defaultValue={schemaName}
          placeholder="Payload schema"
          aria-label="Filter by payload schema"
        />
        <button className="bg-primary text-surface rounded-xl px-4 py-2.5 text-sm font-semibold">
          Apply
        </button>
      </form>
      <p className="text-muted-foreground text-sm">
        {result.page.total_items} action definitions
      </p>
      {result.data.length ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {result.data.map((action) => (
            <ActionCard key={action.id} action={action} />
          ))}
        </div>
      ) : (
        <EmptyState
          title={
            nodeClassId || schemaName ? "No matching actions" : "No actions yet"
          }
          description={
            nodeClassId || schemaName
              ? "Clear or change the filters to broaden the results."
              : "Create an action definition when your role permits it."
          }
        />
      )}
      <Pagination page={result.page} pathname="/actions" searchParams={raw} />
    </main>
  );
}
```

- [ ] **Step 3: Typecheck, lint, format**

Run:
```bash
cd frontend && export PATH=/tmp/mate-node-v24.18.1/bin:$PATH
node_modules/.bin/prettier --write "src/app/(authenticated)/actions/_components/ActionCard.tsx" "src/app/(authenticated)/actions/page.tsx"
node_modules/.bin/eslint "src/app/(authenticated)/actions/_components/ActionCard.tsx" "src/app/(authenticated)/actions/page.tsx"
```
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(authenticated)/actions/_components/ActionCard.tsx" "frontend/src/app/(authenticated)/actions/page.tsx"
git commit -m "frontend: show compatible-class count on action cards, drop node-class dropdown data from actions list page"
```

---

## Task F4: Dispatch dialog and action detail page

**Files:**
- Modify: `frontend/src/app/(authenticated)/actions/_components/DispatchActionDialog.tsx`
- Modify: `frontend/src/app/(authenticated)/actions/[id]/page.tsx`

**Interfaces:**
- Consumes: `getNodeClassActions`/`listNodeClassActions` from Task F1.

- [ ] **Step 1: `DispatchActionDialog.tsx` — filter by a compatible-class-id set**

```tsx
// frontend/src/app/(authenticated)/actions/_components/DispatchActionDialog.tsx
"use client";

import Link from "next/link";
import { useActionState, useId, useState } from "react";

import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import type { ActionResponse } from "@/lib/api/actions";
import type { NodeResponse } from "@/lib/api/nodes";

import {
  dispatchActionFormAction,
  type ActionFormState,
} from "../_lib/actions";

const EMPTY_STATE: ActionFormState = { status: "idle" };

interface DispatchActionDialogProps {
  action: ActionResponse;
  nodes: readonly NodeResponse[];
  compatibleNodeClassIds: ReadonlySet<string>;
}

export default function DispatchActionDialog({
  action,
  nodes,
  compatibleNodeClassIds,
}: DispatchActionDialogProps) {
  const [open, setOpen] = useState(false);
  const [state, formAction, pending] = useActionState(
    dispatchActionFormAction,
    EMPTY_STATE,
  );
  const id = useId();
  const compatibleNodes = nodes.filter((node) =>
    compatibleNodeClassIds.has(node.node_class_id),
  );
  return (
    <>
      <Button type="button" onClick={() => setOpen(true)}>
        Dispatch action
      </Button>
      <Dialog
        open={open}
        onClose={() => setOpen(false)}
        title={`Dispatch ${action.name}`}
        variant="sheet"
      >
        {state.status === "success" && state.executionId ? (
          <div className="space-y-4">
            <p>{state.message}</p>
            <Link
              className="bg-primary text-surface inline-flex rounded-xl px-4 py-2.5 text-sm font-semibold"
              href={`/action-history?execution_id=${encodeURIComponent(state.executionId)}`}
            >
              View execution
            </Link>
          </div>
        ) : (
          <form action={formAction} className="space-y-4">
            <input type="hidden" name="action_id" value={action.id} />
            <div>
              <Label htmlFor={`${id}-node`}>Compatible node</Label>
              <select
                id={`${id}-node`}
                name="node_id"
                required
                className="border-control-border bg-background focus-visible:ring-focus min-h-11 w-full rounded-xl border px-3.5 text-sm focus-visible:ring-2 focus-visible:outline-none"
              >
                <option value="">Select node</option>
                {compatibleNodes.map((node) => (
                  <option key={node.id} value={node.id}>
                    {node.name} · {node.device_id}
                  </option>
                ))}
              </select>
              {compatibleNodes.length === 0 ? (
                <p className="text-warning mt-2 text-sm">
                  No compatible nodes are available.
                </p>
              ) : null}
            </div>
            <div>
              <Label htmlFor={`${id}-payload`}>Payload JSON</Label>
              <textarea
                id={`${id}-payload`}
                name="payload"
                rows={8}
                defaultValue={"{}"}
                className="border-control-border bg-background focus-visible:ring-focus w-full rounded-xl border p-3 font-mono text-sm focus-visible:ring-2 focus-visible:outline-none"
              />
              {state.fieldErrors?.payload ? (
                <p className="text-critical mt-1 text-sm">
                  {state.fieldErrors.payload}
                </p>
              ) : null}
            </div>
            <div>
              <Label htmlFor={`${id}-executed`}>Execute at (optional)</Label>
              <Input
                id={`${id}-executed`}
                name="executed_at"
                type="datetime-local"
              />
            </div>
            <p aria-live="polite" className="text-critical text-sm">
              {state.message}
            </p>
            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="secondary"
                disabled={pending}
                onClick={() => setOpen(false)}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={pending || compatibleNodes.length === 0}>
                {pending ? "Dispatching…" : "Dispatch"}
              </Button>
            </div>
          </form>
        )}
      </Dialog>
    </>
  );
}
```

- [ ] **Step 2: `actions/[id]/page.tsx` — fetch compatible classes, drop the node-class edit wiring, add a read-only compatible-classes list**

```tsx
// frontend/src/app/(authenticated)/actions/[id]/page.tsx
import type { Metadata } from "next";
import { notFound } from "next/navigation";

import ActionForm from "@/app/(authenticated)/actions/_components/ActionForm";
import DispatchActionDialog from "@/app/(authenticated)/actions/_components/DispatchActionDialog";
import Card from "@/components/ui/card";
import PageHeader from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/states";
import { getActionById } from "@/lib/api/actions";
import { ApiError } from "@/lib/api/client";
import { listNodeClassActions } from "@/lib/api/node-classes";
import { listNodes } from "@/lib/api/nodes";
import { listAllPayloadSchemas } from "@/lib/api/payload-schemas";
import { requirePermission } from "@/lib/session";

export const metadata: Metadata = { title: "Action details — Mate Things" };

export default async function ActionDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const [{ id }, { permissions }] = await Promise.all([
    params,
    requirePermission("action:get"),
  ]);
  let action;
  try {
    action = await getActionById(id);
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  }
  const [schemas, compatibleClasses, nodesPage] = await Promise.all([
    permissions.has("payload_schema:get")
      ? listAllPayloadSchemas()
      : Promise.resolve([]),
    permissions.has("node_class_action:get")
      ? listNodeClassActions({ action_id: id, limit: 48 })
      : Promise.resolve({
          data: [],
          page: { page: 1, limit: 48, total_items: 0 },
        }),
    permissions.has("node:get") && permissions.has("action:dispatch")
      ? listNodes({ page: 1, limit: 100 })
      : Promise.resolve({
          data: [],
          page: { page: 1, limit: 100, total_items: 0 },
        }),
  ]);
  const compatibleNodeClassIds = new Set(
    compatibleClasses.data.map((item) => item.node_class.id),
  );
  const canEdit =
    permissions.has("action:set") && permissions.has("payload_schema:get");

  return (
    <main className="mx-auto w-full max-w-5xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <PageHeader
        title={action.name}
        description={action.description || "Fleet action definition"}
        actions={
          <div className="flex flex-wrap gap-2">
            {permissions.has("action:dispatch") &&
            permissions.has("node:get") ? (
              <DispatchActionDialog
                action={action}
                nodes={nodesPage.data}
                compatibleNodeClassIds={compatibleNodeClassIds}
              />
            ) : null}
            {canEdit || permissions.has("action:remove") ? (
              <ActionForm
                action={action}
                schemas={schemas}
                canEdit={canEdit}
                canDelete={permissions.has("action:remove")}
              />
            ) : null}
          </div>
        }
      />
      <Card>
        <dl className="grid gap-4 sm:grid-cols-2">
          <div>
            <dt className="text-muted-foreground text-sm">Payload schema</dt>
            <dd className="mt-1 font-medium">
              {action.payload_schema_name} · version{" "}
              {action.payload_schema_version}
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground text-sm">Created</dt>
            <dd className="mt-1">
              <time dateTime={action.created_at}>
                {new Date(action.created_at).toLocaleString()}
              </time>
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground text-sm">Definition ID</dt>
            <dd className="mt-1 font-mono text-xs break-all">{action.id}</dd>
          </div>
        </dl>
      </Card>
      {permissions.has("node_class_action:get") ? (
        <Card>
          <h2 className="font-display text-primary text-lg">
            Compatible node classes
          </h2>
          {compatibleClasses.data.length ? (
            <ul className="mt-3 space-y-2">
              {compatibleClasses.data.map((item) => (
                <li
                  key={item.node_class.id}
                  className="bg-muted rounded-xl px-3 py-2 text-sm"
                >
                  {item.node_class.name}
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-muted-foreground mt-3 text-sm">
              Not assigned to any node class yet. Manage this from the node
              class's detail page.
            </p>
          )}
        </Card>
      ) : null}
      {permissions.has("action:dispatch") && !permissions.has("node:get") ? (
        <EmptyState
          title="Node access required"
          description="node:get permission is required to choose a compatible dispatch target."
        />
      ) : null}
    </main>
  );
}
```

- [ ] **Step 3: Typecheck, lint, format**

Run:
```bash
cd frontend && export PATH=/tmp/mate-node-v24.18.1/bin:$PATH
node_modules/.bin/prettier --write "src/app/(authenticated)/actions/_components/DispatchActionDialog.tsx" "src/app/(authenticated)/actions/[id]/page.tsx"
node_modules/.bin/eslint "src/app/(authenticated)/actions/_components/DispatchActionDialog.tsx" "src/app/(authenticated)/actions/[id]/page.tsx"
```
Expected: clean.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/(authenticated)/actions/_components/DispatchActionDialog.tsx" "frontend/src/app/(authenticated)/actions/[id]/page.tsx"
git commit -m "frontend: dispatch dialog and action detail page use pivot-based compatibility"
```

---

## Task F5: Node class detail page — compatible-actions toggle checklist

**Files:**
- Create: `frontend/src/app/(authenticated)/node-classes/_components/NodeClassActionChecklist.tsx`
- Modify: `frontend/src/app/(authenticated)/node-classes/_lib/actions.ts`
- Modify: `frontend/src/app/(authenticated)/node-classes/[id]/page.tsx`

**Interfaces:**
- Consumes: `getNodeClassActions`/`assignNodeClassAction`/`revokeNodeClassAction` from Task F1, `listActions` from `lib/api/actions.ts`.

- [ ] **Step 1: New `updateNodeClassActionsAction` server action — mirrors `updateRoleAssignmentsAction`**

In `frontend/src/app/(authenticated)/node-classes/_lib/actions.ts`, add (needs `assignNodeClassAction`, `getNodeClassActions`, `revokeNodeClassAction` imported from `@/lib/api/node-classes`):

```ts
import {
  assignNodeClassAction,
  createNodeClass,
  deleteNodeClass,
  getNodeClassActions,
  revokeNodeClassAction,
  updateNodeClass,
} from "@/lib/api/node-classes";
```

```ts
export async function updateNodeClassActionsAction(
  _previousState: FormActionState,
  formData: FormData,
): Promise<FormActionState> {
  const session = await requireSessionContext();
  const nodeClassId = String(formData.get("node_class_id") ?? "").trim();
  if (!nodeClassId || !session.permissions.has("node_class_action:get")) {
    return permissionDenied();
  }

  const current = new Set(
    (await getNodeClassActions(nodeClassId)).map((item) => item.id),
  );
  const desired = formData.getAll("action_ids").map(String);
  const desiredSet = new Set(desired);
  const assign = desired.filter((id) => !current.has(id));
  const revoke = [...current].filter((id) => !desiredSet.has(id));

  if (assign.length && !session.permissions.has("node_class_action:add")) {
    return permissionDenied();
  }
  if (revoke.length && !session.permissions.has("node_class_action:remove")) {
    return permissionDenied();
  }

  const results = await Promise.allSettled([
    ...assign.map((actionId) => assignNodeClassAction(nodeClassId, actionId)),
    ...revoke.map((actionId) => revokeNodeClassAction(nodeClassId, actionId)),
  ]);
  const failures = results.filter(
    (result) => result.status === "rejected",
  ).length;

  if (failures) {
    return {
      status: "error",
      title: "Assignments partially updated",
      message: `${results.length - failures} changes succeeded and ${failures} failed. The current authoritative assignments were reloaded.`,
    };
  }

  return {
    status: "success",
    title: "Assignments updated",
    message: `${results.length} action changes saved.`,
  };
}
```

(`permissionDenied()`/`FormActionState` already exist in this file per the earlier full dump — reuse them as-is.)

- [ ] **Step 2: New `NodeClassActionChecklist.tsx` — mirrors `PermissionGroups.tsx` + the `<form>`/"Save assignments" wrapper from `RoleDetails.tsx`, flattened (no resource grouping)**

```tsx
// frontend/src/app/(authenticated)/node-classes/_components/NodeClassActionChecklist.tsx
"use client";

import { useActionState, useEffect } from "react";
import { useRouter } from "next/navigation";

import Button from "@/components/ui/button";
import type { ActionResponse } from "@/lib/api/actions";

import { updateNodeClassActionsAction } from "../_lib/actions";
import type { FormActionState } from "../_lib/actions";

const EMPTY_STATE: FormActionState = { status: "idle" };

export default function NodeClassActionChecklist({
  nodeClassId,
  actions,
  selected,
  editable,
}: {
  nodeClassId: string;
  actions: readonly ActionResponse[];
  selected: ReadonlySet<string>;
  editable: boolean;
}) {
  const [state, formAction, pending] = useActionState(
    updateNodeClassActionsAction,
    EMPTY_STATE,
  );
  const router = useRouter();
  useEffect(() => {
    if (state.status === "success") {
      router.refresh();
    }
  }, [state, router]);

  return (
    <form action={formAction} className="space-y-4">
      <input type="hidden" name="node_class_id" value={nodeClassId} />
      <fieldset className="border-border rounded-xl border p-4">
        <legend className="font-display text-primary px-2 text-lg">
          Compatible actions
        </legend>
        <div className="space-y-2">
          {actions.map((action) => (
            <label
              key={action.id}
              className="bg-muted flex gap-3 rounded-lg p-3"
            >
              {editable ? (
                <input
                  type="checkbox"
                  name="action_ids"
                  value={action.id}
                  defaultChecked={selected.has(action.id)}
                  className="accent-primary mt-1 size-4"
                />
              ) : (
                <span aria-hidden="true" className="mt-1">
                  {selected.has(action.id) ? "✓" : "—"}
                </span>
              )}
              <span>
                <span className="block text-sm font-semibold">
                  {action.name}
                </span>
                <span className="text-muted-foreground text-xs">
                  {action.description || "No description."}
                </span>
              </span>
            </label>
          ))}
        </div>
      </fieldset>
      <p
        aria-live="polite"
        className={
          state.status === "error"
            ? "text-critical text-sm"
            : "text-success text-sm"
        }
      >
        {state.message}
      </p>
      {editable ? (
        <Button type="submit" disabled={pending}>
          {pending ? "Saving…" : "Save assignments"}
        </Button>
      ) : null}
    </form>
  );
}
```

- [ ] **Step 3: Wire the checklist into `node-classes/[id]/page.tsx`**

Add imports:

```tsx
import { listActions } from "@/lib/api/actions";
import { getNodeClassActions } from "@/lib/api/node-classes";

import NodeClassActionChecklist from "../_components/NodeClassActionChecklist";
```

(`listActions`/`ActionResponse` are already imported in this file per the earlier dump — reuse the existing import, only adding `getNodeClassActions` and the new component.)

In the `Promise.all` that fetches `nodeClass`/`firmwares`/`nodes`/`actions`, add a parallel fetch for the assignment checklist data, gated on the new permission:

```tsx
  const canManageActions =
    permissions.has("node_class_action:get") && permissions.has("action:get");

  const [nodeClass, firmwares, nodes, actions, allActions, assignedActions] =
    await Promise.all([
      nodeClassRequest,
      canReadFirmware
        ? listFirmwaresByNodeClassId(id, { limit: 12 })
        : Promise.resolve(null),
      canReadNodes
        ? listNodes({ node_class_id: id, limit: 12 })
        : Promise.resolve(null),
      canReadActions
        ? listActions({ node_class_id: id, limit: 12 })
        : Promise.resolve(null),
      canManageActions
        ? listActions({ limit: 100 }).then((result) => result.data)
        : Promise.resolve([]),
      canManageActions ? getNodeClassActions(id) : Promise.resolve([]),
    ]);
```

Add the checklist section after the existing `{actions ? <RelationshipSection ...> ... </RelationshipSection> : null}` block, before the closing `</main>`:

```tsx
      {canManageActions ? (
        <section aria-labelledby="compatible-actions-heading">
          <h2
            id="compatible-actions-heading"
            className="font-display text-primary mb-3 text-2xl tracking-wide"
          >
            Compatible actions
          </h2>
          <NodeClassActionChecklist
            nodeClassId={id}
            actions={allActions}
            selected={new Set(assignedActions.map((action) => action.id))}
            editable={
              permissions.has("node_class_action:add") ||
              permissions.has("node_class_action:remove")
            }
          />
        </section>
      ) : null}
```

(The existing read-only "Actions" `RelationshipSection` stays as-is — it's a different view, a quick-glance card grid linking out to `/actions?node_class_id=`, distinct from this editable checklist. Both coexist, matching how the spec's §8 described adding the checklist "alongside" the existing section.)

- [ ] **Step 4: Typecheck, lint, format**

Run:
```bash
cd frontend && export PATH=/tmp/mate-node-v24.18.1/bin:$PATH
node_modules/.bin/prettier --write "src/app/(authenticated)/node-classes/_components/NodeClassActionChecklist.tsx" "src/app/(authenticated)/node-classes/_lib/actions.ts" "src/app/(authenticated)/node-classes/[id]/page.tsx"
node_modules/.bin/eslint "src/app/(authenticated)/node-classes/_components/NodeClassActionChecklist.tsx" "src/app/(authenticated)/node-classes/_lib/actions.ts" "src/app/(authenticated)/node-classes/[id]/page.tsx"
```
Expected: clean.

- [ ] **Step 5: Full frontend typecheck**

Run: `cd frontend && export PATH=/tmp/mate-node-v24.18.1/bin:$PATH && node_modules/.bin/tsc --noEmit 2>&1 | grep -v "\.test\.\|vitest\|@playwright\|e2e/"`
Expected: clean (or only the pre-existing unrelated missing-dependency errors documented earlier this session).

- [ ] **Step 6: Commit**

```bash
git add "frontend/src/app/(authenticated)/node-classes"
git commit -m "frontend: add compatible-actions toggle checklist to node class detail page"
```

---

## Task T6: Verification

**Files:** none (verification only).

- [ ] **Step 1: Backend build/vet**

Run: `cd backend && go build ./... && go vet ./...`
Expected: clean. Note the pre-existing unrelated `node_log` handler test failures documented earlier this session are not a regression signal here (they predate this work and this plan adds no `_test.go` files).

- [ ] **Step 2: Frontend typecheck/lint/format**

Run:
```bash
cd frontend && export PATH=/tmp/mate-node-v24.18.1/bin:$PATH
node_modules/.bin/tsc --noEmit 2>&1 | grep -v "\.test\.\|vitest\|@playwright\|e2e/"
node_modules/.bin/eslint src
node_modules/.bin/prettier --check src
```
Expected: clean (modulo the documented pre-existing missing-test-dependency errors).

- [ ] **Step 3: Rebuild and start the stack**

Run: `docker compose build --no-cache app && docker compose up -d` (from repo root; `--no-cache` per this session's established practice to avoid stale-cache false positives).

- [ ] **Step 4: Live-verify via a scratch Playwright script**

Using the established pattern (`playwright-core` pointed at `/home/dodol/.cache/ms-playwright/chromium-1234/chrome-linux64/chrome`, `--no-sandbox`, admin/`ChangeMe123!` login):
1. Log in, go to `/node-classes/<base_node id>`, confirm the "Compatible actions" checklist renders with `restart` checked (seeded).
2. Uncheck `restart`, click "Save assignments", confirm success message and that a refresh shows it unchecked.
3. Go to `/actions`, confirm the `restart` card now shows "0 compatible node classes".
4. Go to `/actions/<restart id>`, confirm "Compatible node classes" shows the empty state, and the dispatch dialog shows "No compatible nodes are available."
5. Go back to `/node-classes/<base_node id>`, re-check `restart`, save.
6. Go to `/actions/<restart id>`, dispatch to the seeded node, confirm the dispatch succeeds (proving the pivot-based compatibility check in `execution/usecase.go` works end-to-end).
7. Confirm zero errors in `docker logs mate-things-app-1` across the whole pass.

- [ ] **Step 5: Clean up scratch files, report**

Remove the scratch Playwright directory. This is the final task — no code changes if verification passes; report and stop. If anything fails, treat it as a normal implementation bug in whichever task introduced it (root-cause before patching, per this session's established debugging discipline) rather than patching at the verification layer.
