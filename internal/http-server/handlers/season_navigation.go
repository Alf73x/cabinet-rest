package handlers

import (
	resp "CabinetREST/internal/lib/api/response"
	"CabinetREST/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type ResponseSeasonNavigation struct {
	resp.Response
	ID int `json:"id"`
}

type ISeasonNavigation interface {
	Db_GetDestinationSeasonID(
		id int,
		direction string,
	) (int, error)
}

func NewSeasonNavigation(log *slog.Logger, seasonNavigationI ISeasonNavigation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewSeasonNavigation"

		requestLog := log.With(slog.String("op", _FunctionName), slog.String("request_id", middleware.GetReqID(r.Context())))

		sID := r.URL.Query().Get("id")
		if sID == "" {
			render.JSON(w, r, resp.Error("id is empty"))
			return
		}

		id, err := strconv.Atoi(sID)
		if err != nil {
			render.JSON(w, r, resp.Error("id must be integer"))
			return
		}

		direction := r.URL.Query().Get("direction")
		if direction != "prev" && direction != "next" {
			render.JSON(w, r, resp.Error("direction must be prev or next"))
			return
		}

		destID, err := seasonNavigationI.Db_GetDestinationSeasonID(id, direction)

		if err != nil {
			requestLog.Error("failed to get destination season", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to get destination season"))
			return
		}

		render.JSON(w, r, ResponseSeasonNavigation{Response: resp.OK(), ID: destID})
	}
}
