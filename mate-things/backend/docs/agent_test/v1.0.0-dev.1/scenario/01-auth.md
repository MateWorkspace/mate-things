# 01 — Authentication (`/v1/auth/login`, `/v1/auth/refresh`)

Precondition: `00-setup.md` done. Uses the seeded `admin`/`ChangeMe123!` user.

Base: `http://127.0.0.1:18080/api/v1/auth`

---

### AUTH-01 — Login with valid credentials (positive)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"ChangeMe123!"}'
```
**Expect:** `200`, body contains `user`, `role` (name `super`), `permissions`
(array of 46), `access_token`, `refresh_token`.

### AUTH-02 — Login with wrong password (negative)
Same as above with `"password":"wrong"`.
**Expect:** `401`, body `{"code":"unauthorized", ...}`.

### AUTH-03 — Login with nonexistent username (negative)
`{"username":"nobody","password":"x"}`.
**Expect:** `404`, body `{"code":"not_found", ...}`. ⚠ Note: this leaks
whether a username exists (404 vs 401 distinguishes "no such user" from
"wrong password") — worth flagging as a minor information-disclosure gap,
not fixing now.

### AUTH-04 — Login with empty username (negative/edge)
`{"username":"","password":"ChangeMe123!"}`.
**Expect:** `400` (handler's `RequiredString` on username should reject).

### AUTH-05 — Login with empty password (negative/edge)
`{"username":"admin","password":""}`.
**Expect:** `400`.

### AUTH-06 — Login with malformed JSON body (negative)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/auth/login \
  -H "Content-Type: application/json" -d '{not json'
```
**Expect:** `400` (echo bind failure).

### AUTH-07 — Refresh with valid refresh token (positive)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"
```
**Expect:** `200`, new `access_token`/`refresh_token` pair, same `user`/`role`/`permissions` shape as login.

### AUTH-08 — Refresh with an access token in place of a refresh token (negative)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$ACCESS_TOKEN\"}"
```
**Expect:** `401` — access and refresh tokens are signed with different
secrets (`BE_TOKEN_ACCESS_SECRET` vs `BE_TOKEN_REFRESH_SECRET`), so an
access token should fail refresh-token signature validation.

### AUTH-09 — Refresh with garbage/malformed token (negative)
`{"refresh_token":"not-a-jwt"}`.
**Expect:** `401`.

### AUTH-10 — Access token expiry (positive/timing)
1. Log in fresh, capture `access_token`.
2. Immediately call any authenticated endpoint (e.g. `GET /v1/profile`) — expect `200`.
3. Wait **>2 minutes** (`BE_TOKEN_ACCESS_DURATION=2m`).
4. Call `GET /v1/profile` again with the same (now-expired) `access_token`.
**Expect:** step 2 → `200`; step 4 → `401`, `{"code":"unauthorized", ...}`
(expired and malformed tokens both collapse to the same response shape —
confirmed in code, not a bug to fix, just document the behavior).

### AUTH-11 — Refresh token still works after access token expires (positive)
Immediately after AUTH-10 step 4, call `/v1/auth/refresh` with the
still-valid `refresh_token` (24h duration).
**Expect:** `200`, fresh tokens issued.
