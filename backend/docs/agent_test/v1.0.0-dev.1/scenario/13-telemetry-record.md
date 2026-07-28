# 13 — Telemetry Record (`/v1/telemetry-records`)

## ⚠ Known gap — not a validation gap, a wiring gap

Per code audit: `internal/domain/usecases/telemetry/ingestion.go` +
`internal/application/telemetry/ingestion/usecase.go` is a fully
implemented ingestion usecase (validates payload against a payload schema,
then creates a `telemetry_record` row), and it IS constructed in
`internal/composition/main/application.go` — **but it is never wired to
any HTTP route or MQTT handler.** There is no `POST` route for telemetry
anywhere in `route.go`, and the MQTT `log` topic handler
(`internal/presentation/mqtt/handler/log.go` →
`MessagingCallback.Log`) is a **no-op stub that always returns `nil`**
without calling ingestion at all.

**Conclusion: there is currently no reachable way to create a
`telemetry_record` row through this running application**, via HTTP or
MQTT. Treat this entire scenario as query/cleanup-only against
whatever rows exist from prior manual/seeded data (there will likely be
zero rows in a fresh environment), and record this as a fix candidate
(wire the ingestion usecase to either a new HTTP endpoint or replace the
`Log` no-op) rather than something to test end-to-end right now.

---

### TELE-01 — List with no data (positive/edge)
`GET /v1/telemetry-records` with `Authorization: Bearer $ACCESS_TOKEN`.
**Expect:** `200`, empty array (assuming a fresh environment with nothing
manually inserted).

### TELE-02 — List with pagination params on an empty table (edge)
`GET /v1/telemetry-records?page=1&limit=10`.
**Expect:** `200`, `total: 0`.

### TELE-03 — Delete on an empty table (positive/edge)
`DELETE /v1/telemetry-records`.
**Expect:** `200`/`204`, no error even though nothing exists to delete.

### TELE-04 — Permission enforcement (positive, regression of `02-rbac-authorization.md`)
Confirm `telemetry_record:get`/`telemetry_record:remove` are correctly
required (the `user` role has `telemetry_record:get` but not `:remove` per
the seeder — verify a `user`-role token can `GET` but gets `401` on
`DELETE`).
