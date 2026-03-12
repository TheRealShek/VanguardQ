package queue

import (
	"context"
	"log"
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

func Stop(rdb *redis.Client) {
	_ = rdb.Close()
}

func NewRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379" // fallback for local dev
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil{
		log.Fatal("Redis Connection Failed: ", err)
	}
	return rdb
}

func NewQueueService(rdb *redis.Client) *QueueService {
	return &QueueService{rdb: rdb}
}
