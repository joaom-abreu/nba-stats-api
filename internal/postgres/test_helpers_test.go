//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"nba-stats-api/internal/postgres"
)

func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")

	if databaseURL == "" {
		env, err := godotenv.Read("../../.env")
		if err != nil {
			t.Fatal("set TEST_DATABASE_URL or provide .env at the project root")
		}

		databaseURL = env["DATABASE_URL"]
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	pool, err := postgres.NewPool(ctx, databaseURL)
	cancel()

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)

	return pool
}
