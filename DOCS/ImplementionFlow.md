# VanguardQ - Implementation Flow (Brief)

This is a short, step-by-step build plan. Each step lists what to do and the expected result.

---

## 1) Project skeleton

Folders: `go.mod`, `internal/`, `cmd/`

Do:

- Initialize Go module and basic folders under backend-api.
- Add base configuration loading (env, defaults).

Result:

- Repo builds with `go test ./...` even if empty.

---

## 2) Contracts and models

Folders: `internal/queue/interface.go`, `internal/worker/interface.go`, `internal/api/`, `internal/models/`

Do:

- Implement the Job struct and status enum from ProjectContacts.md.
- Add validation helpers (queue name, payload size, status transitions).

Result:

- Shared types compile and are reused across API, queue, and worker packages.

---

## 3) Redis queue adapter

Folders: `internal/queue/redis.go`, `internal/queue/operations.go`, `internal/queue/*_test.go`

Do:

- Implement the Queue interface using Redis commands from ProjectArchitecture.md.
- Add unit tests for enqueue, reserve, ack, retry, cancel.

Result:

- Queue adapter passes unit tests and matches the locked lifecycle.

---

## 3.5) Persist state on queue ops

Folders: `internal/storage/job.go`, `internal/storage/events.go`

Do:

- Persist job state changes to Postgres during queue operations (enqueue, reserve, ack/fail).

Result:

- Redis and Postgres stay aligned during early development, minimizing drift.

---

## 4) PostgreSQL storage

Folders: `internal/storage/postgres.go`, `migrations/` (optional but recommended)

Do:

- Create migrations for the final schema in ProjectArchitecture.md.
- Implement job insert/update and job_events logging.

Result:

- DB can store and query job state and history.

---

## 5) API server

Folders: `cmd/server/main.go`, `internal/api/handler.go`, `internal/api/request.go`, `internal/api/response.go`

Do:

- Implement REST endpoints: enqueue, delayed enqueue, status, cancel, queue stats.
- Validate inputs and return correct status codes.

Result:

- API accepts jobs and persists them while pushing to Redis.

---

## 6) Worker runtime

Folders: `cmd/worker/main.go`, `internal/worker/runner.go`, `internal/worker/executor.go`

Do:

- Implement worker loop: reserve, execute, ack or fail.
- Check job status before ack (handle cancelled jobs).
- Respect timeouts and retry rules.

Result:

- Jobs process end-to-end with retries and dead letter handling.

---

## 6.5) Worker control plane

Folders: `internal/worker/scaler.go`, `internal/worker/circuit.go`

Do:

- Add dynamic worker auto-scaling (2-20 per queue) based on depth.
- Add circuit breakers per job type.

Result:

- Workers scale safely under load and avoid cascading failures.

---

## 7) Schedulers and recovery

Folders: `cmd/scheduler/main.go`, `internal/queue/scheduler.go`, `internal/queue/recovery.go`

Do:

- Implement delayed scheduler, retry scheduler, processing recovery.
- Add inflight cleanup for BRPOPLPUSH fallback list.

Result:

- Delayed and retry jobs move on schedule; stuck jobs recover.

---

## 8) Observability

Folders: `internal/tracing/otel.go`, `internal/tracing/spans.go`, `internal/metrics/prometheus.go`, `internal/metrics/exporter.go`

Do:

- Add OpenTelemetry tracing and structured logs.
- Add Prometheus metrics and /metrics endpoint.

Result:

- Traces and metrics show queue depth, latency, and worker utilization.

---

## 9) Integration tests

Folders: `*_integration_test.go` (in relevant packages), `test/fixtures/`

Do:

- Add end-to-end tests for enqueue -> process -> success.
- Add retry and recovery tests with simulated worker crashes.

Result:

- Confidence that reliability and lifecycle rules are correct.

---

## 10) Packaging

Folders: `docker-compose.yml`, `.env.example`, `README.md`

Do:

- Add docker-compose for Redis and Postgres.
- Provide sample config and run instructions.

Result:

- One-command local environment for demos and development.
