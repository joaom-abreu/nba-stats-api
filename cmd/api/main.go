package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"nba-stats-api/internal/httpapi"
)

func main() {
	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("starting API", "address", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		slog.Error("API stopped with an error", "error", err)
		os.Exit(1)
	}
}
