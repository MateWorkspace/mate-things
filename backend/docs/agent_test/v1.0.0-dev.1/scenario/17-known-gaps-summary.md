# 17 — Known Gaps Summary — Final Results

This is the post-execution result of the full test pass: every scenario in
`00`–`16` was run against the real external Postgres/Redis/MinIO/HiveMQ Cloud
dev environment described in `.env`, using `docker`, `curl`, and
`mosquitto_pub`/`mosquitto_sub`. Every item below was **fixed and
re-verified live** unless marked otherwise.

## 0 — Critical bug found only by running the suite (not in the original static audit)

| # | Area | Bug | Status |
|---|---|---|---|
| 0 | Auth / Redis cache | `internal/infrastructure/cache/user/redis.go` cached the `domainmodels.User` struct directly via `encoding/json`. `User.PasswordHash` is tagged `json:"-"` (correctly, to keep it out of HTTP responses) — but the same tag also stripped it from the **cached** blob. Result: the *first* login for any user (cache miss) succeeded, but every login after that (cache hit) failed with `500 {"code":"unknown","message":"failed to compare password"}`, because the cached copy's password hash was an empty string. This is as severe as it gets — it silently broke login for the entire system after first use. | **Fixed** — added `internal/infrastructure/cache/user/dto.go` (`cachedUser`), a cache-only DTO with an explicit `password_hash` field, and routed `GetById`/`SetById`/`GetByUsername`/`SetByUsername` through it. Poisoned cache entries were manually flushed from Redis; live-verified with back-to-back logins (cache-miss and cache-hit paths both `200`). |

## 1 — Everything from the original static-audit gap list

| # | Area | Gap | Status |
|---|---|---|---|
| 1 | Auth | Login 404 (no user) vs 401 (wrong password) — username enumeration | Not fixed — low severity, arguably intentional (matches many APIs); left as documented behavior. |
| 2 | RBAC | No distinct 403; missing-auth and missing-permission both 401 | Not fixed — architectural, would be an API-contract change; left as documented behavior. |
| 3 | User | `username` had zero format/length validation | **Fixed** — `RequiredUsername`/`OptionalUsername` (3–64 chars) added to `internal/presentation/http/utils/validation.go`, wired into user create/patch and profile patch. Live-verified. |
| 4 | User/Profile | `password`/`new_password` had zero strength/length validation | **Fixed** — `RequiredPassword` (min 8 chars) wired into user create, admin password reset, and profile password change. Live-verified (`400` where it used to be `201`/`204`). |
| 5 | User/Profile | Password >72 bytes → bcrypt error → **HTTP 500** | **Fixed** — `RequiredPassword` also rejects >72 bytes with a clean `400` before it ever reaches bcrypt. Live-verified on all three call sites (user create, admin reset, profile change). |
| 6 | Profile | `ProfilePatch` didn't validate `username` at all | **Fixed** — `OptionalUsername` wired into `ProfilePatch`. Live-verified: empty-string username patch now `400`, username left unchanged. |
| 7 | Permission/Role/NodeClass/User/Firmware | Unique-name constraints had no partial index excluding soft-deleted rows — a deleted name could never be reused | **Fixed** — new migration `20260727080000_soft_delete_partial_unique_indexes` converts `permissions.name`, `roles.name`, `node_classes.name`, `actions.name`, `users.username`, `firmwares.name` from plain `UNIQUE` constraints to partial unique indexes (`WHERE deleted_at IS NULL`). `payload_schemas(name, version)` and `nodes.device_id` are deliberately **excluded** — both are referenced by other tables' foreign keys (`actions`/`telemetry_records` → `payload_schemas`; `telemetry_records` → `nodes.device_id`), and Postgres foreign keys can only target a real `UNIQUE` constraint, never a partial index. Live-verified: recreating a permission under a soft-deleted name now returns `201` instead of `409`. |
| 8 | Role/NodeClass | Deleting a role/node_class still referenced elsewhere succeeds silently (soft delete bypasses FK) | Not fixed — would need a referential-integrity check added to the delete usecases (block or cascade); scoped as a follow-up, not attempted this pass. |
| 9 | Payload Schema | `version` int32 had no range check | **Fixed** — `RequiredPositiveInt32`/`OptionalPositiveInt32` wired into payload-schema create/patch. Live-verified (negative version now `400`). |
| 10 | Payload Schema | `definition` shape not validated against `PayloadSchemaDefinition` at creation | Not fixed — would need to unmarshal+walk the definition at write time; scoped as a follow-up. |
| 11 | Payload Schema | No `valid_from`/`valid_to` ordering check | **Fixed** — `ValidateTimeWindow` wired into create/patch (only when both bounds are present in the same request). Live-verified (`valid_to` before `valid_from` now `400`). |
| 12 | Node | `device_id` on `NodePatchRequest` had zero validation | Not fixed as originally scoped (remove/lock the field) — left patchable, since removing it would be a route-contract change beyond a bug-fix pass. Flagged for a product decision. |
| 13 | Node | No node-class compatibility check between a node and its assigned firmware | Not fixed — same class of issue as #20 but for firmware assignment rather than action dispatch; not attempted this pass. |
| 14 | Node | Re-registering a device_id after its node was soft-deleted was permanently blocked | **Fixed** — `queryUpsertRegistration` in `internal/infrastructure/repository/node/postgres_query.go` no longer restricts its `ON CONFLICT ... DO UPDATE` to `deleted_at IS NULL` rows, and now explicitly clears `deleted_at`/`deleted_by` on conflict, reviving the row. Live-verified: delete a node, re-publish MQTT registration for the same `device_id`, node comes back `200` (was `404`). |
| 15 | OTA dispatch | `firmware_url` had zero URL-format validation | **Fixed** — new `RequiredURL` validator (requires scheme+host) wired into both OTA dispatch handlers. Live-verified (`400` on a non-URL string). |
| 16 | Firmware | No content-type/extension/magic-byte validation on upload | Not fixed — low priority, arguably fine (firmware blobs are opaque anyway); left as-is. |
| 17 | Firmware | Does renaming a firmware move its MinIO object? | Confirmed via live test: **no** — `binary_path` (the actual storage key) stays fixed at the name given at creation time; only the DB `name` column changes on rename. Not a bug (storage ops correctly use `binary_path`, not `name`, everywhere), just a documented quirk. Not fixed/changed. |
| 18 | Firmware | Does deleting a firmware clean up its MinIO object? | Not re-confirmed this pass (ran out of scope/time) — still an open question, flagged for follow-up. |
| 19 | Action | `payload_schema_version` int32 had no range check on create/patch | **Fixed** — same `RequiredPositiveInt32`/`OptionalPositiveInt32` wired into action create/patch. Live-verified. |
| 20 | Action Dispatch | Dispatching an action against a node of the wrong node class was never rejected | **Fixed** — added an explicit `node.NodeClassId != action.NodeClassId` check in `internal/application/action/execution/usecase.go` `Dispatch`, right after the connectivity check, producing the same `ActionStatusUnexecuted` pattern as the existing node-not-found/disconnected branches. Live-verified end-to-end: registered a node under a different class, dispatched an action bound to `base_node`, got back `action_status: "UNEXECUTED"`, `action_message: "node class does not match action's node class"`. |
| 21 | Action Dispatch | `executed_at` had no bounds check | Not fixed — low priority (a far-future/past timestamp is harmless metadata, not attempted this pass). |
| 22 | Preferences | `preferences` JSON body has zero schema validation per resource type | Not fixed — would need a per-resource schema, out of scope for this pass. |
| 23 | Pagination | Negative/zero `page`/`limit` silently clamped instead of rejected | Not changed — confirmed via live test this is deliberate, working, and reasonable UX (clamps to defaults rather than erroring); not treated as a bug. |
| 24 | Pagination | No upper bound on `limit` | **Fixed** — added `maxLimit = 100` clamp in `internal/presentation/http/utils/pagination.go`. Live-verified: `?limit=1000000` now returns `limit: 100` and exactly 100 (or fewer) rows, not unbounded. |
| 25 | Telemetry | Ingestion usecase fully built but never wired to any HTTP or MQTT entry point | Not fixed — this is a product/roadmap decision (is telemetry ingestion meant to ship yet?), not something to silently wire up without direction. Still flagged as the single largest functional gap in the system. |
| 26 | MQTT Log | `log` topic handler is a no-op stub | Not fixed — same reasoning as #25; likely an intentional placeholder pending the telemetry decision. |
| 27 | FK violations return 409 instead of 400/404 | Confirmed as a systemic, consistent pattern across every resource (`MapPgxError` maps `23503`/`23505` to `ErrTypeConflict` uniformly). Not changed — would be a broad, deliberate API-contract decision, not a one-off fix. |

## Summary

- **1 critical bug** (found only by live execution, not static audit): user
  login cache silently dropping password hashes. Fixed.
- **11 gaps fixed and live-verified**: bcrypt-crash-to-500, username
  validation, password validation (3 call sites), profile username
  validation, soft-delete name reuse (6 tables via new migration), node
  re-registration after delete, OTA URL validation, payload-schema version
  range + validity-window ordering, action version range, action-dispatch
  node-class mismatch, pagination max-limit cap.
- **~14 items intentionally left unfixed**, each with a stated reason
  (product decision needed, architectural/contract change, out of scope for
  a bug-fix pass, or confirmed as correct/intentional behavior rather than a
  bug). None are silent — every one is called out above with its reasoning
  so a follow-up pass can pick them up deliberately.
