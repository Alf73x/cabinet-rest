package handlers

import (
	resp "CabinetREST/internal/lib/api/response"
	"CabinetREST/internal/storage"
	"CabinetREST/internal/storage/sqlite"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
)

type ResponseTournament_Matrix struct {
	resp.Response
	DataType       int                        `json:"datatype"`
	TableFormat    int                        `json:"tableFormat"`
	ResultOf       int                        `json:"resultOf"`
	Points         storage.Points             `json:"points"`
	RoundStandings string                     `json:"roundStandings"`
	InfoText       string                     `json:"infoText"`
	CommentText    string                     `json:"commentText"`
	Data           []storage.TournamentMatrix `json:"list"`
}

type ResponseTournament_Cup struct {
	resp.Response
	DataType    int                     `json:"datatype"`
	InfoText    string                  `json:"infoText"`
	CommentText string                  `json:"commentText"`
	Data        []storage.TournamentCup `json:"list"`
}
type ResponseTournament_PainText struct {
	resp.Response
	DataType    int                           `json:"datatype"`
	InfoText    string                        `json:"infoText"`
	CommentText string                        `json:"commentText"`
	Data        []storage.TournamentPlainText `json:"list"`
}

type IGetTournament_Matrix interface {
	Db_GetSeasons(idsSport string, filterSeason string, filterName string) ([]storage.TblSeason, error)
}

func NewTournament(log *slog.Logger, s *sqlite.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const _FunctionName = "handlers.NewTournament"

		sID := r.URL.Query().Get(Url_Tournament_ID)
		if sID == "" {
			s := Url_Tournament_ID + " is empty"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return

		}
		id, err := strconv.Atoi(sID)
		if err != nil {
			s := Url_Tournament_ID + " must be integer"
			log.Info(s)
			render.JSON(w, r, resp.Error(s))
			return
		}

		info, err := s.DB_GetSeasonVariables(id)

		if info.ViewOpt == storage.KViewOption_Text {
			plain, err := s.ShowDataPlain(id)
			if err != nil {
				log.Info(err.Error())
				render.JSON(w, r, resp.Error(err.Error()))
				return
			}
			responseTournamentPlainOK(w, r, plain, info)
		} else if sqlite.Sport_IsTable(info.Rank) || info.ViewOpt == storage.KViewOption_Table || info.ResultOf > 0 {
			matrix, info, err := s.ShowData_Table(id)
			if err != nil {
				log.Info(err.Error())
				render.JSON(w, r, resp.Error(err.Error()))
				return
			}
			responseTournamentMatrixOK(w, r, matrix, info)
		} else if sqlite.Sport_IsList(info.Rank) || info.ViewOpt == storage.KViewOption_Cup {
			cup, err := s.ShowData_Cup(id)
			if err != nil {
				log.Info(err.Error())
				render.JSON(w, r, resp.Error(err.Error()))
				return
			}
			responseTournamentCupOK(w, r, cup, info)
		}

	}
}

func responseTournamentMatrixOK(w http.ResponseWriter, r *http.Request, data []storage.TournamentMatrix, info storage.TournamentInfo) {
	render.JSON(w, r, ResponseTournament_Matrix{
		Response:       resp.OK(),
		DataType:       1,
		TableFormat:    info.TableFormat,
		ResultOf:       info.ResultOf,
		Points:         info.Points,
		RoundStandings: info.RoundStandings,
		InfoText:       info.PlainText,
		CommentText:    info.RemarkText,
		Data:           data,
	})
}
func responseTournamentCupOK(w http.ResponseWriter, r *http.Request, data []storage.TournamentCup, info storage.TournamentInfo,
) {
	render.JSON(w, r, ResponseTournament_Cup{
		Response:    resp.OK(),
		DataType:    2,
		InfoText:    info.PlainText,
		CommentText: info.RemarkText,
		Data:        data,
	})
}
func responseTournamentPlainOK(w http.ResponseWriter, r *http.Request, data []storage.TournamentPlainText, info storage.TournamentInfo) {
	render.JSON(w, r, ResponseTournament_PainText{
		Response:    resp.OK(),
		DataType:    3,
		InfoText:    info.PlainText,
		CommentText: info.RemarkText,
		Data:        data,
	})
}
