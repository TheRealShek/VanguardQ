# VanguardQ - Production Task Queue System

A **personal learning project** to understand and build a production-grade task queue system like Celery in Go.

VanguardQ enqueues async jobs, processes them with retries/timeouts, and recovers from crashes. Uses Redis (fast), PostgreSQL (durable), and Go goroutines (efficient).

## Docs

- [ProjectSpecs.md](DOCS/ProjectSpecs.md) — Architecture, components, code structure

## Quick Start

```bash
docker-compose up       # Start Redis + PostgreSQL

# Terminal 1: API Server
go run cmd/server/main.go

# Terminal 2: Worker
go run cmd/worker/main.go

# Terminal 3: Schedulers
go run cmd/scheduler/main.go
```

## Test

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"queue":"default","payload":{"action":"test"}}'
```

---
