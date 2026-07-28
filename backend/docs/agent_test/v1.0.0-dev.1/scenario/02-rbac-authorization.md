# 02 — RBAC / Authorization Middleware

Precondition: `01-auth.md` AUTH-01 done (`ACCESS_TOKEN` for `admin`/`super`).

Also needed here: a low-privilege second user. Create via admin token
(uses the endpoint under test in `06-user.md` — fine to do here since this
scenario needs it as a precondition):

```bash
ROLE_USER_ID=$(curl -s http://127.0.0.1:18080/api/v1/admin/roles/by-name/user \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq -r .id)

curl -s -X POST http://127.0.0.1:18080/api/v1/admin/users \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d "{\"role_id\":\"$ROLE_USER_ID\",\"name\":\"Test User\",\"username\":\"rbac_test_user\",\"password\":\"Test1234!\"}"

USER_TOKEN=$(curl -s -X POST http://127.0.0.1:18080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"rbac_test_user","password":"Test1234!"}' | jq -r .access_token)
```

---

### RBAC-01 — No Authorization header (negative)
```bash
curl -s -i http://127.0.0.1:18080/api/v1/profile
```
**Expect:** `401`, `{"code":"unauthorized","message":"authorization is required"}`.

### RBAC-02 — Malformed Authorization header (negative)
```bash
curl -s -i http://127.0.0.1:18080/api/v1/profile -H "Authorization: garbage"
```
**Expect:** `401` (token parse failure — note code treats a bare non-JWT
string as the token itself after stripping an optional `Bearer ` prefix).

### RBAC-03 — Valid token, missing permission → 401 not 403 (behavior note)
Using `USER_TOKEN` (role `user`, which does not have `permission:get`):
```bash
curl -s -i http://127.0.0.1:18080/api/v1/admin/permissions \
  -H "Authorization: Bearer $USER_TOKEN"
```
**Expect:** `401`, `{"code":"unauthorized","message":"permission is denied"}`.
⚠ Document only, not a bug: this codebase has no distinct 403 Forbidden
anywhere — missing-auth and missing-permission both return 401. Confirm
this is intentional with the team before ever "fixing" it, since fixing it
would be an API-contract change for every client.

### RBAC-04 — Valid token, has permission (positive)
```bash
curl -s -i http://127.0.0.1:18080/api/v1/profile \
  -H "Authorization: Bearer $USER_TOKEN"
```
**Expect:** `200` (`user` role has `profile:get`).

### RBAC-05 — Stale permissions in an already-issued token (edge case)
1. Log in as `rbac_test_user` fresh → `USER_TOKEN_2`.
2. As admin, revoke `profile:get` from the `user` role (`DELETE /v1/admin/roles/:role_id/permissions/:permission_id` — see `05-role-permission.md`).
3. Immediately call `GET /v1/profile` with `USER_TOKEN_2` (still unexpired).
**Expect:** `200` — the JWT embeds the permission snapshot at login time,
so revocation does not take effect until the token expires/refreshes. This
is expected behavior given the code (`Auth` middleware never re-queries the
DB), not a bug — document it, and **remember to re-grant `profile:get`
back to the `user` role afterward** so later scenarios aren't affected.

### RBAC-06 — Permissions refresh on `/v1/auth/refresh` (positive, ties to RBAC-05)
After RBAC-05's revoke, call `/v1/auth/refresh` with `rbac_test_user`'s
still-valid refresh token, then retry `GET /v1/profile` with the *new*
access token.
**Expect:** `401` this time — refresh rebuilds the permission snapshot from
the current DB state, so the revoked permission is gone from the new token.
