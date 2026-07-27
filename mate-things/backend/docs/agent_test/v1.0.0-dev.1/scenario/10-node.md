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

### NODE-05 — Patch `device_id` (negative/edge — real gap, likely unintended)
`PATCH /v1/nodes/$NODE_ID` `{"device_id":""}` (empty string).
**Expect:** per audit, `NodePatchRequest.DeviceId *string` is forwarded
**with zero validation, not even `RequiredString`** — an empty string is
plausibly accepted and persisted, which would desync the node from any
future MQTT traffic for its real physical `device_id` (registration/
status/action_ack all key off `device_id`). This is a meaningfully worse
gap than the cosmetic ones elsewhere — flag prominently. Verify actual
behavior and confirm whether `device_id` even *should* be patchable via
this generic endpoint at all (it may be a design mistake that it's exposed
here rather than immutable post-registration).

### NODE-06 — Patch `device_id` to another existing node's device_id (negative)
**Expect:** `409` (`device_id` has a `UNIQUE` constraint) — assuming
NODE-05 didn't already corrupt state; if it did, redo registration first.

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
  -d "{\"firmware_id\":\"$FW_ID_2\",\"firmware_url\":\"https://example.com/fw.bin\"}"
```
**Expect:** `200`/`202` — and confirm via `mosquitto_sub` (see
`16-mqtt-flows.md` MQTT-OTA) that the server actually published to
`/sub/test-device-001/ota`.

### NODE-12 — OTA dispatch with a malformed `firmware_url` (edge, ⚠ known gap)
`{"firmware_id":"$FW_ID_2","firmware_url":"not a url at all!! 🎉"}`.
**Expect:** per audit, **no URL-format validation anywhere** in this path
(`net/url.Parse` never called) — expect this to succeed and get published
to MQTT verbatim as the `firmware_url` string, garbage and all.

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
