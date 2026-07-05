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

type TeamInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ResponseTeam struct {
	resp.Response
	Team TeamInfo          `json:"team"`
	Data []storage.TblTeam `json:"list"`
}

type IGetTeam interface {
	Db_GetTeam(id int) ([]storage.TblTeam, error)
	DB_GetTeamName(id int, mode int) (string, error)
}

func NewTeam(log *slog.Logger, getTeamsI IGetTeam) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewTeam"

		log = log.With(slog.String("op", _FunctionName),
			slog.String("request=id", middleware.GetReqID(r.Context())),
		)

		sIDTeam := r.URL.Query().Get(Url_Team_ID)
		if sIDTeam == "" {
			s := Url_Team_ID + " is empty"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return

		}
		idTeam, err := strconv.Atoi(sIDTeam)
		if err != nil {
			s := Url_Team_ID + " must be integer"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return
		}

		listTeams, err := getTeamsI.Db_GetTeam(idTeam)
		if err != nil {
			log.Error("failed to load team", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to load team"))
			return
		}

		name, err := getTeamsI.DB_GetTeamName(idTeam, 0)
		if err != nil {
			log.Error("failed to load team", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to load team"))
			return
		}

		responseTeamOK(w, r, listTeams, idTeam, name)
	}
}

func responseTeamOK(w http.ResponseWriter, r *http.Request, data []storage.TblTeam, id int, name string) {
	render.JSON(w, r, ResponseTeam{
		Response: resp.OK(),
		Team: TeamInfo{
			ID:   id,
			Name: name,
		},
		Data: data,
	})
}
