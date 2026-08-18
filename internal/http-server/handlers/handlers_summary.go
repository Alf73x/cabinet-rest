package handlers

import (
	resp "CabinetREST/internal/lib/api/response"
	"CabinetREST/internal/lib/logger/sl"
	"CabinetREST/internal/storage"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type ResponseSummaryTable struct {
	resp.Response
	Data storage.TblSummaryTable `json:"data"`
}

type ISummaryTable interface {
	Db_GetSummaryTable(
		category string,
		leagueRanks []int,
		yearFrom string,
		yearTo string,
		sportIDs []int,
	) (storage.TblSummaryTable, error)
}

func NewSummaryTable(log *slog.Logger, summaryTableI ISummaryTable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewSummaryTable"

		log = log.With(
			slog.String("op", _FunctionName),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		category := r.URL.Query().Get(Url_SummaryTables_Category)

		leagueRanks := []int{}

		leagueRanksText := r.URL.Query().Get(Url_SummaryTables_LeagueRanks)
		if leagueRanksText != "" {
			var err error

			leagueRanks, err = ParseLeagueRanks(leagueRanksText)
			if err != nil {
				log.Error("invalid league_ranks", sl.Err(err))
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("invalid league_ranks"))
				return
			}
		}

		yearFrom := r.URL.Query().Get(Url_SummaryTables_YearFrom)
		yearTo := r.URL.Query().Get(Url_SummaryTables_YearTo)

		if yearFrom != "" && yearTo != "" && yearFrom > yearTo {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("year_from must not exceed year_to"))
			return
		}

		sportIDs, err := ParseSportIDs(
			r.URL.Query().Get(Url_SummaryTables_SportIDs),
		)
		if err != nil {
			log.Error("invalid sport_ids", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid sport_ids"))
			return
		}

		data, err := summaryTableI.Db_GetSummaryTable(
			category,
			leagueRanks,
			yearFrom,
			yearTo,
			sportIDs,
		)
		if err != nil {
			log.Error("failed to load summary table", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to load summary table"))
			return
		}

		responseSummaryTableOK(w, r, data)
	}
}

func responseSummaryTableOK(
	w http.ResponseWriter,
	r *http.Request,
	data storage.TblSummaryTable,
) {
	render.JSON(w, r, ResponseSummaryTable{
		Response: resp.OK(),
		Data:     data,
	})
}
