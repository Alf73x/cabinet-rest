package sqlite

import (
	"CabinetREST/internal/storage"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

/*
Summary
*/
func (s *Storage) Db_GetSummaryCategories() (storage.TblSummaryCategories, error) {
	const op = "storage.sqlite.Db_GetSummaryCategories"

	query := fmt.Sprintf(`SELECT DISTINCT TRIM(%[1]s) FROM %[2]s
		WHERE %[1]s IS NOT NULL AND TRIM(%[1]s) <> '' AND IFNULL(%[3]s, 0) <> 1
		ORDER BY TRIM(%[1]s)`,
		storage.Fld_class_season_group, // 1
		storage.Tbl_class_season,       // 2
		storage.Fld_common_private,     // 3
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
	filters = append(filters, fmt.Sprintf("IFNULL(se.%s, 0) <> 1", storage.Fld_common_private))

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
			case 5:
				leagueNames = append(leagueNames, "Чемпионат 5")
			case 0:
				leagueNames = append(leagueNames, "Кубок")
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
	publicTeamFilter := fmt.Sprintf(`
		AND r.%[1]s IN (
			SELECT t.%[2]s FROM %[3]s t
			LEFT JOIN %[4]s c ON c.%[2]s = t.%[5]s
			WHERE IFNULL(t.%[6]s, 0) <> 1 AND IFNULL(c.%[6]s, 0) <> 1
		)
		AND r.%[7]s IN (
			SELECT t.%[2]s FROM %[3]s t
			LEFT JOIN %[4]s c ON c.%[2]s = t.%[5]s
			WHERE IFNULL(t.%[6]s, 0) <> 1 AND IFNULL(c.%[6]s, 0) <> 1
		)`,
		storage.Fld_common_id_team_1, storage.Fld_common_id,
		storage.Tbl_class_team, storage.Tbl_countries,
		storage.Fld_common_id_country, storage.Fld_common_private,
		storage.Fld_common_id_team_2,
	)
	resultsWhereSQL += publicTeamFilter

	query := fmt.Sprintf(`
		SELECT
			tt.tid,
			IFNULL(ct.%[17]s, 0),
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
		WHERE IFNULL(ct.%[18]s, 0) <> 1
		  AND IFNULL(cou.%[18]s, 0) <> 1
		  AND IFNULL(country.%[18]s, 0) <> 1
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
		storage.Fld_common_id_base,         // 17
		storage.Fld_common_private,         // 18
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
			&row.SportID,
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

/*
SeasonInfo
*/
func (s *Storage) Db_GetSeasonInfo(id int) (storage.SeasonInfo, error) {
	result := storage.SeasonInfo{}
	const op = "storage.sqlite.Db_SeasonInfo"

	sqlText := "SELECT IFNULL(" + storage.Fld_class_season_points + ", ''), " +
		" IFNULL(" + storage.Fld_class_season_options_2 + ", '')" +
		" FROM " + storage.Tbl_class_season +
		" WHERE " + storage.Fld_common_id + "=" + strconv.Itoa(id) +
		" AND IFNULL(" + storage.Fld_common_private + ", 0) <> 1"
	rows, err := s.db.Query(sqlText)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	if !rows.Next() {
		return result, nil
	}

	if err := rows.Scan(&result.Points, &result.RankingDistribution); err != nil {
		return storage.SeasonInfo{}, err
	}
	return result, nil
}

/*
TeamInfo
*/
func (s *Storage) Db_GetTeamInfo(id int) (storage.TeamsInfo, error) {
	const op = "storage.sqlite.Db_GetTeamInfo"
	result := storage.TeamsInfo{}

	sl, err := s.GetTeamTreeIDs(id)
	if err != nil {
		return result, err
	}

	if len(sl) == 0 {
		return result, nil
	}

	sIDs := ""
	for i := 0; i < len(sl); i++ {
		if sIDs != "" {
			sIDs += ","
		}
		sIDs += strconv.Itoa(sl[i])
	}

	sqlText := "SELECT " +
		"t." + storage.Fld_common_id + ", " +
		"TRIM(t." + storage.Fld_common_name + " || ' ' || IFNULL(c." + storage.Fld_common_name + ", '')), " +
		"IFNULL(t." + storage.Fld_common_founded_date + ", ''), " +
		"IFNULL(t." + storage.Fld_common_disbanded_date + ", '') " +
		"FROM " + storage.Tbl_class_team + " t " +
		"LEFT JOIN " + storage.Tbl_countries + " c ON c." + storage.Fld_common_id + " = t." + storage.Fld_common_id_country +
		" WHERE t." + storage.Fld_common_id + " IN (" + sIDs + ")" +
		" AND IFNULL(t." + storage.Fld_common_private + ", 0) <> 1" +
		" AND IFNULL(c." + storage.Fld_common_private + ", 0) <> 1"

	rows, err := s.db.Query(sqlText)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var item storage.TeamInfo

		if err := rows.Scan(
			&item.Id,
			&item.Name,
			&item.DateFrom,
			&item.DateTo,
		); err != nil {
			return result, err
		}

		slFrom := strings.Split(item.DateFrom, ",")
		slTo := strings.Split(item.DateTo, ",")

		for i := 0; i < len(slFrom); i++ {
			newItem := storage.TeamInfo{
				Id:       item.Id,
				Name:     item.Name,
				DateFrom: strings.TrimSpace(slFrom[i]),
			}

			if i < len(slTo) {
				newItem.DateTo = strings.TrimSpace(slTo[i])
			}

			result.Teams = append(result.Teams, newItem)
		}
	}

	if err := rows.Err(); err != nil {
		return result, err
	}

	sort.Slice(result.Teams, func(i, j int) bool {
		return teamInfoDateSortValue(result.Teams[i].DateFrom) >
			teamInfoDateSortValue(result.Teams[j].DateFrom)
	})

	return result, nil
}

func teamInfoDateSortValue(s string) int {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "?")

	if s == "" || s == "?" {
		return 0
	}

	sl := strings.Split(s, ".")

	year, _ := strconv.Atoi(sl[0])
	month := 0
	day := 0

	if len(sl) > 1 {
		month, _ = strconv.Atoi(sl[1])
	}
	if len(sl) > 2 {
		day, _ = strconv.Atoi(sl[2])
	}

	return year*10000 + month*100 + day
}
