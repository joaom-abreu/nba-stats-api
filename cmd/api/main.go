package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"nba-stats-api/internal/httpapi"
	"nba-stats-api/internal/postgres"
)

func main() {
	if err := run(); err != nil {
		slog.Error("API stopped with an error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return errors.New("could not load .env file")
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	connectionCtx, cancelConnection := context.WithTimeout(ctx, 5*time.Second)
	pool, err := postgres.NewPool(
		connectionCtx,
		os.Getenv("DATABASE_URL"),
	)
	cancelConnection()

	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	defer pool.Close()

	slog.Info("connected to PostgreSQL")

	teamRepository := postgres.NewTeamRepository(pool)
	gameRepository := postgres.NewGameRepository(pool)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewHandler(pool.Ping, teamRepository, gameRepository),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("starting API", "address", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err

	case <-ctx.Done():
		stop()
		slog.Info("shutting down API")

		shutdownCtx, cancelShutdown := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancelShutdown()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return err
		}

		return nil
	}
}
