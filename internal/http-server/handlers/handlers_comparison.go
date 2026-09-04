package handlers

import (
	resp "CabinetREST/internal/lib/api/response"
	"CabinetREST/internal/lib/logger/sl"
	"CabinetREST/internal/storage"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type ComparisonOpponent struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

type ResponseComparison struct {
	resp.Response
	Data storage.TblComparison `json:"data"`
}

type IComparison interface {
	Db_GetComparison(opponent1Type string, opponent1ID int, opponent2Type string, opponent2ID int, competitionFilter string, sportIDs []int, leagueRanks []int) (storage.TblComparison, error)
}

func NewComparison(log *slog.Logger, comparisonI IComparison) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewComparison"

		requestLog := log.With(slog.String("op", _FunctionName), slog.String("request_id", middleware.GetReqID(r.Context())))

		opponent1Type := r.URL.Query().Get(Url_Comparison_opponent1Type)
		opponent2Type := r.URL.Query().Get(Url_Comparison_opponent2Type)
		opponent1ID, err := strconv.Atoi(r.URL.Query().Get(Url_Comparison_opponent1Id))
		if err != nil || opponent1ID <= 0 {
			err := errInvalidComparisonParameter(Url_Comparison_opponent1Id)

			requestLog.Error(err.Error(), sl.Err(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		opponent2ID, err := strconv.Atoi(r.URL.Query().Get(Url_Comparison_opponent2Id))
		if err != nil || opponent2ID <= 0 {
			err := errInvalidComparisonParameter(Url_Comparison_opponent2Id)

			requestLog.Error(err.Error(), sl.Err(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		if !isValidOpponentType(opponent1Type) {
			err := errInvalidComparisonParameter(Url_Comparison_opponent1Type)

			requestLog.Error(err.Error(), sl.Err(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		if !isValidOpponentType(opponent2Type) {
			err := errInvalidComparisonParameter(Url_Comparison_opponent2Type)

			requestLog.Error(err.Error(), sl.Err(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		competitionFilter := r.URL.Query().Get(Url_Comparison_competitionFilter)
		if competitionFilter == "" {
			competitionFilter = "all"
		}

		sportIDs, err := ParseSportIDs(
			r.URL.Query().Get(Url_Comparison_IDs_Sport),
		)
		if err != nil {
			requestLog.Error("invalid sport_ids", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid sport_ids"))
			return
		}

		leagueRanks := []int{}

		leagueRanksText := r.URL.Query().Get(Url_Comparison_LeagueRanks)
		if leagueRanksText != "" {
			leagueRanks, err = ParseSportIDs(leagueRanksText)
			if err != nil {
				requestLog.Error("invalid league_ranks", sl.Err(err))
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("invalid league_ranks"))
				return
			}
		}

		comparison, err := comparisonI.Db_GetComparison(opponent1Type, opponent1ID, opponent2Type, opponent2ID, competitionFilter, sportIDs, leagueRanks)
		if err != nil {
			requestLog.Error("failed to load comparison", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to load comparison"))
			return
		}

		responseComparisonOK(w, r, comparison)
	}
}
func isValidOpponentType(value string) bool {
	return value == OpponentTypeTerritory || value == OpponentTypeTeam
}

func errInvalidComparisonParameter(name string) error {
	return fmt.Errorf("invalid comparison parameter: %s", name)
}

func responseComparisonOK(w http.ResponseWriter, r *http.Request, data storage.TblComparison) {
	render.JSON(w, r, ResponseComparison{
		Response: resp.OK(),
		Data:     data,
	})
}
