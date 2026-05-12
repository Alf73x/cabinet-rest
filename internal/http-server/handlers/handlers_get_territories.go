package handlers

import (
	resp "CabinetREST/internal/lib/api/response"
	"CabinetREST/internal/lib/logger/sl"
	"CabinetREST/internal/storage"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

/*
	type Request struct {
		Id int `json:"id_country" validate:"required,min=1,max=999999"`
	}
*/
type ResponseTerritories struct {
	resp.Response
	Data []storage.TblCountry `json:"list"`
}

type IGetDb_GetTerritories interface {
	Db_GetTerritories(parentID int) ([]storage.TblCountry, error)
}

type IGetDb_SearchTerritories interface {
	Db_SearchTerritories(filter string) ([]storage.TblCountry, error)
}

type IGetDb_PathTerritories interface {
	Db_PathTerritories(parentID int) ([]int, error)
}

func NewTerritoryChildren(log *slog.Logger, getTerritoriesI IGetDb_GetTerritories) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewTerritoryChildren"

		log = log.With(slog.String("op", _FunctionName),
			slog.String("request=id", middleware.GetReqID(r.Context())),
		)

		/*
			var req Request
			err := render.DecodeJSON(r.Body, &req)
			if errors.Is(err, io.EOF) { // empty body
				log.Error("request body is empty")
				render.JSON(w, r, resp.Error("empty request"))
				return
			}
			if err != nil {
				log.Error("failed to decode request body", sl.Err(err))
				render.JSON(w, r, resp.Error("failed to decode request"))
				return
			}
			log.Info("request body decoded", slog.Any("request", req))

			if err := validator.New().Struct(req); err != nil {
				validateErr := err.(validator.ValidationErrors)
				log.Error("invalid request", sl.Err(err))
				render.JSON(w, r, resp.ValidationError(validateErr))
				return
			}
			listCountries, err := getCountriesI.Db_GetCountries(req.Id)
		*/

		sId := chi.URLParam(r, Url_Territories_ID)
		if sId == "" {
			s := Url_Territories_ID + " is empty"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return

		}
		id, err := strconv.Atoi(sId)
		if err != nil {
			s := Url_Territories_ID + " must be integer"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
		}

		listItems, err := getTerritoriesI.Db_GetTerritories(id)
		/* 		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("url already exists"))

			return
		} */
		if err != nil {
			s := "failed to load tree items"
			log.Error(s, sl.Err(err))
			render.JSON(w, r, resp.Error(s))
			return
		}

		responseTerritoriesOK(w, r, listItems)
	}
}

func NewTerritorySearch(log *slog.Logger, searchTerritoriesI IGetDb_SearchTerritories) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewTerritorySearch"

		log = log.With(slog.String("op", _FunctionName),
			slog.String("request=id", middleware.GetReqID(r.Context())),
		)

		sFilter := strings.TrimSpace(
			r.URL.Query().Get("filter"),
		)
		if sFilter == "" {
			s := "filter is empty"
			log.Info(s)

			render.JSON(w, r, resp.Error(s))
			return
		}

		listItems, err := searchTerritoriesI.Db_SearchTerritories(sFilter)
		if err != nil {
			s := "failed to load tree items"
			log.Error(s, sl.Err(err))
			render.JSON(w, r, resp.Error(s))
			return
		}

		responseTerritoriesOK(w, r, listItems)
	}
}

func NewTerritoryPath(log *slog.Logger, pathTerritoriesI IGetDb_PathTerritories) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewTerritoryPath"

		log = log.With(slog.String("op", _FunctionName),
			slog.String("request=id", middleware.GetReqID(r.Context())),
		)

		sId := strings.TrimSpace(
			r.URL.Query().Get(Url_Territories_ID),
		)

		if sId == "" {
			s := Url_Territories_ID + " is empty"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return

		}
		id, err := strconv.Atoi(sId)
		if err != nil {
			s := Url_Territories_ID + " must be integer"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
		}

		listItems, err := pathTerritoriesI.Db_PathTerritories(id)
		if err != nil {
			s := "failed to find path"
			log.Error(s, sl.Err(err))
			render.JSON(w, r, resp.Error(s))
			return
		}
		render.JSON(w, r, listItems)
	}
}

func responseTerritoriesOK(w http.ResponseWriter, r *http.Request, data []storage.TblCountry) {
	render.JSON(w, r, ResponseTerritories{
		Response: resp.OK(),
		Data:     data,
	})
}
