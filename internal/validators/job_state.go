package validators

import "fmt"

type JobStatus string

const (
	StatusQueued     JobStatus = "queued"
	StatusDelayed    JobStatus = "delayed"
	StatusProcessing JobStatus = "processing"
	StatusRetryWait  JobStatus = "retry_wait"
	StatusSuccess    JobStatus = "success"
	StatusFailed     JobStatus = "failed"
	StatusCancelled  JobStatus = "cancelled"
	StatusDead       JobStatus = "dead"
)

func IsValidStatus(s JobStatus) bool {
	switch s {
	case StatusQueued,
		StatusDelayed,
		StatusProcessing,
		StatusRetryWait,
		StatusSuccess,
		StatusFailed,
		StatusCancelled,
		StatusDead:
		return true
	default:
		return false
	}
}

/*
queued      → [processing, cancelled]
delayed     → [queued, cancelled]
processing  → [success, retry_wait, dead, failed, cancelled]
retry_wait  → [queued]
success     → []   (final)
failed      → []   (final)
cancelled   → []   (final)
dead        → []   (final)
*/
var allowedTransitions = map[JobStatus][]JobStatus{
	StatusQueued: {
		StatusProcessing,
		StatusCancelled,
	},
	StatusDelayed: {
		StatusQueued,
		StatusCancelled,
	},
	StatusProcessing: {
		StatusSuccess,
		StatusRetryWait,
		StatusDead,
		StatusFailed,
		StatusCancelled,
	},
	StatusRetryWait: {
		StatusQueued,
	},
	// terminal states
	StatusSuccess:   {},
	StatusFailed:    {},
	StatusCancelled: {},
	StatusDead:      {},
}

func CanTransition(from, to JobStatus) error {
	allowed, ok := allowedTransitions[from]
	if !ok {
		return fmt.Errorf("unknown state: %s", from)
	}
	// We need to loop as `allowed` can have multiple Values
	for _, s := range allowed {
		if s == to {
			return nil
		}
	}
	return fmt.Errorf("invalid transition %s → %s", from, to)
}

func IsFinalState(s JobStatus) bool {
	switch s {
	case StatusSuccess, StatusFailed, StatusCancelled, StatusDead:
		return true
	default:
		return false
	}
}
