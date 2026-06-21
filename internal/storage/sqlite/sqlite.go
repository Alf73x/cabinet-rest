package sqlite

import (
	"CabinetREST/internal/storage"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const _FunctionName = "storage.sqlite.New"
	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}
	/*
		stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS...`)
		if err != nil {
			return nil, fmt.Errorf("#{op}: #[err}")
		}
		_, err = stmt.Exec()
		if err != nil {
			return nil, fmt.Errorf("#{op}: #[err}")
		}
	*/
	return &Storage{db: db}, nil
}

/*
	func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
		const op = "storage.sqlite.SaveURL"
		stmt, err := s.db.Prepare(`INSERT INTO url (url, alias) VALUES(?, ?)`)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
		res, err := stmt.Exec()
		if err != nil {
			//	if sqliteErr, ok := err.(sqlite.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			//			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
			//
			//			}
			return 0, fmt.Errorf("#{op}: #[err}")
		}
		id, err := res.LastInsertId()

		return id, nil
	}
*/

/*********************************************************************
  Territories for parent.
*********************************************************************/

func (s *Storage) Db_GetTerritories(parentID int) ([]storage.TblCountry, error) {
	const _FunctionName = "storage.sqlite.Db_GetCountries"

	if parentID <= 0 {
		parentID = -1
	}

	sSQL := fmt.Sprintf(`SELECT c.%s, c.%s, 
	    CASE WHEN EXISTS (SELECT 1 FROM %s c2 WHERE c2.%s = c.%s) THEN 1
    	ELSE 0 END AS has_children,
		IFNULL(%s,"")  /* sort_order */
		FROM %s c 
		WHERE IFNULL(c.%s, -1)=%d 
		ORDER BY %s, %s`, storage.Fld_common_id, storage.Fld_common_name, storage.Tbl_countries, storage.Fld_countries_id_parent, storage.Fld_common_id,
		storage.Fld_common_sort_order,
		storage.Tbl_countries,
		storage.Fld_countries_id_parent, parentID, storage.Fld_countries_sort_order, storage.Fld_common_name)

	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}

	defer rows.Close()

	var countries []storage.TblCountry
	for rows.Next() {
		var cntr storage.TblCountry
		err := rows.Scan(
			&cntr.ID,
			&cntr.Name,
			&cntr.HasChildren,
			&cntr.SortOrder,
		)
		if err != nil {
			return nil, err
		}
		countries = append(countries, cntr)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return countries, nil
}

/*********************************************************************
  Territories. Search for text
*********************************************************************/

func (s *Storage) Db_SearchTerritories(filter string) ([]storage.TblCountry, error) {
	const _FunctionName = "storage.sqlite.Db_SearchTerritories"

	sSQL := fmt.Sprintf(`SELECT c.%s, c.%s, 
	    0 AS has_children,
		IFNULL(%s,"")  /* sort_order */
		FROM %s c 
		ORDER BY %s, %s`, storage.Fld_common_id, storage.Fld_common_name,
		storage.Fld_common_sort_order,
		storage.Tbl_countries,
		storage.Fld_countries_sort_order, storage.Fld_common_name)

	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}

	defer rows.Close()

	var countries []storage.TblCountry
	filterLower := strings.ToLower(filter)
	for rows.Next() {
		var cntr storage.TblCountry
		err := rows.Scan(
			&cntr.ID,
			&cntr.Name,
			&cntr.HasChildren,
			&cntr.SortOrder,
		)
		if err != nil {
			return nil, err
		}

		if strings.Contains(strings.ToLower(cntr.Name), filterLower) {
			countries = append(countries, cntr)
		}

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return countries, nil
}

/*********************************************************************
  Db_PathTerritories
**********************************************************************/

func (s *Storage) Db_PathTerritories(id int) ([]int, error) {
	const _FunctionName = "storage.sqlite.Db_PathTerritories"

	sSQL := fmt.Sprintf(`
	WITH RECURSIVE parents AS (
   		SELECT %s, %s, 0 AS level FROM %s WHERE id=%v
    	UNION ALL
    	SELECT t.%s, t.%s, p.level + 1 FROM %s t
    	JOIN parents p ON t.%s = p.%s
	)
	SELECT %s FROM parents ORDER BY level DESC`,
		storage.Fld_common_id, storage.Fld_countries_id_parent, storage.Tbl_countries, id,
		storage.Fld_common_id, storage.Fld_countries_id_parent, storage.Tbl_countries,
		storage.Fld_common_id, storage.Fld_countries_id_parent,
		storage.Fld_common_id,
	)

	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}

	defer rows.Close()

	var path []int
	for rows.Next() {
		var cid int
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		path = append(path, cid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return path, nil
}

/*********************************************************************
  Seasons
**********************************************************************/

func (s *Storage) Db_GetSeasons(idssport string, filterSeason string, filterName string) ([]storage.TblSeason, error) {
	const _FunctionName = "storage.sqlite.Db_GetSeasons"
	sports := ""
	if strings.TrimSpace(idssport) != "" {
		sports = fmt.Sprintf(" AND %s IN (%s)", storage.Fld_common_id_base, idssport)
	}

	sSQL := fmt.Sprintf(`SELECT 
		%s, %s,   
		IFNULL(%s,""), /* prefix */
		%s,
		%s,
		%s,
		IFNULL(%s,0),  /* rank */
		IFNULL(%s,0),  /* sort_order */
		IFNULL(%s,""), /* points */
		IFNULL(%s,""), /* options_1 */
		IFNULL(%s,"")  /* options_2 */
		FROM %s  
		WHERE 1=1 %s
		ORDER BY %s DESC, %s`, storage.Fld_common_id, storage.Fld_class_season_season, storage.Fld_common_prefix, storage.Fld_common_name, storage.Fld_common_id_base,
		storage.Fld_common_group_id, storage.Fld_class_season_league_rank, storage.Fld_common_sort_order,
		storage.Fld_class_season_points, storage.Fld_class_season_options_1, storage.Fld_class_season_options_2,
		storage.Tbl_class_season,
		sports,
		storage.Fld_class_season_season, storage.Fld_common_sort_order,
	)

	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}

	defer rows.Close()

	var seasons []storage.TblSeason
	for rows.Next() {
		var seas storage.TblSeason
		err := rows.Scan(
			&seas.ID,
			&seas.Season,
			&seas.Prefix,
			&seas.Name,
			&seas.SportID,
			&seas.GroupID,
			&seas.LeagueRank,
			&seas.SortOrder,
			&seas.Points,
			&seas.Options1,
			&seas.Options2)
		if err != nil {
			return nil, err
		}
		seasons = append(seasons, seas)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return seasons, nil
}

/*********************************************************************
  Db_GetSports
**********************************************************************/

func (s *Storage) Db_GetSports() ([]storage.TblSport, error) {
	const _FunctionName = "storage.sqlite.Db_GetSports"

	sSQL := fmt.Sprintf(`SELECT %s, %s, %s  
		FROM %s  
		WHERE %s=1
		ORDER BY %s`, storage.Fld_common_id, storage.Fld_common_name, storage.Fld_class_base_options, storage.Tbl_class_base, storage.Fld_class_base_available, storage.Fld_common_id)

	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}

	defer rows.Close()

	var sports []storage.TblSport
	for rows.Next() {
		var sps storage.TblSport
		err := rows.Scan(
			&sps.ID,
			&sps.Name,
			&sps.BaseOption,
		)
		if err != nil {
			return nil, err
		}
		sports = append(sports, sps)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sports, nil
}

/*********************************************************************
  Db_GetTeams
**********************************************************************/

func (s *Storage) Db_GetTeams(idTerritory int, idssport string) ([]storage.TblTeams, error) {
	const _FunctionName = "storage.sqlite.Db_GetTeams"

	sFilter := s.Tree_TreeChildrenFilter(idTerritory, "")
	sSQL := "WITH CTE_Teams AS ( "
	sSQL = sSQL + "   SELECT id FROM " + storage.Tbl_class_team + " WHERE " + sFilter
	sSQL = sSQL + ")"
	sSQL = sSQL + " SELECT "
	sSQL = sSQL + " s." + storage.Fld_common_id + ", "
	sSQL = sSQL + " s." + storage.Fld_common_id_base + ", "
	sSQL = sSQL + " IFNULL(s." + storage.Fld_class_season_season + ", ''), "
	sSQL = sSQL + " s." + storage.Fld_common_name + ", "
	sSQL = sSQL + " tm." + storage.Fld_common_name + " team, "
	sSQL = sSQL + " c." + storage.Fld_common_name + " ctr, "
	sSQL = sSQL + " tm." + storage.Fld_common_id + " teamid, "
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
			&team.SportID,
			&team.Season,
			&team.SeasonName,
			&team.TeamName,
			&team.TeamTerritory,
			&team.TeamID,
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
		team.Place, err = s.GetPlaceAsStr(team.ID, team.StageIndex, team.Place)
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
		match.TeamName1, err = s.GetTeamName(id_team_1, 0)
		if err != nil {
			return nil, err
		}
		match.TeamID2 = id_team_2
		match.TeamName2, err = s.GetTeamName(id_team_2, 0)
		if err != nil {
			return nil, err
		}
		match.Date = date
		match.Score = SportScoreAsTxt(result_type, scored, missed, scored_et, missed_et)

		teamMatches = append(teamMatches, match)
	}

	return teamMatches, nil
}

/*********************************************************************
DB_GetSeasonVariables
**********************************************************************/

func (s *Storage) DB_GetSeasonVariables(id int) (ti storage.TournamentInfo, e error) {
	var info storage.TournamentInfo
	info.IsOk = false

	sSQL := `
	SELECT
		s.` + storage.Fld_common_name + ` AS nm,
		IFNULL(s.` + storage.Fld_common_group_id + `, 0),
		IFNULL(s.` + storage.Fld_class_season_season + `, ''),
		IFNULL(s.` + storage.Fld_class_season_options_1 + `, ''),
		IFNULL(s.` + storage.Fld_class_season_league_rank + `, 0),
		IFNULL(b.` + storage.Fld_common_name + `, '') AS vs,
		IFNULL(s.` + storage.Fld_class_season_plain_text + `, ''),
		IFNULL(s.` + storage.Fld_class_season_remark_text + `, ''),
		IFNULL(s.` + storage.Fld_common_prefix + `, '')
	FROM ` + storage.Tbl_class_season + ` s
	LEFT JOIN ` + storage.Tbl_class_base + ` b
	ON s.` + storage.Fld_common_id_base + ` = b.` + storage.Fld_common_id + `
	WHERE s.` + storage.Fld_common_id + ` = ?`

	var (
		name       sql.NullString
		groupID    sql.NullInt64
		season     sql.NullString
		options1   sql.NullString
		leagueRank sql.NullInt64
		baseName   sql.NullString
		plainText  sql.NullString
		remark     sql.NullString
		prefix     sql.NullString
	)

	err := s.db.QueryRow(sSQL, id).Scan(
		&name,
		&groupID,
		&season,
		&options1,
		&leagueRank,
		&baseName,
		&plainText,
		&remark,
		&prefix,
	)
	if err != nil {
		return info, err
	}

	titlePrefix := ""
	if strings.TrimSpace(prefix.String) != "" {
		titlePrefix = prefix.String + ". "
	}

	info.Title = baseName.String + ". " + titlePrefix + name.String + ". " + season.String
	info.ViewOpt = parseViewOption(options1.String)
	info.Rank = int(leagueRank.Int64)
	info.IsOk = true
	return info, err
}

/*
********************************************************************

	ShowData_Table

*********************************************************************
*/
func (s *Storage) ShowData_Table(idSeason int) ([]storage.TournamentMatrix, error) {
	const fn = "storage.sqlite.ShowData_Table"

	teams, err := s.loadTournamentMatrixTeams(idSeason)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	if len(teams) == 0 {
		return []storage.TournamentMatrix{}, nil
	}

	matches, err := s.loadTournamentMatrixMatches(idSeason)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	result := []storage.TournamentMatrix{
		{
			Teams:   teams,
			Matches: matches,
		},
	}

	return result, nil
}

func (s *Storage) loadTournamentMatrixTeams(idSeason int) ([]storage.TournamentMatrixTeam, error) {
	sqlText := `
		SELECT
			IFNULL(t.` + storage.Fld_common_id_team + `, 0),
			IFNULL(t.` + storage.Fld_common_sport_place + `, 0),
			IFNULL(ct.` + storage.Fld_common_name + `, ''),
			IFNULL(cou.` + storage.Fld_common_name + `, ''),

			IFNULL(t.` + storage.Fld_sport_games_played + `, 0),
			IFNULL(t.` + storage.Fld_sport_points + `, 0),
			IFNULL(t.` + storage.Fld_sport_points_adjustment + `, 0),
			IFNULL(t.` + storage.Fld_sport_wins + `, 0),
			IFNULL(t.` + storage.Fld_sport_wins_et + `, 0),
			IFNULL(t.` + storage.Fld_sport_losses_et + `, 0),
			IFNULL(t.` + storage.Fld_sport_losses + `, 0),
			IFNULL(t.` + storage.Fld_sport_goals_for + `, 0),
			IFNULL(t.` + storage.Fld_sport_goals_against + `, 0)
		FROM ` + storage.Tbl_sport_tables + ` t
		LEFT JOIN ` + storage.Tbl_class_team + ` ct
			ON ct.` + storage.Fld_common_id + ` = t.` + storage.Fld_common_id_team + `
		LEFT JOIN ` + storage.Tbl_countries + ` cou
			ON cou.` + storage.Fld_common_id + ` = ct.` + storage.Fld_common_id_country + `
		WHERE t.` + storage.Fld_common_id_season + ` = ?
		ORDER BY t.` + storage.Fld_common_sport_place

	rows, err := s.db.Query(sqlText, idSeason)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []storage.TournamentMatrixTeam

	for rows.Next() {
		var (
			teamID    int
			place     int
			teamName  string
			cityName  string
			games     int
			points    int
			pointsAdj int
			wins      int
			otWins    int
			otLosses  int
			losses    int
			goalsFor  int
			goalsAg   int
		)

		err := rows.Scan(
			&teamID,
			&place,
			&teamName,
			&cityName,
			&games,
			&points,
			&pointsAdj,
			&wins,
			&otWins,
			&otLosses,
			&losses,
			&goalsFor,
			&goalsAg,
		)
		if err != nil {
			return nil, err
		}

		teams = append(teams, storage.TournamentMatrixTeam{
			ID:           teamID,
			Place:        place,
			Name:         GetSportTeamName(teamName, cityName),
			Games:        games,
			Points:       points + pointsAdj,
			Wins:         wins,
			OTWins:       otWins,
			OTLosses:     otLosses,
			Losses:       losses,
			GoalsFor:     goalsFor,
			GoalsAgainst: goalsAg,
			Diff:         goalsFor - goalsAg,
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
		FROM ` + storage.Tbl_sport_results + `
		WHERE ` + storage.Fld_common_id_season + ` = ?
		ORDER BY ` + storage.Fld_common_id_team_1 + `,
				` + storage.Fld_common_id_team_2 + `,
				` + storage.Fld_common_date

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

	cup := []storage.TournamentCup{
		{
			TblTeamMatches: storage.TblTeamMatches{
				TeamName1: "Спартак Москва",
				TeamID1:   1,
				TeamName2: "Зенит Санкт-Петербург",
				TeamID2:   2,
				Score:     "2:1",
				Date:      "20260515",
			},
			Stage: "Финал",
		},
		{
			TblTeamMatches: storage.TblTeamMatches{
				TeamName1: "ЦСКА Москва",
				TeamID1:   3,
				TeamName2: "Динамо Москва",
				TeamID2:   4,
				Score:     "1:0",
				Date:      "20260510",
			},
			Stage: "1/2 финала",
		},
	}

	return cup, nil
}

/*
********************************************************************

	ShowDataPlain

*********************************************************************
*/
func (s *Storage) ShowDataPlain(ids int) ([]storage.TournamentPlainText, error) {
	const _FunctionName = "storage.sqlite.ShowDataPlain"

	plain := []storage.TournamentPlainText{
		{
			PlainText: `2025-2026 — Вторая лига "Дивизион А".

Первый этап. Группа "Золото".

Команды проводят двухкруговой турнир.
За победу начисляется 3 очка,
за ничью — 1 очко,
за поражение — 0 очков.`,
		},
	}

	return plain, nil
}
