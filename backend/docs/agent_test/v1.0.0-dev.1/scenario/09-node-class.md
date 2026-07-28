# 09 — Node Class (`/v1/node-classes`)

Precondition: `ACCESS_TOKEN`. Seeded `base_node` node class already exists.

---

### NCLASS-01 — Create (positive)
`POST /v1/node-classes` `{"name":"test_class","description":"Test class."}`.
**Expect:** `201`. Save `id` as `NCLASS_ID`.

### NCLASS-02 — Duplicate name (negative)
**Expect:** `409`.

### NCLASS-03 — Empty name (negative)
**Expect:** `400`.

### NCLASS-04 — List / get by id / get by name (positive)
**Expect:** `200` each.

### NCLASS-05 — Get nonexistent by id/name (negative)
**Expect:** `404` each.

### NCLASS-06 — Patch description (positive)
**Expect:** `200`/`204`.

### NCLASS-07 — Get firmwares for a node class with none uploaded yet (edge)
`GET /v1/node-classes/$NCLASS_ID/firmwares`.
**Expect:** `200`, empty array (not 404).

### NCLASS-08 — Delete a node class referenced by nodes/actions (negative/edge)
Attempt to delete `base_node` (referenced by the seeded `restart` action
and, once `10-node.md` runs, by real node rows via `nodes.node_class_id`
FK, plus `actions.node_class_id` FK — neither has `ON DELETE CASCADE`).
**Do not actually run this against the real `base_node`** — it would break
`10-node.md`/`12-action.md`. Instead create a disposable `test_class`
(NCLASS-01), attach a throwaway action or node to it, then attempt delete.
**Expect:** since `NodeClass.DeleteById` is a soft delete (sets
`deleted_at`, doesn't `DROP`), the delete call likely succeeds even with
dependents still referencing the id (soft delete bypasses the FK
entirely) — leaving actions/nodes pointing at a now-invisible node class.
Flag as the same class of data-integrity gap as `04-role.md` ROLE-13
(deleting a role that still has users) — not fixing now.

### NCLASS-09 — Delete then re-create with the same name (edge, ⚠ known gap)
Same "no partial unique index excluding soft-deleted rows" gap as
permissions/roles — confirm `409` on reuse of a soft-deleted name.
