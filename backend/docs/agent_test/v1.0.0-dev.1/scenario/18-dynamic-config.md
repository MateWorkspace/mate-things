# 18 — Dynamic Config over MQTT (`/v1/firmwares/:id/config-parameters`, `/v1/nodes/:id/config`)

Precondition: `ACCESS_TOKEN` (super/admin — only those roles hold
`node_config:get` / `node_config:set`). `NCLASS_ID` from `09-node-class.md`.
`USER_TOKEN` (role `user`) from `02-rbac-authorization.md` for CFG-11.
A registered node (`NODE_ID`, from `16-mqtt-flows.md` MQTT-01) whose
firmware is the one created in CFG-01. An MQTT subscriber on
`mate/+/config` to observe published values:

```bash
head -c 1024 /dev/urandom > /tmp/fw-cfg.bin
mosquitto_sub -h 127.0.0.1 -p 1883 -t 'mate/#' -v &
```

---

### CFG-01 — Create firmware with a `config_schema` (positive, multipart)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/firmwares \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "node_class_id=$NCLASS_ID" -F "name=cfg_firmware_v1" \
  -F "file=@/tmp/fw-cfg.bin" \
  -F 'config_schema=[{"key":"mqtt_host","value_type":"string"},{"key":"mqtt_pass","value_type":"string"},{"key":"sys_rst_aft_ms","value_type":"uint32"},{"key":"wifi_try_init","value_type":"bool"}]'
```
**Expect:** `201`. Save `id` as `CFG_FW_ID`. Note the wire field name is
`value_type`, not `type` — the firmware repo's `upload.py` sends
`value_type` to match.

### CFG-02 — Get config parameters for a firmware with a schema (positive)
```bash
curl -s -i http://127.0.0.1:18080/api/v1/firmwares/$CFG_FW_ID/config-parameters \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```
**Expect:** `200`, a 4-element array of `{key, value_type}` matching CFG-01
exactly (same keys, same types). An empty array here means schema ingestion
silently failed — that is the regression this case guards.

### CFG-03 — Create firmware with an invalid `value_type` (negative, ordering check)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/firmwares \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "node_class_id=$NCLASS_ID" -F "name=cfg_firmware_bad" \
  -F "file=@/tmp/fw-cfg.bin" \
  -F 'config_schema=[{"key":"mqtt_host","value_type":"float64"}]'
```
**Expect:** `400` (allowed types are `string`, `uint32`, `bool`).
Then confirm **no side effect happened**: `GET /v1/firmwares/by-name/cfg_firmware_bad`
returns `404`, and no new MinIO object was created (schema is validated
before the binary is stored). Same check applies to
`PUT /v1/firmwares/$CFG_FW_ID/binary` with a bad `config_schema`.

### CFG-04 — Create firmware with an empty key in the schema (negative)
`config_schema=[{"key":"","value_type":"string"}]`.
**Expect:** `400`, and no firmware row/object created (as in CFG-03).

### CFG-05 — Get config parameters for a nonexistent firmware id (edge, ⚠ known gap)
**Expect:** `200 []` — `ReadByFirmwareId` does not check firmware
existence, so an unknown id is indistinguishable from a firmware with no
schema. Document actual behavior; not fixing now.

### CFG-06 — Get node config before anything is set (positive/edge)
```bash
curl -s -i http://127.0.0.1:18080/api/v1/nodes/$NODE_ID/config \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```
**Expect:** `200 []` — the schema defines what *may* be set; no values
stored yet.

### CFG-07 — Set a config value (positive)
```bash
curl -s -i -X PUT http://127.0.0.1:18080/api/v1/nodes/$NODE_ID/config \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d '{"key":"mqtt_host","value":"broker.example.com"}'
```
**Expect:** `204`. The `mosquitto_sub` window shows a publish on the node's
`config` topic carrying the key/value.

### CFG-08 — Get node config after the set (positive)
Repeat CFG-06.
**Expect:** `200`, one entry `{key:"mqtt_host", value:"broker.example.com", updated_at:...}`.

### CFG-09 — Set an unknown config key (negative)
`{"key":"not_a_real_key","value":"x"}`.
**Expect:** `404` — the key must exist in the node's *current* firmware's
schema.

### CFG-10 — Set a value whose type doesn't match the schema (negative)
`{"key":"sys_rst_aft_ms","value":"soon"}` (schema says `uint32`), then
`{"key":"wifi_try_init","value":"yes"}` (schema says `bool`).
**Expect:** `400` each. `{"key":"wifi_try_init","value":"true"}` → `204`.

### CFG-11 — Get/set config for a nonexistent node UUID (negative, regression)
```bash
GONE=$(python3 -c "import uuid;print(uuid.uuid4())")
curl -s -i http://127.0.0.1:18080/api/v1/nodes/$GONE/config \
  -H "Authorization: Bearer $ACCESS_TOKEN"
curl -s -i -X PUT http://127.0.0.1:18080/api/v1/nodes/$GONE/config \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d '{"key":"mqtt_host","value":"x"}'
```
**Expect:** `404` (`node not found`) on both GET and PUT. GET previously
returned `200 []` for an unknown node because it skipped the node lookup —
this case guards that fix. Use a *random* UUID, not the all-zeros one:
`RequiredUUID` rejects the nil UUID as malformed, so that returns `400`
before the lookup is reached.

### CFG-12 — Config endpoints as a role lacking `node_config:get`/`node_config:set` (negative)
Using `USER_TOKEN` (role `user`, which holds neither permission):
```bash
curl -s -i http://127.0.0.1:18080/api/v1/nodes/$NODE_ID/config \
  -H "Authorization: Bearer $USER_TOKEN"
curl -s -i -X PUT http://127.0.0.1:18080/api/v1/nodes/$NODE_ID/config \
  -H "Authorization: Bearer $USER_TOKEN" -H "Content-Type: application/json" \
  -d '{"key":"mqtt_host","value":"nope.example.com"}'
```
**Expect:** `403` each (per `02-rbac-authorization.md` RBAC-03). Confirm the
PUT did not change the stored value.

### CFG-13 — Node config after a firmware reassignment (edge, lineage semantics)
1. Create a second firmware `cfg_firmware_v2` on the same node class whose
   `config_schema` **omits** `mqtt_host` (e.g. only `sys_rst_aft_ms`).
2. Reassign the node to it (`PATCH /v1/nodes/$NODE_ID/firmware`, see
   `10-node.md` NODE-08).
3. `GET /v1/nodes/$NODE_ID/config`.

**Expect:** the previously-set `mqtt_host` entry is **no longer listed** —
GET filters to keys in the node's current firmware's schema, so read and
write agree on the node's config surface. The row is retained as history
(not deleted); a `PUT mqtt_host` now returns `404`, consistent with the GET.

### CFG-14 — Secret-ish keys are returned in plaintext (edge, accepted tradeoff)
Set `mqtt_pass`, then `GET /v1/nodes/$NODE_ID/config`.
**Expect:** `204`, then the value echoed verbatim in the GET response.
`mqtt_pass` is stored as plaintext `TEXT` and not masked. This is a
documented, accepted tradeoff rather than an auth hole: the endpoint is
gated behind `node_config:get`, granted only to `super`/`admin` (CFG-12).
