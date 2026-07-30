# 10 — Node (`/v1/nodes`, `/v1/nodes/by-device/:device_id`, OTA dispatch)

Precondition: `ACCESS_TOKEN`. **There is no `POST /nodes` HTTP route** —
nodes are only ever created via MQTT device registration. Run
`MQTT-01` from `16-mqtt-flows.md` first to register a device (e.g.
`device_id=test-device-001`) before any test here. Also need a second
firmware row from `11-firmware.md` (`FW_ID_2`) for the firmware-patch test.

```bash
NODE_ID=$(curl -s "http://127.0.0.1:18080/api/v1/nodes/by-device/test-device-001" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq -r .id)
```

---

### NODE-01 — Get list (positive)
`GET /v1/nodes`.
**Expect:** `200`, includes `test-device-001`.

### NODE-02 — Get by id / by device_id (positive)
**Expect:** `200` both.

### NODE-03 — Get nonexistent by id/device_id (negative)
**Expect:** `404` both.

### NODE-04 — Patch name/description (positive)
`PATCH /v1/nodes/$NODE_ID` `{"name":"Renamed Node","description":"..."}`.
**Expect:** `200`/`204`.

### NODE-05 — Patch cannot mutate `device_id` (negative)
Read the node's current `device_id`, then send
`PATCH /v1/nodes/$NODE_ID` `{"device_id":"","description":"identity check"}`.
**Expect:** `200`/`204`; the description changes, but a fresh GET shows the
original non-empty `device_id`. The generic update request intentionally
ignores this unknown field because device identity is immutable after MQTT
registration.

### NODE-06 — Spoof another node's `device_id` (negative)
Send a PATCH containing another registered node's device ID and a harmless
description change.
**Expect:** `200`/`204`; the description may change, but the target node keeps
its original `device_id` and the other node is unchanged.

### NODE-07 — Patch `node_class_id` to a nonexistent class (negative)
**Expect:** `409` (FK violation, same "conflict not not-found" pattern seen
throughout).

### NODE-08 — Patch firmware via dedicated endpoint (positive)
`PATCH /v1/nodes/$NODE_ID/firmware` `{"firmware_id":"$FW_ID_2"}`.
**Expect:** `200`/`204`.

### NODE-09 — Patch firmware with a firmware id belonging to a different node class (edge, ⚠ known gap — verify)
Create a firmware under a *different* node class than the node's own, then
patch it in via NODE-08's endpoint.
**Expect:** verify whether there's any node-class-compatibility check
between a node and the firmware assigned to it — if none, document as a
gap (a node could end up "running" firmware built for an incompatible
class).

### NODE-10 — Get available firmwares for a node (positive)
`GET /v1/nodes/$NODE_ID/firmwares/available`.
**Expect:** `200`.

### NODE-11 — OTA dispatch by node id (positive)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/nodes/$NODE_ID/ota \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d "{\"firmware_id\":\"$FW_ID_2\"}"
```
**Expect:** `204` — and confirm via `mosquitto_sub` (see `16-mqtt-flows.md`
MQTT-OTA) that the server published to `/sub/test-device-001/ota`. The
published URL, size, and checksum must come from the selected backend firmware
row and its presigned stored binary; the client does not supply a URL.

### NODE-12 — OTA dispatch with incompatible firmware (negative)
Create/select a firmware whose `node_class_id` differs from the node, then POST
its `firmware_id` using NODE-11's request.
**Expect:** `400`; no OTA message is published. Compatibility is enforced by
the dispatch usecase before presigning or MQTT publish.

### NODE-13 — OTA dispatch by device_id, nonexistent device (negative)
`POST /v1/nodes/by-device/does-not-exist/ota`.
**Expect:** `404`.

### NODE-14 — Delete node (positive)
`DELETE /v1/nodes/$NODE_ID`.
**Expect:** `200`/`204`.

### NODE-15 — Re-register the same device_id after node deletion (edge)
Re-run the MQTT registration for `test-device-001` (MQTT-01) after NODE-14.
**Expect:** since `UpsertRegistration` does `INSERT ... ON CONFLICT
(device_id) DO UPDATE`, and the prior row is soft-deleted (not hard
deleted), verify what actually happens: does the soft-deleted row get
"revived" (its `deleted_at` cleared) via the upsert's `ON CONFLICT DO
UPDATE`, or does the `UNIQUE` constraint on `device_id` block a fresh
insert entirely because the soft-deleted row still occupies that
`device_id`? This is worth confirming precisely since it directly affects
whether a physically re-flashed/re-registered device can ever come back
after an admin deletes its node record — flag whatever the actual answer
is.
