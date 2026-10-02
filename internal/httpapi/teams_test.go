package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"nba-stats-api/internal/domain"
)

func emptyTeamList(
	ctx context.Context,
	filter domain.TeamFilter,
) ([]domain.Team, error) {
	return nil, nil
}

func newTeamTestHandler(
	listTeams func(context.Context, domain.TeamFilter) ([]domain.Team, error),
) http.Handler {
	return newTestHandler(
		func(context.Context) error { return nil },
		teamReaderStub{list: listTeams},
	)
}

func TestListTeams(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		filter domain.TeamFilter
		teams  []domain.Team
	}{
		{
			name:  "defaults and empty result",
			query: "source=demo",
			filter: domain.TeamFilter{
				Source: "demo",
				Limit:  20,
			},
		},
		{
			name:  "custom filters",
			query: "source=balldontlie&conference=west&limit=2&offset=1",
			filter: domain.TeamFilter{
				Source:     "balldontlie",
				Conference: "west",
				Limit:      2,
				Offset:     1,
			},
			teams: []domain.Team{
				{
					ID: 42, Source: "balldontlie", ExternalID: "7",
					Name: "Test Team", Abbreviation: "TT", Conference: "west",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedFilter domain.TeamFilter

			handler := newTeamTestHandler(
				func(ctx context.Context, filter domain.TeamFilter) ([]domain.Team, error) {
					receivedFilter = filter
					return tt.teams, nil
				},
			)

			request := httptest.NewRequest(
				http.MethodGet,
				"/v1/teams?"+tt.query,
				nil,
			)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", recorder.Code)
			}

			if receivedFilter != tt.filter {
				t.Fatalf("expected filter %+v, got %+v",
					tt.filter, receivedFilter)
			}

			if recorder.Header().Get("X-Request-ID") == "" {
				t.Fatal("expected a request ID")
			}

			var body struct {
				Data []domain.Team `json:"data"`
				Meta struct {
					Limit    int `json:"limit"`
					Offset   int `json:"offset"`
					Returned int `json:"returned"`
				} `json:"meta"`
			}

			if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}

			if body.Data == nil || len(body.Data) != len(tt.teams) {
				t.Fatalf("unexpected data: %+v", body.Data)
			}

			for i, team := range tt.teams {
				if body.Data[i] != team {
					t.Fatalf("expected %+v, got %+v", team, body.Data[i])
				}
			}

			if body.Meta.Limit != tt.filter.Limit ||
				body.Meta.Offset != tt.filter.Offset ||
				body.Meta.Returned != len(tt.teams) {
				t.Fatalf("unexpected pagination: %+v", body.Meta)
			}
		})
	}
}

func TestListTeamsRejectsInvalidQueries(t *testing.T) {
	queries := []string{
		"",
		"source=other",
		"source=demo&source=balldontlie",
		"source=demo&conference=north",
		"source=demo&conference=",
		"source=demo&limit=0",
		"source=demo&limit=101",
		"source=demo&limit=abc",
		"source=demo&limit=",
		"source=demo&offset=-1",
		"source=demo&offset=abc",
		"source=demo&limit=2&limit=3",
		"source=demo&unknown=value",
	}

	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			called := false

			handler := newTeamTestHandler(
				func(ctx context.Context, filter domain.TeamFilter) ([]domain.Team, error) {
					called = true
					return nil, nil
				},
			)

			request := httptest.NewRequest(
				http.MethodGet,
				"/v1/teams?"+query,
				nil,
			)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			assertAPIError(t, recorder, http.StatusBadRequest, "invalid_query")

			if called {
				t.Fatal("invalid query should not reach the repository")
			}
		})
	}
}

func TestAPIErrorResponses(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		status int
		code   string
		calls  int
	}{
		{
			name: "repository failure", method: http.MethodGet,
			path:   "/v1/teams?source=demo",
			status: http.StatusInternalServerError,
			code:   "internal_error", calls: 1,
		},
		{
			name: "method not allowed", method: http.MethodPost,
			path:   "/v1/teams",
			status: http.StatusMethodNotAllowed,
			code:   "method_not_allowed",
		},
		{
			name: "unknown route", method: http.MethodGet,
			path:   "/v1/missing",
			status: http.StatusNotFound,
			code:   "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0

			handler := newTeamTestHandler(
				func(ctx context.Context, filter domain.TeamFilter) ([]domain.Team, error) {
					calls++
					return nil, errors.New("private database details")
				},
			)

			request := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			message := assertAPIError(t, recorder, tt.status, tt.code)

			if calls != tt.calls {
				t.Fatalf("expected %d repository calls, got %d", tt.calls, calls)
			}

			if tt.status == http.StatusInternalServerError &&
				message != "could not list teams" {
				t.Fatalf("unexpected error message: %q", message)
			}

			if tt.status == http.StatusMethodNotAllowed &&
				recorder.Header().Get("Allow") != "GET, HEAD" {
				t.Fatal("expected Allow header")
			}
		})
	}
}

func assertAPIError(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	status int,
	code string,
) string {
	t.Helper()

	if recorder.Code != status {
		t.Fatalf("expected status %d, got %d", status, recorder.Code)
	}

	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatal("expected application/json")
	}

	var body struct {
		Error struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}

	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}

	if body.Error.Code != code || body.Error.Message == "" {
		t.Fatalf("unexpected error: %+v", body.Error)
	}

	id := recorder.Header().Get("X-Request-ID")
	if id == "" || body.Error.RequestID != id {
		t.Fatal("expected matching request IDs in header and body")
	}

	return body.Error.Message
}
