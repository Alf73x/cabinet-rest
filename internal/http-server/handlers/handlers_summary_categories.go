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

type ResponseSummaryCategories struct {
	resp.Response
	Data storage.TblSummaryCategories `json:"data"`
}

type ISummaryCategories interface {
	Db_GetSummaryCategories() (storage.TblSummaryCategories, error)
}

func NewSummaryCategories(log *slog.Logger, summaryCategoriesI ISummaryCategories) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewSummaryCategories"
		requestLog := log.With(slog.String("op", _FunctionName), slog.String("request_id", middleware.GetReqID(r.Context())))
		categories, err := summaryCategoriesI.Db_GetSummaryCategories()
		if err != nil {
			requestLog.Error("failed to load summary categories", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to load summary categories"))
			return
		}

		responseSummaryCategoriesOK(w, r, categories)
	}
}

func responseSummaryCategoriesOK(
	w http.ResponseWriter,
	r *http.Request,
	data storage.TblSummaryCategories,
) {
	render.JSON(w, r, ResponseSummaryCategories{
		Response: resp.OK(),
		Data:     data,
	})
}
