package handlers

import (
	resp "CabinetREST/internal/lib/api/response"
	"CabinetREST/internal/storage"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/render"
)

type ResponseOpponentOptions struct {
	resp.Response
	Cities []storage.OpponentCity `json:"cities"`
	Teams  []storage.OpponentTeam `json:"teams"`
}

type IGetOpponentOptions interface {
	Db_GetOpponentOptions(
		sportIDs []int,
	) (
		[]storage.OpponentCity,
		[]storage.OpponentTeam,
		error,
	)
}

func NewOpponentOptions(log *slog.Logger, getOpponentOptionsI IGetOpponentOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewOpponentOptions"

		sportIDs, err := parseOpponentSportIDs(r.URL.Query().Get(Url_OpponentOptions_SportIDs))
		if err != nil {
			log.Info(err.Error(), slog.String("function", _FunctionName))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		listCities, listTeams, err := getOpponentOptionsI.Db_GetOpponentOptions(sportIDs)
		if err != nil {
			log.Info(
				err.Error(),
				slog.String("function", _FunctionName),
			)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		responseOpponentOptionsOK(w, r, listCities, listTeams)
	}
}

func parseOpponentSportIDs(value string) ([]int, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return []int{}, nil
	}

	parts := strings.Split(value, ",")
	result := make([]int, 0, len(parts))
	used := make(map[int]struct{}, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "" {
			continue
		}

		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, &OpponentSportIDError{
				Value: part,
			}
		}

		if id <= 0 {
			return nil, &OpponentSportIDError{
				Value: part,
			}
		}

		if _, exists := used[id]; exists {
			continue
		}

		used[id] = struct{}{}
		result = append(result, id)
	}

	return result, nil
}

type OpponentSportIDError struct {
	Value string
}

func (e *OpponentSportIDError) Error() string {
	return Url_OpponentOptions_SportIDs + " must contain positive integers: " + e.Value
}

func responseOpponentOptionsOK(w http.ResponseWriter, r *http.Request, cities []storage.OpponentCity, teams []storage.OpponentTeam) {
	if cities == nil {
		cities = make([]storage.OpponentCity, 0)
	}

	if teams == nil {
		teams = make([]storage.OpponentTeam, 0)
	}

	render.JSON(w, r, ResponseOpponentOptions{Response: resp.OK(), Cities: cities, Teams: teams})
}
