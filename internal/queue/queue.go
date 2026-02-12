package queue

import (
	"context"
	"time"

	"github.com/TheRealShek/VanguardQ/internal/models"
)

/*
Queue defines the contract for the task queue system.

Enqueue        → Push immediate job into queue
EnqueueDelayed → Schedule job for future execution
Reserve        → Atomically fetch and reserve next job
Ack            → Mark job as successfully completed
FailWithRetry  → Move job into retry state
FailTerminal   → Mark job permanently failed (no retry)
MoveToDead     → Send job to dead letter queue
Cancel         → Remove job from all queues and mark cancelled
Get            → Fetch job metadata from storage
*/

type Queue interface {
	Enqueue(ctx context.Context, job models.Job) error
	EnqueueDelayed(ctx context.Context, job models.Job, runAt time.Time) error
	Reserve(ctx context.Context, queue string, block time.Duration) (models.Job, error)
	Ack(ctx context.Context, job models.Job) error
	FailWithRetry(ctx context.Context, job models.Job, runAt time.Time, errMsg string) error
	FailTerminal(ctx context.Context, job models.Job, errMsg string) error
	MoveToDead(ctx context.Context, job models.Job, errMsg string) error
	Cancel(ctx context.Context, jobID string, queue string) error
	Get(ctx context.Context, jobID string) (models.Job, error)
}

func (qs *QueueService) Enqueue(ctx context.Context, job models.Job) error {

	return nil
}

func (qs *QueueService) EnqueueDelayed(ctx context.Context, job models.Job, runAt time.Time) error {

	return nil
}

// func (qs *QueueService) Reserve(ctx context.Context, queue string, block time.Duration) (models.Job, error) {
// 	return models.Job{}, nil
// }

// func (qs *QueueService) Ack(ctx context.Context, job models.Job) error {
// 	return nil
// }

// func (qs *QueueService) FailWithRetry(ctx context.Context, job models.Job, runAt time.Time, errMsg string) error {
// 	return nil
// }

// func (qs *QueueService) FailTerminal(ctx context.Context, job models.Job, errMsg string) error {
// 	return nil
// }

// func (qs *QueueService) MoveToDead(ctx context.Context, job models.Job, errMsg string) error {
// 	return nil
// }

// func (qs *QueueService) Cancel(ctx context.Context, jobID string, queue string) error {
// 	return nil
// }

// func (qs *QueueService) Get(ctx context.Context, jobID string) (models.Job, error) {
// 	return models.Job{}, nil
// }
