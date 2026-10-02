package httpapi_test

import (
	"context"

	"nba-stats-api/internal/domain"
)

type teamReaderStub struct {
	list    func(context.Context, domain.TeamFilter) ([]domain.Team, error)
	getByID func(context.Context, int64) (domain.Team, error)
}

func (s teamReaderStub) List(
	ctx context.Context,
	filter domain.TeamFilter,
) ([]domain.Team, error) {
	if s.list != nil {
		return s.list(ctx, filter)
	}

	return emptyTeamList(ctx, filter)
}

func (s teamReaderStub) GetByID(
	ctx context.Context,
	id int64,
) (domain.Team, error) {
	if s.getByID != nil {
		return s.getByID(ctx, id)
	}

	return domain.Team{}, domain.ErrTeamNotFound
}
