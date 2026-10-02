//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"nba-stats-api/internal/postgres"
)

func TestNewPoolCanQuery(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")

	if databaseURL == "" {
		env, err := godotenv.Read("../../.env")
		if err != nil {
			t.Fatal("set TEST_DATABASE_URL or provide .env at the project root")
		}

		databaseURL = env["DATABASE_URL"]
	}

	connectionCtx, cancelConnection := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	pool, err := postgres.NewPool(connectionCtx, databaseURL)
	cancelConnection()

	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	queryCtx, cancelQuery := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancelQuery()

	var result int
	if err := pool.QueryRow(queryCtx, "SELECT 1").Scan(&result); err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if result != 1 {
		t.Fatalf("expected 1, got %d", result)
	}
}
