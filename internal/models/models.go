package models

import (
	"encoding/json"
	"time"
)

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
	Queue         string          `json:"queue"` // This Is the Priorty of the Job
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

/*
Field notes:

- `Queue` can only be set to `high` or `default` or`low`.
- `ScheduledAt` is zero value for immediate jobs.
- `StartedAt` and `CompletedAt` are zero value until set.
- `TimeoutMs` is required; 0 means no timeout.
- `Error` is set only for `failed` or `dead`.
*/
