package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"nba-stats-api/internal/domain"
)

type teamQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type TeamRepository struct {
	db teamQuerier
}

func NewTeamRepository(db teamQuerier) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) List(
	ctx context.Context,
	filter domain.TeamFilter,
) ([]domain.Team, error) {
	const query = `
		SELECT id, source, external_id, name, abbreviation, conference
		FROM teams
		WHERE source = $1
		  AND ($2::text = '' OR conference = $2)
		ORDER BY id ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(
		ctx,
		query,
		filter.Source,
		filter.Conference,
		filter.Limit,
		filter.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query teams: %w", err)
	}
	defer rows.Close()

	teams := make([]domain.Team, 0)

	for rows.Next() {
		var team domain.Team

		if err := rows.Scan(
			&team.ID,
			&team.Source,
			&team.ExternalID,
			&team.Name,
			&team.Abbreviation,
			&team.Conference,
		); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}

		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate teams: %w", err)
	}

	return teams, nil
}

func (r *TeamRepository) GetByID(
	ctx context.Context,
	id int64,
) (domain.Team, error) {
	const query = `
		SELECT id, source, external_id, name, abbreviation, conference
		FROM teams
		WHERE id = $1
	`

	var team domain.Team

	err := r.db.QueryRow(ctx, query, id).Scan(
		&team.ID,
		&team.Source,
		&team.ExternalID,
		&team.Name,
		&team.Abbreviation,
		&team.Conference,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Team{}, domain.ErrTeamNotFound
	}

	if err != nil {
		return domain.Team{}, fmt.Errorf("get team: %w", err)
	}

	return team, nil
}
