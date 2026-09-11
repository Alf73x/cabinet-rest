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

type ResponseSports struct {
	resp.Response
	Data []storage.TblSport `json:"list"`
}

type IGetSports interface {
	Db_GetSports() ([]storage.TblSport, error)
}

func NewSports(log *slog.Logger, getSportsI IGetSports) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewSport"

		requestLog := log.With(slog.String("op", _FunctionName),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		listSports, err := getSportsI.Db_GetSports()
		if err != nil {
			requestLog.Error("failed to load seasons", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to load seasons"))
			return
		}

		responseSportsOK(w, r, listSports)
	}
}

func responseSportsOK(w http.ResponseWriter, r *http.Request, data []storage.TblSport) {
	render.JSON(w, r, ResponseSports{
		Response: resp.OK(),
		Data:     data,
	})
}
