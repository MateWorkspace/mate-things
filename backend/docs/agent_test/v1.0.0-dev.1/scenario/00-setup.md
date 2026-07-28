# 00 — Environment Setup

## Purpose

Get one running backend container, reachable at `http://127.0.0.1:18080`,
pointed at the real external Postgres/Redis/MinIO/MQTT broker described in
`mate-things/.env`, so every other scenario file can just assume the server
is up.

This backend has **no test/staging isolation** — the external Postgres,
Redis, MinIO bucket, and MQTT broker in `.env` are shared dev resources. Data
created during this test pass (nodes, permissions, users, etc.) will persist
there. That's expected for a dev environment, but be aware manual cleanup of
test-created rows may be needed afterward, and running this suite twice
back-to-back will hit unique-constraint conflicts on anything not designed
to be idempotent (most creates aren't — that's fine, expected 409s are noted
per test case).

## Tools required (all confirmed present on this machine)

- `docker`
- `curl`
- `mosquitto_pub` / `mosquitto_sub` (for the MQTT scenarios, `16-mqtt-flows.md`)
- `jq` (optional, for parsing JSON responses in shell — install if missing, or read responses by eye)

## Build + run

From `mate-things/` repo root:

```bash
docker build -t mate-things:test -f Dockerfile .
docker run -d --name mate-things-test --env-file .env -p 18080:80 mate-things:test
```

Wait for it to come up, then tail logs to confirm migrations + seeder + all
three processes started cleanly:

```bash
sleep 5
docker logs mate-things-test 2>&1 | tail -40
```

Expected: migration list, seeder "Seeding completed successfully" (or
"already exists, skipping" lines if this isn't the first run against this
DB), then `http(s) server started` and Next.js `✓ Ready`.

Base URL for every HTTP scenario below: `http://127.0.0.1:18080/api`.

## Getting an access token

Nearly every scenario after auth needs a bearer token. The seeder creates
one bootstrap user: username `admin`, password `ChangeMe123!`, role `super`
(full permissions). Log in once and export the token for reuse:

```bash
LOGIN_RESPONSE=$(curl -s -X POST http://127.0.0.1:18080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"ChangeMe123!"}')

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .access_token)
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .refresh_token)
echo "$ACCESS_TOKEN"
```

Note `BE_TOKEN_ACCESS_DURATION=2m` in `.env` — the access token is only
valid for **2 minutes**. Scenario `01-auth.md` uses this short lifetime
directly for the expiry test instead of needing a special override.

For permission-boundary testing (`02-rbac-authorization.md`), also create a
second user with the low-privilege default `user` role via the `admin`
token, and log in as that user to get a second, restricted token
(`USER_TOKEN` in later scenarios).

## Cleanup

```bash
docker rm -f mate-things-test
```

This only removes the local container — it does **not** clean up rows
created in the external Postgres/MinIO/MQTT resources. If a clean slate is
needed for a full re-run, truncate the relevant tables manually or restore
from a backup before re-running migrations+seeder.
