package httpapi

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"nba-stats-api/internal/domain"
)

type TeamReader interface {
	List(context.Context, domain.TeamFilter) ([]domain.Team, error)
	GetByID(context.Context, int64) (domain.Team, error)
}

type GameReader interface {
	List(context.Context, domain.GameFilter) ([]domain.Game, error)
}

func NewHandler(
	pingDatabase func(context.Context) error,
	teams TeamReader,
	games GameReader,
) http.Handler {
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

	mux.HandleFunc("GET /v1/teams", listTeamsHandler(teams.List))
	mux.HandleFunc("GET /v1/teams/{id}", getTeamHandler(teams.GetByID))
	mux.HandleFunc("GET /v1/games", listGamesHandler(games.List))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			probe := r.Clone(r.Context())
			probe.Method = http.MethodGet

			_, pattern := mux.Handler(probe)

			if pattern != "" && pattern != "/" {
				w.Header().Set("Allow", "GET, HEAD")
				writeError(
					w, r,
					http.StatusMethodNotAllowed,
					"method_not_allowed",
					"method not allowed",
				)
				return
			}
		}

		writeError(
			w, r,
			http.StatusNotFound,
			"not_found",
			"route not found",
		)
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := rand.Text()
		w.Header().Set("X-Request-ID", id)

		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}

func healthLive(w http.ResponseWriter, r *http.Request) {
	writeHealth(w, http.StatusOK, "ok")
}

func writeHealth(w http.ResponseWriter, statusCode int, status string) {
	w.Header().Set("Cache-Control", "no-store")

	response := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	writeJSON(w, statusCode, response)
}
