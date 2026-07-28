# 06 — User Management (`/v1/admin/users`)

Precondition: `ACCESS_TOKEN` (admin/super). Need `user` role's id
(`GET /v1/admin/roles/by-name/user`) as `ROLE_USER_ID`.

This is the resource with the most known validation gaps per the codebase
audit — the user explicitly flagged username/password/role validation as
likely missing. Every gap below is confirmed in code, marked ⚠, and left
**unfixed for this pass** (test plan only, per instructions).

---

### USER-01 — Create user (positive)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/admin/users \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d "{\"role_id\":\"$ROLE_USER_ID\",\"name\":\"Full Test User\",\"username\":\"full_test_user\",\"password\":\"Test1234!\"}"
```
**Expect:** `201`. Save `id` as `USER_ID`.

### USER-02 — Create with nonexistent role_id (negative)
Same body with a random UUID for `role_id`.
**Expect:** FK violation → `409` per `MapPgxError` (same "conflict instead
of not-found" rough edge as RP-03/RP-04 — document, don't fix).

### USER-03 — Create with malformed (non-UUID) role_id (negative)
`role_id: "not-a-uuid"`.
**Expect:** `400` (`RequiredUUID` parse failure — this one IS validated).

### USER-04 — Create with duplicate username (negative)
Repeat USER-01's exact body.
**Expect:** `409` (DB unique constraint on `username` — enforced only at
DB level, no pre-check in the handler, but the end result is still correct).

### USER-05 — Create with a 1-character username (edge, ⚠ known gap)
`username: "a"` (all else valid, different name to avoid USER-04 conflict).
**Expect:** `201` — confirmed **zero format/length validation** on
username anywhere in the handler; any non-empty string is accepted
verbatim. Flag for future fix (e.g. min length, allowed charset).

### USER-06 — Create with a username containing whitespace/unicode/punctuation (edge, ⚠ known gap)
`username: "  weird 🐧 user!! "` (note: not even trimmed — confirm the
stored value keeps the raw whitespace, since `RequiredString` trims only
for the emptiness check, need to verify whether the trimmed or raw value
is what's actually persisted by reading the handler/usecase code path, or
just check `GET` on the created user afterward and see which one comes
back).
**Expect:** `201`. Document actual stored value.

### USER-07 — Create with a 1-character password (edge, ⚠ known gap)
`password: "x"`, unique username.
**Expect:** `201` — **zero password strength/length validation**. Flag for
future fix (minimum length/complexity policy).

### USER-08 — Create with a password >72 bytes (negative/edge — real bug, not just a gap)
```bash
LONGPASS=$(python3 -c "print('a'*100)")
curl -s -i -X POST http://127.0.0.1:18080/api/v1/admin/users \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d "{\"role_id\":\"$ROLE_USER_ID\",\"name\":\"Long Pass\",\"username\":\"longpass_user\",\"password\":\"$LONGPASS\"}"
```
**Expect:** bcrypt's hard 72-byte input limit means `password.Hash` errors
out, wrapped as `ErrTypeFailure` → **HTTP 500**, not a clean 400. This is a
genuine bug (should be validated and rejected with 400 before ever
reaching bcrypt) — flag prominently as a fix candidate, distinct from the
"just unvalidated" gaps above since this one produces a server error, not
merely permissive behavior.

### USER-09 — Create with empty `name` (negative/edge)
`name: ""`.
**Expect:** verify — handler calls `RequiredString` on `Name` per the
audit, so expect `400`. Confirm actual behavior.

### USER-10 — Get list / get by id / get by username (positive)
`GET /v1/admin/users`, `GET /v1/admin/users/$USER_ID`,
`GET /v1/admin/users/by-username/full_test_user`.
**Expect:** all `200`.

### USER-11 — Get by nonexistent id / username (negative)
**Expect:** `404` both.

### USER-12 — Get user's effective permissions (positive)
`GET /v1/admin/users/$USER_ID/permissions`.
**Expect:** `200`, matches the `user` role's 7 seeded permissions.

### USER-13 — Patch username to another user's existing username (negative)
Patch `USER_ID`'s `username` to `"admin"`.
**Expect:** `409`.

### USER-14 — Patch username to empty string via admin patch (edge, ⚠ known gap)
`PATCH /v1/admin/users/$USER_ID` with `{"username":""}` — note this is a
`*string` optional field; sending an explicit empty string (not omitting
the key) may or may not go through `RequiredString`-style validation.
**Expect:** verify actual behavior; document whichever it is (400 if
validated, 200/204-with-empty-username-persisted if not — the latter would
be a genuine bug worth flagging since an empty username breaks login by
username lookup).

### USER-15 — Admin-set password reset, weak password (edge, ⚠ known gap)
`PATCH /v1/admin/users/$USER_ID/password` with `{"password":"1"}`.
**Expect:** `200`/`204` — same zero-strength-validation gap as USER-07.

### USER-16 — Admin-set password reset, >72 bytes (same bug class as USER-08)
**Expect:** `500`.

### USER-17 — Delete user (positive)
`DELETE /v1/admin/users/$USER_ID`.
**Expect:** `200`/`204`.

### USER-18 — Get after delete (negative)
**Expect:** `404`.

### USER-19 — Login as a just-deleted user (negative/edge, security check)
`POST /v1/auth/login` with the deleted user's original username/password.
**Expect:** must be `404`/`401` (soft-deleted users must not be able to
log in) — `ReadByUsername` should filter `deleted_at IS NULL`. This is
important to actually verify, not assume.
