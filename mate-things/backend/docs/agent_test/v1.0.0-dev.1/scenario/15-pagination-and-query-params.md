# 15 — Pagination & Common Query Parameters (cross-cutting)

These apply to every `list` endpoint (`GET /v1/admin/permissions`,
`/roles`, `/users`, `/node-classes`, `/nodes`, `/firmwares`, `/actions`,
`/action-logs`, `/telemetry-records`, `/role-permissions`, etc.). Run this
matrix against at least two representative endpoints
(`/v1/admin/permissions` — has 46+ rows to actually paginate through —
and one smaller list) rather than every single one, unless time allows the
full sweep.

Precondition: `ACCESS_TOKEN`.

---

### PAGE-01 — Default pagination (positive)
`GET /v1/admin/permissions` (no query params).
**Expect:** `200`, response pagination metadata shows `page: 1`, `limit: 10`.

### PAGE-02 — Explicit valid page/limit (positive)
`GET /v1/admin/permissions?page=2&limit=5`.
**Expect:** `200`, 5 items (or fewer on the last page), `page: 2`.

### PAGE-03 — Non-numeric page (negative)
`GET /v1/admin/permissions?page=abc`.
**Expect:** `400`.

### PAGE-04 — Non-numeric limit (negative)
`GET /v1/admin/permissions?limit=abc`.
**Expect:** `400`.

### PAGE-05 — Negative page (edge, ⚠ known gap — silently clamped, not rejected)
`GET /v1/admin/permissions?page=-5`.
**Expect:** `200` with `page` clamped back to `1` in the response metadata
— confirmed in code (`queryPositiveInt` clamps non-positive values to the
default rather than erroring). Document the actual returned `page` value.

### PAGE-06 — Zero limit (edge, ⚠ known gap — same clamping behavior)
`GET /v1/admin/permissions?limit=0`.
**Expect:** `200` with `limit` clamped to `10`.

### PAGE-07 — Extremely large limit, no upper bound (edge, ⚠ known gap — real DoS-shaped issue)
`GET /v1/admin/permissions?limit=1000000`.
**Expect:** `200` — per audit, **there is no max-limit clamp anywhere**;
the value is passed straight to SQL `LIMIT`. With only ~50 rows this is
harmless right now, but flag prominently as a fix candidate (cap `limit`
server-side, e.g. at 100), since it becomes a real resource-exhaustion
vector once tables like `telemetry_records`/`action_logs` grow large.

### PAGE-08 — `search` with SQL wildcard characters (edge)
`GET /v1/admin/permissions?search=%`.
**Expect:** `200` — confirm it behaves like a literal `%` in an ILIKE
pattern (likely matches everything, since `%` is itself the ILIKE
wildcard) rather than erroring; queries are parameterized via `squirrel`
so no injection risk, just confirm the wildcard-matching behavior is
sane/expected and doesn't crash.

### PAGE-09 — `search` with a very long string (edge)
`GET "/v1/admin/permissions?search=$(python3 -c 'print("a"*5000)')"`.
**Expect:** `200`, empty results (no crash/500 from an oversized search term).

### PAGE-10 — `QueryTime` filters with malformed timestamp (where applicable, negative)
Find an endpoint using a time filter (check `payload-schemas` `valid_at`
filter or similar) and pass a non-RFC3339 string, e.g. `?valid_at=2024-01-01`.
**Expect:** `400`, "must be a valid RFC3339 timestamp".

### PAGE-11 — `QueryInt32` filter with negative value (edge)
E.g. `GET /v1/actions?payload_schema_version=-1`.
**Expect:** `200`, empty results (no range validation, just matches nothing).
