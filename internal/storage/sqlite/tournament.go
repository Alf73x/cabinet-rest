package sqlite

import (
	"CabinetREST/internal/storage"
	"database/sql"
	"fmt"
	"sort"
)

/*
********************************************************************

	ShowData_Table

*********************************************************************
*/
func (s *Storage) ShowData_Table(idSeason int) ([]storage.TournamentMatrix, storage.TournamentInfo, error) {

	const fn = "storage.sqlite.ShowData_Table"
	info, err := s.DB_GetSeasonVariables(idSeason)
	if err != nil {
		return nil, storage.TournamentInfo{}, fmt.Errorf("%s: %w", fn, err)
	}

	teams, err := s.loadTournamentMatrixTeams(idSeason)
	if err != nil {
		return nil, info, fmt.Errorf("%s: %w", fn, err)
	}

	if len(teams) == 0 {
		return []storage.TournamentMatrix{}, info, nil
	}

	matches, err := s.loadTournamentMatrixMatches(idSeason)
	if err != nil {
		return nil, info, fmt.Errorf("%s: %w", fn, err)
	}

	result := []storage.TournamentMatrix{
		{
			Teams:   teams,
			Matches: matches,
		},
	}

	return result, info, nil
}

func (s *Storage) loadTournamentMatrixTeams(idSeason int) ([]storage.TournamentMatrixTeam, error) {
	sqlText := `
		SELECT
			IFNULL(t.` + storage.Fld_common_id_team + `, 0),
			IFNULL(t.` + storage.Fld_common_sport_place + `, 0),
			IFNULL(t.` + storage.Fld_sport_result_index + `, 0),
			IFNULL(t.` + storage.Fld_sport_result_index + `_2, 0),
			IFNULL(t.` + storage.Fld_sport_stage_index + `, 0),
			IFNULL(ct.` + storage.Fld_common_name + `, ''),
			IFNULL(cou.` + storage.Fld_common_name + `, ''),
			IFNULL(t.` + storage.Fld_sport_games_played + `, 0),
			IFNULL(t.` + storage.Fld_sport_points + `, 0),
			IFNULL(t.` + storage.Fld_sport_points_adjustment + `, 0),
			IFNULL(t.` + storage.Fld_sport_wins + `, 0),
			IFNULL(t.` + storage.Fld_sport_wins_et + `, 0),
			IFNULL(t.` + storage.Fld_sport_draws + `, 0),
			IFNULL(t.` + storage.Fld_sport_losses_et + `, 0),
			IFNULL(t.` + storage.Fld_sport_losses + `, 0),
			IFNULL(t.` + storage.Fld_sport_goals_for + `, 0),
			IFNULL(t.` + storage.Fld_sport_goals_against + `, 0),

			IFNULL(t.` + storage.Fld_sport_home_games_played + `, 0),
			IFNULL(t.` + storage.Fld_sport_home_points + `, 0),
			IFNULL(t.` + storage.Fld_sport_home_wins + `, 0),
			IFNULL(t.` + storage.Fld_sport_home_wins_et + `, 0),
			IFNULL(t.` + storage.Fld_sport_home_draws + `, 0),
			IFNULL(t.` + storage.Fld_sport_home_losses_et + `, 0),
			IFNULL(t.` + storage.Fld_sport_home_losses + `, 0),
			IFNULL(t.` + storage.Fld_sport_home_goals_for + `, 0),
			IFNULL(t.` + storage.Fld_sport_home_goals_against + `, 0),

			IFNULL(t.` + storage.Fld_sport_away_games_played + `, 0),
			IFNULL(t.` + storage.Fld_sport_away_points + `, 0),
			IFNULL(t.` + storage.Fld_sport_away_wins + `, 0),
			IFNULL(t.` + storage.Fld_sport_away_wins_et + `, 0),
			IFNULL(t.` + storage.Fld_sport_away_draws + `, 0),
			IFNULL(t.` + storage.Fld_sport_away_losses_et + `, 0),
			IFNULL(t.` + storage.Fld_sport_away_losses + `, 0),
			IFNULL(t.` + storage.Fld_sport_away_goals_for + `, 0),
			IFNULL(t.` + storage.Fld_sport_away_goals_against + `, 0)
		FROM ` + storage.Tbl_sport_tables + ` t
		LEFT JOIN ` + storage.Tbl_class_team + ` ct ON ct.` + storage.Fld_common_id + ` = t.` + storage.Fld_common_id_team + `
		LEFT JOIN ` + storage.Tbl_countries + ` cou ON cou.` + storage.Fld_common_id + ` = ct.` + storage.Fld_common_id_country + `
		WHERE t.` + storage.Fld_common_id_season + ` = ?
		  AND IFNULL(ct.` + storage.Fld_common_private + `, 0) <> 1
		  AND IFNULL(cou.` + storage.Fld_common_private + `, 0) <> 1
		ORDER BY t.` + storage.Fld_common_sport_place

	rows, err := s.db.Query(sqlText, idSeason)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []storage.TournamentMatrixTeam

	for rows.Next() {
		var (
			teamID       int
			place        int
			resultIndex  int
			resultIndex2 int
			stageIndex   int
			teamName     string
			cityName     string
			games        int
			points       int
			pointsAdj    int
			wins         int
			otWins       int
			draws        int
			otLosses     int
			losses       int
			goalsFor     int
			goalsAg      int

			homeGames    int
			homePoints   int
			homeWins     int
			homeOTWins   int
			homeDraws    int
			homeOTLosses int
			homeLosses   int
			homeGoalsFor int
			homeGoalsAg  int

			awayGames    int
			awayPoints   int
			awayWins     int
			awayOTWins   int
			awayDraws    int
			awayOTLosses int
			awayLosses   int
			awayGoalsFor int
			awayGoalsAg  int
		)

		err := rows.Scan(
			&teamID,
			&place,
			&resultIndex,
			&resultIndex2,
			&stageIndex,
			&teamName,
			&cityName,
			&games,
			&points,
			&pointsAdj,
			&wins,
			&otWins,
			&draws,
			&otLosses,
			&losses,
			&goalsFor,
			&goalsAg,

			&homeGames,
			&homePoints,
			&homeWins,
			&homeOTWins,
			&homeDraws,
			&homeOTLosses,
			&homeLosses,
			&homeGoalsFor,
			&homeGoalsAg,

			&awayGames,
			&awayPoints,
			&awayWins,
			&awayOTWins,
			&awayDraws,
			&awayOTLosses,
			&awayLosses,
			&awayGoalsFor,
			&awayGoalsAg,
		)
		if err != nil {
			return nil, err
		}

		teams = append(teams, storage.TournamentMatrixTeam{
			ID:           teamID,
			Place:        place,
			ResultIndex:  resultIndex,
			ResultIndex2: resultIndex2,
			StageIndex:   stageIndex,
			Name:         GetSportTeamName(teamName, cityName),
			Games:        games,
			Points:       points + pointsAdj,
			Wins:         wins,
			OTWins:       otWins,
			Draws:        draws,
			OTLosses:     otLosses,
			Losses:       losses,
			GoalsFor:     goalsFor,
			GoalsAgainst: goalsAg,
			Diff:         goalsFor - goalsAg,

			Home: storage.TournamentTeamStat{
				Games:        homeGames,
				Points:       homePoints,
				Wins:         homeWins,
				OTWins:       homeOTWins,
				Draws:        homeDraws,
				OTLosses:     homeOTLosses,
				Losses:       homeLosses,
				GoalsFor:     homeGoalsFor,
				GoalsAgainst: homeGoalsAg,
				Diff:         homeGoalsFor - homeGoalsAg,
			},

			Away: storage.TournamentTeamStat{
				Games:        awayGames,
				Points:       awayPoints,
				Wins:         awayWins,
				OTWins:       awayOTWins,
				Draws:        awayDraws,
				OTLosses:     awayOTLosses,
				Losses:       awayLosses,
				GoalsFor:     awayGoalsFor,
				GoalsAgainst: awayGoalsAg,
				Diff:         awayGoalsFor - awayGoalsAg,
			},
		})
	}

	return teams, rows.Err()
}

func (s *Storage) loadTournamentMatrixMatches(idSeason int) ([]storage.TournamentMatrixMatch, error) {
	sqlText := `SELECT
			IFNULL(` + storage.Fld_common_id_team_1 + `, 0),
			IFNULL(` + storage.Fld_common_id_team_2 + `, 0),
			IFNULL(` + storage.Fld_sport_result_type + `, 0),
			IFNULL(` + storage.Fld_sport_scored + `, 0),
			IFNULL(` + storage.Fld_sport_missed + `, 0),
			IFNULL(` + storage.Fld_sport_scored_et + `, 0),
			IFNULL(` + storage.Fld_sport_missed_et + `, 0),
			IFNULL(` + storage.Fld_sport_tour + `, 0),
			IFNULL(` + storage.Fld_common_date + `, '')
		FROM ` + storage.Tbl_sport_results + ` r
		JOIN ` + storage.Tbl_class_team + ` t1 ON t1.` + storage.Fld_common_id + ` = r.` + storage.Fld_common_id_team_1 + `
		JOIN ` + storage.Tbl_class_team + ` t2 ON t2.` + storage.Fld_common_id + ` = r.` + storage.Fld_common_id_team_2 + `
		LEFT JOIN ` + storage.Tbl_countries + ` c1 ON c1.` + storage.Fld_common_id + ` = t1.` + storage.Fld_common_id_country + `
		LEFT JOIN ` + storage.Tbl_countries + ` c2 ON c2.` + storage.Fld_common_id + ` = t2.` + storage.Fld_common_id_country + `
		WHERE r.` + storage.Fld_common_id_season + ` = ?
		  AND IFNULL(t1.` + storage.Fld_common_private + `, 0) <> 1
		  AND IFNULL(t2.` + storage.Fld_common_private + `, 0) <> 1
		  AND IFNULL(c1.` + storage.Fld_common_private + `, 0) <> 1
		  AND IFNULL(c2.` + storage.Fld_common_private + `, 0) <> 1
		ORDER BY r.` + storage.Fld_common_id_team_1 + `,
				r.` + storage.Fld_common_id_team_2 + `,
				r.` + storage.Fld_common_date

	rows, err := s.db.Query(sqlText, idSeason)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []storage.TournamentMatrixMatch

	orderMap := map[string]int{}

	for rows.Next() {
		var (
			teamID1    int
			teamID2    int
			resultType int
			scored     int
			missed     int
			scoredET   int
			missedET   int
			tour       int
			date       string
		)

		err := rows.Scan(
			&teamID1,
			&teamID2,
			&resultType,
			&scored,
			&missed,
			&scoredET,
			&missedET,
			&tour,
			&date,
		)
		if err != nil {
			return nil, err
		}

		key := fmt.Sprintf("%d:%d", teamID1, teamID2)
		orderMap[key]++

		mask := GetScoresResult(resultType, scored, scoredET, missed, missedET)
		score := SportScoreAsTxt(resultType, scored, missed, scoredET, missedET)

		matches = append(matches, storage.TournamentMatrixMatch{
			TeamID:     teamID1,
			OpponentID: teamID2,
			Score:      score,
			Color:      GetSportScoreColor(mask, "", false),
			Order:      orderMap[key],
			Tour:       tour,
			Date:       SportDateToText(date),
		})
	}

	return matches, rows.Err()
}

/*
********************************************************************
ShowData_Cup
*********************************************************************
*/
func (s *Storage) ShowData_Cup(ids int) ([]storage.TournamentCup, error) {
	const _FunctionName = "storage.sqlite.ShowData_Cup"

	// Get cup winner
	idCupWinner := -1

	sSQL := "SELECT  IFNULL(" + storage.Fld_common_winner_id + ",-1) "
	sSQL += " FROM " + storage.Tbl_class_season
	sSQL += " WHERE " + storage.Fld_common_id + "=? AND IFNULL(" + storage.Fld_common_private + ", 0) <> 1"
	err := s.db.QueryRow(sSQL, ids).Scan(&idCupWinner)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get matches
	sSQL = "SELECT "
	sSQL += "ct1." + storage.Fld_common_id + " AS tid1, "
	sSQL += "ct2." + storage.Fld_common_id + " AS tid2, "
	sSQL += "IFNULL(ct1." + storage.Fld_common_name + ", '') AS tname1, "
	sSQL += "IFNULL(cou1." + storage.Fld_common_name + ", '') AS tcity1, "
	sSQL += "IFNULL(ct2." + storage.Fld_common_name + ", '') AS tname2, "
	sSQL += "IFNULL(cou2." + storage.Fld_common_name + ", '') AS tcity2, "

	sSQL += "IFNULL(r." + storage.Fld_sport_scored + ",0), "
	sSQL += "IFNULL(r." + storage.Fld_sport_scored_et + ",0), "
	sSQL += "IFNULL(r." + storage.Fld_sport_missed + ",0), "
	sSQL += "IFNULL(r." + storage.Fld_sport_missed_et + ",0), "
	sSQL += "IFNULL(r." + storage.Fld_sport_result_type + ",0), "
	sSQL += "IFNULL(r." + storage.Fld_sport_stage_index + ",0), "
	sSQL += "IFNULL(r." + storage.Fld_common_date + ",0) "

	//sSQL += "IFNULL(bs." + storage.Fld_common_id + ", 0) AS bas, "
	// sSQL += "IFNULL(ct1." + storage.Fld_common_favorite + ", 0) AS fv1, "
	// sSQL += "IFNULL(ct2." + storage.Fld_common_favorite + ", 0) AS fv2 "
	sSQL += "FROM " + storage.Tbl_sport_results + " r "
	sSQL += "LEFT JOIN " + storage.Tbl_class_team + " ct1 ON ct1." + storage.Fld_common_id + "=r." + storage.Fld_common_id_team_1 + " "
	sSQL += "LEFT JOIN " + storage.Tbl_countries + " cou1 ON cou1." + storage.Fld_common_id + "=ct1." + storage.Fld_common_id_country + " "
	sSQL += "LEFT JOIN " + storage.Tbl_class_team + " ct2 ON ct2." + storage.Fld_common_id + "=r." + storage.Fld_common_id_team_2 + " "
	sSQL += "LEFT JOIN " + storage.Tbl_countries + " cou2 ON cou2." + storage.Fld_common_id + "=ct2." + storage.Fld_common_id_country + " "
	sSQL += "LEFT JOIN " + storage.Tbl_class_season + " seas ON seas." + storage.Fld_common_id + "=r." + storage.Fld_common_id_season + " "
	//sSQL += "LEFT JOIN " + storage.Tbl_class_base + " bs ON bs." + storage.Fld_common_id + "=seas." + storage.Fld_common_id_base + " "
	sSQL += "WHERE r." + storage.Fld_common_id_season + "=? "
	sSQL += "AND IFNULL(seas." + storage.Fld_common_private + ", 0) <> 1 "
	sSQL += "AND IFNULL(ct1." + storage.Fld_common_private + ", 0) <> 1 "
	sSQL += "AND IFNULL(ct2." + storage.Fld_common_private + ", 0) <> 1 "
	sSQL += "AND IFNULL(cou1." + storage.Fld_common_private + ", 0) <> 1 "
	sSQL += "AND IFNULL(cou2." + storage.Fld_common_private + ", 0) <> 1"
	rows, err := s.db.Query(sSQL, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cup := []storage.TournamentCup{}

	for rows.Next() {
		var match storage.TournamentCup

		var (
			team1       string
			city1       string
			team2       string
			city2       string
			scored      int
			scored_et   int
			missed      int
			missed_et   int
			result_type int
			stage_index int
			date        string
		)
		err := rows.Scan(
			&match.TeamID1,
			&match.TeamID2,
			&team1,
			&city1,
			&team2,
			&city2,
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
		match.TeamName1 = GetSportTeamName(team1, city1)
		match.TeamName2 = GetSportTeamName(team2, city2)
		match.Score = SportScoreAsTxt(result_type, scored, missed, scored_et, missed_et)
		match.Date = date
		match.StageIndex = stage_index
		match.SortOrder = SportGetStageValue(1) - SportGetStageValue(stage_index)

		cup = append(cup, match)
	}

	sort.Slice(cup, func(i, j int) bool {
		return cup[i].SortOrder < cup[j].SortOrder
	})
	return cup, nil
}

/*
********************************************************************

	ShowDataPlain

*********************************************************************
*/
func (s *Storage) ShowDataPlain(ids int) ([]storage.TournamentPlainText, error) {
	const _FunctionName = "storage.sqlite.ShowDataPlain"

	sSQL := " SELECT "
	sSQL = sSQL + " IFNULL(" + storage.Fld_common_text + ", '')"
	sSQL = sSQL + " FROM " + storage.Tbl_class_season
	sSQL = sSQL + " WHERE " + storage.Fld_common_id + "=? AND IFNULL(" + storage.Fld_common_private + ", 0) <> 1"
	var txt string
	err := s.db.QueryRow(sSQL, ids).Scan(&txt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}

	plain := []storage.TournamentPlainText{{PlainText: txt}}
	return plain, nil
}
