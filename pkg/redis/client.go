package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewClient inicializa una conexion a Redis y verifica con un ping.
func NewClient(addr, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return rdb, nil
}
