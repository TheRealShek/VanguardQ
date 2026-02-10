package validators

import "fmt"

type QueueName string

const (
	QueueHigh    QueueName = "high"
	QueueDefault QueueName = "default"
	QueueLow     QueueName = "low"
)

func ValidateQueue(q string) error {
	switch QueueName(q) {
	case QueueHigh, QueueDefault, QueueLow:
		return nil
	default:
		return fmt.Errorf("invalid queue: %s", q)
	}
}
