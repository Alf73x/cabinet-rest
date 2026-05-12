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

type ResponseSeasons struct {
	resp.Response
	Data []storage.TblSeason `json:"list"`
}

type IGetSeasons interface {
	Db_GetSeasons(idsSport string, filterSeason string, filterName string) ([]storage.TblSeason, error)
}

func NewSeasons(log *slog.Logger, getSeasonsI IGetSeasons) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewSeasons"

		log = log.With(slog.String("op", _FunctionName),
			slog.String("request=id", middleware.GetReqID(r.Context())),
		)

		idsSport, err := ParseSportIDs(r.URL.Query().Get("sport_ids"))
		if err != nil {
			log.Error(err.Error(), sl.Err(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		strIdsSport := SportIDsToString(idsSport)

		seasonFilter := r.URL.Query().Get(Url_Seasons_Season_Filter)
		nameFilter := r.URL.Query().Get(Url_Seasons_Name_Filter)

		listSeasons, err := getSeasonsI.Db_GetSeasons(strIdsSport, seasonFilter, nameFilter)
		if err != nil {
			log.Error("failed to load seasons", sl.Err(err))
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
