# L4 TCP Load Balancer

Built **stage by stage** so each step is a clean GitHub commit.

## Roadmap

| Stage | Status | What it adds |
|-------|--------|----------------|
| **1** | done | TCP listen → forward to **one** backend |
| **2** | done | Backend pool + virtual endpoint + **round-robin** |
| **3** | current | Active health monitors (remove / restore) |
| **4** | next | **Least-connections** + active conn tracking |
| **5** | planned | YAML config, load tests, polish |

## Stage 3 — active health monitors

TCP dial probes run in the background. After consecutive failures a backend is
**removed from rotation** (`IsHealthy=false`); after consecutive successes it is
**restored** — the pool list is never edited by hand.

```bash
go run ./cmd/lb \
  -listen 127.0.0.1:8080 \
  -backend 127.0.0.1:9001 \
  -backend 127.0.0.1:9002 \
  -backend 127.0.0.1:9003 \
  -health-interval 1s \
  -health-unhealthy-after 2 \
  -health-healthy-after 2
```

Or: `./scripts/demo.sh` (kills one backend and checks traffic avoids it)

### Layout (Stage 3)

```
internal/health/     active TCP probes + streak thresholds
internal/backend/    healthy flag (atomic)
internal/balancer/   RoundRobin skips unhealthy
…                    (pool, server, proxy unchanged in role)
```

### Commit this stage

```bash
git add .
git commit -m "stage3: active health monitors remove and restore backends"
git push
```

When you're ready, say **Stage 4** for least-connections.
