package postgres

import (
	"context"
	"testing"
	"time"
)

func TestConnect_InvalidDSN(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pool, err := Connect(ctx, "not-a-dsn")
	if err == nil {
		if pool != nil {
			pool.Close()
		}
		t.Fatalf("expected error for invalid DSN")
	}
}
