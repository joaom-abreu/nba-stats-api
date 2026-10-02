package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", healthLive)

	return mux
}

func healthLive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to write health response", "error", err)
	}
}
