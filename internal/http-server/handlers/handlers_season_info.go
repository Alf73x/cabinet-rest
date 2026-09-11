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

type ResponseSeasonInfo struct {
	resp.Response
	Data storage.SeasonInfo `json:"info"`
}

type IGetSeasonInfo interface {
	Db_GetSeasonInfo(id int) (storage.SeasonInfo, error)
}

func NewSeasonInfo(log *slog.Logger, getSeasonInfoI IGetSeasonInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewGetSeasonInfo"

		requestLog := log.With(slog.String("op", _FunctionName),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		sId := r.URL.Query().Get(Url_Info_ID)
		if sId == "" {
			s := Url_Info_ID + " is empty"
			requestLog.Info(s)
			render.JSON(w, r, resp.Error(s))
			return

		}
		id, err := strconv.Atoi(sId)
		if err != nil {
			s := Url_Info_ID + " must be integer"
			requestLog.Info(s)
			render.JSON(w, r, resp.Error(s))
			return
		}

		info, err := getSeasonInfoI.Db_GetSeasonInfo(id)
		if err != nil {
			requestLog.Error("failed to load season info", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to load season info"))
			return
		}

		responseSeasonInfoOK(w, r, info)
	}
}

func responseSeasonInfoOK(w http.ResponseWriter, r *http.Request, data storage.SeasonInfo) {
	render.JSON(w, r, ResponseSeasonInfo{
		Response: resp.OK(),
		Data:     data,
	})
}
