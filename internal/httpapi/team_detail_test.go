package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"nba-stats-api/internal/domain"
	"nba-stats-api/internal/httpapi"
)

func TestGetTeam(t *testing.T) {
	wantTeam := domain.Team{
		ID:           42,
		Source:       "demo",
		ExternalID:   "A",
		Name:         "Demo Team A",
		Abbreviation: "A",
		Conference:   "east",
	}

	tests := []struct {
		name      string
		method    string
		path      string
		err       error
		wantCode  int
		wantError string
		wantID    int64
		wantCalls int
	}{
		{
			name: "found", method: http.MethodGet,
			path: "/v1/teams/42", wantCode: http.StatusOK,
			wantID: 42, wantCalls: 1,
		},
		{
			name: "not found", method: http.MethodGet,
			path:      "/v1/teams/99",
			err:       fmt.Errorf("lookup: %w", domain.ErrTeamNotFound),
			wantCode:  http.StatusNotFound,
			wantError: "team_not_found",
			wantID:    99, wantCalls: 1,
		},
		{
			name: "repository failure", method: http.MethodGet,
			path:      "/v1/teams/42",
			err:       errors.New("private database details"),
			wantCode:  http.StatusInternalServerError,
			wantError: "internal_error",
			wantID:    42, wantCalls: 1,
		},
		{
			name: "zero ID", method: http.MethodGet,
			path:     "/v1/teams/0",
			wantCode: http.StatusBadRequest, wantError: "invalid_id",
		},
		{
			name: "negative ID", method: http.MethodGet,
			path:     "/v1/teams/-1",
			wantCode: http.StatusBadRequest, wantError: "invalid_id",
		},
		{
			name: "text ID", method: http.MethodGet,
			path:     "/v1/teams/abc",
			wantCode: http.StatusBadRequest, wantError: "invalid_id",
		},
		{
			name: "fractional ID", method: http.MethodGet,
			path:     "/v1/teams/1.5",
			wantCode: http.StatusBadRequest, wantError: "invalid_id",
		},
		{
			name: "overflow", method: http.MethodGet,
			path:     "/v1/teams/9223372036854775808",
			wantCode: http.StatusBadRequest, wantError: "invalid_id",
		},
		{
			name: "query parameters", method: http.MethodGet,
			path:     "/v1/teams/42?source=demo",
			wantCode: http.StatusBadRequest, wantError: "invalid_query",
		},
		{
			name: "method not allowed", method: http.MethodPost,
			path:      "/v1/teams/42",
			wantCode:  http.StatusMethodNotAllowed,
			wantError: "method_not_allowed",
		},
		{
			name: "missing ID", method: http.MethodGet,
			path:     "/v1/teams/",
			wantCode: http.StatusNotFound, wantError: "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedID int64
			calls := 0

			handler := httpapi.NewHandler(
				func(context.Context) error { return nil },
				teamReaderStub{
					getByID: func(ctx context.Context, id int64) (domain.Team, error) {
						receivedID = id
						calls++

						if tt.err != nil {
							return domain.Team{}, tt.err
						}

						return wantTeam, nil
					},
				},
			)

			request := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			if calls != tt.wantCalls || receivedID != tt.wantID {
				t.Fatalf("expected %d calls with ID %d, got %d calls with ID %d",
					tt.wantCalls, tt.wantID, calls, receivedID)
			}

			if tt.wantCode != http.StatusOK {
				message := assertAPIError(t, recorder, tt.wantCode, tt.wantError)

				if tt.wantCode == http.StatusInternalServerError &&
					message != "could not get team" {
					t.Fatalf("unexpected error message: %q", message)
				}

				if tt.wantCode == http.StatusMethodNotAllowed &&
					recorder.Header().Get("Allow") != "GET, HEAD" {
					t.Fatal("expected Allow header")
				}

				return
			}

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", recorder.Code)
			}

			if recorder.Header().Get("Content-Type") != "application/json" {
				t.Fatal("expected application/json")
			}

			if recorder.Header().Get("X-Request-ID") == "" {
				t.Fatal("expected a request ID")
			}

			var body struct {
				Data domain.Team `json:"data"`
			}

			if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}

			if body.Data != wantTeam {
				t.Fatalf("expected %+v, got %+v", wantTeam, body.Data)
			}
		})
	}
}
