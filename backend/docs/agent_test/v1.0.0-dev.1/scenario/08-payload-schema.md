# 08 — Payload Schema (`/v1/admin/payload-schemas`)

Precondition: `ACCESS_TOKEN` (admin/super). The seeded `restart` schema
(name=`restart`, version=1) already exists — use different names for
create tests to avoid unrelated conflicts.

---

### PSCHEMA-01 — Create schema (positive)
```bash
curl -s -i -X POST http://127.0.0.1:18080/api/v1/admin/payload-schemas \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" \
  -d '{"name":"test_schema","version":1,"definition":{"type":"object","required":[],"properties":{}}}'
```
**Expect:** `201`.

### PSCHEMA-02 — Create with duplicate (name, version) (negative)
Repeat PSCHEMA-01.
**Expect:** `409` (composite unique constraint).

### PSCHEMA-03 — Create same name, different version (positive — versioning works)
`{"name":"test_schema","version":2,"definition":{...}}`.
**Expect:** `201` — confirms `(name, version)` composite key, not `name` alone.

### PSCHEMA-04 — Create with negative version (edge, ⚠ known gap)
`{"name":"test_schema_neg","version":-1,"definition":{"type":"object","required":[],"properties":{}}}`.
**Expect:** `201` — **zero range validation** on `version int32`. Flag for
future fix (should probably require `version >= 1`).

### PSCHEMA-05 — Create with malformed `definition` JSON (negative)
`{"name":"x","version":1,"definition": not-json}` (send genuinely invalid
JSON syntax in the body, not just an odd shape).
**Expect:** `400` (`RequiredRawJSON` only checks well-formedness, but
malformed JSON in the whole request body fails at bind time first).

### PSCHEMA-06 — Create with a `definition` that doesn't match the bespoke schema-definition shape at all (edge, ⚠ known gap)
`{"name":"junk_schema","version":1,"definition":{"foo":"bar"}}` — valid
JSON, but not a `PayloadSchemaDefinition`-shaped object (missing `type`).
**Expect:** `201` — creation only checks `json.Valid()`, never validates
the definition's shape against `PayloadSchemaDefinition` at write time.
The mismatch only surfaces later, when something tries to actually
validate a payload against this schema (see `12-action.md` ACT-xx dispatch
tests) — at which point the validator's `switch` on `Type` hits the
`default` branch and every dispatch attempt against this schema will fail
with "unsupported schema type". Flag as a gap: schema shape should
probably be validated at creation time, not deferred to first use.

### PSCHEMA-07 — Create with `valid_to` before `valid_from` (edge, ⚠ known gap)
`{"name":"backwards_validity","version":1,"definition":{"type":"object","required":[],"properties":{}},"valid_from":"2030-01-01T00:00:00Z","valid_to":"2020-01-01T00:00:00Z"}`.
**Expect:** `201` — no ordering check between `valid_from`/`valid_to`.

### PSCHEMA-08 — Get latest by name (positive)
`GET /v1/admin/payload-schemas/latest?name=test_schema`.
**Expect:** `200`, returns version 2 (the highest).

### PSCHEMA-09 — Get by name+version (positive)
`GET /v1/admin/payload-schemas/by-name-version?name=test_schema&version=1`.
**Expect:** `200`.

### PSCHEMA-10 — Get by name+version with negative version query param (edge, ⚠ known gap)
`GET /v1/admin/payload-schemas/by-name-version?name=test_schema&version=-5`.
**Expect:** `400` if non-numeric fails first, but `-5` IS a valid int32 —
per audit, `RequiredInt32` has no range check, so this parses fine and
simply returns `404` (no such row), not a `400` for the bad range. Confirm.

### PSCHEMA-11 — List / get by id / patch / delete (positive, standard CRUD)
Standard round-trip, same shape as prior resources — `200`/`201`/`204` as expected.
