package sqlite

import (
	"CabinetREST/internal/storage"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

/*********************************************************************
  Db_GetTeams for selected territory
**********************************************************************/

func (s *Storage) Db_GetTeams(idTerritory int, idssport string) ([]storage.TblTeams, error) {
	const _FunctionName = "storage.sqlite.Db_GetTeams"

	sFilter := s.Tree_TreeChildrenFilter(idTerritory, "")
	sSQL := "WITH CTE_Teams AS ( "
	sSQL = sSQL + "   SELECT id FROM " + storage.Tbl_class_team + " WHERE " + sFilter
	sSQL = sSQL + ")"
	sSQL = sSQL + " SELECT "
	sSQL = sSQL + " tm." + storage.Fld_common_id + " teamid, "
	sSQL = sSQL + " s." + storage.Fld_common_id + ", "
	sSQL = sSQL + " s." + storage.Fld_common_id_base + ", "
	sSQL = sSQL + " IFNULL(s." + storage.Fld_class_season_season + ", ''), "
	sSQL = sSQL + " s." + storage.Fld_common_name + ", "
	sSQL = sSQL + " tm." + storage.Fld_common_name + " team, "
	sSQL = sSQL + " c." + storage.Fld_common_name + " ctr, "
	sSQL = sSQL + " IFNULL(s." + storage.Fld_common_group_id + ", 0) grp, "
	sSQL = sSQL + " IFNULL(s." + storage.Fld_class_season_league_rank + ", 0), "
	sSQL = sSQL + " IFNULL(t." + storage.Fld_common_sport_place + ", 0), "
	sSQL = sSQL + " IFNULL(t." + storage.Fld_sport_stage_index + ", 0), "
	sSQL = sSQL + " t." + storage.Fld_sport_wins + ", "
	sSQL = sSQL + " t." + storage.Fld_sport_wins_et + ", "
	sSQL = sSQL + " t." + storage.Fld_sport_draws + ", "
	sSQL = sSQL + " t." + storage.Fld_sport_losses_et + ", "
	sSQL = sSQL + " t." + storage.Fld_sport_losses + ", "
	sSQL = sSQL + " t." + storage.Fld_sport_goals_for + ", "
	sSQL = sSQL + " t." + storage.Fld_sport_goals_against + ", "
	sSQL = sSQL + " IFNULL(s." + storage.Fld_common_winner_id + ", 0), "
	sSQL = sSQL + " IFNULL(s." + storage.Fld_class_season_options_1 + ", '') "
	sSQL = sSQL + " FROM " + storage.Tbl_class_season + " s "
	sSQL = sSQL + " LEFT JOIN " + storage.Tbl_sport_tables + " t ON t." + storage.Fld_common_id_season + "=s." + storage.Fld_common_id + " AND t." + storage.Fld_common_id_team + " IN (SELECT * FROM CTE_Teams)"
	sSQL = sSQL + " LEFT JOIN " + storage.Tbl_class_team + " tm ON tm." + storage.Fld_common_id + "=t." + storage.Fld_common_id_team
	sSQL = sSQL + " LEFT JOIN " + storage.Tbl_countries + " c ON c." + storage.Fld_common_id + "=tm." + storage.Fld_common_id_country
	sSQL = sSQL + " WHERE tm." + storage.Fld_common_id + " IN (SELECT * FROM CTE_Teams) "
	sSQL = sSQL + " AND " + GetBaseFilter(idssport, "s.")
	sSQL = sSQL + " ORDER BY " + storage.Fld_class_season_season + " DESC, s." + storage.Fld_common_sort_order + " DESC "

	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}
	defer rows.Close()

	var teams []storage.TblTeams
	for rows.Next() {
		var team storage.TblTeams
		err := rows.Scan(
			&team.ID,
			&team.SeasonID,
			&team.SportID,
			&team.Season,
			&team.SeasonName,
			&team.TeamName,
			&team.TeamTerritory,
			&team.GroupID,
			&team.LeagueRank,
			&team.Place,
			&team.StageIndex,
			&team.Wins,
			&team.WinsET,
			&team.Draws,
			&team.LossesET,
			&team.Losses,
			&team.Goals_For,
			&team.Goals_Against,
			&team.WinnerID,
			&team.Options,
		)
		if err != nil {
			return nil, err
		}

		if strings.Contains(strings.ToLower(team.Options), strings.ToLower(storage.KOptionsResultsOf)) {
			if pos := strings.Index(team.SeasonName, "."); pos >= 0 {
				team.SeasonName = strings.TrimSpace(team.SeasonName[:pos])
			}
		}
		team.Place, err = s.GetPlaceAsStr(team.SeasonID, team.StageIndex, team.Place)
		leagueRank, err := strconv.Atoi(team.LeagueRank)
		if err != nil {
			leagueRank = 0
		}
		team.LeagueRank = s.GetSportRankAsInt(leagueRank)
		teams = append(teams, team)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

func (s *Storage) DB_GetTeamName(id int, mode int) (string, error) {
	result := ""
	str := "SELECT " +
		"IFNULL(t." + storage.Fld_common_name + ",'') tm, " +
		"IFNULL(c." + storage.Fld_common_name + ",'') cn " +
		"FROM " + storage.Tbl_class_team + " t " +
		"LEFT JOIN " + storage.Tbl_countries + " c ON t." +
		storage.Fld_common_id_country + "=c." + storage.Fld_common_id + " " +
		"WHERE t." + storage.Fld_common_id + "=" + strconv.Itoa(id)
	row := s.db.QueryRow(str)
	var teamName string
	var countryName string
	err := row.Scan(&teamName, &countryName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	switch mode {
	case 0:
		result = teamName + " " + countryName
	case 1:
		result = teamName
	case 2:
		result = countryName
	}
	return result, nil
}

/*********************************************************************
  Db_GetTeam
**********************************************************************/

func (s *Storage) Db_GetTeam(id int) ([]storage.TblTeam, error) {
	const _FunctionName = "storage.sqlite.Db_GetTeam"

	sl, err := s.GetTeamTreeIDs(id)
	var sIDs string
	for i := 0; i < len(sl); i++ {
		if sIDs != "" {
			sIDs += ","
		}
		sIDs += strconv.Itoa(sl[i])
	}

	sSQL := " SELECT "
	sSQL = sSQL + " IFNULL(tm." + storage.Fld_common_id + ", 0) teamid, "
	sSQL = sSQL + " IFNULL(tm." + storage.Fld_common_name + ", '') team, "
	sSQL = sSQL + " IFNULL(c." + storage.Fld_common_name + ", '') ctr, "
	sSQL = sSQL + " IFNULL(s." + storage.Fld_common_id + ", '') seasonid, "
	sSQL = sSQL + " IFNULL(s." + storage.Fld_class_season_season + ", '') season, "
	sSQL = sSQL + " IFNULL(s." + storage.Fld_common_name + ", '') seas, "
	sSQL = sSQL + " IFNULL(s." + storage.Fld_class_season_league_rank + ", 0), "
	sSQL = sSQL + " IFNULL(t." + storage.Fld_common_sport_place + ", 0), "
	sSQL = sSQL + " IFNULL(t." + storage.Fld_sport_stage_index + ", 0), "
	sSQL = sSQL + " IFNULL(t." + storage.Fld_sport_wins + ", 0), "
	sSQL = sSQL + " IFNULL(t." + storage.Fld_sport_wins_et + ", 0), "
	sSQL = sSQL + " IFNULL(t." + storage.Fld_sport_draws + ", 0), "
	sSQL = sSQL + " IFNULL(t." + storage.Fld_sport_losses_et + ", 0), "
	sSQL = sSQL + " IFNULL(t." + storage.Fld_sport_losses + ", 0), "
	sSQL = sSQL + " IFNULL(t." + storage.Fld_sport_goals_for + ", 0), "
	sSQL = sSQL + " IFNULL(t." + storage.Fld_sport_goals_against + ", 0), "
	sSQL = sSQL + " IFNULL(s." + storage.Fld_class_season_options_1 + ", '') "
	sSQL = sSQL + " FROM " + storage.Tbl_class_season + " s "
	sSQL = sSQL + " LEFT JOIN " + storage.Tbl_class_base + " b ON b." + storage.Fld_common_id + "=s." + storage.Fld_common_id_base
	sSQL = sSQL + " LEFT JOIN " + storage.Tbl_sport_tables + " t ON t." + storage.Fld_common_id_season + "=s." + storage.Fld_common_id + " AND t." + storage.Fld_common_id_team + " IN (" + sIDs + ")"
	sSQL = sSQL + " LEFT JOIN " + storage.Tbl_class_team + " tm ON tm." + storage.Fld_common_id + "=t." + storage.Fld_common_id_team
	sSQL = sSQL + " LEFT JOIN " + storage.Tbl_countries + " c ON c." + storage.Fld_common_id + "=tm." + storage.Fld_common_id_country
	sSQL = sSQL + " WHERE tm." + storage.Fld_common_id + " IN (" + sIDs + ") AND IFNULL(s." + storage.Fld_common_private + ", 0) <> 1 AND IFNULL(tm." + storage.Fld_common_private + ", 0) <> 1  "
	sSQL = sSQL + " ORDER BY " + storage.Fld_class_season_season + " DESC, s." + storage.Fld_common_sort_order + " DESC "

	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}

	defer rows.Close()
	var list []storage.TblTeam
	for rows.Next() {
		var team storage.TblTeam
		var teamName string
		var ctrName string
		var seasonOptions string
		err := rows.Scan(
			&team.ID,
			&teamName,
			&ctrName,
			&team.SeasonID,
			&team.Season,
			&team.SeasonName,
			&team.LeagueRank,
			&team.Place,
			&team.StageIndex,
			&team.Wins,
			&team.WinsET,
			&team.Draws,
			&team.LossesET,
			&team.Losses,
			&team.Goals_For,
			&team.Goals_Against,
			&seasonOptions,
		)

		if err != nil {
			return nil, err
		}
		team.Name = GetSportTeamName(teamName, ctrName)
		team.Place, err = s.GetPlaceAsStr(team.SeasonID, team.StageIndex, team.Place)
		leagueRank, err := strconv.Atoi(team.LeagueRank)
		if err != nil {
			leagueRank = 0
		}
		team.LeagueRank = s.GetSportRankAsInt(leagueRank)

		if strings.Contains(strings.ToLower(seasonOptions), strings.ToLower(storage.KOptionsResultsOf)) {
			if pos := strings.Index(team.SeasonName, "."); pos >= 0 {
				team.SeasonName = strings.TrimSpace(team.SeasonName[:pos])
			}
		}

		list = append(list, team)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
