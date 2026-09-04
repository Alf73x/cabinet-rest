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
