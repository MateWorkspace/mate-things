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
- **Telemetry ingestion** — devices publish telemetry records that the API
  exposes for querying.
- **RBAC** — permissions, roles, and role↔permission assignments gate every
  admin-facing endpoint; users authenticate via JWT access/refresh tokens.
- **Preferences** — arbitrary per-resource key/value preferences (e.g. UI
  state) attached to any entity.

The backend is the product's core; the frontend is currently an
unstarted Next.js scaffold (see below).

## Project structure

```
mate-things/
├── backend/           Go API + MQTT service (see below)
├── frontend/           Next.js admin dashboard — scaffold only, not yet built out
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
  - `models/` — plain domain structs (e.g. `Node`, `Firmware`, `PayloadSchema`).
  - `contracts/` — interfaces the application layer depends on and
    infrastructure implements: `repository/` (Postgres access per entity),
    `cache/`, `storage/` (firmware binaries via MinIO), `node/` (MQTT
    publish/subscribe), `utility/` (password hashing, tokens, payload
    schema validation), `logger/`.
  - `usecases/` — one interface per feature area (`admin`, `node`, `action`,
    `auth`, `profile`, `preferences`, `telemetry`, `seeder`, and an internal
    `repocache` interface set used by the cache-decorator layer) plus their
    request/response DTOs.
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
    device registration, status, action acks, and logs.
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
- **`database/migrations/`** — golang-migrate SQL migrations, timestamp-prefixed.
- **`database/seeder/`** — baseline RBAC + example node class/action/user
  seed data as JSON, embedded into the seeder binary via `go:embed`.
- **`docs/swagger/`** — generated by `swag init` (see
  [Regenerating swagger docs](#regenerating-swagger-docs)); never hand-edit.
- **`docs/agent_test/`** — a manual/scripted test checklist per release
  (`checklist.md` + one `scenario/NN-*.md` file per feature area), used to
  verify a build end-to-end against a real Postgres/Redis/MinIO/MQTT
  environment before release.
- **`pkg/pgxdt/`** — a small standalone pgx transaction-context helper,
  importable outside this module if ever needed.

### Frontend (`frontend/`)

A default `create-next-app` scaffold (Next.js 16, App Router, React 19) —
`src/app/page.tsx` is still the stock starter page. `src/core/` mirrors the
backend's layering (`application/domain/infrastructure/composition`) as a
placeholder for when the admin dashboard is actually built, and
`src/shared/` holds cross-cutting `components/hooks/lib`. Nothing here talks
to the backend API yet.

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
3. starts the backend (binds the container's public port directly —
   `BE_HTTP_SERVER_ADDRESS` is forced to `:80`) and the Next.js server,
   under one process group.

The backend is the sole ingress: it serves `/api/*` itself, reverse-proxies
everything else to Next.js, and reverse-proxies `/minio-proxy/*` to MinIO
for presigned downloads. There is no separate proxy process.

### Running locally without Docker

- Backend: `cd backend && go run ./cmd/main` (reads `.env`-equivalent
  variables from the real environment — export them or use a tool like
  `direnv`/`dotenv`). Defaults to listening on `:8080` when
  `BE_HTTP_SERVER_ADDRESS` isn't set.
- Seeder: `cd backend && go run ./cmd/seeder` (run once against a fresh DB,
  after migrations).
- Migrations: use the `migrate` CLI directly against
  `database/migrations/`, e.g.
  `migrate -path backend/database/migrations -database "$DATABASE_URL" up`.
- Frontend: `cd frontend && npm run dev`.

### Verification checklist for any backend change

1. `cd backend && go build ./... && go vet ./... && gofmt -l .` — must be
   clean before anything else.
2. If any `@Success`/`@Param`/etc. swagger annotation changed, regenerate
   docs (see below).
3. For anything touching HTTP/MQTT behavior end-to-end, build the Docker
   image and run it against the real `.env`-configured services, then
   exercise the change with `curl`/`mosquitto_pub` — unit tests don't exist
   in this codebase; `docs/agent_test/<version>/` is the closest thing to a
   test suite and is meant to be run manually/scripted against a live
   container.

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

## Validation convention

This codebase enforces a strict split between two kinds of validation:

- **Presentation layer** (`internal/presentation/http/utils/validation.go`)
  only validates *raw request shape*: is this a well-formed UUID, a parseable
  int, valid JSON, non-empty. It never enforces business rules like length,
  charset, or format.
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
