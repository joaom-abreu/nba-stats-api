//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"nba-stats-api/internal/domain"
	"nba-stats-api/internal/postgres"
)

func TestTeamRepositoryList(t *testing.T) {
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
		CREATE TEMP TABLE teams (
			LIKE public.teams INCLUDING ALL
		) ON COMMIT DROP
	`)
	if err != nil {
		t.Fatal(err)
	}

	fixtures := []domain.Team{
		{
			ID: 30, Source: "demo", ExternalID: "A",
			Name: "Test A", Abbreviation: "TA", Conference: "east",
		},
		{
			ID: 10, Source: "demo", ExternalID: "B",
			Name: "Test B", Abbreviation: "TB", Conference: "east",
		},
		{
			ID: 40, Source: "demo", ExternalID: "C",
			Name: "Test C", Abbreviation: "TC", Conference: "west",
		},
		{
			ID: 20, Source: "demo", ExternalID: "D",
			Name: "Test D", Abbreviation: "TD", Conference: "west",
		},
		{
			ID: 50, Source: "balldontlie", ExternalID: "A",
			Name: "Provider A", Abbreviation: "PA", Conference: "east",
		},
	}

	expectedTeams := make(map[int64]domain.Team)

	for _, team := range fixtures {
		expectedTeams[team.ID] = team

		_, err := tx.Exec(ctx, `
			INSERT INTO teams (
				id, source, external_id, name, abbreviation, conference
			)
			OVERRIDING SYSTEM VALUE
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
			team.ID,
			team.Source,
			team.ExternalID,
			team.Name,
			team.Abbreviation,
			team.Conference,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	repository := postgres.NewTeamRepository(tx)

	tests := []struct {
		name       string
		source     string
		conference string
		limit      int
		offset     int
		wantIDs    []int64
	}{
		{
			name: "all demo teams", source: "demo",
			limit: 20, wantIDs: []int64{10, 20, 30, 40},
		},
		{
			name: "east conference", source: "demo", conference: "east",
			limit: 20, wantIDs: []int64{10, 30},
		},
		{
			name: "west conference", source: "demo", conference: "west",
			limit: 20, wantIDs: []int64{20, 40},
		},
		{
			name: "first page", source: "demo",
			limit: 2, wantIDs: []int64{10, 20},
		},
		{
			name: "second page", source: "demo",
			limit: 2, offset: 2, wantIDs: []int64{30, 40},
		},
		{
			name: "other source", source: "balldontlie",
			limit: 20, wantIDs: []int64{50},
		},
		{
			name: "empty page", source: "demo",
			limit: 20, offset: 20, wantIDs: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teams, err := repository.List(ctx, domain.TeamFilter{
				Source:     tt.source,
				Conference: tt.conference,
				Limit:      tt.limit,
				Offset:     tt.offset,
			})
			if err != nil {
				t.Fatal(err)
			}

			if teams == nil {
				t.Fatal("expected an empty slice instead of nil")
			}

			if len(teams) != len(tt.wantIDs) {
				t.Fatalf("expected %d teams, got %d",
					len(tt.wantIDs), len(teams))
			}

			for i, id := range tt.wantIDs {
				want := expectedTeams[id]

				if teams[i] != want {
					t.Fatalf("position %d: expected %+v, got %+v",
						i, want, teams[i])
				}
			}
		})
	}
}
