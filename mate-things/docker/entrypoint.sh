#!/usr/bin/env bash
set -Euo pipefail

urlencode() {
    jq -rn --arg v "$1" '$v|@uri'
}

db_host="${BE_POSTGRES_HOST:-127.0.0.1}"
db_port="${BE_POSTGRES_PORT:-5432}"
db_user="${BE_POSTGRES_USERNAME:-postgres}"
db_pass="${BE_POSTGRES_PASSWORD:-postgres}"
db_name="${BE_POSTGRES_DATABASE:-nusapala_things}"
db_sslmode="${BE_POSTGRES_SSL_MODE:-disable}"
db_url="postgres://$(urlencode "$db_user"):$(urlencode "$db_pass")@${db_host}:${db_port}/${db_name}?sslmode=${db_sslmode}"

echo "[entrypoint] running database migrations"
if ! migrate -path /app/migrations -database "$db_url" up; then
    echo "[entrypoint] migrations failed" >&2
    exit 1
fi

pids=()

terminate() {
    trap - TERM INT
    code="${1:-0}"
    for pid in "${pids[@]:-}"; do
        kill -TERM "$pid" 2>/dev/null || true
    done
    wait 2>/dev/null || true
    exit "$code"
}
trap 'terminate 0' TERM INT

/app/backend/backend &
pids+=("$!")

(cd /app/frontend && PORT=3000 HOSTNAME=0.0.0.0 node server.js) &
pids+=("$!")

nginx -g 'daemon off;' &
pids+=("$!")

wait -n
exit_code=$?
terminate "$exit_code"
