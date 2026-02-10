package validators

import (
	"fmt"

	"github.com/TheRealShek/VanguardQ/internal/models" // update to your actual module path
)

func IsValidStatus(s models.JobStatus) bool {
	switch s {
	case models.JobQueued,
		models.JobDelayed,
		models.JobProcessing,
		models.JobRetryWait,
		models.JobSuccess,
		models.JobFailed,
		models.JobCancelled,
		models.JobDead:
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
var allowedTransitions = map[models.JobStatus][]models.JobStatus{
	models.JobQueued: {
		models.JobProcessing,
		models.JobCancelled,
	},
	models.JobDelayed: {
		models.JobQueued,
		models.JobCancelled,
	},
	models.JobProcessing: {
		models.JobSuccess,
		models.JobRetryWait,
		models.JobDead,
		models.JobFailed,
		models.JobCancelled,
	},
	models.JobRetryWait: {
		models.JobQueued,
	},
	// terminal states
	models.JobSuccess:   {},
	models.JobFailed:    {},
	models.JobCancelled: {},
	models.JobDead:      {},
}

func CanTransition(from, to models.JobStatus) error {
	allowed, ok := allowedTransitions[from]
	if !ok {
		return fmt.Errorf("unknown state: %s", from)
	}
	for _, s := range allowed {
		if s == to {
			return nil
		}
	}
	return fmt.Errorf("invalid transition %s → %s", from, to)
}

func IsFinalState(s models.JobStatus) bool {
	switch s {
	case models.JobSuccess, models.JobFailed, models.JobCancelled, models.JobDead:
		return true
	default:
		return false
	}
}
