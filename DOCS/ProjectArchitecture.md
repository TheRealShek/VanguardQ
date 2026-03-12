# VanguardQ - Locked Architecture

This document is the single source of truth for the system architecture.
All lifecycle states, Redis command choices, failure rules, and DB schema are finalized here.

---

## Job Lifecycle (Exact States)

These are the only allowed values for job `status`:

1. `queued`
2. `delayed`
3. `processing`
4. `retry_wait`
5. `success`
6. `failed`
7. `cancelled`
8. `dead`

State meanings:

- `queued`: Ready to be consumed immediately.
- `delayed`: Scheduled for a future run time.
- `processing`: Reserved by a worker and currently executing.
- `retry_wait`: Waiting for backoff delay before re-queue.
- `success`: Completed without error.
- `failed`: Terminal error but still retryable (used only if `max_retries` is 0).
- `cancelled`: Explicitly cancelled by API.
- `dead`: Moved to dead letter after exhausting retries.

Allowed transitions:

- `queued` -> `processing`
- `queued` -> `cancelled`
- `delayed` -> `queued`
- `delayed` -> `cancelled`
- `processing` -> `success`
- `processing` -> `retry_wait`
- `processing` -> `dead`
- `processing` -> `failed`
- `processing` -> `cancelled`
- `retry_wait` -> `queued`

No other transitions are valid.

---

## Redis Keys and Data Model

Key prefixes:

- `queue:{name}` (list) - active queue
- `processing:{name}` (sorted set) - in-flight reservations with score = deadline unix timestamp
- `delayed` (sorted set) - scheduled jobs with score = run_at unix timestamp
- `retry` (sorted set) - retry schedule with score = run_at unix timestamp
- `job:{id}` (hash) - job metadata
- `unique:{fingerprint}` (string) - uniqueness lock (optional)
- `dead:{name}` (list) - dead letter queue per priority

`job:{id}` fields (required):

- `id`
- `correlation_id`
- `queue`
- `payload`
- `status`
- `created_at`
- `scheduled_at`
- `max_retries`
- `retries`
- `timeout_ms`

---

## Redis Commands Per Step

### Enqueue (immediate)

1. `HSET job:{id} ... status queued ...`
2. `LPUSH queue:{name} {id}`

### Enqueue (delayed)

1. `HSET job:{id} ... status delayed ... scheduled_at {ts} ...`
2. `ZADD delayed {ts} {id}`

### Fetch (reserve for processing)

Worker command sequence:

1. `BRPOPLPUSH queue:{name} queue:{name}:inflight 5`
2. `ZADD processing:{name} {deadline_ts} {id}`
3. `LREM queue:{name}:inflight 1 {id}`

Notes:

- `queue:{name}:inflight` is an ephemeral list used only to make BRPOPLPUSH atomic.
- If step 2 fails, re-push to `queue:{name}`.

### Ack (success)

1. `ZREM processing:{name} {id}`
2. `HSET job:{id} status success`

### Fail with retry

1. `HINCRBY job:{id} retries 1`
2. `HSET job:{id} status retry_wait`
3. `ZREM processing:{name} {id}`
4. `ZADD retry {run_at_ts} {id}`

### Fail without retry (terminal)

1. `ZREM processing:{name} {id}`
2. `HSET job:{id} status failed`

### Move to dead letter

1. `ZREM processing:{name} {id}`
2. `HSET job:{id} status dead`
3. `LPUSH dead:{name} {id}`

### Cancel job

1. `HSET job:{id} status cancelled`
2. `LREM queue:{name} 0 {id}`
3. `ZREM delayed {id}`
4. `ZREM retry {id}`
5. `ZREM processing:{name} {id}`

---

## Scheduler Loops

### Delayed Scheduler (every 5s)

1. `ZRANGEBYSCORE delayed 0 {now_ts}`
2. For each id:
   - `ZREM delayed {id}`
   - `HSET job:{id} status queued`
   - `LPUSH queue:{queue} {id}`

### Retry Scheduler (every 5s)

1. `ZRANGEBYSCORE retry 0 {now_ts}`
2. For each id:
   - `ZREM retry {id}`
   - `HSET job:{id} status queued`
   - `LPUSH queue:{queue} {id}`

### Processing Recovery (every 60s)

1. `ZRANGEBYSCORE processing:{name} 0 {now_ts}`
2. For each id:
   - `ZREM processing:{name} {id}`
   - `HSET job:{id} status queued`
   - `LPUSH queue:{name} {id}`

---

## Failure Cases and Recovery Rules

### Worker crash after fetch

- Job remains in `processing:{name}` with a deadline.
- Recovery loop re-queues it when deadline passes.
- Result: at-least-once delivery.

### Worker crash before `ZADD processing`

- Job may be stuck in `queue:{name}:inflight`.
- Cleanup task runs every 60s:
  - `LRANGE queue:{name}:inflight 0 -1`
  - For each id: `LPUSH queue:{name} {id}` then `LREM queue:{name}:inflight 1 {id}`

### Redis restart

- Job metadata lives in Redis, but PostgreSQL is the source of truth for job history.
- On Redis recovery, a bootstrap task re-enqueues jobs in DB with status in (`queued`, `delayed`, `retry_wait`, `processing`).

### PostgreSQL outage

- API returns 503 on enqueue if DB insert fails.
- Workers continue processing jobs already in Redis but do not update status; a later reconciliation task updates DB from Redis job hashes.

### Duplicate job submission

- Use `SETNX unique:{fingerprint} {id} EX {ttl}`.
- If SETNX fails, the API returns conflict and does not enqueue.

### Cancel during processing

- Cancellation sets `status=cancelled` but does not kill the running worker.
- Worker checks status before completion; if `cancelled`, it skips success update and still removes from `processing`.

### Retry exhaustion

- If `retries >= max_retries`, job transitions to `dead` and is pushed to `dead:{name}`.

---

## Final Database Schema (No Changes Later)

PostgreSQL is the durable history store. This schema is final.

```sql,
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

Status values must match the Job Lifecycle list exactly.

---

## Anything Else (Locked Decisions)

- At-least-once delivery is required; exactly-once is out of scope.
- Worker timeouts are enforced by context deadlines and `processing:{name}` deadlines.
- Backoff strategy is exponential: base 1s, doubling each retry, capped at 1m.
- API must validate queue names: `high`, `default`, `low` only.
- Max payload size: 256 KB (reject larger jobs at enqueue).
- Correlation ID is mandatory and generated if missing.
