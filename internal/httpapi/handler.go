package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func NewHandler(pingDatabase func(context.Context) error) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", healthLive)

	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := pingDatabase(ctx); err != nil {
			writeHealth(w, http.StatusServiceUnavailable, "unavailable")
			return
		}

		writeHealth(w, http.StatusOK, "ok")
	})

	return mux
}

func healthLive(w http.ResponseWriter, r *http.Request) {
	writeHealth(w, http.StatusOK, "ok")
}

func writeHealth(w http.ResponseWriter, statusCode int, status string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(statusCode)

	response := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to write health response", "error", err)
	}
}
