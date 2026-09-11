package sqlite

import (
	"CabinetREST/internal/storage"
	"database/sql"
	"fmt"
	"strings"
)

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
		IFNULL(%s,""), /* options_2 */
		IFNULL(%s, 0), /* icon_index */
		IFNULL(%s,""), /* plain_text */
        IFNULL(%s,"")  /* remark_text */
		FROM %s  
		WHERE IFNULL(%s, 0) <> 1 %s %s %s
		ORDER BY %s DESC, %s`, storage.Fld_common_id, storage.Fld_class_season_season, storage.Fld_common_prefix, storage.Fld_common_name, storage.Fld_common_id_base,
		storage.Fld_common_group_id, storage.Fld_class_season_league_rank, storage.Fld_common_sort_order,
		storage.Fld_class_season_points, storage.Fld_class_season_options_1, storage.Fld_class_season_options_2, storage.Fld_common_icon_index, storage.Fld_class_season_plain_text, storage.Fld_class_season_remark_text,
		storage.Tbl_class_season,
		storage.Fld_common_private, sports, seasonFilter, nameFilter,
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
			&seas.IconIndex,
			&seas.PlainText,
			&seas.RemarkText)
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

	sSQL := fmt.Sprintf(`SELECT DISTINCT %s FROM %s WHERE IFNULL(%s, 0) <> 1 %s ORDER BY %s DESC`,
		storage.Fld_class_season_season, // 1
		storage.Tbl_class_season,        // 2
		storage.Fld_common_private,      // 3
		sports,                          // 4
		storage.Fld_class_season_season, // 5
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
	WHERE s.` + storage.Fld_common_id + ` = ?
	AND IFNULL(s.` + storage.Fld_common_private + `, 0) <> 1`

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
	info.PlainText = plainText.String
	info.RemarkText = remark.String

	return info, err
}

/*********************************************************************
Db_GetSeasonByID
**********************************************************************/

func (s *Storage) Db_GetSeasonByID(id int) (storage.TblSeason, error) {
	const _FunctionName = "storage.sqlite.Db_GetSeasonByID"

	sSQL := fmt.Sprintf(`
		SELECT
			%s,
			%s,
			IFNULL(%s, ''),
			%s,
			%s,
			%s,
			IFNULL(%s, 0),
			IFNULL(%s, 0),
			IFNULL(%s, ''),
			IFNULL(%s, ''),
			IFNULL(%s, ''),
			IFNULL(%s, 0),
			IFNULL(%s, ''),
			IFNULL(%s, '')
		FROM %s
		WHERE %s = ?
		  AND IFNULL(%s, 0) <> 1
	`,
		storage.Fld_common_id,
		storage.Fld_class_season_season,
		storage.Fld_common_prefix,
		storage.Fld_common_name,
		storage.Fld_common_id_base,
		storage.Fld_common_group_id,
		storage.Fld_class_season_league_rank,
		storage.Fld_common_sort_order,
		storage.Fld_class_season_points,
		storage.Fld_class_season_options_1,
		storage.Fld_class_season_options_2,
		storage.Fld_common_icon_index,
		storage.Fld_class_season_plain_text,
		storage.Fld_class_season_remark_text,
		storage.Tbl_class_season,
		storage.Fld_common_id,
		storage.Fld_common_private,
	)

	var season storage.TblSeason

	err := s.db.QueryRow(sSQL, id).Scan(
		&season.ID,
		&season.Season,
		&season.Prefix,
		&season.Name,
		&season.SportID,
		&season.GroupID,
		&season.LeagueRank,
		&season.SortOrder,
		&season.Points,
		&season.Options1,
		&season.Options2,
		&season.IconIndex,
		&season.PlainText,
		&season.RemarkText,
	)
	if err != nil {
		return storage.TblSeason{},
			fmt.Errorf("%s: %w", _FunctionName, err)
	}

	return season, nil
}

/*
********************************************************************
Db_GetDestinationSeasonID
*********************************************************************
*/
func (s *Storage) Db_GetDestinationSeasonID(
	id int,
	direction string,
) (int, error) {
	const _FunctionName = "storage.sqlite.Db_GetDestinationSeasonID"

	/*
		1. Получаем параметры текущего турнира.
	*/
	var (
		idBase     int
		season     string
		seasonName string
		leagueRank int
	)

	query := fmt.Sprintf(`
		SELECT
			%s,
			%s,
			%s,
			IFNULL(%s, 0)
		FROM %s
		WHERE %s = ?
	`,
		storage.Fld_common_id_base,
		storage.Fld_class_season_season,
		storage.Fld_common_name,
		storage.Fld_class_season_league_rank,
		storage.Tbl_class_season,
		storage.Fld_common_id,
	)

	err := s.db.QueryRow(query, id).Scan(
		&idBase,
		&season,
		&seasonName,
		&leagueRank,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}

		return 0, fmt.Errorf("%s: get current season: %w", _FunctionName, err)
	}

	/*
		2. Ищем соседний сезон с тем же league_rank.
	*/
	var (
		operator string
		order    string
	)

	switch direction {
	case "prev":
		// Предыдущий сезон:
		// 2006 -> 2005
		operator = "<"
		order = "DESC"
	case "next":
		// Следующий сезон:
		// 2006 -> 2007
		operator = ">"
		order = "ASC"
	default:
		return 0, fmt.Errorf("%s: invalid direction %q", _FunctionName, direction)
	}
	var destSeason string

	query = fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?
		  AND %s %s ?
		  AND IFNULL(%s, 0) = ?
		ORDER BY %s %s, %s %s
		LIMIT 1
	`,
		storage.Fld_class_season_season,
		storage.Tbl_class_season,

		storage.Fld_common_id_base,

		storage.Fld_class_season_season,
		operator,

		storage.Fld_class_season_league_rank,

		storage.Fld_class_season_season,
		order,

		storage.Fld_common_sort_order,
		order,
	)

	err = s.db.QueryRow(query, idBase, season, leagueRank).Scan(&destSeason)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Предыдущего/следующего сезона нет.
		}

		return 0, fmt.Errorf("%s: find destination season: %w", _FunctionName, err)
	}

	/*
		3. Сначала пытаемся найти в найденном сезоне  турнир с тем же названием.
		Это аналог:
		q1.Filter := season = sDestSeason  AND league_rank = league_rank  AND name = season_name
	*/
	var destID int

	query = fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?
		  AND IFNULL(%s, 0) = ?
		  AND %s = ?
		ORDER BY %s
		LIMIT 1
	`,
		storage.Fld_common_id,
		storage.Tbl_class_season,

		storage.Fld_class_season_season,
		storage.Fld_class_season_league_rank,
		storage.Fld_common_name,

		storage.Fld_common_sort_order,
	)

	err = s.db.QueryRow(query, destSeason, leagueRank, seasonName).Scan(&destID)
	if err == nil {
		return destID, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("%s: find destination season by name: %w", _FunctionName, err)
	}

	/*
		4. Турнира с таким же названием нет.

		   Delphi тогда берёт первый турнир найденного   сезона с тем же league_rank.
	*/
	query = fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?
		  AND IFNULL(%s, 0) = ?
		ORDER BY %s
		LIMIT 1
	`,
		storage.Fld_common_id,
		storage.Tbl_class_season,

		storage.Fld_class_season_season,
		storage.Fld_class_season_league_rank,

		storage.Fld_common_sort_order,
	)

	err = s.db.QueryRow(query, destSeason, leagueRank).Scan(&destID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}

		return 0, fmt.Errorf("%s: find destination season fallback: %w", _FunctionName, err)
	}

	return destID, nil
}
