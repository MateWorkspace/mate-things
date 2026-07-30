# Typed, Self-Describing Error Handling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every backend error response carries a message that's specific to what actually happened (sourced from the error itself, not a hand-typed per-endpoint guess), with conflict errors identifying exactly which field collided.

**Architecture:** Add 11 new `ErrorType` sentinels for user-triggerable unique-constraint conflicts; teach the shared Postgres error mapper to resolve `23505` violations to the specific sentinel via `strings.Contains` on the constraint name; fix generic not-found messages at their repo-layer source; rewrite the central HTTP `Error()` helper to derive title+message from the error itself via a static mapping table (dropping the per-call-site message argument entirely — a breaking signature change that forces every one of 224 call sites to be updated); simplify the frontend to match.

**Tech Stack:** Go (backend), `github.com/jackc/pgx/v5`, `github.com/labstack/echo/v5`, TypeScript/Next.js (frontend, `_lib/actions.ts` files only).

## Global Constraints

- **No test-writing.** Per explicit user instruction: skip all test steps (no new `_test.go` files, no updates to existing tests unless a change breaks compilation of an existing test — in which case fix only the compile error, don't expand test coverage). Verification is `go build ./...`, `go vet ./...`, and manual curl/browser checks.
- Use `strings.Contains` on `pgErr.ConstraintName` to resolve which conflict sentinel applies — per explicit user instruction, not an exact map lookup.
- 5xx-class errors (`ErrTypeFailure`, `ErrTypeUnknown`, plus token/timeout/unimplemented) always return a fixed, safe generic message regardless of the real internal error text — never regress this existing safety property.
- `BadRequest(c, message)` in `presentation/http/utils/error.go` is unchanged (still takes an explicit message) — it fires before a usecase runs (request-shape/JSON-parse failures), so there's no domain error to derive from.
- Follow existing code style: `domainmodels.NewError(message, errType, source)` for wrapping, `MapPgxError` package `infrastructurerepositoryshared`, existing import aliasing conventions (`domainmodels`, `infrastructurerepositoryshared`, etc.) exactly as already used in each file.

---

### Task 1: Domain error types + conflict-aware Postgres error mapper

**Files:**
- Modify: `backend/internal/domain/models/error.go`
- Modify: `backend/internal/infrastructure/repository/shared/error_pgx.go`

**Interfaces:**
- Produces: 11 new `domainmodels.ErrorType` sentinels; `infrastructurerepositoryshared.ConflictMatch{Contains string; Type domainmodels.ErrorType}`; `MapPgxError(message string, err error, conflicts ...ConflictMatch) error` (backward-compatible — existing 2-arg call sites keep compiling unchanged).

- [ ] **Step 1: Add the 11 new sentinels**

In `backend/internal/domain/models/error.go`, add to the existing `var (...)` block (after `ErrTypeUnknown`):

```go
	ErrTypeUsernameExists             = errors.New("USERNAME_EXISTS")
	ErrTypeRoleNameExists              = errors.New("ROLE_NAME_EXISTS")
	ErrTypePermissionNameExists        = errors.New("PERMISSION_NAME_EXISTS")
	ErrTypeActionNameExists            = errors.New("ACTION_NAME_EXISTS")
	ErrTypeNodeClassNameExists         = errors.New("NODE_CLASS_NAME_EXISTS")
	ErrTypeFirmwareNameExists          = errors.New("FIRMWARE_NAME_EXISTS")
	ErrTypeNodeDeviceIdExists          = errors.New("NODE_DEVICE_ID_EXISTS")
	ErrTypePayloadSchemaVersionExists  = errors.New("PAYLOAD_SCHEMA_VERSION_EXISTS")
	ErrTypeRolePermissionExists        = errors.New("ROLE_PERMISSION_EXISTS")
	ErrTypeFirmwareConfigKeyExists     = errors.New("FIRMWARE_CONFIG_KEY_EXISTS")
	ErrTypeNodeConfigKeyExists         = errors.New("NODE_CONFIG_KEY_EXISTS")
```

- [ ] **Step 2: Rewrite `MapPgxError` to resolve conflicts via `strings.Contains`**

In `backend/internal/infrastructure/repository/shared/error_pgx.go`, add `"strings"` to the imports, add the `ConflictMatch` type, and replace the function body:

```go
package infrastructurerepositoryshared

import (
	"errors"
	"strings"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ConflictMatch resolves a Postgres unique-violation to a specific
// ErrorType when the fired constraint's name contains Contains. Used at
// Create/Update call sites where a single statement can violate more than
// one unique constraint on the same table, so the generic 23505 code alone
// doesn't say which one fired.
type ConflictMatch struct {
	Contains string
	Type     domainmodels.ErrorType
}

func MapPgxError(message string, err error, conflicts ...ConflictMatch) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domainmodels.NewError(message, domainmodels.ErrTypeNotFound, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			for _, m := range conflicts {
				if strings.Contains(pgErr.ConstraintName, m.Contains) {
					return domainmodels.NewError(message, m.Type, err)
				}
			}
			return domainmodels.NewError(message, domainmodels.ErrTypeConflict, err)
		case "23503":
			return domainmodels.NewError(message, domainmodels.ErrTypeConflict, err)
		case "22001", "22P02", "23502", "23514":
			return domainmodels.NewError(message, domainmodels.ErrTypeValidation, err)
		}
	}

	return domainmodels.NewError(message, domainmodels.ErrTypeUnknown, err)
}

func QueryBuildError(message string, err error) error {
	return domainmodels.NewError(message, domainmodels.ErrTypeFailure, err)
}

func NotFound(message string, err error) error {
	return domainmodels.NewError(message, domainmodels.ErrTypeNotFound, err)
}
```

- [ ] **Step 3: Build**

```bash
cd backend && go build ./... && go vet ./...
```
Expected: succeeds (no call sites are broken — the new parameter is variadic).

- [ ] **Step 4: Commit**

```bash
git add internal/domain/models/error.go internal/infrastructure/repository/shared/error_pgx.go
git commit -m "feat(backend): add per-field conflict error types and constraint-aware Postgres error mapping"
```

---

### Task 2: Repository layer — admin cluster (user, role, role_permission, permission, payload_schema)

**Files:**
- Modify: `backend/internal/infrastructure/repository/user/postgres.go`
- Modify: `backend/internal/infrastructure/repository/role/postgres.go`
- Modify: `backend/internal/infrastructure/repository/role_permission/postgres.go`
- Modify: `backend/internal/infrastructure/repository/permission/postgres.go`
- Modify: `backend/internal/infrastructure/repository/payload_schema/postgres.go`

**Interfaces:**
- Consumes: `infrastructurerepositoryshared.ConflictMatch`, `domainmodels.ErrType{Username,RoleName,PermissionName,PayloadSchemaVersion,RolePermission}Exists` from Task 1.

- [ ] **Step 1: `user/postgres.go` — wire the username conflict, fix not-found messages**

Line 50 (`Create`), change:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create user", err)
```
to:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create user", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "username", Type: domainmodels.ErrTypeUsernameExists},
		)
```

Line 161 (`Update` — username can be changed), change:
```go
		return infrastructurerepositoryshared.MapPgxError("failed to update user", err)
```
to:
```go
		return infrastructurerepositoryshared.MapPgxError(
			"failed to update user", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "username", Type: domainmodels.ErrTypeUsernameExists},
		)
```

Lines 67 and 84 (the two single-row reads — `ReadById`/`ReadByUsername`), change both instances of:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read user", err)
```
to:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("user not found", err)
```

- [ ] **Step 2: `role/postgres.go` — wire the role-name conflict, fix not-found messages**

At line 57 and line 72 (both currently `"failed to create role"`, inside/outside the same insert path — open the file and confirm which one wraps the actual `INSERT` on the `roles` table; it's the one whose surrounding code executes the `squirrel.Insert("roles")...` query), change that one to:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create role", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypeRoleNameExists},
		)
```
(Leave the other "failed to create role"/"failed to unset default roles" wrapping calls as plain `MapPgxError(message, err)` — they wrap different statements in the same transaction that can't violate the name constraint.)

Line 197 (`Update`), change:
```go
			return infrastructurerepositoryshared.MapPgxError("failed to update role", err)
```
to:
```go
			return infrastructurerepositoryshared.MapPgxError(
				"failed to update role", err,
				infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypeRoleNameExists},
			)
```

Lines 89, 106, 123 (three single-row reads), change all three instances of:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read role", err)
```
to:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("role not found", err)
```

- [ ] **Step 3: `role_permission/postgres.go` — wire the pair conflict, fix not-found messages**

Line 46 (`Create`), change:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create role permission", err)
```
to:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create role permission", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "role_id_permission_id", Type: domainmodels.ErrTypeRolePermissionExists},
		)
```

Lines 66 and 87 (single-row reads), change both instances of:
```go
		return nil, nil, nil, infrastructurerepositoryshared.MapPgxError("failed to read role permission", err)
```
to:
```go
		return nil, nil, nil, infrastructurerepositoryshared.MapPgxError("this permission is not assigned to the role", err)
```

- [ ] **Step 4: `permission/postgres.go` — wire the name conflict, fix not-found messages**

Line 47 (`Create`), change:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create permission", err)
```
to:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create permission", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypePermissionNameExists},
		)
```

Line 134 (`Update`), change:
```go
		return infrastructurerepositoryshared.MapPgxError("failed to update permission", err)
```
to:
```go
		return infrastructurerepositoryshared.MapPgxError(
			"failed to update permission", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypePermissionNameExists},
		)
```

Lines 64 and 81 (single-row reads), change both instances of:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read permission", err)
```
to:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("permission not found", err)
```

- [ ] **Step 5: `payload_schema/postgres.go` — wire the name+version conflict, fix not-found messages**

Line 51 (`Create`), change:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create payload schema", err)
```
to:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create payload schema", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name_version", Type: domainmodels.ErrTypePayloadSchemaVersionExists},
		)
```

Line 163 (`Update`) — open the file and check whether `name`/`version` are mutable fields on update. If they are, apply the same `ConflictMatch` as above to that call site. If the update only ever touches non-unique fields (e.g. just `preferences`/description-style fields), leave line 163 unchanged.

Lines 68, 89, 106 (single-row reads), change all three instances of:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read payload schema", err)
```
to:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("payload schema not found", err)
```

- [ ] **Step 6: Build**

```bash
cd backend && go build ./... && go vet ./...
```
Expected: succeeds.

- [ ] **Step 7: Commit**

```bash
git add internal/infrastructure/repository/user internal/infrastructure/repository/role internal/infrastructure/repository/role_permission internal/infrastructure/repository/permission internal/infrastructure/repository/payload_schema
git commit -m "feat(backend): specific conflict types and not-found messages for admin-cluster repositories"
```

---

### Task 3: Repository layer — node cluster (node, node_class, firmware, firmware_config_parameter, node_config_value)

**Files:**
- Modify: `backend/internal/infrastructure/repository/node/postgres.go`
- Modify: `backend/internal/infrastructure/repository/node_class/postgres.go`
- Modify: `backend/internal/infrastructure/repository/firmware/postgres.go`
- Modify: `backend/internal/infrastructure/repository/firmware_config_parameter/postgres.go`
- Modify: `backend/internal/infrastructure/repository/node_config_value/postgres.go`

**Interfaces:**
- Consumes: `infrastructurerepositoryshared.ConflictMatch`, `domainmodels.ErrType{NodeDeviceId,NodeClassName,FirmwareName,FirmwareConfigKey,NodeConfigKey}Exists` from Task 1.

- [ ] **Step 1: `node/postgres.go` — wire the device_id conflict, fix not-found messages**

Line 52 (`Create`), change:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create node", err)
```
to:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create node", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "device_id", Type: domainmodels.ErrTypeNodeDeviceIdExists},
		)
```

Line 93 (`upsert node registration`), change:
```go
		return nil, false, infrastructurerepositoryshared.MapPgxError("failed to upsert node registration", err)
```
to:
```go
		return nil, false, infrastructurerepositoryshared.MapPgxError(
			"failed to upsert node registration", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "device_id", Type: domainmodels.ErrTypeNodeDeviceIdExists},
		)
```

Line 187 (`Update`) — open the file and check whether `device_id` is mutable on update. If yes, apply the same `ConflictMatch`. If `device_id` is immutable post-registration, leave unchanged.

Lines 110 and 127 (single-row reads), change both instances of:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read node", err)
```
to:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("node not found", err)
```

- [ ] **Step 2: `node_class/postgres.go` — wire the name conflict, fix not-found messages**

Line 47 (`Create`), change:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create node class", err)
```
to:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create node class", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypeNodeClassNameExists},
		)
```

Line 134 (`Update`), change:
```go
		return infrastructurerepositoryshared.MapPgxError("failed to update node class", err)
```
to:
```go
		return infrastructurerepositoryshared.MapPgxError(
			"failed to update node class", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypeNodeClassNameExists},
		)
```

Lines 64 and 81 (single-row reads), change both instances of:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read node class", err)
```
to:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("node class not found", err)
```

- [ ] **Step 3: `firmware/postgres.go` — wire the name conflict, fix not-found messages**

Line 50 (`Create`), change:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create firmware", err)
```
to:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create firmware", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypeFirmwareNameExists},
		)
```

Line 151 (`Update`), change:
```go
		return infrastructurerepositoryshared.MapPgxError("failed to update firmware", err)
```
to:
```go
		return infrastructurerepositoryshared.MapPgxError(
			"failed to update firmware", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypeFirmwareNameExists},
		)
```

Lines 67 and 84 (single-row reads), change both instances of:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read firmware", err)
```
to:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("firmware not found", err)
```

- [ ] **Step 4: `firmware_config_parameter/postgres.go` — wire the firmware_id+key conflict**

Line 58 (`upsert firmware config parameter`), change:
```go
				return infrastructurerepositoryshared.MapPgxError("failed to upsert firmware config parameter", err)
```
to:
```go
				return infrastructurerepositoryshared.MapPgxError(
					"failed to upsert firmware config parameter", err,
					infrastructurerepositoryshared.ConflictMatch{Contains: "firmware_id_key", Type: domainmodels.ErrTypeFirmwareConfigKeyExists},
				)
```
(Lines 49, 77, 83 are delete/read/scan of the *collection* of parameters — leave unchanged.)

- [ ] **Step 5: `node_config_value/postgres.go` — wire the node_id+key conflict, fix not-found message**

Line 85 (`upsert node config value`), change:
```go
		return infrastructurerepositoryshared.MapPgxError("failed to upsert node config value", err)
```
to:
```go
		return infrastructurerepositoryshared.MapPgxError(
			"failed to upsert node config value", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "node_id_key", Type: domainmodels.ErrTypeNodeConfigKeyExists},
		)
```

Line 65 (single-row read), change:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read node config value", err)
```
to:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("node config value not found", err)
```

- [ ] **Step 6: Build**

```bash
cd backend && go build ./... && go vet ./...
```
Expected: succeeds.

- [ ] **Step 7: Commit**

```bash
git add internal/infrastructure/repository/node internal/infrastructure/repository/node_class internal/infrastructure/repository/firmware internal/infrastructure/repository/firmware_config_parameter internal/infrastructure/repository/node_config_value
git commit -m "feat(backend): specific conflict types and not-found messages for node-cluster repositories"
```

---

### Task 4: Repository layer — action, and confirming the remaining repos need no change

**Files:**
- Modify: `backend/internal/infrastructure/repository/action/postgres.go`

**Interfaces:**
- Consumes: `infrastructurerepositoryshared.ConflictMatch`, `domainmodels.ErrTypeActionNameExists` from Task 1.

- [ ] **Step 1: `action/postgres.go` — wire the name conflict, fix not-found messages**

Line 50 (`Create`), change:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError("failed to create action", err)
```
to:
```go
		return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
			"failed to create action", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypeActionNameExists},
		)
```

Line 143 (`Update`), change:
```go
		return infrastructurerepositoryshared.MapPgxError("failed to update action", err)
```
to:
```go
		return infrastructurerepositoryshared.MapPgxError(
			"failed to update action", err,
			infrastructurerepositoryshared.ConflictMatch{Contains: "name", Type: domainmodels.ErrTypeActionNameExists},
		)
```

Lines 67 and 84 (single-row reads), change both instances of:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("failed to read action", err)
```
to:
```go
		return nil, infrastructurerepositoryshared.MapPgxError("action not found", err)
```

- [ ] **Step 2: Confirm no changes needed in `action_log`, `telemetry_record`, `node_log`**

These three tables have no user-facing unique constraint (verified against the migrations in the design spec) and no single-row `Read*` method that surfaces `ErrTypeNotFound` to a user in practice (only list/count/scan/create/delete). No edits needed — this step is just a confirmation, not a code change. Run:
```bash
grep -n "MapPgxError(" internal/infrastructure/repository/action_log/postgres.go internal/infrastructure/repository/telemetry_record/postgres.go internal/infrastructure/repository/node_log/postgres.go
```
and confirm none of the listed lines are single-row reads (i.e. none return a bare `nil, ...` for one object — they all return `nil, 0, ...` for lists or are create/update/delete). If you find one that was missed, fix its message the same way as the other tasks (`"failed to X"` → `"<resource> not found"`).

- [ ] **Step 3: Build**

```bash
cd backend && go build ./... && go vet ./...
```
Expected: succeeds.

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/repository/action
git commit -m "feat(backend): specific conflict type and not-found messages for action repository"
```

---

### Task 5: Application layer audit

**Files:**
- Read/potentially modify: all files under `backend/internal/application/**` (34 usecase files)

**Interfaces:**
- Consumes: nothing new — this is a verification/cleanup pass over existing code.

- [ ] **Step 1: Sweep for bare/unwrapped errors**

```bash
cd backend
grep -rLn "domainmodels" internal/application --include="*.go" | grep -v _test
```
For any file listed (meaning it never references `domainmodels` at all — a signal it may be returning raw errors without any type), open it and confirm every returned error either originates from a repo call (already correctly typed after Tasks 1-4) or a call to `applicationshared` validation helpers (already correctly typed). Fix any found bare `return nil, err`-without-wrapping paths by wrapping with `domainmodels.NewError(<specific message>, <correct ErrType>, err)`.

- [ ] **Step 2: Sweep for inline business-rule errors with weak messages**

```bash
grep -rn "domainmodels.NewError(" internal/application --include="*.go" | grep -v _test
```
Read each result. For messages that are vague ("operation not allowed", "invalid state", etc.) or that don't say *why* (e.g. don't name the resource/reason), rewrite the message to be specific and user-safe while keeping the same `ErrType`. Example pattern to look for and the fix: a message like `"cannot delete"` for a default-role-deletion guard should read `"the default role cannot be deleted; set another role as default first"`.

- [ ] **Step 3: Build**

```bash
cd backend && go build ./... && go vet ./...
```
Expected: succeeds.

- [ ] **Step 4: Commit**

```bash
git add internal/application
git commit -m "fix(backend): sharpen application-layer error messages found during audit"
```
(If Steps 1-2 found nothing to change, skip the commit — note in the task ledger that the audit found no gaps.)

---

### Task 6: Presentation layer — central error mapping, response shape, and all 224 handler call sites

**Files:**
- Modify: `backend/internal/presentation/http/utils/error.go`
- Modify: `backend/internal/presentation/http/response/common.go`
- Modify: `backend/internal/presentation/http/handler/action/handler.go`
- Modify: `backend/internal/presentation/http/handler/admin/handler.go`
- Modify: `backend/internal/presentation/http/handler/auth/handler.go`
- Modify: `backend/internal/presentation/http/handler/node/handler.go`
- Modify: `backend/internal/presentation/http/handler/node_log/handler.go`
- Modify: `backend/internal/presentation/http/handler/preferences/handler.go`
- Modify: `backend/internal/presentation/http/handler/profile/handler.go`
- Modify: `backend/internal/presentation/http/handler/telemetry/handler.go`

**Interfaces:**
- Consumes: all 11 new `ErrType*Exists` sentinels from Task 1.
- Produces: `presentationhttputils.Error(c *echo.Context, err error) error` (message parameter removed — every caller in the codebase must be updated in this same task to keep the build green).

- [ ] **Step 1: Rewrite `error.go`**

Replace the full contents of `backend/internal/presentation/http/utils/error.go`:

```go
package presentationhttputils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	"github.com/labstack/echo/v5"
)

func BadRequest(c *echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, presentationhttpresponse.ErrorResponse{
		Error:   "Bad Request",
		Message: message,
	})
}

type errorMapping struct {
	status  int
	title   string
	message string // non-empty: always used, overriding the domain error's own message
}

var errorMappings = []struct {
	errType domainmodels.ErrorType
	mapping errorMapping
}{
	{domainmodels.ErrTypeNotFound, errorMapping{http.StatusNotFound, "Not Found", ""}},
	{domainmodels.ErrTypeUsernameExists, errorMapping{http.StatusConflict, "Already Exists", "This username is already taken."}},
	{domainmodels.ErrTypeRoleNameExists, errorMapping{http.StatusConflict, "Already Exists", "A role with this name already exists."}},
	{domainmodels.ErrTypePermissionNameExists, errorMapping{http.StatusConflict, "Already Exists", "A permission with this name already exists."}},
	{domainmodels.ErrTypeActionNameExists, errorMapping{http.StatusConflict, "Already Exists", "An action with this name already exists."}},
	{domainmodels.ErrTypeNodeClassNameExists, errorMapping{http.StatusConflict, "Already Exists", "A node class with this name already exists."}},
	{domainmodels.ErrTypeFirmwareNameExists, errorMapping{http.StatusConflict, "Already Exists", "A firmware with this name already exists."}},
	{domainmodels.ErrTypeNodeDeviceIdExists, errorMapping{http.StatusConflict, "Already Exists", "A node with this device ID is already registered."}},
	{domainmodels.ErrTypePayloadSchemaVersionExists, errorMapping{http.StatusConflict, "Already Exists", "This payload schema name and version already exists."}},
	{domainmodels.ErrTypeRolePermissionExists, errorMapping{http.StatusConflict, "Already Exists", "This permission is already assigned to the role."}},
	{domainmodels.ErrTypeFirmwareConfigKeyExists, errorMapping{http.StatusConflict, "Already Exists", "This config key already exists for the firmware."}},
	{domainmodels.ErrTypeNodeConfigKeyExists, errorMapping{http.StatusConflict, "Already Exists", "This config key already exists for the node."}},
	{domainmodels.ErrTypeConflict, errorMapping{http.StatusConflict, "Already Exists", ""}},
	{domainmodels.ErrTypeBadArgs, errorMapping{http.StatusBadRequest, "Invalid Format", ""}},
	{domainmodels.ErrTypeValidation, errorMapping{http.StatusBadRequest, "Invalid Format", ""}},
	{domainmodels.ErrTypeBadState, errorMapping{http.StatusPreconditionFailed, "Invalid State", ""}},
	{domainmodels.ErrTypeForbidden, errorMapping{http.StatusForbidden, "Access Denied", ""}},
	{domainmodels.ErrTypeUnauthorized, errorMapping{http.StatusUnauthorized, "Unauthorized", ""}},
	{domainmodels.ErrTypeTokenExpired, errorMapping{http.StatusUnauthorized, "Session Expired", "Your session has expired. Please sign in again."}},
	{domainmodels.ErrTypeTokenInvalid, errorMapping{http.StatusUnauthorized, "Invalid Token", "Your session is no longer valid. Please sign in again."}},
	{domainmodels.ErrTypeTimeout, errorMapping{http.StatusGatewayTimeout, "Request Timeout", "The request took too long. Please try again."}},
	{domainmodels.ErrTypeUnimplemented, errorMapping{http.StatusNotImplemented, "Not Implemented", "This feature isn't available yet."}},
	{domainmodels.ErrTypeFailure, errorMapping{http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."}},
	{domainmodels.ErrTypeUnknown, errorMapping{http.StatusInternalServerError, "Internal Server Error", "Something went wrong on our end. Please try again later."}},
}

const genericServerErrorMessage = "Something went wrong on our end. Please try again later."

func Error(c *echo.Context, err error) error {
	if err == nil {
		return nil
	}

	for _, m := range errorMappings {
		if errors.Is(err, m.errType) {
			message := m.mapping.message
			if message == "" {
				message = domainMessage(err)
			}
			return c.JSON(m.mapping.status, presentationhttpresponse.ErrorResponse{
				Error:   m.mapping.title,
				Message: message,
			})
		}
	}

	return c.JSON(http.StatusInternalServerError, presentationhttpresponse.ErrorResponse{
		Error:   "Internal Server Error",
		Message: genericServerErrorMessage,
	})
}

func domainMessage(err error) string {
	var domainErr *domainmodels.Error
	if errors.As(err, &domainErr) && strings.TrimSpace(domainErr.Message) != "" {
		return domainErr.Message
	}
	return genericServerErrorMessage
}

func MissingResponse(name string) error {
	return domainmodels.NewError(
		fmt.Sprintf("%s is missing from usecase response", name),
		domainmodels.ErrTypeUnknown,
		nil,
	)
}
```

Note: `ErrTypeUsernameExists`, etc. must be checked *before* the base `ErrTypeConflict` in the slice (they are, above) — `errors.Is` only matches the exact sentinel each `*domainmodels.Error.Is` compares against, so ordering doesn't actually affect correctness here (each specific sentinel and the base sentinel are distinct, non-overlapping values), but keeping specific-before-generic in the source is clearer to read.

- [ ] **Step 2: Drop the `Details` field from `ErrorResponse`**

In `backend/internal/presentation/http/response/common.go`, change:
```go
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid Format"`
	Message string `json:"message" example:"Please select a valid node class."`
	Details string `json:"details" example:"node_class_id is required"`
}
```
to:
```go
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid Format"`
	Message string `json:"message" example:"Please select a valid node class."`
}
```

- [ ] **Step 3: Update all 224 handler call sites — scripted pass**

Run this from `backend/`:

```bash
python3 - <<'EOF'
import re
import pathlib

files = [
    "internal/presentation/http/handler/action/handler.go",
    "internal/presentation/http/handler/admin/handler.go",
    "internal/presentation/http/handler/auth/handler.go",
    "internal/presentation/http/handler/node/handler.go",
    "internal/presentation/http/handler/node_log/handler.go",
    "internal/presentation/http/handler/preferences/handler.go",
    "internal/presentation/http/handler/profile/handler.go",
    "internal/presentation/http/handler/telemetry/handler.go",
]

# Matches: `, "<message ending the Error(...) call>")` at end of line,
# where the string is a plain Go double-quoted literal (no embedded raw
# newlines). Replaces it with just `)`, stripping the trailing message arg.
pattern = re.compile(r',\s*"(?:[^"\\]|\\.)*"\)\s*$', re.MULTILINE)

for rel in files:
    path = pathlib.Path(rel)
    text = path.read_text()
    new_text = pattern.sub(")", text)
    path.write_text(new_text)
    print(rel, "changed" if new_text != text else "UNCHANGED")
EOF
```

- [ ] **Step 4: Fix the one known multi-line call by hand**

In `backend/internal/presentation/http/handler/node/handler.go`, find (around line 1179 pre-edit, may have shifted slightly after Step 3):
```go
		return presentationhttputils.Error(
			c,
			domainmodels.NewError("value is required", domainmodels.ErrTypeValidation, nil),
			"Please provide a config value.",
		)
```
Change it to:
```go
		return presentationhttputils.Error(
			c,
			domainmodels.NewError("value is required", domainmodels.ErrTypeValidation, nil),
		)
```

- [ ] **Step 5: Verify no `Error(` call retains a trailing message argument**

```bash
cd backend
grep -rn 'presentationhttputils\.Error([^)]*,\s*"' internal/presentation/http/handler
```
Expected: no output. If any lines are printed, they're calls the script's regex didn't match (e.g. a different multi-line shape) — open each and remove the trailing message argument by hand, following the same pattern as Step 4.

- [ ] **Step 6: Build**

```bash
cd backend && go build ./... && go vet ./...
```
Expected: succeeds — this is the real verification that every one of the 224 call sites was updated correctly (a leftover 3-argument call is now a compile error against the new 2-argument signature).

- [ ] **Step 7: Commit**

```bash
git add internal/presentation/http/utils/error.go internal/presentation/http/response/common.go internal/presentation/http/handler
git commit -m "feat(backend): derive HTTP error responses from the error itself, drop per-call-site messages"
```

---

### Task 7: Swagger regeneration

**Files:**
- Modify (generated): `backend/docs/swagger/docs.go`, `backend/docs/swagger/swagger.json`, `backend/docs/swagger/swagger.yaml`

**Interfaces:** none — pure regeneration.

- [ ] **Step 1: Regenerate**

```bash
cd backend
swag init -g cmd/main/main.go -o docs/swagger --parseInternal --parseDependency
```

- [ ] **Step 2: Confirm the `ErrorResponse` schema no longer lists `details`**

```bash
grep -A5 '"ErrorResponse"' docs/swagger/swagger.json
```
Expected: only `error` and `message` properties.

- [ ] **Step 3: Build (confirm generated code compiles)**

```bash
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add docs/swagger
git commit -m "docs(backend): regenerate swagger for the simplified ErrorResponse shape"
```

---

### Task 8: Frontend — simplify error handling to match the single-message response

**Files:**
- Modify: `frontend/src/lib/api/client.ts`
- Modify: `frontend/src/app/(authenticated)/admin/users/_lib/actions.ts`
- Modify: `frontend/src/app/(authenticated)/admin/access-control/_lib/actions.ts`
- Modify: `frontend/src/app/(authenticated)/admin/payload-schemas/_lib/actions.ts`
- Modify: `frontend/src/components/preferences/preferences-actions.ts`

**Interfaces:**
- Produces: `ApiError` with only `status`, `title`, `message` (no `details`).

- [ ] **Step 1: Drop `details` from `ApiError` and its construction**

In `frontend/src/lib/api/client.ts`, change:
```ts
export class ApiError extends Error {
  readonly status: number;
  readonly title: string;
  readonly details?: string;

  constructor(
    status: number,
    title: string,
    message: string,
    details?: string,
  ) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.title = title;
    this.details = details;
  }
}
```
to:
```ts
export class ApiError extends Error {
  readonly status: number;
  readonly title: string;

  constructor(status: number, title: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.title = title;
  }
}
```

And change the throw site:
```ts
    throw new ApiError(
      response.status,
      errorBody?.error ?? "Something Went Wrong",
      errorBody?.message ?? response.statusText,
      errorBody?.details,
    );
```
to:
```ts
    throw new ApiError(
      response.status,
      errorBody?.error ?? "Something Went Wrong",
      errorBody?.message ?? response.statusText,
    );
```

- [ ] **Step 2: Simplify each `failure()` helper**

In each of the four files below, find the `failure()` function and change the `message` line from `error.details?.trim() || error.message` to `error.message`.

`frontend/src/app/(authenticated)/admin/users/_lib/actions.ts:31` — change:
```ts
        ? error.details?.trim() || error.message
```
to:
```ts
        ? error.message
```

`frontend/src/app/(authenticated)/admin/access-control/_lib/actions.ts:37` — inside its `failure()` function, change:
```ts
        ? error.details?.trim() || error.message
```
to:
```ts
        ? error.message
```

`frontend/src/app/(authenticated)/admin/payload-schemas/_lib/actions.ts:29` — same change:
```ts
        ? error.details?.trim() || error.message
```
to:
```ts
        ? error.message
```

`frontend/src/components/preferences/preferences-actions.ts:74` — same change:
```ts
          ? error.details?.trim() || error.message
```
to:
```ts
          ? error.message
```

(Verified: `admin/users/_lib/actions.ts:31`, `admin/access-control/_lib/actions.ts:37`,
`admin/payload-schemas/_lib/actions.ts:29`, and `preferences-actions.ts:74`
all currently contain exactly this `error.details?.trim() || error.message`
line — all four need the edit, none should be skipped.)

- [ ] **Step 3: Typecheck**

```bash
cd frontend
export PATH=/tmp/mate-node-v24.18.1/bin:$PATH
node_modules/.bin/tsc --noEmit -p tsconfig.json 2>&1 | grep -iE "client.ts|_lib/actions.ts|preferences-actions.ts"
```
Expected: no output (no new errors in the touched files — pre-existing unrelated errors about missing test packages are fine to ignore).

- [ ] **Step 4: Commit**

```bash
git add src/lib/api/client.ts "src/app/(authenticated)/admin/users/_lib/actions.ts" "src/app/(authenticated)/admin/access-control/_lib/actions.ts" "src/app/(authenticated)/admin/payload-schemas/_lib/actions.ts" src/components/preferences/preferences-actions.ts
git commit -m "refactor(frontend): drop the redundant details fallback now that message is always specific"
```

---

### Task 9: End-to-end manual verification

**Files:** none — verification only.

- [ ] **Step 1: Rebuild and restart the container**

```bash
cd /home/dodol/Repositories/mate/mate-things
docker compose build app
docker compose up -d --force-recreate app
sleep 5
docker compose logs app --tail 30
```
Expected: no startup errors.

- [ ] **Step 2: Log in and exercise each conflict type via curl**

Using the same login-then-action curl technique established earlier in this project (extract `$ACTION_1:0`/`$ACTION_KEY` from the rendered form, or call the underlying `/api/v1/...` endpoints directly with a Bearer token from `/api/v1/auth/login`), confirm for at least these cases:
- Duplicate username on user creation → `409` with `{"error":"Already Exists","message":"This username is already taken."}`.
- Duplicate role name → `409` with the role-specific message.
- Duplicate permission name → `409` with the permission-specific message.
- Duplicate node device_id registration → `409` with the device-specific message.
- A nonexistent user ID read → `404` with `{"error":"Not Found","message":"user not found"}`.
- A validation failure (e.g. missing required `username` on create) → `400` with the actual validation message (not a generic override).

- [ ] **Step 3: Confirm 5xx safety net**

Temporarily point `BE_POSTGRES_HOST` at an unreachable host in `.env`, restart the container, and confirm a request that hits the database returns `500` with the generic `"Something went wrong on our end. Please try again later."` message — not any raw connection error text. Revert `.env` afterward and restart the container again.

- [ ] **Step 4: Frontend toast check**

In the admin/users UI, create a user with a duplicate username; confirm the toast (from the earlier fix) now reads "This username is already taken." instead of "failed to create user".

- [ ] **Step 5: Final state check**

```bash
cd /home/dodol/Repositories/mate/mate-things
git status --porcelain
```
Confirm everything expected is committed (no stray uncommitted changes) across both `backend/` and `frontend/` in the `mate-things` repo.
