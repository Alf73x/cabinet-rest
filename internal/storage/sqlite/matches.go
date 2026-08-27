package sqlite

import (
	"CabinetREST/internal/storage"
	"strconv"
	"strings"
)

/*********************************************************************
  Db_GetTeamMatches
**********************************************************************/

func (s *Storage) Db_GetTeamMatches(idTeam int, idSeason int) ([]storage.TblTeamMatches, error) {
	const _FunctionName = "storage.sqlite.Db_GetTeams"
	var teamMatches []storage.TblTeamMatches

	sl, err := s.GetTeamTreeIDs(idTeam)
	var sIDs string
	for i := 0; i < len(sl); i++ {
		if sIDs != "" {
			sIDs += ","
		}
		sIDs += strconv.Itoa(sl[i])
	}

	sSeasonID := strconv.Itoa(idSeason)
	sExtraTeamsFilter := ""

	sqlText := "SELECT IFNULL(" + storage.Fld_class_season_options_1 + ", '')" +
		" FROM " + storage.Tbl_class_season +
		" WHERE " + storage.Fld_common_id + "=" + sSeasonID
	rows, err := s.db.Query(sqlText)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, err
	}

	var sOpt string
	if err := rows.Scan(&sOpt); err != nil {
		return nil, err
	}

	slOptions := strings.Split(sOpt, ";")

	sJoinSeasonIDs := ""
	iResultsOf := 0

	for _, opt := range slOptions {
		str := strings.ToUpper(strings.TrimSpace(opt))

		if strings.HasPrefix(str, storage.KOptionsJoin+"=") {
			sJoinSeasonIDs = strings.TrimSpace(strings.ToUpper(str[5:]))

			rowsTeams, err := s.db.Query(
				"SELECT " + storage.Fld_common_id_team +
					" FROM " + storage.Tbl_sport_tables +
					" WHERE " + storage.Fld_common_id_season + "=" + sSeasonID,
			)
			if err != nil {
				return nil, err
			}
			var teamIDs []string
			for rowsTeams.Next() {
				var teamID string
				if err := rowsTeams.Scan(&teamID); err != nil {
					rowsTeams.Close()
					return nil, err
				}
				teamIDs = append(teamIDs, teamID)
			}
			rowsTeams.Close()
			if len(teamIDs) > 0 {
				sExtraTeamsFilter = " AND (" + storage.Fld_common_id_team_1 + " IN (" + strings.Join(teamIDs, ",") + ")" + " AND " +
					storage.Fld_common_id_team_2 + " IN (" + strings.Join(teamIDs, ",") + ")" + ")"
			}

		} else if strings.HasPrefix(str, storage.KOptionsResultsOf+"=") {
			value := strings.TrimSpace(str[len(storage.KOptionsResultsOf)+1:])

			if n, err := strconv.Atoi(value); err == nil {
				iResultsOf = n
			} else {
				iResultsOf = -1
			}
		}
	}

	if iResultsOf > 0 {
		sSeasonID = strconv.Itoa(iResultsOf)
	} else if strings.Contains(
		strings.ToUpper(sOpt),
		storage.KOptionsMode+"="+storage.KOptionsViewModeResults,
	) {
		if sJoinSeasonIDs != "" {
			sSeasonID = sJoinSeasonIDs
		}
	}

	str := "SELECT " +
		"IFNULL(" + storage.Fld_common_id_team_1 + ",0), " +
		"IFNULL(" + storage.Fld_common_id_team_2 + ",0), " +
		"IFNULL(" + storage.Fld_sport_scored + ",0), " +
		"IFNULL(" + storage.Fld_sport_scored_et + ",0), " +
		"IFNULL(" + storage.Fld_sport_missed + ",0), " +
		"IFNULL(" + storage.Fld_sport_missed_et + ",0), " +
		"IFNULL(" + storage.Fld_sport_result_type + ",0), " +
		"IFNULL(" + storage.Fld_sport_stage_index + ",0), " +
		"IFNULL(" + storage.Fld_common_date + ",'')"
	str += " FROM " + storage.Tbl_sport_results
	str += " WHERE id_season IN (" + sSeasonID + ")"
	str += " AND (" + storage.Fld_common_id_team_1 + " IN (" + sIDs + ")" + " OR " + storage.Fld_common_id_team_2 + " IN (" + sIDs + "))"
	str += sExtraTeamsFilter
	str += " ORDER BY " + storage.Fld_common_date + " DESC "

	rows, err = s.db.Query(str)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var match storage.TblTeamMatches

		var (
			id_team_1   int
			id_team_2   int
			scored      int
			scored_et   int
			missed      int
			missed_et   int
			result_type int
			stage_index int
			date        string
		)
		err := rows.Scan(
			&id_team_1,
			&id_team_2,
			&scored,
			&scored_et,
			&missed,
			&missed_et,
			&result_type,
			&stage_index,
			&date,
		)
		if err != nil {
			return nil, err
		}
		match.TeamID1 = id_team_1
		match.TeamName1, err = s.DB_GetTeamName(id_team_1, 0)
		if err != nil {
			return nil, err
		}
		match.TeamID2 = id_team_2
		match.TeamName2, err = s.DB_GetTeamName(id_team_2, 0)
		if err != nil {
			return nil, err
		}
		match.Date = date
		match.Score = SportScoreAsTxt(result_type, scored, missed, scored_et, missed_et)

		teamMatches = append(teamMatches, match)
	}

	return teamMatches, nil
}
