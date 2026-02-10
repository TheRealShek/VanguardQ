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
