package queue

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/TheRealShek/VanguardQ/internal/models"
)

// Convert Job struct to Redis hash fields (HSET args)
func jobToHash(job models.Job) map[string]string {
	return map[string]string{
		"id":             job.ID,
		"correlation_id": job.CorrelationID,
		"queue":          job.Queue,
		"payload":        string(job.Payload), // json.RawMessage is []byte
		"status":         string(job.Status),
		"created_at":     formatTime(job.CreatedAt),
		"scheduled_at":   formatTime(job.ScheduledAt),
		"started_at":     formatTime(job.StartedAt),
		"completed_at":   formatTime(job.CompletedAt),
		"timeout_ms":     strconv.Itoa(job.TimeoutMs),
		"retries":        strconv.Itoa(job.Retries),
		"max_retries":    strconv.Itoa(job.MaxRetries),
		"error":          job.Error,
	}
}

// Deserialize HGETALL result back to Job struct
func hashToJob(m map[string]string) models.Job {
	timeoutMs, _ := strconv.Atoi(m["timeout_ms"])
	retries, _ := strconv.Atoi(m["retries"])
	maxRetries, _ := strconv.Atoi(m["max_retries"])

	return models.Job{
		ID:            m["id"],
		CorrelationID: m["correlation_id"],
		Queue:         m["queue"],
		Payload:       json.RawMessage(m["payload"]),
		Status:        models.JobStatus(m["status"]),
		CreatedAt:     parseTime(m["created_at"]),
		ScheduledAt:   parseTime(m["scheduled_at"]),
		StartedAt:     parseTime(m["started_at"]),
		CompletedAt:   parseTime(m["completed_at"]),
		TimeoutMs:     timeoutMs,
		Retries:       retries,
		MaxRetries:    maxRetries,
		Error:         m["error"],
	}
}

// formatTime converts time to millisecond unix timestamp string.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return strconv.FormatInt(t.UnixMilli(), 10)
}

// parseTime converts millisecond string to time.
func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}

	ms, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{} // or handle error properly
	}

	return time.UnixMilli(ms)
}
