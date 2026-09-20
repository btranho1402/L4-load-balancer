#!/usr/bin/env bash
# Stage 3 demo: RR across 3 backends, then kill one and show it leaves rotation.
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
B2_PID=$!
PIDS+=($B2_PID)
/tmp/l4-backend -name backend-3 -addr 127.0.0.1:9003 -mode http &
PIDS+=($!)
sleep 0.2

/tmp/l4-lb \
  -listen 127.0.0.1:8080 \
  -backend 127.0.0.1:9001 \
  -backend 127.0.0.1:9002 \
  -backend 127.0.0.1:9003 \
  -health-interval 200ms \
  -health-timeout 100ms \
  -health-unhealthy-after 1 \
  -health-healthy-after 1 &
PIDS+=($!)

for _ in $(seq 1 50); do
  if curl -sf "http://127.0.0.1:8080/" >/dev/null 2>&1; then
    break
  fi
  sleep 0.1
done

echo "== Before failover (expect all three) =="
for i in $(seq 1 6); do
  curl -sf "http://127.0.0.1:8080/"
done

echo "== Killing backend-2 (9002) =="
kill "$B2_PID" 2>/dev/null || true
# Drop from PIDS cleanup list so we don't double-kill
PIDS=("${PIDS[@]/$B2_PID}")
sleep 0.6

echo "== After failover (must not see backend-2) =="
for i in $(seq 1 8); do
  out=$(curl -sf "http://127.0.0.1:8080/")
  echo -n "$out"
  if [[ "$out" == *backend-2* ]]; then
    echo "FAIL: traffic still hit backend-2"
    exit 1
  fi
done
echo "OK"
