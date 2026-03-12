# VanguardQ - Contracts (Locked)

This document defines the final contracts for job data, worker behavior, and the queue abstraction.
These contracts are locked and must not change without explicit approval.

---

## Job Contract (Final Fields)

Status values must match the lifecycle in ProjectArchitecture.md.

```go
// JobStatus is the only allowed job status set.
type JobStatus string

const (
	JobQueued     JobStatus = "queued"
	JobDelayed    JobStatus = "delayed"
	JobProcessing JobStatus = "processing"
	JobRetryWait  JobStatus = "retry_wait"
	JobSuccess    JobStatus = "success"
	JobFailed     JobStatus = "failed"
	JobCancelled  JobStatus = "cancelled"
	JobDead       JobStatus = "dead"
)

// Job is the canonical representation used by API, queue, and worker layers.
type Job struct {
	ID            string          `json:"id"`
	CorrelationID string          `json:"correlation_id"`
	Queue         string          `json:"queue"`
	Payload       json.RawMessage `json:"payload"`

	Status      JobStatus `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ScheduledAt time.Time `json:"scheduled_at"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`

	TimeoutMs  int `json:"timeout_ms"`
	Retries    int `json:"retries"`
	MaxRetries int `json:"max_retries"`

	Error string `json:"error"`
}
```

Field notes:

- `ScheduledAt` is zero value for immediate jobs.
- `StartedAt` and `CompletedAt` are zero value until set.
- `TimeoutMs` is required; 0 means no timeout.
- `Error` is set only for `failed` or `dead`.

---

## Worker Contract

Workers execute a single job and return a typed result.

```go
// Worker executes a job payload.
// It must be idempotent because jobs can be re-delivered.
type Worker interface {
	// Name returns a stable worker name for logging and metrics.
	Name() string

	// Execute runs the job. Returning a RetryableError triggers retry.
	Execute(ctx context.Context, job Job) error
}

// RetryableError indicates a transient failure.
type RetryableError interface {
	error
	Retryable() bool
}
```

Rules:

- Any non-nil error that is not `RetryableError` is terminal.
- Workers must respect `ctx` cancellation and deadlines.
- Workers must not mutate the job payload.

---

## Queue Contract (Redis Abstraction)

The queue interface shields Redis details and is used by API, scheduler, and workers.

[ API ] \
[ Scheduler ] ---> Queue interface ---> Redis (hidden)
[ Worker ] /

- Everything talks to Queue, not Redis.

```go
// Queue provides atomic job operations backed by Redis.
type Queue interface {
	// Enqueue pushes an immediate job to its queue.
	Enqueue(ctx context.Context, job Job) error

	// EnqueueDelayed schedules a job at a specific time.
	EnqueueDelayed(ctx context.Context, job Job, runAt time.Time) error

	// Reserve blocks and reserves the next job for processing.
	Reserve(ctx context.Context, queue string, block time.Duration) (Job, error)

	// Ack marks a job as successfully completed.
	Ack(ctx context.Context, job Job) error

	// FailWithRetry moves a job into retry wait state.
	FailWithRetry(ctx context.Context, job Job, runAt time.Time, errMsg string) error

	// FailTerminal marks a job as failed without retry.
	FailTerminal(ctx context.Context, job Job, errMsg string) error

	// MoveToDead pushes a job to the dead letter queue.
	MoveToDead(ctx context.Context, job Job, errMsg string) error

	// Cancel removes a job from all queues and marks it cancelled.
	Cancel(ctx context.Context, jobID string, queue string) error

	// Get fetches job metadata from Redis.
	Get(ctx context.Context, jobID string) (Job, error)
}
```

Guarantees:

- `Reserve` is at-least-once and must set `processing` deadlines.
- `Ack`, `FailWithRetry`, `FailTerminal`, and `MoveToDead` must remove the job from `processing`.
- `Cancel` is best-effort if the job is already executing.

---

## API Contract (Minimal)

These endpoints are required and their behaviors are fixed.

```
GET    /health            - Fetch health 200 OK
POST   /jobs              - Enqueue immediate job
POST   /jobs/delayed      - Enqueue delayed job
GET    /jobs/:id          - Fetch job state
DELETE /jobs/:id          - Cancel job
GET    /queues/:name/stats - Queue metrics
```

Response rules:

- Enqueue returns `202 Accepted` with job ID.
- Status returns `200 OK` with job JSON.
- Cancel returns `202 Accepted` if job exists, `404` otherwise.
