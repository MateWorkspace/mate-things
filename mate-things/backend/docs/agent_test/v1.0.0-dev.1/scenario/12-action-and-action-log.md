# 12 — Action, Action Dispatch, Action Log

Precondition: `ACCESS_TOKEN`. `NCLASS_ID` (a node class), seeded
`restart` payload schema (name=`restart`, version=1) or a fresh one from
`08-payload-schema.md`. A registered+connected node from `10-node.md`
(`NODE_ID`, `device_id=test-device-001`) for the dispatch tests.

---

## Action CRUD (`/v1/actions`)

### ACT-01 — Create action (positive)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/actions \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d "{\"node_class_id\":\"$NCLASS_ID\",\"name\":\"test_action\",\"payload_schema_name\":\"restart\",\"payload_schema_version\":1}"
```
**Expect:** `201`. Save `id` as `ACTION_ID`.

### ACT-02 — Create with duplicate name (negative)
**Expect:** `409` (`actions.name` unique).

### ACT-03 — Create referencing a nonexistent `(payload_schema_name, payload_schema_version)` pair (negative)
`payload_schema_name: "does_not_exist", payload_schema_version: 1`.
**Expect:** `409` (composite FK violation).

### ACT-04 — Create with `payload_schema_version: 0` (edge, ⚠ known gap)
**Expect:** per audit, zero range validation on `PayloadSchemaVersion
int32` at create time — if a schema with version 0 doesn't actually exist,
expect `409` (FK failure) rather than a `400` for the bad range itself;
the "gap" is that the *field itself* isn't range-checked, only its
existence as an FK target is enforced indirectly.

### ACT-05 — Get list / by id / by name (positive)
**Expect:** `200` each.

### ACT-06 — Patch / Delete (positive, standard CRUD)
**Expect:** `200`/`204` as expected.

---

## Action Dispatch (`POST /actions/:id/dispatch`)

### ACT-DISPATCH-01 — Dispatch valid payload to a connected node (positive)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/actions/$ACTION_ID/dispatch \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d "{\"node_id\":\"$NODE_ID\",\"payload\":{\"delay_ms\":1000}}"
```
**Expect:** `201`, action log body with `action_status: "UNRESPONDED"`.
Confirm via `mosquitto_sub` on `/sub/test-device-001/action` (see
`16-mqtt-flows.md`) that the exact payload
`{"execution_id":"...","action":"test_action","payload":{"delay_ms":1000}}`
was actually published. Save `EXECUTION_ID` from the response for the
action_ack MQTT test.

### ACT-DISPATCH-02 — Dispatch with a payload that fails schema validation (negative)
`{"node_id":"$NODE_ID","payload":{"delay_ms":"not-a-number"}}` (string
instead of integer, per the `restart` schema's `type: "integer"` field).
**Expect:** `400`, and confirm **no action log row is created** for this
failure path (per code audit — this differs from the node-not-found case
below, which DOES create a row). Verify via `GET /v1/action-logs` that the
count didn't increase.

### ACT-DISPATCH-03 — Dispatch to a nonexistent node (edge — verify surprising 201 behavior)
`{"node_id":"00000000-0000-0000-0000-000000000000","payload":{}}`.
**Expect:** `201` (not 404!) — per audit, an action log row IS created
with `node_id: null`, `action_status: "UNEXECUTED"`,
`action_message: "node not found"`. This is a deliberate design choice in
the code (always log the dispatch attempt), but confirm the actual
response status/body match this exactly since it's counterintuitive for
an API consumer expecting 404.

### ACT-DISPATCH-04 — Dispatch to a disconnected node (edge)
Publish an MQTT `status: OFFLINE` for `test-device-001` first (see
`16-mqtt-flows.md` MQTT-03), then dispatch.
**Expect:** `201`, action log `action_status: "UNEXECUTED"`,
`action_message: "node is not connected"`.

### ACT-DISPATCH-05 — Dispatch an action against a node of the WRONG node class (negative/edge, ⚠ known gap)
Create a second node class (`test_class` from `09-node-class.md`) and a
node under it (register a second device via MQTT), then dispatch
`test_action` (bound to `NCLASS_ID`/`base_node`) against that other node.
**Expect:** per audit, **there is no node-class compatibility check
anywhere in the dispatch flow** — this should proceed exactly like
ACT-DISPATCH-01 (schema validation + publish), fully ignoring the
class mismatch. Flag prominently: an action meant only for one hardware
class can currently be dispatched to any node regardless of its class.

### ACT-DISPATCH-06 — Dispatch to a nonexistent action id (negative)
`POST /v1/actions/00000000-0000-0000-0000-000000000000/dispatch`.
**Expect:** `404` — no action log row created at all (differs from
ACT-DISPATCH-03's node-not-found case, which does log — confirm this
asymmetry is real).

### ACT-DISPATCH-07 — Dispatch with a custom `executed_at` far in the future (edge, ⚠ known gap)
`{"node_id":"$NODE_ID","payload":{},"executed_at":"2099-01-01T00:00:00Z"}`.
**Expect:** `201` — no bounds check on `executed_at` at all.

---

## Action Log (`/v1/action-logs`)

### ALOG-01 — List (positive)
`GET /v1/action-logs`.
**Expect:** `200`, includes rows from the dispatch tests above.

### ALOG-02 — Delete (bulk/filtered) (positive)
`DELETE /v1/action-logs` (check whether this takes filter query params or
deletes unconditionally — read the handler/request shape if unclear from
this doc, and document exactly what filter parameters it accepts).
**Expect:** `200`/`204`.
