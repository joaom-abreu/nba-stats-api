package httpapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"nba-stats-api/internal/domain"
)

func listTeamsHandler(
	listTeams func(context.Context, domain.TeamFilter) ([]domain.Team, error),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := parseTeamFilter(r.URL.RawQuery)
		if err != nil {
			writeError(
				w, r,
				http.StatusBadRequest,
				"invalid_query",
				err.Error(),
			)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		teams, err := listTeams(ctx, filter)
		if err != nil {
			slog.Error(
				"failed to list teams",
				"request_id", requestID(r),
				"error", err,
			)

			writeError(
				w, r,
				http.StatusInternalServerError,
				"internal_error",
				"could not list teams",
			)
			return
		}

		if teams == nil {
			teams = make([]domain.Team, 0)
		}

		response := struct {
			Data []domain.Team  `json:"data"`
			Meta paginationMeta `json:"meta"`
		}{
			Data: teams,
			Meta: paginationMeta{
				Limit:    filter.Limit,
				Offset:   filter.Offset,
				Returned: len(teams),
			},
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func parseTeamFilter(rawQuery string) (domain.TeamFilter, error) {
	filter := domain.TeamFilter{
		Limit: 20,
	}

	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return filter, errors.New("invalid query parameters")
	}

	for name, values := range query {
		switch name {
		case "source", "conference", "limit", "offset":
		default:
			return filter, fmt.Errorf("unsupported query parameter: %s", name)
		}

		if len(values) != 1 {
			return filter, fmt.Errorf("provide %s only once", name)
		}
	}

	filter.Source = query.Get("source")

	if filter.Source != "demo" && filter.Source != "balldontlie" {
		return filter, errors.New("source must be demo or balldontlie")
	}

	if query.Has("conference") {
		filter.Conference = query.Get("conference")

		if filter.Conference != "east" && filter.Conference != "west" {
			return filter, errors.New("conference must be east or west")
		}
	}

	if query.Has("limit") {
		filter.Limit, err = strconv.Atoi(query.Get("limit"))

		if err != nil || filter.Limit < 1 || filter.Limit > 100 {
			return filter, errors.New("limit must be an integer between 1 and 100")
		}
	}

	if query.Has("offset") {
		filter.Offset, err = strconv.Atoi(query.Get("offset"))

		if err != nil || filter.Offset < 0 {
			return filter, errors.New("offset must be a non-negative integer")
		}
	}

	return filter, nil
}

func getTeamHandler(
	getTeam func(context.Context, int64) (domain.Team, error),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			writeError(
				w, r,
				http.StatusBadRequest,
				"invalid_id",
				"id must be a positive integer",
			)
			return
		}

		if r.URL.RawQuery != "" {
			writeError(
				w, r,
				http.StatusBadRequest,
				"invalid_query",
				"this endpoint does not accept query parameters",
			)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		team, err := getTeam(ctx, id)

		if errors.Is(err, domain.ErrTeamNotFound) {
			writeError(
				w, r,
				http.StatusNotFound,
				"team_not_found",
				"team not found",
			)
			return
		}

		if err != nil {
			slog.Error(
				"failed to get team",
				"request_id", requestID(r),
				"team_id", id,
				"error", err,
			)

			writeError(
				w, r,
				http.StatusInternalServerError,
				"internal_error",
				"could not get team",
			)
			return
		}

		response := struct {
			Data domain.Team `json:"data"`
		}{
			Data: team,
		}

		writeJSON(w, http.StatusOK, response)
	}
}
