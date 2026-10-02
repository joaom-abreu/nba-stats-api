package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"nba-stats-api/internal/domain"
	"nba-stats-api/internal/httpapi"
)

func newGameTestHandler(
	listGames func(context.Context, domain.GameFilter) ([]domain.Game, error),
) http.Handler {
	return httpapi.NewHandler(
		func(context.Context) error { return nil },
		teamReaderStub{},
		gameReaderStub{list: listGames},
	)
}

func TestListGames(t *testing.T) {
	dateFrom := time.Date(2025, 4, 20, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2025, 4, 22, 0, 0, 0, 0, time.UTC)
	startTime := time.Date(2025, 4, 21, 23, 0, 0, 0, time.UTC)

	homeScore := 100
	awayScore := 90

	game := domain.Game{
		ID:          42,
		Source:      "balldontlie",
		ExternalID:  "provider-game-42",
		Season:      2024,
		Phase:       "playoffs",
		GameDate:    "2025-04-21",
		StartTime:   &startTime,
		HomeTeamID:  7,
		AwayTeamID:  8,
		HomeScore:   &homeScore,
		AwayScore:   &awayScore,
		Status:      domain.GameFinished,
		CollectedAt: time.Date(2025, 4, 22, 1, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name       string
		query      string
		result     []domain.Game
		wantFilter domain.GameFilter
	}{
		{
			name:  "defaults and empty result",
			query: "source=demo",
			wantFilter: domain.GameFilter{
				Source: "demo",
				Limit:  20,
			},
		},
		{
			name: "all filters",
			query: "source=balldontlie&season=2024&phase=playoffs" +
				"&date_from=2025-04-20&date_to=2025-04-22" +
				"&team_id=7&status=finished&limit=2&offset=1",
			result: []domain.Game{game},
			wantFilter: domain.GameFilter{
				Source:   "balldontlie",
				Season:   2024,
				Phase:    "playoffs",
				DateFrom: &dateFrom,
				DateTo:   &dateTo,
				TeamID:   7,
				Status:   domain.GameFinished,
				Limit:    2,
				Offset:   1,
			},
		},
		{
			name: "same start and end date",
			query: "source=demo&date_from=2025-04-20" +
				"&date_to=2025-04-20",
			wantFilter: domain.GameFilter{
				Source:   "demo",
				DateFrom: &dateFrom,
				DateTo:   &dateFrom,
				Limit:    20,
			},
		},
		{
			name:  "only start date",
			query: "source=demo&date_from=2025-04-20",
			wantFilter: domain.GameFilter{
				Source:   "demo",
				DateFrom: &dateFrom,
				Limit:    20,
			},
		},
		{
			name:  "only end date",
			query: "source=demo&date_to=2025-04-22",
			wantFilter: domain.GameFilter{
				Source: "demo",
				DateTo: &dateTo,
				Limit:  20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0

			handler := newGameTestHandler(func(
				ctx context.Context,
				filter domain.GameFilter,
			) ([]domain.Game, error) {
				calls++

				if !reflect.DeepEqual(filter, tt.wantFilter) {
					t.Fatalf(
						"filter = %#v; want %#v",
						filter, tt.wantFilter,
					)
				}

				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("repository context has no deadline")
				}

				return tt.result, nil
			})

			request := httptest.NewRequest(
				http.MethodGet,
				"/v1/games?"+tt.query,
				nil,
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf(
					"status = %d; body = %s",
					recorder.Code, recorder.Body.String(),
				)
			}

			if calls != 1 {
				t.Fatalf("repository calls = %d; want 1", calls)
			}

			if recorder.Header().Get("Content-Type") != "application/json" {
				t.Fatal("expected JSON content type")
			}

			if recorder.Header().Get("X-Request-ID") == "" {
				t.Fatal("missing request ID")
			}

			var response struct {
				Data []domain.Game `json:"data"`
				Meta struct {
					Limit    int `json:"limit"`
					Offset   int `json:"offset"`
					Returned int `json:"returned"`
				} `json:"meta"`
			}

			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}

			wantData := tt.result
			if wantData == nil {
				wantData = make([]domain.Game, 0)
			}

			if !reflect.DeepEqual(response.Data, wantData) {
				t.Fatalf(
					"data = %#v; want %#v",
					response.Data, wantData,
				)
			}

			if response.Meta.Limit != tt.wantFilter.Limit ||
				response.Meta.Offset != tt.wantFilter.Offset ||
				response.Meta.Returned != len(wantData) {
				t.Fatalf("unexpected metadata: %+v", response.Meta)
			}
		})
	}
}

func TestListGamesRejectsInvalidQuery(t *testing.T) {
	queries := []string{
		"",
		"source=",
		"source=other",
		"source=demo&source=balldontlie",
		"source=%ZZ",
		"source=demo;season=2024",
		"source=demo&conference=east",
		"source=demo&season=",
		"source=demo&season=0",
		"source=demo&season=-1",
		"source=demo&season=abc",
		"source=demo&season=1.5",
		"source=demo&season=2147483648",
		"source=demo&season=2024&season=2023",
		"source=demo&phase=",
		"source=demo&phase=finals",
		"source=demo&phase=playoffs&phase=regular_season",
		"source=demo&date_from=",
		"source=demo&date_from=2025-02-30",
		"source=demo&date_from=2025-4-20",
		"source=demo&date_from=0000-01-01",
		"source=demo&date_to=",
		"source=demo&date_to=abc",
		"source=demo&date_to=2025-13-01",
		"source=demo&date_from=2025-04-22&date_to=2025-04-20",
		"source=demo&date_from=2025-04-20&date_from=2025-04-21",
		"source=demo&date_to=2025-04-20&date_to=2025-04-21",
		"source=demo&team_id=",
		"source=demo&team_id=0",
		"source=demo&team_id=-1",
		"source=demo&team_id=abc",
		"source=demo&team_id=9223372036854775808",
		"source=demo&team_id=7&team_id=8",
		"source=demo&status=",
		"source=demo&status=done",
		"source=demo&status=finished&status=scheduled",
		"source=demo&limit=",
		"source=demo&limit=0",
		"source=demo&limit=101",
		"source=demo&limit=abc",
		"source=demo&limit=2&limit=3",
		"source=demo&offset=",
		"source=demo&offset=-1",
		"source=demo&offset=abc",
		"source=demo&offset=9223372036854775808",
		"source=demo&offset=1&offset=2",
	}

	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			calls := 0

			handler := newGameTestHandler(func(
				context.Context,
				domain.GameFilter,
			) ([]domain.Game, error) {
				calls++
				return nil, nil
			})

			request := httptest.NewRequest(
				http.MethodGet, "/v1/games", nil,
			)
			request.URL.RawQuery = query

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			assertAPIError(
				t, recorder, http.StatusBadRequest, "invalid_query",
			)

			if calls != 0 {
				t.Fatalf("repository calls = %d; want 0", calls)
			}
		})
	}
}

func TestListGamesRepositoryError(t *testing.T) {
	calls := 0

	handler := newGameTestHandler(func(
		context.Context,
		domain.GameFilter,
	) ([]domain.Game, error) {
		calls++
		return nil, errors.New("private database error")
	})

	request := httptest.NewRequest(
		http.MethodGet, "/v1/games?source=demo", nil,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	message := assertAPIError(
		t, recorder, http.StatusInternalServerError, "internal_error",
	)

	if message != "could not list games" {
		t.Fatalf("unexpected error message: %q", message)
	}

	if calls != 1 {
		t.Fatalf("repository calls = %d; want 1", calls)
	}
}

func TestListGamesRejectsPost(t *testing.T) {
	calls := 0

	handler := newGameTestHandler(func(
		context.Context,
		domain.GameFilter,
	) ([]domain.Game, error) {
		calls++
		return nil, nil
	})

	request := httptest.NewRequest(
		http.MethodPost, "/v1/games?source=demo", nil,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	assertAPIError(
		t, recorder, http.StatusMethodNotAllowed, "method_not_allowed",
	)

	if recorder.Header().Get("Allow") != "GET, HEAD" {
		t.Fatal("unexpected Allow header")
	}

	if calls != 0 {
		t.Fatalf("repository calls = %d; want 0", calls)
	}
}
