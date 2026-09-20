# L4 TCP Load Balancer

Built **stage by stage** so each step is a clean GitHub commit.

## Roadmap

| Stage | Status | What it adds |
|-------|--------|----------------|
| **1** | done | TCP listen → forward to **one** backend |
| **2** | current | Backend pool + virtual endpoint + **round-robin** |
| **3** | next | Active health monitors (remove / restore) |
| **4** | planned | **Least-connections** + active conn tracking |
| **5** | planned | YAML config, load tests, polish |

## Stage 2 — pool + round-robin

One virtual listen address, multiple backends, connections distributed in order.

```bash
# Backends
go run ./cmd/backend -name backend-1 -addr 127.0.0.1:9001 -mode http
go run ./cmd/backend -name backend-2 -addr 127.0.0.1:9002 -mode http
go run ./cmd/backend -name backend-3 -addr 127.0.0.1:9003 -mode http

# Load balancer
go run ./cmd/lb \
  -listen 127.0.0.1:8080 \
  -backend 127.0.0.1:9001 \
  -backend 127.0.0.1:9002 \
  -backend 127.0.0.1:9003

# Clients — expect backend-1, backend-2, backend-3, …
curl http://127.0.0.1:8080/
```

Or: `./scripts/demo.sh`

### Layout (Stage 2)

```
cmd/lb/main.go              wire listen + backends → pool + listener
cmd/backend/main.go         demo upstream
internal/backend/           upstream identity
internal/balancer/          RoundRobin
internal/pool/              backend pool
internal/server/            virtual listen endpoint
internal/proxy/             bidirectional TCP forward (Stage 1)
scripts/demo.sh
```

### Commit this stage

```bash
git add .
git commit -m "stage2: backend pool with round-robin"
git push
```

When you're ready, say **Stage 3** for active health monitors.
