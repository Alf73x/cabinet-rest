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

type ResponseTeamInfo struct {
	resp.Response
	Data storage.TeamsInfo `json:"info"`
}

type IGetTeamInfo interface {
	Db_GetTeamInfo(id int) (storage.TeamsInfo, error)
}

func NewTeamInfo(log *slog.Logger, getTeamInfoI IGetTeamInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewTeamInfo"

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
		}

		info, err := getTeamInfoI.Db_GetTeamInfo(id)
		if err != nil {
			requestLog.Error("failed to load team info", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to load team info"))
			return
		}

		responseTeamInfoOK(w, r, info)
	}
}

func responseTeamInfoOK(w http.ResponseWriter, r *http.Request, data storage.TeamsInfo) {
	render.JSON(w, r, ResponseTeamInfo{
		Response: resp.OK(),
		Data:     data,
	})
}
