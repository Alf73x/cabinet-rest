package sqlite

import (
	"CabinetREST/internal/config"
	"CabinetREST/internal/storage"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

/*********************************************************************
  Db_GetOpponentOptions
**********************************************************************/

func (s *Storage) Db_GetOpponentOptions(sportIDs []int) ([]storage.OpponentCity, []storage.OpponentTeam, error) {
	const _FunctionName = "sqlite.Db_GetOpponentOptions"

	cities, err := s.db_GetOpponentCities()
	if err != nil {
		return nil, nil, fmt.Errorf("%s: get cities: %w", _FunctionName, err)
	}

	teams, err := s.db_GetOpponentTeams(sportIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: get teams: %w", _FunctionName, err)
	}

	return cities, teams, nil
}

func (s *Storage) db_GetOpponentCities() ([]storage.OpponentCity, error) {
	query := fmt.Sprintf(` SELECT %s,  %s  FROM %s WHERE IFNULL(%s, 0)<>1 ORDER BY %s`,
		storage.Fld_common_id,
		storage.Fld_common_name,
		storage.Tbl_countries,
		storage.Fld_common_private,
		storage.Fld_common_name,
	)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]storage.OpponentCity, 0)
	for rows.Next() {
		var item storage.OpponentCity
		err = rows.Scan(&item.ID, &item.Name)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Storage) db_GetOpponentTeams(sportIDs []int) ([]storage.OpponentTeam, error) {
	const _FunctionName = "sqlite.db_GetOpponentTeams"

	var caseSQL strings.Builder

	caseSQL.WriteString("CASE t.")
	caseSQL.WriteString(storage.Fld_common_group_id)
	for i := 2; i < len(config.GsSportGroups); i++ {
		r := []rune(config.GsSportGroups[i])
		if len(r) == 0 {
			continue
		}
		caseSQL.WriteString(fmt.Sprintf(" WHEN %d THEN ' (%s)'", i, string(r[0])))
	}
	caseSQL.WriteString(" ELSE '' END")

	query := fmt.Sprintf(`SELECT t.%s, t.%s AS team, IFNULL(c.%s, '') AS city, IFNULL(b.%s, '') AS bname, t.%s AS bsid,
		%s AS %s FROM %s t
		LEFT JOIN %s c ON t.%s = c.%s
		LEFT JOIN %s b ON b.%s = t.%s`,
		storage.Fld_common_id, storage.Fld_common_name, storage.Fld_common_name, storage.Fld_common_name, storage.Fld_common_id_base,
		caseSQL.String(), storage.Fld_tmp_name, storage.Tbl_class_team,
		storage.Tbl_countries, storage.Fld_common_id_country, storage.Fld_common_id,
		storage.Tbl_class_base, storage.Fld_common_id, storage.Fld_common_id_base,
	)
	args := make([]any, 0, len(sportIDs))
	if len(sportIDs) > 0 {
		query += fmt.Sprintf(` WHERE b.%s IN (%s)`, storage.Fld_common_id, makeOpponentPlaceholders(len(sportIDs)))
		for _, sportID := range sportIDs {
			args = append(args, sportID)
		}
	} else {
		query += fmt.Sprintf(` WHERE b.%s = -1`, storage.Fld_common_id)
	}
	query += fmt.Sprintf(` AND IFNULL(t.%s, 0) <> 1`, storage.Fld_common_private)
	query += fmt.Sprintf(` ORDER BY t.%s, c.%s`, storage.Fld_common_name, storage.Fld_common_name)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", _FunctionName, err)
	}
	defer rows.Close()

	result := make([]storage.OpponentTeam, 0)

	for rows.Next() {
		var (
			id          int
			teamName    string
			cityName    string
			baseName    string
			baseID      int
			groupSuffix string
		)
		err = rows.Scan(
			&id,
			&teamName,
			&cityName,
			&baseName,
			&baseID,
			&groupSuffix,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: scan: %w", _FunctionName, err)
		}

		displayName := strings.Join(nonEmptyStrings(teamName, cityName, groupSuffix, baseName), " ")
		displayName = strings.ReplaceAll(displayName, `"`, "")

		result = append(result, storage.OpponentTeam{
			ID:       id,
			Name:     displayName,
			CityName: strings.TrimSpace(cityName),
			SportID:  baseID,
		})

	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", _FunctionName, err)
	}

	return result, nil
}

func makeOpponentPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)

		if value != "" {
			result = append(result, value)
		}
	}

	return result
}

/*
********************************************************************

	Db_GetComparison

*********************************************************************
*/

func (s *Storage) Db_GetComparison(opponent1Type string, opponent1ID int, opponent2Type string, opponent2ID int, competitionFilter string, sportIDs []int, leagueRanks []int) (storage.TblComparison, error) {
	const op = "storage.Db_GetComparison"

	result := storage.TblComparison{
		Data: make([]storage.ComparisonRow, 0),
	}

	teamIDs1, err := s.getComparisonTeamIDs(opponent1Type, opponent1ID, sportIDs)
	if err != nil {
		return storage.TblComparison{}, fmt.Errorf("%s: resolve opponent 1: %w", op, err)
	}

	teamIDs2, err := s.getComparisonTeamIDs(opponent2Type, opponent2ID, sportIDs)
	if err != nil {
		return storage.TblComparison{}, fmt.Errorf("%s: resolve opponent 2: %w", op, err)
	}

	for _, teamID1 := range teamIDs1 {
		teamName1, err := s.GetTeamNameByID(teamID1)
		if err != nil {
			return storage.TblComparison{}, fmt.Errorf("%s: get team 1 name: %w", op, err)
		}

		for _, teamID2 := range teamIDs2 {
			if teamID1 == teamID2 {
				continue
			}

			teamName2, err := s.GetTeamNameByID(teamID2)
			if err != nil {
				return storage.TblComparison{}, fmt.Errorf("%s: get team 2 name: %w", op, err)
			}

			row, err := s.getComparisonRow(teamID1, teamName1, teamID2, teamName2, competitionFilter, leagueRanks)
			if err != nil {
				return storage.TblComparison{}, fmt.Errorf("%s: compare teams %d and %d: %w", op, teamID1, teamID2, err)
			}

			if row.Total.IsEmpty() {
				continue
			}

			result.Data = append(result.Data, row)
			result.Totals.AddRow(row)
		}
	}

	return result, nil
}
func (s *Storage) getComparisonTeamIDs(opponentType string, opponentID int, sportIDs []int) ([]int, error) {
	switch opponentType {
	case "territory":
		territoryIDs := s.Tree_FindTreeChildrenIDs(opponentID)
		if territoryIDs == "" {
			return []int{}, nil
		}

		if len(sportIDs) == 0 {
			return []int{}, nil
		}

		placeholders := make([]string, len(sportIDs))
		args := make([]any, len(sportIDs))

		for i, sportID := range sportIDs {
			placeholders[i] = "?"
			args[i] = sportID
		}

		query := fmt.Sprintf(` SELECT %[1]s FROM %[2]s WHERE %[3]s IN (%[4]s) AND %[5]s IN (%[6]s) ORDER BY %[1]s`,
			storage.Fld_common_id,           // 1
			storage.Tbl_class_team,          // 2
			storage.Fld_common_id_country,   // 3
			territoryIDs,                    // 4
			storage.Fld_common_id_base,      // 5
			strings.Join(placeholders, ","), // 6
		)

		rows, err := s.db.Query(query, args...)
		if err != nil {
			return nil, fmt.Errorf("query comparison team IDs for territory %d: %w", opponentID, err)
		}
		defer rows.Close()

		teamIDs := make([]int, 0)

		for rows.Next() {
			var teamID int

			if err := rows.Scan(&teamID); err != nil {
				return nil, fmt.Errorf("scan comparison team ID: %w", err)
			}

			teamIDs = append(teamIDs, teamID)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate comparison team IDs: %w", err)
		}

		return teamIDs, nil

	case "team":
		return []int{opponentID}, nil

	default:
		return nil, fmt.Errorf("unknown opponent type %q", opponentType)
	}
}

func (s *Storage) getTeamIDForSports(teamID int, sportIDs []int) ([]int, error) {
	if len(sportIDs) == 0 {
		return []int{}, nil
	}

	placeholders := make([]string, len(sportIDs))
	args := make([]any, 0, len(sportIDs)+1)
	args = append(args, teamID)

	for i, sportID := range sportIDs {
		placeholders[i] = "?"
		args = append(args, sportID)
	}

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ? AND %s IN (%s) LIMIT 1	`,
		storage.Fld_common_id,
		storage.Tbl_class_team,
		storage.Fld_common_id,
		storage.Fld_common_id_base,
		strings.Join(placeholders, ","),
	)

	var id int

	err := s.db.QueryRow(query, args...).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		// Команда не относится ни к одному выбранному виду спорта.
		return []int{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("check team %d sport: %w", teamID, err)
	}

	return []int{id}, nil
}

func (s *Storage) getComparisonRow(teamID1 int, teamName1 string, teamID2 int, teamName2 string, competitionFilter string, leagueRanks []int) (storage.ComparisonRow, error) {
	sportID, err := s.getTeamSportID(teamID1)
	if err != nil {
		return storage.ComparisonRow{}, err
	}

	home, err := s.getDirectComparisonStat(teamID1, teamID2, competitionFilter, leagueRanks)
	if err != nil {
		return storage.ComparisonRow{}, err
	}

	reverse, err := s.getDirectComparisonStat(teamID2, teamID1, competitionFilter, leagueRanks)
	if err != nil {
		return storage.ComparisonRow{}, err
	}

	away := storage.Stat{
		Wins:          reverse.Losses,
		WinsET:        reverse.LossesET,
		Draws:         reverse.Draws,
		LossesET:      reverse.WinsET,
		Losses:        reverse.Wins,
		Goals_For:     reverse.Goals_Against,
		Goals_Against: reverse.Goals_For,
	}

	away.CalculateGames()

	total := home
	total.Add(away)

	return storage.ComparisonRow{
		SportID: sportID,
		Team1:   teamName1,
		Team2:   teamName2,
		IDTeam1: teamID1,
		IDTeam2: teamID2,
		Total:   total,
		Home:    home,
		Away:    away,
	}, nil
}

func (s *Storage) getDirectComparisonStat(teamID1 int, teamID2 int, competitionFilter string, leagueRanks []int) (storage.Stat, error) {
	leagueFilterSQL, leagueFilterArgs, err := s.buildComparisonCompetitionFilter(competitionFilter)
	if err != nil {
		return storage.Stat{}, err
	}

	rankFilterSQL, rankFilterArgs := s.buildSportComparisonRankFilter(leagueRanks)

	query := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(W + Wpm), 0) AS W,
			COALESCE(SUM(D + Dpm + Dex), 0) AS D,
			COALESCE(SUM(L + Lpm), 0) AS L,
			COALESCE(SUM(Wet + Wex1 + Wex2), 0) AS We,
			COALESCE(SUM(Let + Lex1 + Lex2), 0) AS Le,
			COALESCE(SUM(S), 0) AS S,
			COALESCE(SUM(M), 0) AS M,
			COALESCE(SUM(Cnt), 0) AS Cnt
		FROM (
			SELECT
				r.%[1]s,
				se.%[2]s,
				1 AS Cnt,

				CASE WHEN r.%[3]s = ? AND r.%[4]s > r.%[5]s THEN 1 ELSE 0 END AS W,
				CASE WHEN r.%[3]s = ? AND r.%[4]s < r.%[5]s THEN 1 ELSE 0 END AS L,
				CASE WHEN r.%[3]s = ? AND r.%[4]s = r.%[5]s AND r.%[6]s > r.%[7]s THEN 1 ELSE 0 END AS Wet,
				CASE WHEN r.%[3]s = ? AND r.%[4]s = r.%[5]s AND r.%[6]s < r.%[7]s THEN 1 ELSE 0 END AS Let,
				CASE WHEN r.%[3]s = ? AND r.%[4]s = r.%[5]s AND r.%[6]s = r.%[7]s THEN 1 ELSE 0 END AS D,

				CASE WHEN r.%[3]s IN (?, ?, ?, ?, ?) AND r.%[4]s > r.%[5]s THEN 1 ELSE 0 END AS Wex1,
				CASE WHEN r.%[3]s IN (?, ?, ?, ?, ?) AND r.%[4]s < r.%[5]s THEN 1 ELSE 0 END AS Lex1,
				CASE WHEN r.%[3]s IN (?, ?, ?, ?, ?) AND r.%[4]s = r.%[5]s AND r.%[6]s > r.%[7]s THEN 1 ELSE 0 END AS Wex2,
				CASE WHEN r.%[3]s IN (?, ?, ?, ?, ?) AND r.%[4]s = r.%[5]s AND r.%[6]s < r.%[7]s THEN 1 ELSE 0 END AS Lex2,
				CASE WHEN r.%[3]s IN (?, ?, ?, ?, ?) AND r.%[4]s = r.%[5]s AND r.%[6]s = r.%[7]s THEN 1 ELSE 0 END AS Dex,

				CASE WHEN r.%[3]s = ? THEN 1 ELSE 0 END AS Wpm,
				CASE WHEN r.%[3]s = ? THEN 1 ELSE 0 END AS Lpm,
				CASE WHEN r.%[3]s = ? THEN 1 ELSE 0 END AS Dpm,

				CASE WHEN r.%[4]s = -1 THEN 0 ELSE r.%[4]s END AS S,
				CASE WHEN r.%[5]s = -1 THEN 0 ELSE r.%[5]s END AS M
			FROM %[8]s r
			LEFT JOIN %[9]s se ON se.%[10]s = r.%[1]s
			WHERE r.%[11]s = ?
			  AND r.%[12]s = ?
			  %s
			  %s
		) comparison_data`,
		storage.Fld_common_id_season,         // 1
		storage.Fld_class_season_league_rank, // 2
		storage.Fld_sport_result_type,        // 3
		storage.Fld_sport_scored,             // 4
		storage.Fld_sport_missed,             // 5
		storage.Fld_sport_scored_et,          // 6
		storage.Fld_sport_missed_et,          // 7
		storage.Tbl_sport_results,            // 8
		storage.Tbl_class_season,             // 9
		storage.Fld_common_id,                // 10
		storage.Fld_common_id_team_1,         // 11
		storage.Fld_common_id_team_2,         // 12
		leagueFilterSQL,
		rankFilterSQL,
	)

	args := make([]any, 0, 50)

	args = append(args,
		storage.RtScoreNormal,
		storage.RtScoreNormal,
		storage.RtScoreNormal,
		storage.RtScoreNormal,
		storage.RtScoreNormal,
	)

	extraResultTypes := []any{
		storage.RtScoreOT,
		storage.RtScoreB,
		storage.RtScoreP,
		storage.RtScoreEt,
		storage.RtScoreAllExtra,
	}

	args = append(args, extraResultTypes...)
	args = append(args, extraResultTypes...)
	args = append(args, extraResultTypes...)
	args = append(args, extraResultTypes...)
	args = append(args, extraResultTypes...)

	args = append(args,
		storage.RtScorePlusMinus,
		storage.RtScoreMinusPlus,
		storage.RtScoreMinusMinus,
	)

	args = append(args, teamID1, teamID2)
	args = append(args, leagueFilterArgs...)
	args = append(args, rankFilterArgs...)

	var stat storage.Stat

	err = s.db.QueryRow(query, args...).Scan(
		&stat.Wins,
		&stat.Draws,
		&stat.Losses,
		&stat.WinsET,
		&stat.LossesET,
		&stat.Goals_For,
		&stat.Goals_Against,
		&stat.Games,
	)
	if err != nil {
		return storage.Stat{}, fmt.Errorf(
			"get direct comparison stat for teams %d and %d: %w",
			teamID1,
			teamID2,
			err,
		)
	}

	return stat, nil
}

func (s *Storage) buildComparisonCompetitionFilter(competitionFilter string) (string, []any, error) {
	switch competitionFilter {
	case "", "all", "0":
		return "", nil, nil

	case "league", "1":
		filter := fmt.Sprintf(`AND ((se.%[1]s >= ? AND se.%[1]s <= ?) OR  (se.%[1]s >= ? AND se.%[1]s <= ?) OR  (se.%[1]s >= ? AND se.%[1]s <= ?))`, storage.Fld_class_season_league_rank)
		return filter, []any{
			storage.KSeasonsRankMin,
			storage.KSeasonsRankMax,
			storage.KPlayoffMin,
			storage.KPlayoffMax,
			storage.KPlayoffMatchesMin,
			storage.KPlayoffMatchesMax,
		}, nil

	case "cup", "2":
		filter := fmt.Sprintf(`AND (se.%[1]s = ? OR se.%[1]s = ?)`, storage.Fld_class_season_league_rank)
		return filter, []any{
			storage.KSeasonsRankCup,
			storage.KSeasonsRankCupTournament,
		}, nil

	case "other", "3":
		ranks := make([]int, 0)

		for rank := storage.KSeasonsRankMin; rank <= storage.KSeasonsRankMax; rank++ {
			ranks = append(ranks, rank)
		}

		for rank := storage.KPlayoffMin; rank <= storage.KPlayoffMax; rank++ {
			ranks = append(ranks, rank)
		}

		for rank := storage.KPlayoffMatchesMin; rank <= storage.KPlayoffMatchesMax; rank++ {
			ranks = append(ranks, rank)
		}

		ranks = append(
			ranks,
			storage.KSeasonsRankCup,
			storage.KSeasonsRankCupTournament,
		)

		placeholders := make([]string, len(ranks))
		args := make([]any, len(ranks))

		for i, rank := range ranks {
			placeholders[i] = "?"
			args[i] = rank
		}

		filter := fmt.Sprintf(
			`AND se.%s NOT IN (%s)`,
			storage.Fld_class_season_league_rank,
			strings.Join(placeholders, ","),
		)

		return filter, args, nil

	default:
		return "", nil, fmt.Errorf("unknown competition filter %q", competitionFilter)
	}
}

func (s *Storage) buildSportComparisonRankFilter(leagueRanks []int) (string, []any) {
	if len(leagueRanks) == 0 {
		return "", nil
	}
	placeholders := make([]string, len(leagueRanks))
	args := make([]any, len(leagueRanks))

	for i, rank := range leagueRanks {
		placeholders[i] = "?"
		args[i] = rank
	}

	filter := fmt.Sprintf("AND se.%s IN (%s)", storage.Fld_class_season_league_rank, strings.Join(placeholders, ","))
	return filter, args
}

func (s *Storage) getTeamSportID(teamID int) (int, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE %s = ?`,
		storage.Fld_common_id_base,
		storage.Tbl_class_team,
		storage.Fld_common_id,
	)

	var sportID int

	if err := s.db.QueryRow(query, teamID).Scan(&sportID); err != nil {
		return 0, fmt.Errorf("get sport ID for team %d: %w", teamID, err)
	}

	return sportID, nil
}

/*

Db_GetComparisonMatches

*/

func (s *Storage) Db_GetComparisonMatches(team1ID int, team2ID int) ([]storage.TblTeamMatches, error) {
	const op = "storage.sqlite.Db_GetComparisonMatches"

	query := fmt.Sprintf(`
		SELECT
			IFNULL(r.%[1]s, 0),
			IFNULL(r.%[2]s, 0),
			IFNULL(t1.%[3]s, ''),
			IFNULL(c1.%[3]s, ''),
			IFNULL(t2.%[3]s, ''),
			IFNULL(c2.%[3]s, ''),
			IFNULL(r.%[4]s, 0),
			IFNULL(r.%[5]s, 0),
			IFNULL(r.%[6]s, 0),
			IFNULL(r.%[7]s, 0),
			IFNULL(r.%[8]s, 0),
			IFNULL(r.%[9]s, '')
		FROM %[10]s r
		LEFT JOIN %[11]s t1 ON t1.%[12]s = r.%[1]s
		LEFT JOIN %[13]s c1 ON c1.%[12]s = t1.%[14]s
		LEFT JOIN %[11]s t2 ON t2.%[12]s = r.%[2]s
		LEFT JOIN %[13]s c2 ON c2.%[12]s = t2.%[14]s
		WHERE
			(r.%[1]s = ? AND r.%[2]s = ?)
			OR
			(r.%[1]s = ? AND r.%[2]s = ?)
		ORDER BY r.%[9]s DESC`,
		storage.Fld_common_id_team_1,  // 1
		storage.Fld_common_id_team_2,  // 2
		storage.Fld_common_name,       // 3
		storage.Fld_sport_scored,      // 4
		storage.Fld_sport_scored_et,   // 5
		storage.Fld_sport_missed,      // 6
		storage.Fld_sport_missed_et,   // 7
		storage.Fld_sport_result_type, // 8
		storage.Fld_common_date,       // 9
		storage.Tbl_sport_results,     // 10
		storage.Tbl_class_team,        // 11
		storage.Fld_common_id,         // 12
		storage.Tbl_countries,         // 13
		storage.Fld_common_id_country, // 14
	)

	rows, err := s.db.Query(query, team1ID, team2ID, team2ID, team1ID)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	result := make([]storage.TblTeamMatches, 0)

	for rows.Next() {
		var (
			match      storage.TblTeamMatches
			teamName1  string
			cityName1  string
			teamName2  string
			cityName2  string
			scored     int
			scoredET   int
			missed     int
			missedET   int
			resultType int
		)

		err = rows.Scan(
			&match.TeamID1,
			&match.TeamID2,
			&teamName1,
			&cityName1,
			&teamName2,
			&cityName2,
			&scored,
			&scoredET,
			&missed,
			&missedET,
			&resultType,
			&match.Date,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}

		match.TeamName1 = GetSportTeamName(teamName1, cityName1)
		match.TeamName2 = GetSportTeamName(teamName2, cityName2)
		match.Score = SportScoreAsTxt(resultType, scored, missed, scoredET, missedET)

		result = append(result, match)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}

	return result, nil
}
