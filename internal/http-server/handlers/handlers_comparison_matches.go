package handlers

import (
	resp "CabinetREST/internal/lib/api/response"
	"CabinetREST/internal/lib/logger/sl"
	"CabinetREST/internal/storage"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type ResponseComparisonMatches struct {
	resp.Response
	Data []storage.TblTeamMatches `json:"data"`
}

type IComparisonMatches interface {
	Db_GetComparisonMatches(team1ID int, team2ID int) ([]storage.TblTeamMatches, error)
}

func NewComparisonMatches(log *slog.Logger, comparisonMatchesI IComparisonMatches) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewComparisonMatches"

		log = log.With(
			slog.String("op", _FunctionName),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		team1ID, err := strconv.Atoi(r.URL.Query().Get(Url_ComparisonMatches_Team1ID))
		if err != nil || team1ID <= 0 {
			err := errInvalidComparisonParameter(Url_ComparisonMatches_Team1ID)

			log.Error(err.Error(), sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		team2ID, err := strconv.Atoi(r.URL.Query().Get(Url_ComparisonMatches_Team2ID))
		if err != nil || team2ID <= 0 {
			err := errInvalidComparisonParameter(Url_ComparisonMatches_Team2ID)

			log.Error(err.Error(), sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		matches, err := comparisonMatchesI.Db_GetComparisonMatches(team1ID, team2ID)
		if err != nil {
			log.Error("failed to load comparison matches", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to load comparison matches"))
			return
		}

		responseComparisonMatchesOK(w, r, matches)
	}
}

func responseComparisonMatchesOK(w http.ResponseWriter, r *http.Request, data []storage.TblTeamMatches) {
	render.JSON(w, r, ResponseComparisonMatches{
		Response: resp.OK(),
		Data:     data,
	})
}
