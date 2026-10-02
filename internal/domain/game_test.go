package domain_test

import (
	"errors"
	"testing"

	"nba-stats-api/internal/domain"
)

func score(value int) *int {
	return &value
}

func TestGameValidateResult(t *testing.T) {
	tests := []struct {
		name      string
		status    domain.GameStatus
		homeScore *int
		awayScore *int
		wantErr   bool
	}{
		{"finished game", domain.GameFinished, score(110), score(100), false},
		{"zero is a valid score", domain.GameFinished, score(0), score(100), false},
		{"scheduled without scores", domain.GameScheduled, nil, nil, false},
		{"postponed without scores", domain.GamePostponed, nil, nil, false},
		{"cancelled without scores", domain.GameCancelled, nil, nil, false},
		{"in progress with tied scores", domain.GameInProgress, score(60), score(60), false},
		{"in progress without scores", domain.GameInProgress, nil, nil, false},
		{"unknown status", "unknown", score(110), score(100), true},
		{"empty status", "", nil, nil, true},
		{"finished without home score", domain.GameFinished, nil, score(100), true},
		{"finished without away score", domain.GameFinished, score(110), nil, true},
		{"finished without both scores", domain.GameFinished, nil, nil, true},
		{"finished with tied scores", domain.GameFinished, score(100), score(100), true},
		{"finished with zero to zero", domain.GameFinished, score(0), score(0), true},
		{"negative home score", domain.GameFinished, score(-1), score(100), true},
		{"negative away score", domain.GameFinished, score(110), score(-1), true},
		{"negative score before finish", domain.GameScheduled, score(-1), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := domain.Game{
				HomeTeamID: 1,
				AwayTeamID: 2,
				HomeScore:  tt.homeScore,
				AwayScore:  tt.awayScore,
				Status:     tt.status,
			}

			err := game.Validate()

			if tt.wantErr {
				if !errors.Is(err, domain.ErrInvalidGame) {
					t.Fatalf("expected ErrInvalidGame, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected valid game, got %v", err)
			}
		})
	}
}

func TestGameValidateTeams(t *testing.T) {
	tests := []struct {
		name       string
		homeTeamID int64
		awayTeamID int64
	}{
		{"same team", 1, 1},
		{"zero home ID", 0, 2},
		{"negative home ID", -1, 2},
		{"zero away ID", 1, 0},
		{"negative away ID", 1, -2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := domain.Game{
				HomeTeamID: tt.homeTeamID,
				AwayTeamID: tt.awayTeamID,
				HomeScore:  score(110),
				AwayScore:  score(100),
				Status:     domain.GameFinished,
			}

			if err := game.Validate(); !errors.Is(err, domain.ErrInvalidGame) {
				t.Fatalf("expected ErrInvalidGame, got %v", err)
			}
		})
	}
}
