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

type ResponseSeasons struct {
	resp.Response
	Data []storage.TblSeason `json:"list"`
}

type ResponseSeasonNames struct {
	resp.Response
	Data []string `json:"list"`
}

type IGetSeasons interface {
	Db_GetSeasons(idsSport string, filterSeason string, filterName string) ([]storage.TblSeason, error)
	Db_GetSeasonNames(idsSport string) ([]string, error)
	Db_GetSeasonByID(id int) (storage.TblSeason, error)
}

func NewSeasons(log *slog.Logger, getSeasonsI IGetSeasons) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewSeasons"

		requestLog := log.With(
			slog.String("op", _FunctionName),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		/*
		 * Получение одного турнира/сезона по ID.
		 *
		 * Например:
		 * GET /seasons?id=300
		 */
		sID := r.URL.Query().Get(Url_Seasons_ID)
		if sID != "" {
			id, err := strconv.Atoi(sID)
			if err != nil {
				requestLog.Info(Url_Seasons_ID + " must be integer")
				render.JSON(w, r, resp.Error(Url_Seasons_ID+" must be integer"))
				return
			}

			season, err := getSeasonsI.Db_GetSeasonByID(id)

			if err != nil {
				requestLog.Error("failed to load season by id", sl.Err(err))
				render.JSON(w, r, resp.Error("failed to load season"))
				return
			}

			/*
			 * Возвращаем тот же формат list,
			 * что и Db_GetSeasons.
			 *
			 * Это позволяет frontend использовать
			 * один ApiResponse<SeasonItem>.
			 */
			responseSeasonsOK(w, r, []storage.TblSeason{season})
			return
		}

		idsSport, err := ParseSportIDs(r.URL.Query().Get(Url_Seasons_IDs_Sport))
		if err != nil {
			requestLog.Error(err.Error(), sl.Err(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		strIdsSport := SportIDsToString(idsSport)

		namesOnly := r.URL.Query().Get(Url_Seasons_Names) == "1"

		if namesOnly {
			// Если names=1 — возвращаем только список названий сезонов
			listSeasonNames, err := getSeasonsI.Db_GetSeasonNames(strIdsSport)
			if err != nil {
				requestLog.Error("failed to load season names", sl.Err(err))
				render.JSON(w, r, resp.Error("failed to load season names"))
				return
			}

			responseSeasonNamesOK(w, r, listSeasonNames)
			return
		}

		// Обычный режим — возвращаем турниры
		seasonFilter := r.URL.Query().Get(Url_Seasons_Season_Filter)
		nameFilter := r.URL.Query().Get(Url_Seasons_Name_Filter)

		listSeasons, err := getSeasonsI.Db_GetSeasons(
			strIdsSport,
			seasonFilter,
			nameFilter,
		)
		if err != nil {
			requestLog.Error("failed to load seasons", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to load seasons"))
			return
		}

		responseSeasonsOK(w, r, listSeasons)
	}
}

func responseSeasonsOK(w http.ResponseWriter, r *http.Request, data []storage.TblSeason) {
	render.JSON(w, r, ResponseSeasons{
		Response: resp.OK(),
		Data:     data,
	})
}

func responseSeasonNamesOK(w http.ResponseWriter, r *http.Request, data []string) {
	render.JSON(w, r, ResponseSeasonNames{
		Response: resp.OK(),
		Data:     data,
	})
}
