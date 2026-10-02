//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"
)

func TestNewPoolCanQuery(t *testing.T) {
	pool := openTestPool(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	var result int

	if err := pool.QueryRow(ctx, "SELECT 1").Scan(&result); err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if result != 1 {
		t.Fatalf("expected 1, got %d", result)
	}
}
