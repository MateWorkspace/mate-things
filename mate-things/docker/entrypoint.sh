#!/usr/bin/env bash
set -Euo pipefail

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
