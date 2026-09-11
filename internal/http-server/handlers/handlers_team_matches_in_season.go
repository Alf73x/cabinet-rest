package handlers

import (
	resp "CabinetREST/internal/lib/api/response"
	"CabinetREST/internal/lib/logger/sl"
	"CabinetREST/internal/storage"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
)

type ResponseTeamMatches struct {
	resp.Response
	Data []storage.TblTeamMatches `json:"list"`
}

type IGetTeamMatches interface {
	Db_GetTeamMatches(idTeam int, idSeason int) ([]storage.TblTeamMatches, error)
}

func NewTeamMatches(log *slog.Logger, getTeamMatchesI IGetTeamMatches) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewTeamMatches"

		sIDTeam := r.URL.Query().Get(Url_Team_Matches_ID_Team)
		if sIDTeam == "" {
			s := Url_Team_Matches_ID_Team + " is empty"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return

		}
		idTeam, err := strconv.Atoi(sIDTeam)
		if err != nil {
			s := Url_Team_Matches_ID_Team + " must be integer"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return
		}

		sIDSeason := r.URL.Query().Get(Url_Team_Matches_ID_Season)
		if sIDSeason == "" {
			s := Url_Team_Matches_ID_Season + " is empty"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return

		}
		idSeason, err := strconv.Atoi(sIDSeason)
		if err != nil {
			s := Url_Team_Matches_ID_Season + " must be integer"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return
		}

		listTeams, err := getTeamMatchesI.Db_GetTeamMatches(idTeam, idSeason)
		if err != nil {
			log.Error("failed to load teams", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to load team matches"))
			return
		}

		responseTeamsMatchesOk(w, r, listTeams)
	}
}

func responseTeamsMatchesOk(w http.ResponseWriter, r *http.Request, data []storage.TblTeamMatches) {
	render.JSON(w, r, ResponseTeamMatches{
		Response: resp.OK(),
		Data:     data,
	})
}
