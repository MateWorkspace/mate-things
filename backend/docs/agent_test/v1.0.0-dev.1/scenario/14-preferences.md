# 14 — Preferences (`PATCH /v1/preferences/:resource/:id`)

Precondition: `ACCESS_TOKEN`. One valid id per resource type (reuse
`PERM_ID`, `ROLE_ID`, `NCLASS_ID`, `FW_ID`, `ACTION_ID`, `USER_ID` from
earlier scenarios, whichever are still alive — recreate throwaway ones if
needed since several were deleted in their own scenario's cleanup steps).

Supported `:resource` values, confirmed via the handler's hardcoded
switch: `action`, `firmware`, `node`, `node_class`, `payload_schema`,
`permission`, `role`, `user`. (Confirmed **no SQL/table-name injection
risk** — `:resource` is purely a switch key dispatching to statically
typed usecase methods, never interpolated into a query.)

---

### PREF-01 — Patch preferences on a valid resource+id (positive)
```bash
curl -s -i -X PATCH http://127.0.0.1:18080/api/v1/preferences/permission/$PERM_ID \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d '{"preferences":{"color":"blue","pinned":true}}'
```
**Expect:** `200`/`204`. Confirm via `GET /v1/admin/permissions/$PERM_ID`
that `preferences` now reflects the patched value.

### PREF-02 — Unknown resource string (negative)
`PATCH /v1/preferences/widget/$PERM_ID`.
**Expect:** `400`, `{"message":"resource is not supported", ...}`.

### PREF-03 — Resource string with wrong case (negative/edge)
`PATCH /v1/preferences/Permission/$PERM_ID` (capitalized).
**Expect:** `400` — switch is case-sensitive, confirmed in code.

### PREF-04 — Valid resource, nonexistent id (negative)
`PATCH /v1/preferences/permission/00000000-0000-0000-0000-000000000000`.
**Expect:** `404`.

### PREF-05 — Valid resource, malformed id (negative)
`PATCH /v1/preferences/permission/not-a-uuid`.
**Expect:** `400`.

### PREF-06 — Malformed JSON body (negative)
`-d '{not json'`.
**Expect:** `400`.

### PREF-07 — Empty body / missing `preferences` key (negative)
`-d '{}'`.
**Expect:** `400`, `{"message":"preferences is required", ...}` (or
similar — `RequiredRawJSON` treats a missing/null field as absent).

### PREF-08 — Arbitrary/unstructured JSON accepted regardless of resource semantics (edge, ⚠ known gap)
`PATCH /v1/preferences/role/$ROLE_ID` with
`{"preferences":{"totally_unrelated_junk":[1,2,3],"nested":{"a":{"b":{"c":"deep"}}}}}`.
**Expect:** `200`/`204` — confirmed **zero schema validation on
`preferences` content for any resource type**, just JSON well-formedness.
Flag as a gap (each resource could reasonably want a constrained
preferences shape) but not fixing now.

### PREF-09 — Permission enforcement (regression check)
Confirm `preferences:set` is required — `user`-role token should get `401`
on any of the above.
