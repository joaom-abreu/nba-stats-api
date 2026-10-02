package httpapi_test

import (
	"context"
	"net/http"

	"nba-stats-api/internal/domain"
	"nba-stats-api/internal/httpapi"
)

type gameReaderStub struct {
	list func(context.Context, domain.GameFilter) ([]domain.Game, error)
}

func (s gameReaderStub) List(
	ctx context.Context,
	filter domain.GameFilter,
) ([]domain.Game, error) {
	if s.list != nil {
		return s.list(ctx, filter)
	}

	return nil, nil
}

func newTestHandler(
	pingDatabase func(context.Context) error,
	teams httpapi.TeamReader,
) http.Handler {
	return httpapi.NewHandler(
		pingDatabase,
		teams,
		gameReaderStub{},
	)
}
