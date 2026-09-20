#!/usr/bin/env bash
# Stage 2 demo: three backends + round-robin through :8080.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PIDS=()
cleanup() {
  for pid in "${PIDS[@]:-}"; do
    kill "$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT

go build -o /tmp/l4-backend ./cmd/backend
go build -o /tmp/l4-lb ./cmd/lb

/tmp/l4-backend -name backend-1 -addr 127.0.0.1:9001 -mode http &
PIDS+=($!)
/tmp/l4-backend -name backend-2 -addr 127.0.0.1:9002 -mode http &
PIDS+=($!)
/tmp/l4-backend -name backend-3 -addr 127.0.0.1:9003 -mode http &
PIDS+=($!)
sleep 0.2

/tmp/l4-lb \
  -listen 127.0.0.1:8080 \
  -backend 127.0.0.1:9001 \
  -backend 127.0.0.1:9002 \
  -backend 127.0.0.1:9003 &
PIDS+=($!)

for _ in $(seq 1 50); do
  if curl -sf "http://127.0.0.1:8080/" >/dev/null 2>&1; then
    break
  fi
  sleep 0.1
done

echo "== Stage 2: round-robin across 3 backends =="
for i in $(seq 1 6); do
  curl -sf "http://127.0.0.1:8080/"
done
echo "OK"
