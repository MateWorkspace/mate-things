# 16 — MQTT Flows

The backend connects to the **real external HiveMQ Cloud broker**
configured in `.env` (`BE_MQTT_BROKER_URL=ssl://...:8883`,
`BE_MQTT_USERNAME=matemqtt`). These tests use `mosquitto_pub`/`mosquitto_sub`
against that same broker, so the running backend container observes/reacts
to them in real time. Source the same credentials from `.env`:

```bash
ENV_FILE=/home/dhonan-wsl/Repositories/mate/mate-things/.env  # repo root, adjust if run from elsewhere
MQTT_HOST=$(grep BE_MQTT_BROKER_URL "$ENV_FILE" | sed -E 's#.*://([^:]+):.*#\1#')
MQTT_PORT=8883
MQTT_USER=$(grep BE_MQTT_USERNAME "$ENV_FILE" | cut -d= -f2)
MQTT_PASS=$(grep BE_MQTT_PASSWORD "$ENV_FILE" | cut -d= -f2)
```

Common `mosquitto_pub`/`mosquitto_sub` flags for all commands below:
```
-h $MQTT_HOST -p $MQTT_PORT --cafile /etc/ssl/certs/ca-certificates.crt \
-u "$MQTT_USER" -P "$MQTT_PASS"
```

---

### MQTT-01 — Device registration (positive)
```bash
mosquitto_pub -h $MQTT_HOST -p $MQTT_PORT --cafile /etc/ssl/certs/ca-certificates.crt \
  -u "$MQTT_USER" -P "$MQTT_PASS" \
  -t "/pub/registration" \
  -m '{"device_id":"test-device-001","device_info":"esp32-test-rig","firmware_name":"restart-fw"}'
```
**Expect:** within a couple seconds, `curl GET /v1/nodes/by-device/test-device-001`
returns `200` with a new node row. Also expect a `registration_ack` message
(empty payload) on `/sub/test-device-001/registration_ack` — confirm with
a subscriber running concurrently:
```bash
mosquitto_sub -h $MQTT_HOST -p $MQTT_PORT --cafile /etc/ssl/certs/ca-certificates.crt \
  -u "$MQTT_USER" -P "$MQTT_PASS" -t "/sub/test-device-001/#" -v &
```
Note: `firmware_name":"restart-fw"` must match an existing `firmwares.name`
row for `UpsertRegistration`'s join to succeed — if no such firmware
exists yet, expect registration to silently fail server-side (repo-level
`NotFound("firmware not found")`) with **no error surfaced back to the
device at all** (no ack, no error topic) — create a firmware named
`restart-fw` first via `11-firmware.md` FW-01 if needed, or use whatever
firmware name already exists in your environment.

### MQTT-02 — Registration with missing required field (negative/edge)
```bash
mosquitto_pub ... -t "/pub/registration" -m '{"device_id":"test-device-002"}'
```
**Expect:** dropped silently server-side (warning logged, confirmed in
code — `device_info`/`firmware_name` both required non-empty after trim).
No node created; no ack sent. Verify via `curl` that
`/v1/nodes/by-device/test-device-002` stays `404`.

### MQTT-03 — Status update ONLINE/OFFLINE (positive)
```bash
mosquitto_pub ... -t "/pub/test-device-001/status" -m '{"status":"ONLINE"}'
```
**Expect:** `GET /v1/nodes/by-device/test-device-001` → `is_connected: true`.
```bash
mosquitto_pub ... -t "/pub/test-device-001/status" -m '{"status":"OFFLINE"}'
```
**Expect:** `is_connected: false`.

### MQTT-04 — Status with invalid value (negative/edge)
`-m '{"status":"KINDA_ONLINE"}'`.
**Expect:** dropped, warning logged, node's `is_connected` unchanged from
its prior value.

### MQTT-05 — Status for an unregistered device_id (negative/edge)
`-t "/pub/does-not-exist/status" -m '{"status":"ONLINE"}'`.
**Expect:** logged error (node lookup by device_id fails), no crash, no
row created.

### MQTT-06 — Log message (edge — confirms no-op, see `13-telemetry-record.md`)
`-t "/pub/test-device-001/log" -m 'hello from device'`.
**Expect:** message is received/subscribed (confirm no error in backend
logs), but has **zero effect** — the handler is a no-op stub. Nothing to
assert beyond "doesn't crash the connection."

### MQTT-07 — Action dispatch reaches the device (positive, ties to `12-action-and-action-log.md` ACT-DISPATCH-01)
While subscribed to `/sub/test-device-001/action`, trigger
`POST /v1/actions/:id/dispatch` via curl (see ACT-DISPATCH-01).
**Expect:** the subscriber receives exactly one message:
`{"execution_id":"<uuid>","action":"test_action","payload":{"delay_ms":1000}}`.

### MQTT-08 — Action ack, SUCCESS (positive)
Using the `EXECUTION_ID` captured from ACT-DISPATCH-01:
```bash
mosquitto_pub ... -t "/pub/test-device-001/action_ack" \
  -m "{\"execution_id\":\"$EXECUTION_ID\",\"status\":\"SUCCESS\",\"message\":\"done\"}"
```
**Expect:** `GET /v1/action-logs` shows that row's `action_status` now
`SUCCESS`, `action_message: "done"`.

### MQTT-09 — Action ack, FAILED (positive, separate dispatch)
Dispatch a new action, then ack with `"status":"FAILED","message":"could not restart"`.
**Expect:** action log status `FAILED`.

### MQTT-10 — Action ack with invalid status value (negative/edge)
`"status":"MAYBE"`.
**Expect:** dropped silently (only `SUCCESS`/`FAILED` accepted,
case-insensitive/trimmed per code), action log stays `UNRESPONDED`.

### MQTT-11 — Action ack with unknown execution_id (negative/edge)
Random UUID not tied to any real dispatch.
**Expect:** logged not-found error server-side, no crash, obviously no
matching row updated (there isn't one).

### MQTT-12 — OTA message reaches the device (positive, ties to `10-node.md` NODE-11)
Subscribe to `/sub/test-device-001/ota`, then trigger the OTA dispatch via
curl.
**Expect:** message
`{"firmware_url":"...","firmware_size":<int>,"firmware_checksum":"..."}`
received exactly once.

### MQTT-13 — Resubscription after broker reconnect (edge)
```bash
docker restart mate-things-test
sleep 8
mosquitto_pub ... -t "/pub/test-device-001/status" -m '{"status":"ONLINE"}'
```
**Expect:** after the container's MQTT client reconnects
(`OnConnect`→`Resubscribe`), the status update is still processed
correctly (confirm via curl) — i.e. per-device topic subscriptions for
every existing node survive a reconnect, not just the global
`/pub/registration` topic.
