package domain

import (
	"errors"
	"fmt"
)

type GameStatus string

const (
	GameScheduled  GameStatus = "scheduled"
	GameInProgress GameStatus = "in_progress"
	GameFinished   GameStatus = "finished"
	GamePostponed  GameStatus = "postponed"
	GameCancelled  GameStatus = "cancelled"
)

var ErrInvalidGame = errors.New("invalid game")

type Game struct {
	HomeTeamID int64
	AwayTeamID int64
	HomeScore  *int
	AwayScore  *int
	Status     GameStatus
}

// Validate checks the rules that can be verified from the game fields.
func (g Game) Validate() error {
	if g.HomeTeamID <= 0 || g.AwayTeamID <= 0 {
		return fmt.Errorf("%w: team IDs must be positive", ErrInvalidGame)
	}

	if g.HomeTeamID == g.AwayTeamID {
		return fmt.Errorf("%w: teams must be different", ErrInvalidGame)
	}

	switch g.Status {
	case GameScheduled, GameInProgress, GameFinished,
		GamePostponed, GameCancelled:
	default:
		return fmt.Errorf("%w: unknown status %q", ErrInvalidGame, g.Status)
	}

	if g.HomeScore != nil && *g.HomeScore < 0 {
		return fmt.Errorf("%w: home score must not be negative", ErrInvalidGame)
	}

	if g.AwayScore != nil && *g.AwayScore < 0 {
		return fmt.Errorf("%w: away score must not be negative", ErrInvalidGame)
	}

	if g.Status == GameFinished {
		if g.HomeScore == nil || g.AwayScore == nil {
			return fmt.Errorf("%w: finished game must have both scores", ErrInvalidGame)
		}

		if *g.HomeScore == *g.AwayScore {
			return fmt.Errorf("%w: finished game must not be tied", ErrInvalidGame)
		}
	}

	return nil
}
