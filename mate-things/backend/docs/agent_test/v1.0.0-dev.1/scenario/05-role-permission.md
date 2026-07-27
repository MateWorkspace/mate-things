# 05 — Role-Permission Assignment (`/v1/admin/role-permissions`, `/v1/admin/roles/:role_id/permissions/:permission_id`)

Precondition: `ACCESS_TOKEN` (admin/super). Reuse `PERM_ID` from
`03-permission.md` if still present, or create a fresh throwaway
permission. Reuse a throwaway role (e.g. re-create `test_role` from
`04-role.md`, since it was deleted at the end of that scenario).

---

### RP-01 — Assign a permission to a role (positive)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/admin/roles/$ROLE_ID/permissions/$PERM_ID \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```
**Expect:** `201`, body has the `role_permission` row id.

### RP-02 — Assign the same pair again (negative)
Repeat RP-01.
**Expect:** `409` (unique `(role_id, permission_id)`).

### RP-03 — Assign with a nonexistent permission id (negative)
`POST /v1/admin/roles/$ROLE_ID/permissions/00000000-0000-0000-0000-000000000000`.
**Expect:** `409` or `400`/`404` depending on FK-violation mapping — verify
exact status returned (`role_permission.permission_id` has an FK to
`permissions(id)`; a nonexistent id should trigger a `23503` FK-violation,
mapped to `ErrTypeConflict` per `MapPgxError` → expect `409`, **not** 404 —
this is a real UX rough edge worth noting: the client gets "conflict" for
what's really "the referenced permission doesn't exist").

### RP-04 — Assign with a nonexistent role id (negative)
Same as RP-03 but with a bogus `:role_id`. Same expected `409` reasoning.

### RP-05 — Get role's permission list (positive)
`GET /v1/admin/roles/$ROLE_ID/permissions`.
**Expect:** `200`, array containing the permission from RP-01.

### RP-06 — List all role_permission rows (positive)
`GET /v1/admin/role-permissions`.
**Expect:** `200`, paginated.

### RP-07 — Get by pair (positive)
`GET /v1/admin/role-permissions/by-pair?role_id=$ROLE_ID&permission_id=$PERM_ID`.
**Expect:** `200`.

### RP-08 — Get by pair, nonexistent pair (negative)
Same query with a permission id never assigned to that role.
**Expect:** `404`.

### RP-09 — Get by id (positive)
`GET /v1/admin/role-permissions/:id` using the id captured from RP-01.
**Expect:** `200`.

### RP-10 — Remove assignment by pair (positive)
```bash
curl -s -i -X DELETE http://127.0.0.1:18080/api/v1/admin/roles/$ROLE_ID/permissions/$PERM_ID \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```
**Expect:** `200`/`204`.

### RP-11 — Remove an assignment that doesn't exist (negative/edge)
Repeat RP-10 immediately.
**Expect:** verify actual behavior — `DeleteByRoleIdAndPermissionId` may be
a no-op-success (`204` even if zero rows affected) rather than `404`,
since it's a delete-by-filter not delete-by-id. Document whichever it
actually returns; not necessarily a bug either way.

### RP-12 — Cross-check via `05-role.md`'s ROLE-11 (regression)
After RP-10, re-`GET /v1/admin/roles/$ROLE_ID/permissions` — should be
empty again.
