# L4 TCP Load Balancer

Built **stage by stage** so each step is a clean GitHub commit.

## Roadmap

| Stage | Status | What it adds |
|-------|--------|----------------|
| **1** | current | TCP listen → forward every connection to **one** backend |
| **2** | next | Backend pool + virtual endpoint + **round-robin** |
| **3** | planned | Active health monitors (remove / restore) |
| **4** | planned | **Least-connections** + active conn tracking |
| **5** | planned | YAML config, load tests, polish |

## Stage 1 — single-backend TCP proxy

Prove L4 basics: accept TCP clients, dial one upstream, splice bytes both ways with **goroutine-per-connection**.

```bash
# Terminal 1 — backend
go run ./cmd/backend -name backend-1 -addr 127.0.0.1:9001 -mode http

# Terminal 2 — proxy
go run ./cmd/lb -listen 127.0.0.1:8080 -backend 127.0.0.1:9001

# Terminal 3 — client
curl http://127.0.0.1:8080/
# → backend-1
```

Or: `./scripts/demo.sh`

### Layout (Stage 1)

```
cmd/lb/main.go          accept loop + flags
cmd/backend/main.go     demo upstream
internal/proxy/         bidirectional TCP forward
scripts/demo.sh         one-shot smoke demo
```

### Commit this stage

```bash
cd ~/Projects/l4-load-balancer
git init -b main   # if not already
git add .
git commit -m "stage1: TCP proxy to a single backend"
```

When you're ready, say **Stage 2** and we'll add pools + round-robin.
