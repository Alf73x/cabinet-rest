package sqlite

import (
	"CabinetREST/internal/config"
	"CabinetREST/internal/http-server/handlers/auth"
	"CabinetREST/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
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

/*
********************************************************************

	GetUserByLogin

********************************************************************
*/
func (s *Storage) GetUserByLogin(ctx context.Context, loginName string) (auth.User, error) {
	query := fmt.Sprintf(
		`SELECT %s, %s, %s, %s
		 FROM %s
		 WHERE %s = ?
		 LIMIT 1`,
		storage.Fld_common_id,
		storage.Fld_user_management_login_name,
		storage.Fld_user_management_user_password_hash,
		storage.Fld_user_management_enabled,
		storage.Tbl_user_management_users,
		storage.Fld_user_management_login_name,
	)

	var user auth.User
	var enabled int

	err := s.db.QueryRowContext(
		ctx,
		query,
		loginName,
	).Scan(
		&user.ID,
		&user.LoginName,
		&user.PasswordHash,
		&enabled,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, auth.ErrUserNotFound
	}

	if err != nil {
		return auth.User{}, fmt.Errorf(
			"get user by login: %w",
			err,
		)
	}
	user.Enabled = enabled != 0
	return user, nil
}

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
	seasonValue := strings.ReplaceAll(filterSeason, "'", "''")

	seasonFilter := ""
	if strings.TrimSpace(seasonValue) != "" {
		seasonFilter = fmt.Sprintf(
			" AND %s = '%s'",
			storage.Fld_class_season_season,
			seasonValue,
		)
	}

	nameValue := strings.ReplaceAll(filterName, "'", "''")

	nameFilter := ""
	if strings.TrimSpace(nameValue) != "" {
		nameFilter = fmt.Sprintf(
			" AND %s LIKE '%%%s%%'",
			storage.Fld_common_name,
			nameValue,
		)
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
		IFNULL(%s,"") , /* options_2 */
		IFNULL(%s, 0) /* icon_index */
		FROM %s  
		WHERE 1=1 %s %s %s
		ORDER BY %s DESC, %s`, storage.Fld_common_id, storage.Fld_class_season_season, storage.Fld_common_prefix, storage.Fld_common_name, storage.Fld_common_id_base,
		storage.Fld_common_group_id, storage.Fld_class_season_league_rank, storage.Fld_common_sort_order,
		storage.Fld_class_season_points, storage.Fld_class_season_options_1, storage.Fld_class_season_options_2, storage.Fld_common_icon_index,
		storage.Tbl_class_season,
		sports, seasonFilter, nameFilter,
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
			&seas.Options2,
			&seas.IconIndex)
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

func (s *Storage) Db_GetSeasonNames(idssport string) ([]string, error) {
	const _FunctionName = "storage.sqlite.Db_GetSeasonNames"

	sports := ""
	if strings.TrimSpace(idssport) != "" {
		sports = fmt.Sprintf(" AND %s IN (%s)", storage.Fld_common_id_base, idssport)
	}

	sSQL := fmt.Sprintf(`SELECT DISTINCT %s FROM %s WHERE 1=1 %s ORDER BY %s DESC`,
		storage.Fld_class_season_season, // 1
		storage.Tbl_class_season,        // 2
		sports,                          // 3
		storage.Fld_class_season_season, // 4
	)

	rows, err := s.db.Query(sSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}
	defer rows.Close()

	var seasons []string

	for rows.Next() {
		var season string

		if err := rows.Scan(&season); err != nil {
			return nil, fmt.Errorf("%s: %w", _FunctionName, err)
		}

		seasons = append(seasons, season)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
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
		IFNULL(s.` + storage.Fld_common_prefix + `, ''),
		IFNULL(s.` + storage.Fld_class_season_points + `, ''),
		IFNULL(s.` + storage.Fld_class_round_standings + `, '')
	FROM ` + storage.Tbl_class_season + ` s
	LEFT JOIN ` + storage.Tbl_class_base + ` b
	ON s.` + storage.Fld_common_id_base + ` = b.` + storage.Fld_common_id + `
	WHERE s.` + storage.Fld_common_id + ` = ?`

	var (
		name           sql.NullString
		groupID        sql.NullInt64
		season         sql.NullString
		options1       sql.NullString
		leagueRank     sql.NullInt64
		baseName       sql.NullString
		plainText      sql.NullString
		remark         sql.NullString
		prefix         sql.NullString
		points         sql.NullString
		roundStandings sql.NullString
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
		&points,
		&roundStandings,
	)
	if err != nil {
		return info, err
	}

	titlePrefix := ""
	if strings.TrimSpace(prefix.String) != "" {
		titlePrefix = prefix.String + ". "
	}

	info.Title = baseName.String + ". " + titlePrefix + name.String + ". " + season.String
	info.ViewOpt, info.ResultOf = parseViewOption(options1.String)
	info.Rank = int(leagueRank.Int64)
	info.IsOk = true
	info.RoundStandings = roundStandings.String
	info.TableFormat, info.Points = parseSeasonPoints(points.String)

	return info, err
}

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

	// Get cup winner
	idCupWinner := -1

	sSQL := "SELECT  IFNULL(" + storage.Fld_common_winner_id + ",-1) "
	sSQL += " FROM " + storage.Tbl_class_season
	sSQL += " WHERE " + storage.Fld_common_id + "=?"
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
	sSQL += "WHERE r." + storage.Fld_common_id_season + "=?"
	rows, err := s.db.Query(sSQL, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}

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
	sSQL = sSQL + " WHERE " + storage.Fld_common_id + "=?"
	var txt string
	err := s.db.QueryRow(sSQL, ids).Scan(&txt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", _FunctionName, err)
	}

	plain := []storage.TournamentPlainText{{PlainText: txt}}
	return plain, nil
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
	sSQL = sSQL + " IFNULL(t." + storage.Fld_sport_goals_against + ", 0) "
	sSQL = sSQL + " FROM " + storage.Tbl_class_season + " s "
	sSQL = sSQL + " LEFT JOIN " + storage.Tbl_class_base + " b ON b." + storage.Fld_common_id + "=s." + storage.Fld_common_id_base
	sSQL = sSQL + " LEFT JOIN " + storage.Tbl_sport_tables + " t ON t." + storage.Fld_common_id_season + "=s." + storage.Fld_common_id + " AND t." + storage.Fld_common_id_team + " IN (" + sIDs + ")"
	sSQL = sSQL + " LEFT JOIN " + storage.Tbl_class_team + " tm ON tm." + storage.Fld_common_id + "=t." + storage.Fld_common_id_team
	sSQL = sSQL + " LEFT JOIN " + storage.Tbl_countries + " c ON c." + storage.Fld_common_id + "=tm." + storage.Fld_common_id_country
	sSQL = sSQL + " WHERE tm." + storage.Fld_common_id + " IN (" + sIDs + ") "
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
		)
		if err != nil {
			return nil, err
		}
		team.Name = GetSportTeamName(teamName, ctrName)
		team.Place, err = s.GetPlaceAsStr(team.ID, team.StageIndex, team.Place)
		leagueRank, err := strconv.Atoi(team.LeagueRank)
		if err != nil {
			leagueRank = 0
		}
		team.LeagueRank = s.GetSportRankAsInt(leagueRank)
		list = append(list, team)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

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

	Db_GetOpponentOptions

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

/*
Summary
*/
func (s *Storage) Db_GetSummaryCategories() (storage.TblSummaryCategories, error) {
	const op = "storage.sqlite.Db_GetSummaryCategories"

	query := fmt.Sprintf(`SELECT DISTINCT TRIM(%[1]s) FROM %[2]s
		WHERE %[1]s IS NOT NULL AND TRIM(%[1]s) <> ''
		ORDER BY TRIM(%[1]s)`,
		storage.Fld_class_season_group, // 1
		storage.Tbl_class_season,       // 2
	)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	result := storage.TblSummaryCategories{
		{ID: 0, Name: "Всего"},
	}
	id := 1

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		result = append(result, storage.SummaryCategory{ID: id, Name: name})

		id++
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows: %w", op, err)
	}

	return result, nil
}

func (s *Storage) buildSummarySeasonFilter(category string, leagueRanks []int, yearFrom string, yearTo string, sportIDs []int) (string, []any) {
	filters := make([]string, 0, 4)
	args := make([]any, 0, 16)

	if category != "" {
		filters = append(filters, fmt.Sprintf("se.%s = ?", storage.Fld_class_season_group))
		args = append(args, category)
	}

	if len(leagueRanks) > 0 {
		placeholders := make([]string, len(leagueRanks))

		for i, rank := range leagueRanks {
			placeholders[i] = "?"
			args = append(args, rank)
		}

		filters = append(filters, fmt.Sprintf("se.%s IN (%s)", storage.Fld_class_season_league_rank, strings.Join(placeholders, ",")))
	}

	if yearFrom != "" {
		filters = append(filters, fmt.Sprintf("se.%s >= ?", storage.Fld_class_season_season))
		args = append(args, yearFrom)
	}

	if yearTo != "" {
		filters = append(filters, fmt.Sprintf("se.%s <= ?", storage.Fld_class_season_season))
		args = append(args, yearTo)
	}

	if len(sportIDs) > 0 {
		placeholders := make([]string, len(sportIDs))

		for i, sportID := range sportIDs {
			placeholders[i] = "?"
			args = append(args, sportID)
		}

		filters = append(filters, fmt.Sprintf("se.%s IN (%s)", storage.Fld_common_id_base, strings.Join(placeholders, ",")))
	}
	if len(filters) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(filters, " AND "), args
}

func buildSummaryTableTitle(
	category string,
	leagueRanks []int,
	yearFrom string,
	yearTo string,
) string {
	parts := make([]string, 0, 4)

	if category != "" {
		parts = append(parts, category)
	}

	if len(leagueRanks) > 0 {
		leagueNames := make([]string, 0, len(leagueRanks))

		for _, rank := range leagueRanks {
			switch rank {
			case 1:
				leagueNames = append(leagueNames, "Чемпионат 1")
			case 2:
				leagueNames = append(leagueNames, "Чемпионат 2")
			case 3:
				leagueNames = append(leagueNames, "Чемпионат 3")
			case 4:
				leagueNames = append(leagueNames, "Чемпионат 4")
			default:
				leagueNames = append(leagueNames, strconv.Itoa(rank))
			}
		}

		parts = append(parts, strings.Join(leagueNames, ", "))
	}

	if yearFrom != "" {
		parts = append(parts, "От: "+yearFrom)
	}

	if yearTo != "" {
		parts = append(parts, "До: "+yearTo)
	}

	return strings.Join(parts, ". ")
}

func (s *Storage) Db_GetSummaryTable(
	category string,
	leagueRanks []int,
	yearFrom string,
	yearTo string,
	sportIDs []int,
) (storage.TblSummaryTable, error) {
	const op = "storage.sqlite.Db_GetSummaryTable"

	seasonWhereSQL, seasonArgs := s.buildSummarySeasonFilter(
		category,
		leagueRanks,
		yearFrom,
		yearTo,
		sportIDs,
	)

	resultsWhereSQL := ""
	if seasonWhereSQL != "" {
		resultsWhereSQL = fmt.Sprintf(`
			WHERE r.%[1]s IN (
				SELECT se.%[2]s
				FROM %[3]s se
				%[4]s
			)`,
			storage.Fld_common_id_season, // 1
			storage.Fld_common_id,        // 2
			storage.Tbl_class_season,     // 3
			seasonWhereSQL,               // 4
		)
	}

	query := fmt.Sprintf(`
		SELECT
			tt.tid,
			tt.scnt,
			tt.sw,
			tt.sw0,
			tt.sw1,
			tt.sw2,
			tt.d,
			tt.sl,
			tt.slet,
			tt.slet1,
			tt.slet2,
			tt.ss,
			tt.sm,
			IFNULL(ct.%[1]s, ''),
			IFNULL(cou.%[1]s, ''),
			IFNULL(country.%[1]s, ''),
			IFNULL(ct.%[2]s, 0)
		FROM (
			SELECT
				tid,
				SUM(scnt) AS scnt,
				SUM(sw + sw0 + sw1 + sw2) AS w,
				SUM(d) AS d,
				SUM(sl + slet + slet1 + slet2) AS l,
				SUM(sw) AS sw,
				SUM(sw0) AS sw0,
				SUM(sw1) AS sw1,
				SUM(sw2) AS sw2,
				SUM(sl) AS sl,
				SUM(slet) AS slet,
				SUM(slet1) AS slet1,
				SUM(slet2) AS slet2,
				SUM(ss) AS ss,
				SUM(sm) AS sm
			FROM (
				SELECT
					q1.%[3]s AS tid,
					SUM(q1.Cnt) AS scnt,
					SUM(q1.W + q1.Wpm) AS sw,
					SUM(q1.L + q1.Lpm) AS sl,
					SUM(q1.Wet) AS sw0,
					SUM(q1.Let) AS slet,
					SUM(q1.D + q1.Dex + q1.Dpm) AS d,
					SUM(q1.Wex1) AS sw1,
					SUM(q1.Lex1) AS slet1,
					SUM(q1.Wex2) AS sw2,
					SUM(q1.Lex2) AS slet2,
					SUM(q1.S) AS ss,
					SUM(q1.M) AS sm
				FROM (
					SELECT
						r.%[3]s,
						1 AS Cnt,

						CASE WHEN r.%[4]s = ? AND r.%[5]s > r.%[6]s THEN 1 ELSE 0 END AS W,
						CASE WHEN r.%[4]s = ? AND r.%[5]s < r.%[6]s THEN 1 ELSE 0 END AS L,
						CASE WHEN r.%[4]s = ? AND r.%[5]s = r.%[6]s AND r.%[7]s > r.%[8]s THEN 1 ELSE 0 END AS Wet,
						CASE WHEN r.%[4]s = ? AND r.%[5]s = r.%[6]s AND r.%[7]s < r.%[8]s THEN 1 ELSE 0 END AS Let,
						CASE WHEN r.%[4]s = ? AND r.%[5]s = r.%[6]s AND r.%[7]s = r.%[8]s THEN 1 ELSE 0 END AS D,

						CASE WHEN r.%[4]s IN (?, ?, ?, ?, ?) AND r.%[5]s > r.%[6]s THEN 1 ELSE 0 END AS Wex1,
						CASE WHEN r.%[4]s IN (?, ?, ?, ?, ?) AND r.%[5]s < r.%[6]s THEN 1 ELSE 0 END AS Lex1,
						CASE WHEN r.%[4]s IN (?, ?, ?, ?, ?) AND r.%[5]s = r.%[6]s AND r.%[7]s > r.%[8]s THEN 1 ELSE 0 END AS Wex2,
						CASE WHEN r.%[4]s IN (?, ?, ?, ?, ?) AND r.%[5]s = r.%[6]s AND r.%[7]s < r.%[8]s THEN 1 ELSE 0 END AS Lex2,
						CASE WHEN r.%[4]s IN (?, ?, ?, ?, ?) AND r.%[5]s = r.%[6]s AND r.%[7]s = r.%[8]s THEN 1 ELSE 0 END AS Dex,

						CASE WHEN r.%[4]s = ? THEN 1 ELSE 0 END AS Wpm,
						CASE WHEN r.%[4]s = ? THEN 1 ELSE 0 END AS Lpm,
						CASE WHEN r.%[4]s = ? THEN 1 ELSE 0 END AS Dpm,

						CASE WHEN r.%[5]s = -1 THEN 0 ELSE r.%[5]s END AS S,
						CASE WHEN r.%[6]s = -1 THEN 0 ELSE r.%[6]s END AS M
					FROM %[9]s r
					%[10]s
				) q1
				GROUP BY q1.%[3]s

				UNION

				SELECT
					q2.%[11]s AS tid,
					SUM(q2.Cnt) AS scnt,
					SUM(q2.W + q2.Wpm) AS sw,
					SUM(q2.L + q2.Lpm) AS sl,
					SUM(q2.Wet) AS sw0,
					SUM(q2.Let) AS slet,
					SUM(q2.D + q2.Dex + q2.Dpm) AS d,
					SUM(q2.Wex1) AS sw1,
					SUM(q2.Lex1) AS slet1,
					SUM(q2.Wex2) AS sw2,
					SUM(q2.Lex2) AS slet2,
					SUM(q2.S) AS ss,
					SUM(q2.M) AS sm
				FROM (
					SELECT
						r.%[11]s,
						1 AS Cnt,

						CASE WHEN r.%[4]s = ? AND r.%[5]s > r.%[6]s THEN 1 ELSE 0 END AS L,
						CASE WHEN r.%[4]s = ? AND r.%[5]s < r.%[6]s THEN 1 ELSE 0 END AS W,
						CASE WHEN r.%[4]s = ? AND r.%[5]s = r.%[6]s AND r.%[7]s > r.%[8]s THEN 1 ELSE 0 END AS Let,
						CASE WHEN r.%[4]s = ? AND r.%[5]s = r.%[6]s AND r.%[7]s < r.%[8]s THEN 1 ELSE 0 END AS Wet,
						CASE WHEN r.%[4]s = ? AND r.%[5]s = r.%[6]s AND r.%[7]s = r.%[8]s THEN 1 ELSE 0 END AS D,

						CASE WHEN r.%[4]s IN (?, ?, ?, ?, ?) AND r.%[5]s > r.%[6]s THEN 1 ELSE 0 END AS Lex1,
						CASE WHEN r.%[4]s IN (?, ?, ?, ?, ?) AND r.%[5]s < r.%[6]s THEN 1 ELSE 0 END AS Wex1,
						CASE WHEN r.%[4]s IN (?, ?, ?, ?, ?) AND r.%[5]s = r.%[6]s AND r.%[7]s > r.%[8]s THEN 1 ELSE 0 END AS Lex2,
						CASE WHEN r.%[4]s IN (?, ?, ?, ?, ?) AND r.%[5]s = r.%[6]s AND r.%[7]s < r.%[8]s THEN 1 ELSE 0 END AS Wex2,
						CASE WHEN r.%[4]s IN (?, ?, ?, ?, ?) AND r.%[5]s = r.%[6]s AND r.%[7]s = r.%[8]s THEN 1 ELSE 0 END AS Dex,

						CASE WHEN r.%[4]s = ? THEN 1 ELSE 0 END AS Lpm,
						CASE WHEN r.%[4]s = ? THEN 1 ELSE 0 END AS Wpm,
						CASE WHEN r.%[4]s = ? THEN 1 ELSE 0 END AS Dpm,

						CASE WHEN r.%[5]s = -1 THEN 0 ELSE r.%[5]s END AS M,
						CASE WHEN r.%[6]s = -1 THEN 0 ELSE r.%[6]s END AS S
					FROM %[9]s r
					%[10]s
				) q2
				GROUP BY q2.%[11]s
			)
			GROUP BY tid
		) tt
		LEFT JOIN %[12]s ct ON ct.%[13]s = tt.tid
		LEFT JOIN %[14]s cou ON cou.%[13]s = ct.%[15]s
		LEFT JOIN %[14]s country ON country.%[13]s = ct.%[16]s
		ORDER BY
			w DESC,
			(sw + sw0 + sw1 + sw2) DESC,
			(sw0 + sw1 + sw2) DESC,
			d DESC,
			(slet + slet1 + slet2) DESC,
			(sl + slet + slet1 + slet2) ASC`,
		storage.Fld_common_name,            // 1
		storage.Fld_common_favorite,        // 2
		storage.Fld_common_id_team_1,       // 3
		storage.Fld_sport_result_type,      // 4
		storage.Fld_sport_scored,           // 5
		storage.Fld_sport_missed,           // 6
		storage.Fld_sport_scored_et,        // 7
		storage.Fld_sport_missed_et,        // 8
		storage.Tbl_sport_results,          // 9
		resultsWhereSQL,                    // 10
		storage.Fld_common_id_team_2,       // 11
		storage.Tbl_class_team,             // 12
		storage.Fld_common_id,              // 13
		storage.Tbl_countries,              // 14
		storage.Fld_common_id_country,      // 15
		storage.Fld_common_id_country+"_2", // 16
	)

	queryArgs := make([]any, 0, 80)

	appendSideArgs := func() {
		queryArgs = append(queryArgs,
			storage.RtScoreNormal,
			storage.RtScoreNormal,
			storage.RtScoreNormal,
			storage.RtScoreNormal,
			storage.RtScoreNormal,
		)

		for range 5 {
			queryArgs = append(queryArgs,
				storage.RtScoreOT,
				storage.RtScoreB,
				storage.RtScoreP,
				storage.RtScoreEt,
				storage.RtScoreAllExtra,
			)
		}

		queryArgs = append(queryArgs,
			storage.RtScorePlusMinus,
			storage.RtScoreMinusPlus,
			storage.RtScoreMinusMinus,
		)

		queryArgs = append(queryArgs, seasonArgs...)
	}

	appendSideArgs()
	appendSideArgs()

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return storage.TblSummaryTable{}, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	result := storage.TblSummaryTable{
		Rows: make([]storage.SummaryTableRow, 0),
	}

	for rows.Next() {
		var (
			row   storage.SummaryTableRow
			sw0   int
			sw1   int
			sw2   int
			slet  int
			slet1 int
			slet2 int
			fav   int
		)

		if err = rows.Scan(
			&row.TeamID,
			&row.Games,
			&row.Wins,
			&sw0,
			&sw1,
			&sw2,
			&row.Draws,
			&row.Losses,
			&slet,
			&slet1,
			&slet2,
			&row.GoalsFor,
			&row.GoalsAgainst,
			&row.TeamName,
			&row.TerritoryName,
			&row.CountryName,
			&fav,
		); err != nil {
			return storage.TblSummaryTable{}, fmt.Errorf("%s: scan: %w", op, err)
		}

		row.Place = len(result.Rows) + 1
		row.WinsET = sw0 + sw1 + sw2
		row.LossesET = slet + slet1 + slet2
		row.GoalDiff = row.GoalsFor - row.GoalsAgainst
		row.Favorite = fav > 0

		if row.Games > 0 {
			row.WinPercent = math.Round(
				float64(row.Wins+row.WinsET)*1000/float64(row.Games),
			) / 10

			row.LossPercent = math.Round(
				float64(row.Losses+row.LossesET)*1000/float64(row.Games),
			) / 10
		}

		result.Rows = append(result.Rows, row)
	}

	if err = rows.Err(); err != nil {
		return storage.TblSummaryTable{}, fmt.Errorf("%s: rows: %w", op, err)
	}

	result.Title = buildSummaryTableTitle(
		category,
		leagueRanks,
		yearFrom,
		yearTo,
	)

	return result, nil
}
