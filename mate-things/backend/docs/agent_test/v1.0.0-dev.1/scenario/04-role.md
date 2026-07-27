# 04 — Role Management (`/v1/admin/roles`)

Precondition: `ACCESS_TOKEN` (admin/super).

---

### ROLE-01 — Create role (positive)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/admin/roles \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d '{"name":"test_role","description":"Role for feature testing."}'
```
**Expect:** `201`. Save `id` as `ROLE_ID`. `is_default` should be `false`
(create endpoint never sets it true — confirmed in code, only
`PATCH .../default` can).

### ROLE-02 — Create with duplicate name (negative)
Repeat ROLE-01.
**Expect:** `409`.

### ROLE-03 — Create with empty name (negative)
`{"name":""}`.
**Expect:** `400`.

### ROLE-04 — Get list (positive)
`GET /v1/admin/roles`.
**Expect:** `200`, includes seeded `super`/`admin`/`user` + `test_role`.

### ROLE-05 — Get default role (positive)
`GET /v1/admin/roles/default`.
**Expect:** `200`, returns the seeded `user` role (`is_default: true`).

### ROLE-06 — Get by name / by id (positive)
`GET /v1/admin/roles/by-name/test_role`, `GET /v1/admin/roles/$ROLE_ID`.
**Expect:** both `200`.

### ROLE-07 — Get nonexistent role by id (negative)
**Expect:** `404`.

### ROLE-08 — Set `test_role` as the new default (positive, exclusivity check)
```bash
curl -s -i -X PATCH http://127.0.0.1:18080/api/v1/admin/roles/$ROLE_ID/default \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```
**Expect:** `200`/`204`. Then `GET /v1/admin/roles/default` → should now
return `test_role`, and `GET /v1/admin/roles/by-name/user` → `is_default`
should now be `false`. This confirms the DB-level exclusivity (partial
unique index `uq_roles_is_default_true`) and the transaction that unsets
the previous default.

### ROLE-09 — Restore `user` as default (cleanup, required for later scenarios)
`PATCH /v1/admin/roles/{user_role_id}/default`.
**Expect:** `200`/`204`. Re-verify `GET /v1/admin/roles/default` returns `user`.
⚠ **Do this before moving on** — several other scenarios (new-user
creation, RBAC low-privilege tests) implicitly assume `user` is the
default role for any test-created account that omits an explicit role.

### ROLE-10 — Patch description (positive)
`PATCH /v1/admin/roles/$ROLE_ID` with `{"description":"Updated."}`.
**Expect:** `200`/`204`.

### ROLE-11 — `GET .../:id/permissions` on a role with zero assignments (positive/edge)
`GET /v1/admin/roles/$ROLE_ID/permissions` (fresh `test_role`, nothing
assigned yet).
**Expect:** `200`, empty array (not 404).

### ROLE-12 — Delete role (positive)
`DELETE /v1/admin/roles/$ROLE_ID`.
**Expect:** `200`/`204`.

### ROLE-13 — Delete a role that still has users assigned (negative/edge — verify FK behavior)
Attempt to delete the seeded `user` role (has real seeded/test users
referencing it via `users.role_id NOT NULL REFERENCES roles(id)`, no
`ON DELETE` clause specified in migration → defaults to `NO ACTION`).
```bash
curl -s -i -X DELETE http://127.0.0.1:18080/api/v1/admin/roles/{user_role_id} \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```
**Expect:** since `roles` uses **soft delete** (`DeleteById` sets
`deleted_at`, doesn't actually `DROP` the row), this will likely succeed
at the DB level (soft delete doesn't touch the FK) even though it leaves
users pointing at a "deleted" role. **Do NOT actually run this against the
real `user` role** — it would break every other scenario relying on it.
Instead, create a disposable throwaway role, assign a throwaway user to
it, then soft-delete the role and confirm: (a) the delete call succeeds,
(b) `GET` on the deleted role returns 404, (c) the orphaned user's
`GET /v1/admin/users/:id` still returns 200 with a `role_id` pointing at a
now-invisible role. Flag this as a data-integrity gap worth a future fix
(e.g. block role deletion while users reference it), not fixing now.

### ROLE-14 — Re-create with a soft-deleted role's name (edge, ⚠ known gap)
Same class of gap as PERM-15: `roles.name` unique constraint has no
`WHERE deleted_at IS NULL` partial index, so re-using a soft-deleted role's
exact name will `409`. Confirm with the disposable role from ROLE-13's setup.
