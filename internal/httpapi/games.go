package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"nba-stats-api/internal/domain"
)

func listGamesHandler(
	listGames func(context.Context, domain.GameFilter) ([]domain.Game, error),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := parseGameFilter(r.URL.RawQuery)
		if err != nil {
			writeError(
				w, r, http.StatusBadRequest,
				"invalid_query", err.Error(),
			)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		games, err := listGames(ctx, filter)
		if err != nil {
			slog.Error(
				"failed to list games",
				"request_id", requestID(r),
				"error", err,
			)

			writeError(
				w, r, http.StatusInternalServerError,
				"internal_error", "could not list games",
			)
			return
		}

		if games == nil {
			games = make([]domain.Game, 0)
		}

		response := struct {
			Data []domain.Game  `json:"data"`
			Meta paginationMeta `json:"meta"`
		}{
			Data: games,
			Meta: paginationMeta{
				Limit:    filter.Limit,
				Offset:   filter.Offset,
				Returned: len(games),
			},
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func parseGameFilter(rawQuery string) (domain.GameFilter, error) {
	filter := domain.GameFilter{
		Limit: 20,
	}

	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return filter, fmt.Errorf("invalid query parameters")
	}

	for name, values := range query {
		switch name {
		case "source", "season", "phase", "date_from", "date_to",
			"team_id", "status", "limit", "offset":
		default:
			return filter, fmt.Errorf(
				"unsupported query parameter: %s", name,
			)
		}

		if len(values) != 1 {
			return filter, fmt.Errorf("provide %s only once", name)
		}
	}

	switch query.Get("source") {
	case "demo", "balldontlie":
		filter.Source = query.Get("source")
	default:
		return filter, fmt.Errorf("source must be demo or balldontlie")
	}

	if query.Has("season") {
		season, err := strconv.ParseInt(query.Get("season"), 10, 32)
		if err != nil || season < 1 {
			return filter, fmt.Errorf(
				"season must be an integer between 1 and 2147483647",
			)
		}
		filter.Season = int(season)
	}

	if query.Has("phase") {
		switch query.Get("phase") {
		case "regular_season", "playoffs":
			filter.Phase = query.Get("phase")
		default:
			return filter, fmt.Errorf(
				"phase must be regular_season or playoffs",
			)
		}
	}

	filter.DateFrom, err = parseFilterDate(query, "date_from")
	if err != nil {
		return filter, err
	}

	filter.DateTo, err = parseFilterDate(query, "date_to")
	if err != nil {
		return filter, err
	}

	if filter.DateFrom != nil && filter.DateTo != nil {
		if filter.DateFrom.After(*filter.DateTo) {
			return filter, fmt.Errorf(
				"date_from must be before or equal to date_to",
			)
		}
	}

	if query.Has("team_id") {
		teamID, err := strconv.ParseInt(query.Get("team_id"), 10, 64)
		if err != nil || teamID < 1 {
			return filter, fmt.Errorf(
				"team_id must be a positive integer",
			)
		}
		filter.TeamID = teamID
	}

	if query.Has("status") {
		status := domain.GameStatus(query.Get("status"))

		switch status {
		case domain.GameScheduled,
			domain.GameInProgress,
			domain.GameFinished,
			domain.GamePostponed,
			domain.GameCancelled:
			filter.Status = status
		default:
			return filter, fmt.Errorf(
				"status must be scheduled, in_progress, finished, postponed or cancelled",
			)
		}
	}

	if query.Has("limit") {
		limit, err := strconv.Atoi(query.Get("limit"))
		if err != nil || limit < 1 || limit > 100 {
			return filter, fmt.Errorf(
				"limit must be an integer between 1 and 100",
			)
		}
		filter.Limit = limit
	}

	if query.Has("offset") {
		offset, err := strconv.Atoi(query.Get("offset"))
		if err != nil || offset < 0 {
			return filter, fmt.Errorf(
				"offset must be a non-negative integer",
			)
		}
		filter.Offset = offset
	}

	return filter, nil
}

func parseFilterDate(query url.Values, name string) (*time.Time, error) {
	if !query.Has(name) {
		return nil, nil
	}

	date, err := time.Parse(time.DateOnly, query.Get(name))
	if err != nil || date.Year() < 1 {
		return nil, fmt.Errorf(
			"%s must be a valid date in YYYY-MM-DD format", name,
		)
	}

	return &date, nil
}
