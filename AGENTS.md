# mate-things

## Project goals

`mate-things` (internally also referred to as "Nusapala Things" in a few
places — see [Naming note](#naming-note)) is an IoT fleet-management
platform for ESP32-class devices. It provides:

- **Device registration & lifecycle** — nodes register themselves over MQTT,
  are grouped into node classes, and report online/offline status.
- **Firmware OTA** — firmware binaries are uploaded per node class and
  dispatched to devices for over-the-air updates.
- **Action dispatch** — admins define actions (validated against a
  payload schema) and dispatch them to specific nodes over MQTT; dispatch
  results are recorded as action logs.
- **Telemetry ingestion & live broadcast** — devices publish telemetry
  records that the API exposes for querying, plus a websocket
  (`GET /v1/telemetry/broadcast`, filtered by node/metric) that streams new
  records as they arrive and pushes the latest known reading immediately on
  connect so the UI isn't stuck waiting for the device's next publish.
- **Node-log ingestion** — devices forward UTC log lines over MQTT; the
  backend parses and stores their level, tag, message, and embedded timestamp
  in a compressed TimescaleDB hypertable exposed through GET/DELETE endpoints.
- **RBAC** — permissions, roles, and role↔permission assignments gate every
  admin-facing endpoint; users authenticate via JWT access/refresh tokens, or
  via a per-user `X-Api-Key` header (one key per user, optional expiry,
  admin-managed at `/admin/api-keys`) as a lighter alternative for
  app-to-app callers.
- **Preferences** — arbitrary per-resource key/value preferences (e.g. UI
  state) attached to any entity.

The Go backend is the authority for business rules and device orchestration.
The Next.js frontend is a complete permission-aware operations cockpit for
fleet, observability, and administration workflows.

## Project structure

```
mate-things/
├── backend/           Go API + MQTT service (see below)
├── frontend/           Next.js operations cockpit (see frontend/AGENTS.md)
├── Dockerfile           multi-stage build: backend + frontend into one runtime image
├── entrypoint.sh        container entrypoint: migrate → seed → run
├── compose.yml           minimal single-service compose file for local/prod runs
└── .env.example         documents every BE_* environment variable
```

### Backend (`backend/`)

Go module `github.com/MateWorkspace/mate-things/backend`, structured as
clean/layered architecture with a strict dependency direction:

```
presentation → application → domain ← infrastructure
                                ↑
                          composition (wires everything together)
```

- **`internal/domain/`** — the core, dependency-free layer.
  - `models/` — plain domain structs (e.g. `Node`, `Firmware`, `PayloadSchema`,
    `NodeLog`).
  - `contracts/` — interfaces the application layer depends on and
    infrastructure implements: `repository/` (Postgres access per entity),
    `cache/`, `storage/` (firmware binaries via MinIO), `node/` (MQTT
    publish/subscribe), `broadcaster/` (per-domain-typed live websocket
    fan-out, e.g. `Telemetry`), `utility/` (password hashing, tokens, API
    key generation/hashing, payload schema validation), `logger/`.
  - `usecases/` — one interface per feature area (`admin`, `node`, `node_log`,
    `action`, `auth`, `profile`, `preferences`, `telemetry`, `seeder`, and an
    internal `repocache` interface set used by the cache-decorator layer) plus
    their request/response DTOs.
- **`internal/application/`** — one package per usecase interface,
  implementing the business logic: orchestrates repositories/contracts,
  and is where **value validation** lives (see
  [Validation convention](#validation-convention) below). `application/shared/`
  holds the reusable value-validation helpers (`RequiredUsername`,
  `RequiredPersonName`, `RequiredPayloadSchemaDefinition`, etc.).
  `application/repocache/` is a Redis-backed caching decorator sitting in
  front of the raw Postgres repositories, implementing the `repocache`
  usecase interfaces domain-side.
- **`internal/infrastructure/`** — concrete implementations of domain
  contracts: `repository/` (Postgres via squirrel + pgx), `cache/` (Redis),
  `storage/firmware/` (MinIO, presigned downloads), `node/publish` +
  `node/subscriptions` (MQTT via paho), `utility/` (bcrypt password hashing,
  JWT tokens, the hand-rolled payload-schema validator), `logger/leveled`
  (zerolog/slog dual backend).
- **`internal/presentation/`** — the only layer allowed to touch raw
  HTTP/MQTT request shapes:
  - `http/handler/` — one package per resource area, Echo handlers.
  - `http/request/`, `http/response/` — request DTOs and response shaping.
  - `http/route/route.go` — all HTTP route registration in one place.
  - `http/middleware/` — JWT auth + permission-checking middleware.
  - `http/utils/` — raw-shape validation only (UUID/int/JSON parsing,
    presence checks) — see [Validation convention](#validation-convention).
  - `http/proxy/` — reverse proxies: one to the Next.js frontend, one to
    MinIO for presigned firmware downloads (`/minio-proxy/*`). The Go
    backend is the container's sole ingress; there is no nginx.
  - `mqtt/handler/`, `mqtt/dto/`, `mqtt/event/` — MQTT topic handlers for
    device registration, status, action acks, telemetry, and raw device logs.
- **`internal/composition/`** — dependency wiring only, no business logic.
  `main/` wires the full HTTP+MQTT service (`driver.go` sets up
  Postgres/Redis/MinIO/MQTT clients and the Echo instance; `infrastructure.go`
  constructs repositories/caches; `application.go` constructs usecases;
  `presentation.go` constructs handlers, proxies, and registers routes;
  `launcher.go` is `Launch() int`, the single entrypoint). `seeder/` is a
  parallel, smaller composition for the one-shot seeder binary (Postgres
  only — no HTTP, Redis, or MQTT).
- **`internal/config/`** — all `BE_*` environment variables, loaded once via
  `config.LoadEnv()` at process start; every setting has a hardcoded default
  in `env.go`. Never read `os.Getenv` outside this package.
- **`cmd/main/`** — the API+MQTT binary entrypoint (`compositionmain.Launch()`);
  carries the swagger `@title`/`@BasePath`/etc. annotations.
- **`cmd/seeder/`** — the one-shot DB seeder binary
  (`compositionseeder.Launch()`), run once by `entrypoint.sh` before the
  main binary starts.
- **`database/migrations/`** — golang-migrate SQL migrations,
  timestamp-prefixed. Time-series tables use the existing TimescaleDB
  hypertable conventions.
- **`database/seeder/`** — baseline RBAC + example node class/action/user
  seed data as JSON, embedded into the seeder binary via `go:embed`.
- **`docs/swagger/`** — generated by `swag init` (see
  [Regenerating swagger docs](#regenerating-swagger-docs)); never hand-edit.
- **`docs/agent_test/`** — a manual/scripted test checklist per release
  (`checklist.md` + one `scenario/NN-*.md` file per feature area), used to
  verify a build end-to-end against a real Postgres/Redis/MinIO/MQTT
  environment before release.
- **`*_test.go` files** — focused automated tests now cover node-log parsing,
  query delegation, HTTP behavior, route registration, and seed RBAC data.
  Keep new behavior similarly protected even though older packages remain
  test-free.
- **`pkg/pgxdt/`** — a small standalone pgx transaction-context helper,
  importable outside this module if ever needed.

### Frontend (`frontend/`)

Next.js 16 App Router + React 19 operations cockpit. Read
`frontend/AGENTS.md` before changing it; that file is authoritative for
frontend structure, styling, Server Component/Action boundaries, and
verification.

- `src/app/(authenticated)/` owns the protected product routes and shared
  shell: dashboard, nodes, node classes, firmware, actions, action history,
  telemetry, node logs, users, access control, and payload schemas.
- Route-local UI and mutation code is colocated in `_components/` and
  `_lib/`; reusable collection, layout, profile, preference, record,
  refresh, and primitive UI components live in `src/components/`.
- `src/lib/api/` is the only frontend backend-API boundary. Server
  Components perform reads; feature-local Server Actions recheck permissions
  before mutations.
- Navigation is grouped by purpose and filtered by exact permissions in
  `src/config/navigation.ts`. Hiding a link is not authorization: protected
  pages and every Server Action must independently enforce access.
- Collection pages use URL-owned filters/pagination and responsive resource
  cards rather than desktop-only tables. Record-heavy pages share bounded
  time filters, JSON inspection, scoped deletion, and visibility-aware smart
  refresh.
- Authentication tokens remain in HTTP-only cookies. The Go service is the
  production ingress and reverse-proxies non-API traffic to the standalone
  Next.js server.

## Development flow

### Environment

Copy `.env.example` to `.env` and fill in real credentials for Postgres,
Redis, MinIO, and an MQTT broker (HiveMQ Cloud in practice) — none of these
are bundled/started by this repo; they're expected to already exist
somewhere reachable. `BE_BASE_URL` must be the public URL the app is served
at (used to build presigned firmware-download links); `BE_MQTT_CLIENT_ID`
must be unique per running instance (a shared broker kicks whichever client
connects second with the same ID).

### Running everything (Docker)

```bash
docker compose up --build
```

This builds one image (Go backend binary + seeder binary + Next.js
standalone build) and runs a single container whose `entrypoint.sh`:
1. runs `migrate ... up` against `BE_POSTGRES_*`,
2. runs the seeder binary (idempotent — safe on every restart),
3. starts the backend (binds `:80` directly, hardcoded — it's the
   container's public ingress) and the Next.js server, under one process
   group.

The backend is the sole ingress: it serves `/api/*` itself, reverse-proxies
everything else to Next.js, and reverse-proxies `/minio-proxy/*` to MinIO
for presigned downloads. There is no separate proxy process.

### Running locally without Docker

- Backend: `cd backend && go run ./cmd/main` (reads `.env`-equivalent
  variables from the real environment — export them or use a tool like
  `direnv`/`dotenv`). Always binds `:80` (hardcoded, not configurable) —
  binding it without root/`CAP_NET_BIND_SERVICE` will fail on most systems,
  so running the raw binary outside Docker currently needs elevated
  privileges (e.g. `sudo go run ./cmd/main`, or `sudo setcap
  'cap_net_bind_service=+ep'` on the built binary).
- Seeder: `cd backend && go run ./cmd/seeder` (run once against a fresh DB,
  after migrations).
- Migrations: use the `migrate` CLI directly against
  `database/migrations/`, e.g.
  `migrate -path backend/database/migrations -database "$DATABASE_URL" up`.
- Frontend: `cd frontend && npm run dev`.

### Verification checklist for any backend change

1. `cd backend && go test -count=1 ./... && go build ./... && go vet ./...`
   — all must pass. `gofmt -l .` and `git diff --check` must print nothing.
2. If any `@Success`/`@Param`/etc. swagger annotation changed, regenerate
   docs (see below).
3. For anything touching HTTP/MQTT behavior end-to-end, build the Docker
   image and run it against the real `.env`-configured services, then
   exercise the change with `curl`/`mosquitto_pub`. Automated tests do not
   replace `docs/agent_test/<version>/` or a live-service integration check
   for cross-boundary behavior.

### Verification checklist for any frontend change

From `frontend/`, run:

```bash
npm run typecheck
npm run lint
npm test
npm run build
```

Use `npm run test:e2e` for authentication, shell, focus, or other
browser-level behavior. A backend contract change consumed by the frontend
must be verified on both sides; mocked frontend tests alone do not prove
permission or request-shape compatibility.

### Automated-test convention

- Write tests against observable behavior and real package boundaries.
  Small test-only fakes may record repository, logger, handler, or middleware
  calls; keep them in `_test.go` files and assert the exact contract.
- For time-dependent fallback behavior, bound server time between values
  captured immediately before and after the call instead of asserting an
  exact timestamp.
- Route tests must exercise Echo registration with requests so method, path,
  and permission strings can regress independently.
- Seeder tests must decode the embedded JSON data and assert permissions and
  role grants as data; do not grep source text.
- When a production error path requires logging, assert both error propagation
  and the logger metadata that makes the failure diagnosable.

### Regenerating swagger docs

```bash
cd backend
swag init -g cmd/main/main.go -o docs/swagger --parseInternal --parseDependency
```

Run this after changing any handler's swagger annotations. The output
(`docs.go`, `swagger.json`, `swagger.yaml`) is generated — never hand-edit it.

### Adding a migration

Create a new timestamp-prefixed pair in `database/migrations/`
(`YYYYMMDDHHMMSS_description.up.sql` / `.down.sql`), matching the existing
naming convention exactly so `migrate` orders them correctly.

For append-only time-series data, mirror `telemetry_records`, `action_logs`,
and `node_logs`: identity `BIGINT`, a composite primary key containing the
time column, a 7-day hypertable chunk interval, and no update/delete audit
machinery. Compression retention/policy is a separate TimescaleDB setting
from chunk size. Down migrations must remove policies before dropping the
table/type.

### Node-log contract

- Firmware log lines have the exact shape
  `DD/MM/YYYY HH:MM:SS.mmm [LEVEL] [TAG] message`; the timestamp is UTC.
- Valid levels are exactly `NONE`, `ERROR`, `WARN`, `INFO`, and `DEBUG`.
- Any syntax mismatch, impossible timestamp, out-of-range component, or
  unsupported level uses the full fallback: `NONE`, empty tag, the entire raw
  line as the message, and server receive time in UTC. Never drop malformed
  device logs.
- Keep parsing in the application layer. Repository contracts accept only the
  structured fields and must not grow a redundant raw-line column.
- HTTP level filters must reject unsupported values as validation errors
  before PostgreSQL sees the enum.

### Cross-layer fleet contracts

- Node `device_id` is the MQTT/device identity and is immutable through the
  node update workflow. Changing it requires a deliberately designed identity
  migration, not an ordinary edit form.
- Persisted node and node-class names may contain safe underscores and
  hyphens (for example `node_<device-id>` and `base_node`). Keep create/update
  validation compatible with those stored names while rejecting path-like or
  unsafe input.
- OTA dispatch accepts a `firmware_id`. The backend—not the browser—loads the
  authoritative node and firmware, checks node-class compatibility, resolves
  the firmware binary URL, and publishes authoritative checksum/size metadata.
  The route is usable with `ota:dispatch` alone; do not add an accidental
  `firmware:get` dependency to the mutation.
- Firmware deletion requires the expected firmware name in the JSON request.
  The backend compares it with the authoritative stored name before deleting;
  the typed-name dialog is a safety interlock, not a source of truth.
- Firmware binary replacement distinguishes an omitted `config_schema`
  (preserve the current schema) from an explicit empty array (clear it).
- String node-config values are stored verbatim, including empty or
  whitespace-only strings. Validate presence separately from value content;
  numeric and boolean types retain their own validation.
- Registration ack (`/sub/<device_id>/registration_ack`) carries a real
  `{"success": bool}` body, not an empty payload. `messaging_callback`'s
  `Register` usecase always publishes exactly one ack per registration
  attempt via a `defer` (`true` only on full success; `false` on any
  earlier failure) — on the firmware side, `success:false`, a missing
  `success` field, and an unparseable payload all restart the device
  immediately (see `mate-espidf-base/AGENTS.md`'s MQTT protocol contracts).
  Keep both repos' handling of this payload shape in sync.

### RBAC seed and cache rollout

`database/seeder/*.json` is written through raw Postgres repositories, while
effective user permissions are cached through Redis in the running service.
After adding or changing seeded permissions, validate the role matrix in
tests and ensure deployment invalidates/increments the relevant permission
cache version (or restarts/clears the cache) before testing with freshly
issued authentication. A successful seeder run alone does not prove cached
authorization has refreshed.

## Validation convention

This codebase enforces a strict split between two kinds of validation:

- **Presentation layer** (`internal/presentation/http/utils/validation.go`)
  only validates *raw request shape*: is this a well-formed UUID, a parseable
  int, valid JSON, non-empty. Closed wire enums may also be rejected here
  when accepting an unknown token would otherwise reach a PostgreSQL enum and
  turn a client error into a 500 (the node-log level filter is the precedent).
  Presentation does not enforce business rules like length or charset.
- **Application layer** (`internal/application/shared/validation.go`) owns
  *value validation*: every business rule about whether a value is
  acceptable (username charset/length, password strength, snake_case
  payload-schema names, node-class name charset, a payload schema
  definition's structural shape, a device_id's MAC format, and so on) lives
  here as paired `RequiredX`/`OptionalX` functions, called from each
  usecase's `Create`/`UpdateById` as the first thing they do, before any
  repository or storage call.

When adding a new writable field anywhere in the API, follow this same
split: presentation parses/binds, application validates the value.

## Naming note

The module path, Docker image name, and most `BE_*` env defaults use
`mate-things` / `mate-things-backend`, but the swagger `@title` and several
config fallback defaults still say "Nusapala Things" / `nusapala-*` — an
earlier project name that was partially renamed. Both refer to the same
project; don't be surprised by the inconsistency, and prefer `mate-things`
in anything new.
