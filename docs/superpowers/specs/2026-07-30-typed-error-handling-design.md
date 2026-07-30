# Typed, self-describing error handling — design

## Context

Every HTTP error response today is `{"error": "...", "message": "...", "details": "..."}`,
built centrally by `presentation/http/utils/error.go`'s `Error(c, err, message)`.
The `message` is a hand-typed string literal supplied by *each of 224 handler
call sites*, identical regardless of which specific error actually occurred —
so a duplicate-username conflict and a malformed-role-id error on the same
endpoint can produce the exact same sentence. Underneath, `message` also gets
attached generically at 108 `MapPgxError(...)` call sites in the repository
layer (e.g. `"failed to create user"`, `"failed to read user"`), so even the
`Details` field (meant to carry the real explanation) usually just repeats
something equally generic.

Two concrete symptoms drove this:
- Creating a user with a duplicate username surfaces "failed to create user"
  instead of anything saying the username is taken.
- A not-found lookup says "failed to read user" instead of "user not found".

Validation errors are the exception — `application/shared/validation.go`
already produces specific, correct messages (`"username is required"`,
length/charset messages, etc.) but they get silently discarded by the
handler's hardcoded override anyway.

## Goals

1. Every error response's `message` is genuinely specific to what happened,
   sourced from the error itself — never a call-site guess.
2. The presentation layer is the single place that turns a typed error into
   a user-presentable response, with no per-endpoint hand-written text.
3. 5xx-class errors never leak internal detail; the real cause stays in logs.
4. Conflict errors say *which* field/constraint was violated (a single
   `Create` call can violate more than one unique constraint on the same
   table, and the generic Postgres `23505` code alone doesn't disambiguate
   which one fired).

## Design

### 1. Domain models — 11 new conflict sentinels

`domain/models/error.go` gains one `ErrorType` per user-triggerable unique
constraint in the schema (verified against the migrations):

| New sentinel | Table.constraint |
|---|---|
| `ErrTypeUsernameExists` | `users.uq_users_username` |
| `ErrTypeRoleNameExists` | `roles.uq_roles_name` |
| `ErrTypePermissionNameExists` | `permissions.uq_permissions_name` |
| `ErrTypeActionNameExists` | `actions.uq_actions_name` |
| `ErrTypeNodeClassNameExists` | `node_classes.uq_node_classes_name` |
| `ErrTypeFirmwareNameExists` | `firmwares.uq_firmwares_name` |
| `ErrTypeNodeDeviceIdExists` | `nodes.nodes_device_id_key` |
| `ErrTypePayloadSchemaVersionExists` | `payload_schemas.uq_payload_schemas_name_version` |
| `ErrTypeRolePermissionExists` | `role_permission.uq_role_permission_role_id_permission_id` |
| `ErrTypeFirmwareConfigKeyExists` | `firmware_config_parameters.uq_firmware_config_parameters_firmware_id_key` |
| `ErrTypeNodeConfigKeyExists` | `node_config_values.uq_node_config_values_node_id_key` |

`roles.uq_roles_is_default_true` (the single-default-role invariant) stays
mapped to the existing generic `ErrTypeConflict` — it's an internal
invariant, not a user-facing "already exists" case.

Each is declared exactly like the existing sentinels:
```go
ErrTypeUsernameExists = errors.New("USERNAME_EXISTS")
```

### 2. Repository layer — constraint-aware conflict resolution + message cleanup

`infrastructure/repository/shared/error_pgx.go`'s `MapPgxError` gains an
optional constraint-name lookup, backward-compatible with its existing
signature via a variadic param so the ~96 call sites that can't produce a
disambiguation-needing conflict (reads, deletes, single-unique-constraint
tables) don't change at all:

```go
func MapPgxError(
	message string,
	err error,
	conflicts ...map[string]domainmodels.ErrorType,
) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domainmodels.NewError(message, domainmodels.ErrTypeNotFound, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			for _, m := range conflicts {
				if t, ok := m[pgErr.ConstraintName]; ok {
					return domainmodels.NewError(message, t, err)
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
```

At the 12 `Create`/`Update` call sites identified (one per table with a
user-triggerable unique constraint), pass the constraint map:

```go
// internal/infrastructure/repository/user/postgres.go
id, err := ...
if err != nil {
	return uuid.Nil, infrastructurerepositoryshared.MapPgxError(
		"failed to create user", err,
		map[string]domainmodels.ErrorType{
			"uq_users_username": domainmodels.ErrTypeUsernameExists,
		},
	)
}
```

Separately (mechanical text-only edits, no logic change): every
`MapPgxError` call site whose message is a vague "failed to X" gets
corrected to state what actually happened in plain, user-safe language —
e.g. `"failed to read user"` → `"user not found"`. This is the only reason
NotFound doesn't need per-resource sentinel types: each repo method already
operates on exactly one resource, so a corrected string is unambiguous
without a type split. Conflict needed types instead of just message-fixing
because a single `INSERT` can hit more than one unique index on the same
table, and the Postgres error code alone can't tell you which — only
`pgErr.ConstraintName`, read at the repo layer, can.

### 3. Application layer — audit pass

Usecases already largely propagate errors correctly (validation is already
well-messaged). This is a verification pass across all 34 usecase files:
confirm every returned error is either propagated as a well-formed
`domainmodels.Error` (now with a good repo-layer message per above) or,
where a usecase constructs its own error inline (e.g. "cannot delete the
default role"), that the message is specific and safe. Fix any found gaps
in place — no structural changes expected here beyond isolated fixes.

### 4. Presentation layer — one mapping table, no hand-typed messages

`presentation/http/utils/error.go`:

```go
type errorMapping struct {
	status  int
	title   string
	message string // non-empty: always used, overriding the domain message
}

var errorMappings = []struct {
	errType domainmodels.ErrorType
	mapping errorMapping
}{
	{domainmodels.ErrTypeNotFound, errorMapping{http.StatusNotFound, "Not Found", ""}},
	{domainmodels.ErrTypeConflict, errorMapping{http.StatusConflict, "Already Exists", ""}},
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
	// Unmatched/unwrapped error - safe generic 500, never leak internals.
	return c.JSON(http.StatusInternalServerError, presentationhttpresponse.ErrorResponse{
		Error:   "Internal Server Error",
		Message: "Something went wrong on our end. Please try again later.",
	})
}

func domainMessage(err error) string {
	var domainErr *domainmodels.Error
	if errors.As(err, &domainErr) && strings.TrimSpace(domainErr.Message) != "" {
		return domainErr.Message
	}
	return "Something went wrong on our end. Please try again later."
}
```

Rules:
- Types with a fixed `message` (the 11 conflict subtypes, plus the existing
  token/timeout/unimplemented/failure/unknown cases) always use that exact
  canned text — the domain error's own message is ignored for these (for
  5xx-class, this is a safety requirement so internals never leak).
- Types with an empty `message` (NotFound, base Conflict, BadArgs,
  Validation, BadState, Forbidden, Unauthorized) use the domain error's own
  `Message` directly — now specific and safe per the repo-layer fixes above.
- An error that reaches `Error()` without matching any mapping (shouldn't
  happen, but "catch every error possibility" means it must be handled) —
  falls through to a safe generic 500. This satisfies the "the HTTP
  presentation handler catches every error possibility" requirement even
  for errors nobody anticipated.

`BadRequest(c, message)` is unchanged — it's used for request-*shape*
failures (JSON parse errors) before a usecase ever runs, so there's no
domain error to derive from; it keeps its own message parameter.

**`Error`'s signature drops `message` entirely.** All 224 call sites across
the 8 handler files change from:
```go
return presentationhttputils.Error(c, err, "Unable to create the permission. Please check your input and try again.")
```
to:
```go
return presentationhttputils.Error(c, err)
```
This is a breaking signature change on purpose — it's the forcing function
that guarantees no call site keeps stale, overriding text after this change
lands (the compiler won't build until every one is updated).

### 5. Response shape cleanup

`presentation/http/response/common.go`'s `ErrorResponse` drops the `Details`
field — it existed to carry the "real" message that `message` itself
usually failed to provide; now that `message` is always specific, `Details`
is redundant.

Frontend: `frontend/src/app/(authenticated)/admin/users/_lib/actions.ts`'s
`failure()` (and the same pattern in the sibling `_lib/actions.ts` files for
access-control, payload-schemas, and `preferences-actions.ts`) currently do
`error.details?.trim() || error.message` — simplify to just `error.message`
in each, matching the resulting single-field API response.

## Scope (file-level)

- `backend/internal/domain/models/error.go` — 11 new sentinels.
- `backend/internal/infrastructure/repository/shared/error_pgx.go` —
  `MapPgxError` conflict-map param.
- `backend/internal/infrastructure/repository/{user,role,permission,action,node_class,firmware,node,payload_schema,role_permission,firmware_config_parameter,node_config_value}/postgres.go`
  — pass constraint maps at the relevant `Create`/`Update` call site(s);
  across these and the remaining files under `infrastructure/repository/*`
  (`action_log`, `node_log`, `telemetry_record`, which have no user-facing
  unique constraints), correct generic `MapPgxError` message strings (108
  call sites total, most just message-text edits).
- `backend/internal/application/**` (34 usecase files) — audit pass, fix
  isolated gaps found.
- `backend/internal/presentation/http/utils/error.go` — rewrite per above.
- `backend/internal/presentation/http/response/common.go` — drop `Details`.
- `backend/internal/presentation/http/handler/**` (8 files, 224 call
  sites) — update every `Error(c, err, ...)` call to `Error(c, err)`.
- `backend/docs/swagger/*` — regenerate (`swag init`) since `@Failure`
  response shape changes.
- `frontend/src/app/(authenticated)/admin/users/_lib/actions.ts`,
  `frontend/src/app/(authenticated)/admin/access-control/_lib/actions.ts`,
  `frontend/src/app/(authenticated)/admin/payload-schemas/_lib/actions.ts`,
  `frontend/src/components/preferences/preferences-actions.ts` — simplify
  `failure()` to drop the `.details` fallback.

## Verification

1. `cd backend && go build ./... && go vet ./...` — confirms every one of
   the 224 handler call sites compiles against the new `Error(c, err)`
   signature (the forcing function working as intended).
2. Rebuild and run the container; exercise real failure paths end-to-end:
   - Create a user with a duplicate username → 409,
     `{"error":"Already Exists","message":"This username is already taken."}`.
   - Read a nonexistent user by ID → 404, `{"error":"Not Found","message":"user not found"}`.
   - Trigger each of the other 10 conflict subtypes at least once (duplicate
     role name, permission name, action name, node class name, firmware
     name, node device_id, payload schema name+version, duplicate
     role-permission assignment, duplicate firmware/node config key).
   - A validation failure (e.g. missing required field) → 400 with the
     actual validation message (e.g. "username is required"), not a generic
     override.
   - Force a 500 → confirm the generic safe message, and confirm the real
     error still appears in server logs (unchanged logging behavior).
3. Regenerate swagger, spot-check `http://localhost:8080/api/docs` shows the
   new two-field `ErrorResponse` schema (no `details`).
4. Frontend: create a duplicate user via the admin/users UI, confirm the
   toast (from the earlier fix) now shows "This username is already taken."
   instead of "failed to create user".
