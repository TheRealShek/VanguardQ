# VanguardQ - Production Task Queue System

---

## Example: Celery in Action at Instagram

When you upload a photo:

- Web request saves the raw image and returns immediately (you see "uploading...")
- Celery workers handle in the background:
  - Generate multiple thumbnail sizes
  - Apply filters/optimizations
  - Extract metadata
  - Run ML models for content moderation
  - Trigger push notifications to followers
  - Update recommendation feeds

## Without Celery, you'd wait 5-10 seconds staring at a loading screen. With it, upload feels instant.

## Stack

- **Go** (goroutines, channels, context)
- **Redis** (job queue, sorted sets)
- **PostgreSQL** (job history, results)
- **OpenTelemetry** (distributed tracing)
- **Prometheus** (metrics)
- **Docker Compose** (local dev)

---

## Code Structure

```
vanguardq/
├── cmd/
│   ├── server/
│   │   └── main.go              # Start REST API server on :8080
│   │                             # Enqueue, Status, Cancel endpoints
│   │
│   ├── worker/
│   │   └── main.go              # Start worker pool
│   │                             # BRPOPLPUSH reserve, execute, ack/fail
│   │                             # Auto-scaling, circuit breakers
│   │
│   └── scheduler/
│       └── main.go              # Start delayed & retry schedulers
│                                 # Runs every 5s (delayed, retry)
│                                 # Runs every 60s (recovery)
│
├── internal/
│   ├── api/
│   │   ├── handler.go           # HTTP handlers for 5 endpoints
│   │   ├── request.go           # Request DTOs (queue, payload, etc)
│   │   └── response.go          # Response envelopes (202, 200, 404)
│   │
│   ├── queue/
│   │   ├── interface.go         # Queue interface (from ProjectContacts.md)
│   │   ├── redis.go             # Redis impl: LPUSH, BRPOPLPUSH, ZADD, etc
│   │   └── operations.go        # Helpers: enqueue, reserve, ack, fail, retry
│   │
│   ├── storage/
│   │   ├── postgres.go          # DB connection & migrations
│   │   ├── job.go               # Job insert/update queries
│   │   └── events.go            # Log job events (queued, processing, etc)
│   │
│   ├── worker/
│   │   ├── interface.go         # Worker interface (from ProjectContacts.md)
│   │   ├── runner.go            # Main loop: reserve, execute, ack/fail
│   │   ├── executor.go          # Call user-defined handlers
│   │   ├── scaler.go            # Auto-scale 2-20 workers per queue
│   │   └── circuit.go           # Circuit breaker per job type
│   │
│   ├── metrics/
│   │   ├── prometheus.go        # Register gauges, histograms, counters
│   │   └── exporter.go          # Expose /metrics endpoint
│   │
│   └── tracing/
│       ├── otel.go              # Initialize OpenTelemetry
│       └── spans.go             # Create spans with correlation ID
│
├── docker-compose.yml           # Redis + PostgreSQL local dev
├── go.mod                        # Module dependencies
└── README.md                     # Setup, run, API examples
```

---

## Architecture

```
Producer API (REST/gRPC)
    ↓
Redis Queue (LPUSH/BRPOP)
    ↓
Worker Pool (goroutines)
    ↓
PostgreSQL (job results)
```

---

## Core Components

### 1. Producer API

**Endpoints:**

```
POST   /jobs              - Enqueue job
POST   /jobs/delayed      - Schedule for later
GET    /jobs/:id          - Job status
DELETE /jobs/:id          - Cancel job
GET    /queues/:name/stats - Queue metrics
```

**Job struct (authoritative in ProjectContacts.md):**

```go
type Job struct {
    ID            string
    CorrelationID string
    Queue         string
    Payload       json.RawMessage

    Status      string
    CreatedAt   time.Time
    ScheduledAt time.Time
    StartedAt   time.Time
    CompletedAt time.Time

    TimeoutMs  int
    Retries    int
    MaxRetries int

    Error string
}
```

### 2. Redis Schema

**Active queues (lists):**

```
queue:high       - LPUSH/BRPOP
queue:default
queue:low
```

**Processing set:**

```
processing:{queue}  - ZADD with timeout timestamp
```

**Delayed jobs (sorted set):**

```
delayed  - ZADD with scheduled timestamp
```

**Job metadata (hash):**

```
job:{id}  - payload, retries, timeout, etc.
```

### 3. Worker Pool

**Per queue:**

- Dynamic concurrency (auto-scale 2-20 workers based on queue depth)
- BRPOPLPUSH for reliability
- Heartbeat every 30s
- Graceful shutdown (finish current jobs)
- Circuit breakers per job type

**Scaling logic:**

```go
// Monitor every 30s
if queueDepth > 100 && currentWorkers < maxWorkers {
    spawnWorker()
}
if queueDepth < 10 && currentWorkers > minWorkers {
    shutdownWorker()
}
```

**Recovery:**

- Check `processing:*` sets every 60s
- Re-queue jobs past timeout

### 4. PostgreSQL Schema

The authoritative schema is in ProjectArchitecture.md. This is aligned to it.

```sql
CREATE TABLE jobs (
    id              VARCHAR PRIMARY KEY,
    correlation_id  VARCHAR NOT NULL,
    queue           VARCHAR NOT NULL,
    payload         JSONB NOT NULL,
    status          VARCHAR NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    scheduled_at    TIMESTAMP,
    started_at      TIMESTAMP,
    completed_at    TIMESTAMP,
    error           TEXT,
    retries         INT NOT NULL DEFAULT 0,
    max_retries     INT NOT NULL DEFAULT 0,
    timeout_ms      INT NOT NULL DEFAULT 0
);

CREATE TABLE job_events (
    id          BIGSERIAL PRIMARY KEY,
    job_id      VARCHAR NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    event       VARCHAR NOT NULL,
    detail      TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_queue_status ON jobs(queue, status);
CREATE INDEX idx_jobs_created_at ON jobs(created_at DESC);
CREATE INDEX idx_jobs_correlation_id ON jobs(correlation_id);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_scheduled_at ON jobs(scheduled_at);
CREATE INDEX idx_job_events_job_id ON job_events(job_id);
CREATE INDEX idx_job_events_created_at ON job_events(created_at DESC);
```

---

## Features

### Job Management

- [x] Multiple priority queues
- [x] Delayed execution (cron-like)
- [x] Exponential backoff retry (1s, 2s, 4s, 8s...)
- [x] Job cancellation
- [x] Unique jobs (fingerprint with Redis SETNX)

### Worker Pool

- [x] Dynamic concurrency per queue (auto-scale 2-20 workers)
- [x] Graceful shutdown (context cancellation)
- [x] Worker heartbeat monitoring
- [x] Auto-scale based on queue depth
- [x] Circuit breakers per job type

### Reliability

- [x] Job timeout handling
- [x] Dead letter queue (after max retries)
- [x] At-least-once delivery
- [x] Job result storage (success/failure logs)

### Observability

- [x] OpenTelemetry distributed tracing
- [x] Structured logging with correlation IDs
- [x] Queue depth gauge
- [x] Job latency histogram (enqueue → complete)
- [x] Worker utilization
- [x] Failed job counter
- [x] Prometheus metrics export
- [x] Admin API (GET /metrics, GET /queues)

---

## Implementation Details

### Job Enqueue Flow

1. Generate job ID (UUID)
2. Store metadata in `job:{id}` hash
3. If delayed: ZADD to `delayed` sorted set
4. Else: LPUSH to `queue:{name}`
5. Insert to PostgreSQL with status=queued

### Worker Loop

```go
for {
    select {
    case <-ctx.Done():
        // Graceful shutdown
        return
    default:
        // BRPOPLPUSH queue:default processing:default 5s
        jobID := redis.BRPopLPush(queue, processing, 5*time.Second)
        if jobID == "" {
            continue
        }
        go processJob(jobID)
    }
}
```

### Job Processing

1. Load metadata from `job:{id}`
2. Start OpenTelemetry span with correlation ID
3. Update PostgreSQL: status=processing, started_at=now
4. Execute payload (with timeout context)
5. Log with structured fields (job_id, correlation_id, queue, duration)
6. If success:
   - Remove from `processing:{queue}`
   - Update PostgreSQL: status=success, completed_at=now
   - End span with success status
7. If failure:
   - Increment retry counter
   - If retries < max: re-queue with backoff delay
   - Else: move to dead letter queue
   - Update PostgreSQL: status=failed, error=msg
   - End span with error status

**Structured log example:**

```go
logger.Info("job completed",
    zap.String("job_id", job.ID),
    zap.String("correlation_id", job.CorrelationID),
    zap.String("queue", job.Queue),
    zap.Duration("duration", time.Since(start)),
    zap.String("status", "success"),
)
```

### Delayed Job Scheduler

```go
// Runs every 5 seconds
jobs := redis.ZRangeByScore("delayed", 0, now.Unix())
for _, jobID := range jobs {
    redis.LPush("queue:"+job.Queue, jobID)
    redis.ZRem("delayed", jobID)
}
```

### Job Uniqueness

```go
fingerprint := sha256(queue + payload)
if !redis.SetNX("unique:"+fingerprint, jobID, ttl) {
    return ErrDuplicateJob
}
```

### Worker Recovery

```go
// Runs every 60 seconds
jobs := redis.ZRangeByScore("processing:default", 0, now-timeout)
for _, jobID := range jobs {
    redis.LPush("queue:default", jobID)
    redis.ZRem("processing:default", jobID)
}
```

---

## Metrics to Showcase

- "Handles **50k jobs/hour** with 10 workers"
- "**99.99% job delivery** with 3 retry attempts"
- "**Sub-10ms** job enqueue latency"
- "Recovers jobs within **30s** of worker crash"
- "**95% automatic retry success** rate"

---

## Build Timeline (2-3 weeks)

### Week 1: Core

- [ ] Producer API (enqueue, status, cancel)
- [ ] Redis queue (LPUSH/BRPOP)
- [ ] Basic worker pool (single queue)
- [ ] PostgreSQL job storage

### Week 2: Reliability

- [ ] Retry logic with exponential backoff
- [ ] Delayed jobs (sorted set + scheduler)
- [ ] Multiple priority queues
- [ ] Job timeout + recovery

### Week 3: Production

- [ ] Unique jobs (fingerprinting)
- [ ] Dead letter queue
- [ ] OpenTelemetry distributed tracing
- [ ] Structured logging with correlation IDs
- [ ] Dynamic worker auto-scaling
- [ ] Circuit breakers per job type
- [ ] Prometheus metrics
- [ ] Admin API (stats, queue health)
- [ ] Docker Compose setup
- [ ] README with architecture diagram

---

## Interview Talking Points

**Q: How do you prevent job loss if worker crashes?**
A: BRPOPLPUSH atomically moves job to processing set with timeout. Recovery goroutine re-queues stale jobs every 60s.

**Q: How do you handle duplicate jobs?**
A: Job fingerprinting (SHA-256 of queue+payload) with Redis SETNX + TTL. Returns error if duplicate within window.

**Q: How do you scale workers?**
A: Monitor queue depth with Prometheus. If depth > threshold for 5min, spawn workers up to max. Scale down on idle.

**Q: How do you trace job execution across systems?**
A: Generate correlation ID on enqueue. Propagate through Redis → Worker → PostgreSQL. OpenTelemetry spans link API call → job execution → completion.

**Q: Why Redis + PostgreSQL?**
A: Redis for fast queue ops (<1ms), PostgreSQL for durable history + analytics. Decouple ephemeral state from permanent records.

**Q: How do you ensure at-least-once delivery?**
A: Jobs stay in processing set until explicitly removed. Timeout recovery re-queues. Retries with exponential backoff.

**Q: What happens if Redis crashes?**
A: Jobs in PostgreSQL with status=queued/processing can be re-enqueued on Redis recovery. Accept brief unavailability over data loss.

**Q: What are circuit breakers for?**
A: If job type X fails 10 times in 5min, stop processing that type temporarily. Prevents cascading failures from bad jobs flooding workers.

---

## Testing Strategy

**Unit tests:**

- Queue operations (enqueue, dequeue, cancel)
- Worker logic (process, retry, timeout)
- Job fingerprinting

**Integration tests:**

- End-to-end job flow (enqueue → process → success)
- Retry logic with mock failures
- Worker crash recovery

**Load tests:**

- Enqueue 10k jobs, measure latency
- 50+ concurrent workers, measure throughput
- Simulate worker crashes during processing

---

## Extensions (After MVP)

- [ ] Job chaining (run job B after job A)
- [ ] Batch jobs (process 100 items together)
- [ ] Job priority within queue (sorted set instead of list)
- [ ] Web UI for job monitoring
- [ ] gRPC API (in addition to REST)
- [ ] Multi-tenancy (namespace jobs by customer)

---

## Why This Crushes Interviews

1. **Production patterns**: Retry, timeout, recovery, observability, circuit breakers
2. **Distributed systems**: At-least-once delivery, idempotency, crash recovery, distributed tracing
3. **Go concurrency**: Goroutines, channels, context cancellation, dynamic worker pools
4. **System design**: Redis vs PostgreSQL tradeoffs, queue depth monitoring, auto-scaling
5. **Observability**: OpenTelemetry, structured logging, correlation IDs across systems
6. **Relatable**: Every company uses task queues

This shows you can build **real backend infrastructure**, not just CRUD APIs.
