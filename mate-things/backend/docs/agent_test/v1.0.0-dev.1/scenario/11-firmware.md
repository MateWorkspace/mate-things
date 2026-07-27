# 11 — Firmware (`/v1/firmwares`, binary upload/download)

Precondition: `ACCESS_TOKEN`. `NCLASS_ID` from `09-node-class.md` (or use
seeded `base_node`'s id). Need a small dummy binary file:

```bash
head -c 1024 /dev/urandom > /tmp/fw-test.bin
head -c 1024 /dev/urandom > /tmp/fw-test-v2.bin
```

---

### FW-01 — Create firmware with binary upload (positive, multipart)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/firmwares \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "node_class_id=$NCLASS_ID" -F "name=test_firmware_v1" -F "file=@/tmp/fw-test.bin"
```
**Expect:** `201`, body includes computed `checksum` (SHA-256) and `size`.
Save `id` as `FW_ID`.

### FW-02 — Create with duplicate name (negative)
Repeat FW-01 with the same `name`.
**Expect:** `409`. Also confirm no orphaned MinIO object was left behind
from this failed attempt (per audit, the code deletes the just-stored
MinIO object if the DB insert fails — worth spot-checking via MinIO if
console/mc access is available, otherwise just trust the code review note).

### FW-03 — Create with missing `file` part (negative)
Same form without `-F "file=@..."`.
**Expect:** `400` (`RequiredFormFile`).

### FW-04 — Create with an empty (zero-byte) file (negative/edge)
`-F "file=@/dev/null"`.
**Expect:** `400` (`RequiredFormFile` checks `Size > 0`).

### FW-05 — Create with a non-firmware file, e.g. a `.txt` (edge, ⚠ known gap)
```bash
echo "this is not firmware" > /tmp/fake.txt
curl -s -i -X POST http://127.0.0.1:18080/api/v1/firmwares \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "node_class_id=$NCLASS_ID" -F "name=fake_firmware" -F "file=@/tmp/fake.txt"
```
**Expect:** `201` — **zero content-type/extension/magic-byte validation**;
any bytes are accepted as "firmware". Flag as a gap; not fixing now.

### FW-06 — Create with nonexistent `node_class_id` (negative)
**Expect:** `409` (FK violation pattern, consistent with other resources).

### FW-07 — Get list / by id / by name (positive)
**Expect:** `200` each.

### FW-08 — Get nonexistent by id/name (negative)
**Expect:** `404` each.

### FW-09 — Get binary by id / by name (positive)
```bash
curl -s -i http://127.0.0.1:18080/api/v1/firmwares/$FW_ID/binary \
  -H "Authorization: Bearer $ACCESS_TOKEN" -o /tmp/fw-downloaded.bin
```
**Expect:** `200`, headers include `Content-Disposition`,
`X-Firmware-Checksum`, `X-Firmware-Name`. Diff `/tmp/fw-downloaded.bin`
against `/tmp/fw-test.bin` — must be byte-identical.

### FW-10 — Binary stat (GET) (positive)
`GET /v1/firmwares/by-name/test_firmware_v1/binary/stat`.
**Expect:** `200`, JSON `{binary_path, size, checksum}` matching FW-01's response.

### FW-11 — Binary stat (HEAD) (positive)
```bash
curl -s -I http://127.0.0.1:18080/api/v1/firmwares/by-name/test_firmware_v1/binary \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```
**Expect:** `200`, no body, `Content-Length`/`X-Firmware-Checksum` headers present.

### FW-12 — Get binary for nonexistent firmware (negative)
**Expect:** `404`.

### FW-13 — Replace binary (`PUT .../binary`) (positive)
```bash
curl -s -i -X PUT http://127.0.0.1:18080/api/v1/firmwares/$FW_ID/binary \
  -H "Authorization: Bearer $ACCESS_TOKEN" -F "file=@/tmp/fw-test-v2.bin"
```
**Expect:** `200`/`204`. Re-download (FW-09-style) and confirm the content
+ checksum now match `/tmp/fw-test-v2.bin`, not the original.

### FW-14 — Replace binary for nonexistent firmware id (negative)
**Expect:** `404`.

### FW-15 — Patch firmware metadata (name/node_class_id) (positive)
`PATCH /v1/firmwares/$FW_ID` `{"name":"test_firmware_v1_renamed"}`.
**Expect:** `200`/`204`. ⚠ Note: confirm whether renaming the DB row also
renames/moves the underlying MinIO object key (the object key was
originally derived from the firmware's name per the audit) — if not, a
mismatch between the DB `name` and the actual stored object path could
occur. Document actual behavior either way.

### FW-16 — Delete firmware (positive)
`DELETE /v1/firmwares/$FW_ID`.
**Expect:** `200`/`204`. ⚠ Confirm whether the underlying MinIO object is
actually deleted too, or just the DB row (soft-deleted) — if MinIO objects
are never cleaned up on firmware delete, that's a storage-leak gap worth
flagging (not fixing now).

### FW-17 — Get deleted firmware's binary (negative)
**Expect:** `404`.
