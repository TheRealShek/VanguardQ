package queue

import (
	"os"

	"github.com/redis/go-redis/v9"
)

type QueueService struct {
	rdb *redis.Client
}

func Start() {
	client := NewRedisClient()
	queueSvc := NewQueueService(client)
}

func NewRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379" // fallback for local dev
	}
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func NewQueueService(rdb *redis.Client) *QueueService {
	return &QueueService{rdb: rdb}
}
