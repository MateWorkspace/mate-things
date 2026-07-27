# mate-things

An IoT fleet-management platform for ESP32-class devices: device
registration over MQTT, firmware OTA, remote action dispatch, telemetry
ingestion, and RBAC-gated admin APIs.

- **Backend** — Go (Echo + MQTT), Postgres, Redis, MinIO. Fully built out.
- **Frontend** — Next.js admin dashboard. Scaffold only, not yet built.

## Prerequisites

You need existing, reachable instances of:

- PostgreSQL
- Redis
- MinIO (or any S3-compatible object storage)
- An MQTT broker (e.g. HiveMQ Cloud)

This repo doesn't run any of these for you — only the app itself.

## Quick start

```bash
cp .env.example .env
# fill in your Postgres / Redis / MinIO / MQTT credentials and BE_BASE_URL

docker compose up --build
```

This builds a single image and runs one container that, on start:

1. runs database migrations,
2. seeds baseline RBAC data (roles, permissions) and an example
   node class/action/user (**default login: `admin` / `ChangeMe123!` —
   change it immediately**),
3. serves the API, MQTT handlers, and the frontend behind one port.

Once running:

- API: `http://localhost:8080/api/v1`
- Swagger docs: `http://localhost:8080/api/docs`
- Frontend: `http://localhost:8080/`

## Running without Docker

```bash
# backend
cd backend && go run ./cmd/main

# seeder (once, after migrations)
cd backend && go run ./cmd/seeder

# frontend
cd frontend && npm run dev
```

## More

See [`AGENTS.md`](./AGENTS.md) for project structure, architectural
conventions, and the full development workflow.
