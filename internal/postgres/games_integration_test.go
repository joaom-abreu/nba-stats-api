//go:build integration

package postgres_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"nba-stats-api/internal/domain"
	"nba-stats-api/internal/postgres"
)

func gameDatePointer(t *testing.T, value string) *time.Time {
	t.Helper()

	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatal(err)
	}

	return &date
}

func gameScorePointer(value int) *int {
	return &value
}

func TestGameRepositoryList(t *testing.T) {
	pool := openTestPool(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanupCtx, cancelCleanup := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancelCleanup()

		_ = tx.Rollback(cleanupCtx)
	})

	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE games (
			LIKE public.games INCLUDING ALL
		) ON COMMIT DROP
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO games (
			id, source, external_id, season, phase, game_date,
			home_team_id, away_team_id,
			home_score, away_score, status, start_time
		)
		OVERRIDING SYSTEM VALUE
		VALUES
			(10, 'demo', '101', 2024, 'regular_season', DATE '2025-01-02',
			 11, 12, 110, 100, 'finished', TIMESTAMPTZ '2025-01-02 15:00:00-03'),
			(20, 'demo', '102', 2024, 'regular_season', DATE '2025-01-04',
			 13, 11, 105, 90, 'finished', NULL),
			(30, 'demo', '103', 2024, 'regular_season', DATE '2025-01-06',
			 11, 14, 100, 95, 'finished', NULL),
			(40, 'demo', '104', 2024, 'regular_season', DATE '2025-01-08',
			 11, 12, NULL, NULL, 'scheduled', NULL),
			(50, 'demo', '105', 2024, 'regular_season', DATE '2025-01-10',
			 13, 11, 60, 62, 'in_progress', NULL),
			(60, 'demo', '106', 2024, 'regular_season', DATE '2025-01-10',
			 12, 14, 0, 1, 'finished', NULL),
			(70, 'demo', '107', 2024, 'playoffs', DATE '2025-01-11',
			 11, 12, 120, 110, 'finished', NULL),
			(80, 'demo', '108', 2023, 'regular_season', DATE '2025-01-12',
			 11, 12, 112, 110, 'finished', NULL),
			(90, 'balldontlie', '101', 2024, 'regular_season', DATE '2025-01-13',
			 21, 22, 105, 99, 'finished', NULL)
	`)
	if err != nil {
		t.Fatal(err)
	}

	repository := postgres.NewGameRepository(tx)

	jan2 := gameDatePointer(t, "2025-01-02")
	jan4 := gameDatePointer(t, "2025-01-04")
	jan8 := gameDatePointer(t, "2025-01-08")
	jan10 := gameDatePointer(t, "2025-01-10")

	base := domain.GameFilter{
		Source: "demo",
		Season: 2024,
		Phase:  "regular_season",
		Limit:  20,
	}

	tests := []struct {
		name    string
		change  func(*domain.GameFilter)
		wantIDs []int64
	}{
		{
			name:    "regular season",
			wantIDs: []int64{60, 50, 40, 30, 20, 10},
		},
		{
			name: "source only",
			change: func(f *domain.GameFilter) {
				f.Season = 0
				f.Phase = ""
			},
			wantIDs: []int64{80, 70, 60, 50, 40, 30, 20, 10},
		},
		{
			name: "finished",
			change: func(f *domain.GameFilter) {
				f.Status = domain.GameFinished
			},
			wantIDs: []int64{60, 30, 20, 10},
		},
		{
			name: "scheduled",
			change: func(f *domain.GameFilter) {
				f.Status = domain.GameScheduled
			},
			wantIDs: []int64{40},
		},
		{
			name: "in progress",
			change: func(f *domain.GameFilter) {
				f.Status = domain.GameInProgress
			},
			wantIDs: []int64{50},
		},
		{
			name: "home and away",
			change: func(f *domain.GameFilter) {
				f.TeamID = 11
			},
			wantIDs: []int64{50, 40, 30, 20, 10},
		},
		{
			name: "inclusive date range",
			change: func(f *domain.GameFilter) {
				f.DateFrom = jan4
				f.DateTo = jan8
			},
			wantIDs: []int64{40, 30, 20},
		},
		{
			name: "same date ordered by ID",
			change: func(f *domain.GameFilter) {
				f.DateFrom = jan10
				f.DateTo = jan10
			},
			wantIDs: []int64{60, 50},
		},
		{
			name: "date from",
			change: func(f *domain.GameFilter) {
				f.DateFrom = jan8
			},
			wantIDs: []int64{60, 50, 40},
		},
		{
			name: "date to",
			change: func(f *domain.GameFilter) {
				f.DateTo = jan4
			},
			wantIDs: []int64{20, 10},
		},
		{
			name: "pagination",
			change: func(f *domain.GameFilter) {
				f.Limit = 2
				f.Offset = 2
			},
			wantIDs: []int64{40, 30},
		},
		{
			name: "playoffs",
			change: func(f *domain.GameFilter) {
				f.Phase = "playoffs"
			},
			wantIDs: []int64{70},
		},
		{
			name: "other season",
			change: func(f *domain.GameFilter) {
				f.Season = 2023
			},
			wantIDs: []int64{80},
		},
		{
			name: "other source",
			change: func(f *domain.GameFilter) {
				f.Source = "balldontlie"
			},
			wantIDs: []int64{90},
		},
		{
			name: "empty page",
			change: func(f *domain.GameFilter) {
				f.Offset = 100
			},
			wantIDs: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := base
			if tt.change != nil {
				tt.change(&filter)
			}

			games, err := repository.List(ctx, filter)
			if err != nil {
				t.Fatal(err)
			}

			if games == nil || len(games) != len(tt.wantIDs) {
				t.Fatalf("expected %d games, got %+v", len(tt.wantIDs), games)
			}

			for i, id := range tt.wantIDs {
				game := games[i]

				if game.ID != id || game.Source != filter.Source {
					t.Fatalf("position %d: unexpected game %+v", i, game)
				}

				if game.CollectedAt.IsZero() ||
					game.CollectedAt.Location() != time.UTC {
					t.Fatal("expected collection time in UTC")
				}

				if id == 40 && (game.HomeScore != nil ||
					game.AwayScore != nil || game.StartTime != nil) {
					t.Fatal("expected absent scores and start time")
				}

				if id == 60 && (game.HomeScore == nil || *game.HomeScore != 0) {
					t.Fatal("expected a present score of zero")
				}
			}
		})
	}

	t.Run("all fields", func(t *testing.T) {
		filter := base
		filter.DateFrom = jan2
		filter.DateTo = jan2

		games, err := repository.List(ctx, filter)
		if err != nil {
			t.Fatal(err)
		}

		if len(games) != 1 {
			t.Fatalf("expected one game, got %d", len(games))
		}

		startTime := time.Date(2025, 1, 2, 18, 0, 0, 0, time.UTC)

		want := domain.Game{
			ID:          10,
			Source:      "demo",
			ExternalID:  "101",
			Season:      2024,
			Phase:       "regular_season",
			GameDate:    "2025-01-02",
			StartTime:   &startTime,
			HomeTeamID:  11,
			AwayTeamID:  12,
			HomeScore:   gameScorePointer(110),
			AwayScore:   gameScorePointer(100),
			Status:      domain.GameFinished,
			CollectedAt: games[0].CollectedAt,
		}

		if !reflect.DeepEqual(games[0], want) {
			t.Fatalf("expected %+v, got %+v", want, games[0])
		}
	})
}
