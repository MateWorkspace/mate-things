# 07 — Profile (self-service, `/v1/profile`)

Precondition: use a disposable non-admin token so tests don't corrupt the
`admin` account — reuse `USER_TOKEN`/`rbac_test_user` from
`02-rbac-authorization.md`, or create a fresh one per `06-user.md` USER-01.

---

### PROF-01 — Get own profile (positive)
`GET /v1/profile` with `Authorization: Bearer $USER_TOKEN`.
**Expect:** `200`, own user record (no `password_hash` field — confirm
`json:"-"` tag actually excludes it from the response body).

### PROF-02 — Get own permissions (positive)
`GET /v1/profile/permissions`.
**Expect:** `200`, matches the token's role permissions.

### PROF-03 — Patch own name/bio (positive)
`PATCH /v1/profile` with `{"name":"New Name","bio":"New bio."}`.
**Expect:** `200`/`204`; re-`GET /v1/profile` reflects the change.

### PROF-04 — Patch own username to empty string (edge, ⚠ known gap)
`PATCH /v1/profile` with `{"username":""}`.
**Expect:** confirmed in code audit — `ProfilePatch` handler passes
`Name`/`Bio`/`Username` straight through with **no validation calls at
all**, since they're optional `*string` fields the handler doesn't
individually check. An explicit empty-string username may get persisted
as-is. Verify actual behavior with a `GET /v1/profile` afterward — if the
username is now empty, this is a real bug (breaks future login-by-username
for this account) worth flagging prominently, distinct from "just
unvalidated". Use a throwaway test account here, not `admin` or anything
load-bearing for later scenarios.

### PROF-05 — Patch own username to another existing user's username (negative)
**Expect:** `409`.

### PROF-06 — Attempt to patch `role_id`/other privileged fields via profile patch (negative/security check)
Send `PATCH /v1/profile` with an extra `"role_id":"<super-role-id>"` field
not in `ProfilePatchRequest`'s schema.
**Expect:** field should be silently ignored by JSON binding (not part of
the struct) — confirm the user's role does NOT change after this call.
This verifies there's no privilege-escalation path via extra JSON fields.

### PROF-07 — Change own password, correct current password (positive)
`PATCH /v1/profile/password` with
`{"current_password":"Test1234!","new_password":"NewPass1234!"}`.
**Expect:** `200`/`204`. Then log in with the new password to confirm.

### PROF-08 — Change own password, wrong current password (negative)
`{"current_password":"wrong","new_password":"whatever"}`.
**Expect:** `401`.

### PROF-09 — Change own password, weak new password (edge, ⚠ known gap)
`{"current_password":"NewPass1234!","new_password":"1"}` (continuing from PROF-07).
**Expect:** `200`/`204` — no strength/length validation, same class of gap
as USER-07/USER-15.

### PROF-10 — Change own password, new password >72 bytes (bug, same class as USER-08)
**Expect:** `500` instead of a clean `400`.
