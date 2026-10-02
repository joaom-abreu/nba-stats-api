package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nba-stats-api/internal/domain"
)

type gameQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

type GameRepository struct {
	db gameQuerier
}

func NewGameRepository(db gameQuerier) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) List(
	ctx context.Context,
	filter domain.GameFilter,
) ([]domain.Game, error) {
	const query = `
		SELECT
			id, source, external_id, season, phase,
			game_date, start_time,
			home_team_id, away_team_id,
			home_score, away_score, status, collected_at
		FROM games
		WHERE source = $1
		  AND ($2::integer = 0 OR season = $2)
		  AND ($3::text = '' OR phase = $3)
		  AND ($4::date IS NULL OR game_date >= $4)
		  AND ($5::date IS NULL OR game_date <= $5)
		  AND (
			  $6::bigint = 0
			  OR home_team_id = $6
			  OR away_team_id = $6
		  )
		  AND ($7::text = '' OR status = $7)
		ORDER BY game_date DESC, id DESC
		LIMIT $8 OFFSET $9
	`

	rows, err := r.db.Query(
		ctx,
		query,
		filter.Source,
		filter.Season,
		filter.Phase,
		filter.DateFrom,
		filter.DateTo,
		filter.TeamID,
		string(filter.Status),
		filter.Limit,
		filter.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query games: %w", err)
	}
	defer rows.Close()

	games := make([]domain.Game, 0)

	for rows.Next() {
		var game domain.Game
		var gameDate time.Time

		if err := rows.Scan(
			&game.ID,
			&game.Source,
			&game.ExternalID,
			&game.Season,
			&game.Phase,
			&gameDate,
			&game.StartTime,
			&game.HomeTeamID,
			&game.AwayTeamID,
			&game.HomeScore,
			&game.AwayScore,
			&game.Status,
			&game.CollectedAt,
		); err != nil {
			return nil, fmt.Errorf("scan game: %w", err)
		}

		game.GameDate = gameDate.Format(time.DateOnly)
		game.CollectedAt = game.CollectedAt.UTC()

		if game.StartTime != nil {
			startTime := game.StartTime.UTC()
			game.StartTime = &startTime
		}

		games = append(games, game)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate games: %w", err)
	}

	return games, nil
}
