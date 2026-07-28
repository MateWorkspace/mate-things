# 03 — Permission Management (`/v1/admin/permissions`)

Precondition: `ACCESS_TOKEN` (admin/super) from `01-auth.md`.
Header used throughout: `-H "Authorization: Bearer $ACCESS_TOKEN"`.

---

### PERM-01 — Create permission (positive)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/admin/permissions \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d '{"name":"test:resource","description":"Test permission for feature testing."}'
```
**Expect:** `201`, body has `id`, `name`, `description`. Save `id` as `PERM_ID`.

### PERM-02 — Create permission with empty name (negative)
`{"name":"","description":"x"}`.
**Expect:** `400`.

### PERM-03 — Create permission with duplicate name (negative)
Repeat PERM-01's exact body.
**Expect:** `409` (unique constraint on `permissions.name`).

### PERM-04 — Create permission with no description field at all (edge, ⚠ known gap)
`{"name":"test:resource2"}` (omit `description` entirely).
**Expect:** `201` — `description` is fully optional with zero validation
(no length cap, no format). Document as expected-but-unvalidated; not a bug
to fix unless product wants a max length.

### PERM-05 — Get list (positive)
```bash
curl -s http://127.0.0.1:18080/api/v1/admin/permissions -H "Authorization: Bearer $ACCESS_TOKEN"
```
**Expect:** `200`, paginated list including the 46 seeded + newly created.

### PERM-06 — Get by id (positive)
`GET /v1/admin/permissions/$PERM_ID`.
**Expect:** `200`.

### PERM-07 — Get by id, nonexistent UUID (negative)
`GET /v1/admin/permissions/00000000-0000-0000-0000-000000000000`.
**Expect:** `404`.

### PERM-08 — Get by id, malformed UUID (negative/edge)
`GET /v1/admin/permissions/not-a-uuid`.
**Expect:** `400` (`RequiredUUID`/path binding fails to parse).

### PERM-09 — Get by name (positive)
`GET /v1/admin/permissions/by-name/test:resource`.
**Expect:** `200`. Note: name contains a colon — confirm URL-encoding
(`test%3Aresource`) isn't required/doesn't break routing either way.

### PERM-10 — Get by name, nonexistent (negative)
`GET /v1/admin/permissions/by-name/does-not-exist`.
**Expect:** `404`.

### PERM-11 — Patch description (positive)
```bash
curl -s -i -X PATCH http://127.0.0.1:18080/api/v1/admin/permissions/$PERM_ID \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d '{"description":"Updated description."}'
```
**Expect:** `200`/`204`.

### PERM-12 — Patch name to an already-taken name (negative)
Patch `PERM_ID`'s name to `"profile:get"` (an existing seeded permission).
**Expect:** `409`.

### PERM-13 — Delete (positive)
`DELETE /v1/admin/permissions/$PERM_ID`.
**Expect:** `200`/`204`.

### PERM-14 — Get after delete (negative, confirms soft-delete semantics)
`GET /v1/admin/permissions/$PERM_ID` again.
**Expect:** `404` (soft-deleted rows excluded from reads).

### PERM-15 — Re-create with the same name as a soft-deleted permission (edge, ⚠ known gap)
Re-run PERM-01's create body (`name":"test:resource"`) after PERM-13/14.
**Expect:** `409` — the `permissions.name` unique constraint has **no
partial index excluding soft-deleted rows** (confirmed in migration SQL),
so a soft-deleted permission's name permanently blocks reuse. This is a
real design gap worth flagging for a future fix (partial unique index
`WHERE deleted_at IS NULL`), not fixing now.
