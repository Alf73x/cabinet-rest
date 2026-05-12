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

type ResponseTeams struct {
	resp.Response
	Data []storage.TblTeams `json:"list"`
}

type IGetTeams interface {
	Db_GetTeams(idTerritory int, idssport string) ([]storage.TblTeams, error)
}

func NewTeams(log *slog.Logger, getTeamsI IGetTeams) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewTeams"

		log = log.With(slog.String("op", _FunctionName),
			slog.String("request=id", middleware.GetReqID(r.Context())),
		)

		sTerritoryID := r.URL.Query().Get(Url_Teams_ID_Territory)
		if sTerritoryID == "" {
			s := Url_Teams_ID_Territory + " is empty"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return

		}
		idTerritory, err := strconv.Atoi(sTerritoryID)
		if err != nil {
			s := Url_Teams_ID_Territory + " must be integer"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return
		}

		idsSport, err := ParseSportIDs(r.URL.Query().Get("sport_ids"))
		if err != nil {
			log.Error(err.Error(), sl.Err(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		strIdsSport := SportIDsToString(idsSport)

		listTeams, err := getTeamsI.Db_GetTeams(idTerritory, strIdsSport)
		if err != nil {
			log.Error("failed to load teams", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to load teams"))
			return
		}

		responseTeamsOK(w, r, listTeams)
	}
}

func responseTeamsOK(w http.ResponseWriter, r *http.Request, data []storage.TblTeams) {
	render.JSON(w, r, ResponseTeams{
		Response: resp.OK(),
		Data:     data,
	})
}
