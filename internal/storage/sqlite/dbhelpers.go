package sqlite

import (
	"CabinetREST/internal/config"
	"CabinetREST/internal/storage"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	kWin     = 1
	kLoose   = 2
	kWinET   = 3
	kLooseET = 4
	kDraw    = 5

	kExtraMaskBorder = 1000

	kColorWin       = "#d4edda"
	kColorLoose     = "#f8d7da"
	kColorWinET     = "#cce5ff"
	kColorLooseET   = "#ffe5b4"
	kColorDarkWin   = "#1f5f2c"
	kColorDarkLoose = "#7a1f1f"
)

var treeChildrenInitialized bool = false
var gListTree []storage.TreeID
var maxPlaceCache = make(map[int]string)

func GetBaseFilter(AsIDs, AsPrefix string) string {

	return "(" + AsPrefix + storage.Fld_common_id_base + " IN (" + AsIDs + " ))"
}

func (s *Storage) Tree_TreeChildrenFilter(AiID int, AsPrefix string) string {
	if AiID == -1 {
		return "(" + AsPrefix + storage.Fld_common_id_country + " IN (-1))"
	} else {
		return "(" + AsPrefix + storage.Fld_common_id_country + " IN (" + s.Tree_FindTreeChildrenIDs(AiID) + "))"
	}
}

func (s *Storage) Tree_FindTreeChildrenIDs(AiID int) string {
	sOutputIDs := ""
	if AiID == -1 {
		return "-1"
	}

	if !treeChildrenInitialized {
		gListTree = gListTree[:0]

		sSQL := "SELECT " + storage.Fld_common_id +
			",IFNULL(" + storage.Fld_countries_id_parent + ",-1)" +
			",IFNULL(" + storage.Fld_countries_issues + ",0)" +
			",IFNULL(" + storage.Fld_countries_id_parent + "_2,-1)"

		for i := 1; i <= storage.KSubIDsCount; i++ {
			sSQL = sSQL + ",IFNULL(" + storage.Fld_common_sub_id + "_" + strconv.Itoa(i) + ",-1)"
		}
		sSQL = sSQL + " FROM " + storage.Tbl_countries + " ORDER BY " + storage.Fld_countries_id_parent
		rows, err := s.db.Query(sSQL)
		if err == nil {
			defer rows.Close()

			for rows.Next() {
				var t storage.TreeID

				args := []any{
					&t.ID,
					&t.ParentID,
					&t.Issues,
					&t.SecondParentID,
				}

				for i := 0; i < storage.KSubIDsCount; i++ {
					t.SubIDs[i] = -1
					args = append(args, &t.SubIDs[i])
				}

				if err := rows.Scan(args...); err != nil {
					return "-1"
				}

				gListTree = append(gListTree, t)
			}
		}
		treeChildrenInitialized = true
	}
	sOutputIDs = strconv.Itoa(AiID)

	s.FindTreeChildrenIDs(AiID, &sOutputIDs)
	if sOutputIDs == "" {
		return "-1"
	}
	return sOutputIDs

}

func (s *Storage) FindTreeChildrenIDs(aiID int, outputIDs *string) string {
	low := 0
	high := len(gListTree) - 1

	addToList := func(k int) {
		idText := strconv.Itoa(gListTree[k].ID)

		if !strings.Contains(*outputIDs+",", ","+idText+",") {
			*outputIDs += "," + idText
			s.FindTreeChildrenIDs(gListTree[k].ID, outputIDs)
		}
	}

	for low <= high {
		mid := (low + high) / 2
		if mid > len(gListTree)-1 {
			return *outputIDs
		}
		idParent := gListTree[mid].ParentID
		if idParent == aiID {
			k := mid
			for k >= 0 && gListTree[k].ParentID == aiID {
				addToList(k)
				k--
			}
			k = mid + 1
			for k < len(gListTree) && gListTree[k].ParentID == aiID {
				addToList(k)
				k++
			}
			return *outputIDs
		} else if idParent > aiID {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return *outputIDs
}

func StageIndexToText(aiIndex int) string {
	switch aiIndex {
	case 1:
		return "Финал"
	case 2:
		return "1/2"
	case 3:
		return "1/4"
	case 4:
		return "1/8"
	case 5:
		return "1/16"
	case 6:
		return "1/32"
	case 7:
		return "1/64"
	case 8:
		return "1/128"
	case 9:
		return "1/256"
	case 10:
		return "1/512"
	case 11:
		return "1/1024"
	case 50:
		return "За 5-8 место"
	case 51:
		return "За 7 место"
	case 52:
		return "За 5 место"
	case 53, 54:
		return "За 3 место"
	case 57:
		return "За 2 место"
	case 65:
		return "За 1 место"
	case 66:
		return "За 9 место"
	case 67:
		return "За 11 место"
	case 68:
		return "За 13 место"
	case 69:
		return "За 15 место"
	case 85:
		return "За 16 место"
	case 70:
		return "За 17 место"
	case 71:
		return "За 19 место"
	case 72:
		return "За 21 место"
	case 73:
		return "За 23 место"
	case 74:
		return "За 25 место"
	case 75:
		return "За 27 место"
	case 76:
		return "За 29 место"
	case 77:
		return "За 31 место"
	case 78:
		return "За 33 место"
	case 79:
		return "За 35 место"
	case 80:
		return "За 37 место"
	case 81:
		return "За 39 место"

	case 180:
		return "Путь регионов. 1-й раунд"
	case 181:
		return "Путь регионов. 2-й раунд"
	case 182:
		return "Путь регионов. 3-й раунд"
	case 183:
		return "Путь регионов. 4-й раунд"
	case 184:
		return "Путь регионов. 5-й раунд"
	case 185:
		return "Путь регионов. 6-й раунд"

	case 210:
		return "Путь РПЛ. Группа A"
	case 211:
		return "Путь РПЛ. Группа B"
	case 212:
		return "Путь РПЛ. Группа C"
	case 213:
		return "Путь РПЛ. Группа D"
	case 214:
		return "Путь РПЛ. 1/4 финала"
	case 215:
		return "Путь РПЛ. 1/2 финала"
	case 218:
		return "Путь РПЛ. Финал"

	case 225:
		return "Путь регионов. 1/4 финала. 1-й этап"
	case 226:
		return "Путь регионов. 1/4 финала. 2-й этап"
	case 227:
		return "Путь регионов. 1/2 финала. 1-й этап"
	case 228:
		return "Путь регионов. 1/2 финала. 2-й этап"
	case 229:
		return "Путь регионов. Финал"

	case 320:
		return "Группа A"
	case 321:
		return "Группа B"
	case 322:
		return "Группа C"
	case 323:
		return "Группа D"
	case 324:
		return "Группа E"
	case 325:
		return "Группа F"
	case 326:
		return "Группа G"
	case 327:
		return "Группа H"

	case 390:
		return "Этап победителей"
	}

	if aiIndex >= 150 && aiIndex <= 175 {
		return "Элитный групповой раунд. Группа " + strconv.Itoa(aiIndex-150+1)
	}

	if aiIndex >= 361 && aiIndex <= 370 {
		return strconv.Itoa(aiIndex-361+1) + " группа"
	}

	if aiIndex >= 400 && aiIndex <= 3699 {
		sFirstPart := ""
		sSecondPart := ""

		switch {
		case aiIndex >= 400 && aiIndex <= 499:
			sFirstPart = ""
		case aiIndex >= 500 && aiIndex <= 599:
			sFirstPart = "РСФСР"
		case aiIndex >= 600 && aiIndex <= 699:
			sFirstPart = "УССР"
		case aiIndex >= 700 && aiIndex <= 799:
			sFirstPart = "УССР. Зона Закарпатья"
		case aiIndex >= 800 && aiIndex <= 899:
			sFirstPart = "УССР. Зона Крыма"
		case aiIndex >= 900 && aiIndex <= 999:
			sFirstPart = "Средняя Азия"
		case aiIndex >= 1000 && aiIndex <= 1099:
			sFirstPart = "Средняя Азия и Казахстан"
		case aiIndex >= 1100 && aiIndex <= 1199:
			sFirstPart = "Союзные республики"
		case aiIndex >= 1200 && aiIndex <= 1299:
			sFirstPart = "Центральная зона"
		case aiIndex >= 1300 && aiIndex <= 1399:
			sFirstPart = "УССР"
		case aiIndex >= 1400 && aiIndex <= 1499:
			sFirstPart = "Украинская зона"
		case aiIndex >= 1500 && aiIndex <= 1599:
			sFirstPart = "Закавказская зона"
		case aiIndex >= 1600 && aiIndex <= 1699:
			sFirstPart = "Среднеазиатская зона"
		case aiIndex >= 1700 && aiIndex <= 1799:
			sFirstPart = "I зона, 1 группа. Москва"
		case aiIndex >= 1800 && aiIndex <= 1899:
			sFirstPart = "I зона, 2 группа. Москва"
		case aiIndex >= 1900 && aiIndex <= 1999:
			sFirstPart = "II зона, Ленинград"
		case aiIndex >= 2000 && aiIndex <= 2099:
			sFirstPart = "III зона, Воронеж"
		case aiIndex >= 2100 && aiIndex <= 2199:
			sFirstPart = "IV зона, Хабаровск"
		case aiIndex >= 2200 && aiIndex <= 2299:
			sFirstPart = "V зона, Новосибирск"
		case aiIndex >= 2300 && aiIndex <= 2399:
			sFirstPart = "VI зона, Свердловск"
		case aiIndex >= 2400 && aiIndex <= 2499:
			sFirstPart = "VII зона, Горький"
		case aiIndex >= 2500 && aiIndex <= 2599:
			sFirstPart = "VIII зона, Нижневолжская"
		case aiIndex >= 2600 && aiIndex <= 2699:
			sFirstPart = "IX зона, Ростов"
		case aiIndex >= 2700 && aiIndex <= 2799:
			sFirstPart = "X зона, Тбилиси"
		case aiIndex >= 2800 && aiIndex <= 2899:
			sFirstPart = "XI зона, Баку"
		case aiIndex >= 2900 && aiIndex <= 2999:
			sFirstPart = "XII зона, Ташкент"
		case aiIndex >= 3000 && aiIndex <= 3099:
			sFirstPart = "XIII зона, Минск"
		case aiIndex >= 3100 && aiIndex <= 3199:
			sFirstPart = "XIV зона, Харьков (1 зона УССР)"
		case aiIndex >= 3200 && aiIndex <= 3299:
			sFirstPart = "XV зона, Киев (2 зона УССР)"
		case aiIndex >= 3300 && aiIndex <= 3399:
			sFirstPart = "XVI зона, Одесса (3 зона УССР)"
		case aiIndex >= 3400 && aiIndex <= 3499:
			sFirstPart = "XVII зона, Днепропетровск (4 зона УССР)"
		case aiIndex >= 3500 && aiIndex <= 3599:
			sFirstPart = "XVIII зона, Сталино (5 зона УССР)"
		case aiIndex >= 3600 && aiIndex <= 3690:
			sFirstPart = "XIX зона, Симферополь"
		case aiIndex == 3696:
			sFirstPart = "XIV зона, Харьков (1 зона УССР). 1/512 финала"
		case aiIndex == 3697:
			sFirstPart = "XVIII зона, Сталино (5 зона УССР). 1/512 финала"
		case aiIndex == 3698:
			sFirstPart = "V зона, Новосибирск. 1/512 финала"
		case aiIndex == 3699:
			sFirstPart = "Стыковые игры победителей зон"
		}

		if aiIndex != 3696 && aiIndex != 3697 && aiIndex != 3698 && aiIndex != 3699 {
			iMod := aiIndex % 10

			switch iMod {
			case 1:
				sSecondPart = "Финал"
			case 2:
				sSecondPart = "1/2"
			case 3:
				sSecondPart = "1/4"
			case 4:
				sSecondPart = "1/8"
			case 5:
				sSecondPart = "1/16"
			case 6:
				sSecondPart = "1/32"
			case 7:
				sSecondPart = "1/64"
			case 8:
				sSecondPart = "1/128"
			case 9:
				sSecondPart = "1/256"
			}

			if sFirstPart != "" {
				sFirstPart += ". "
			}

			iDiv := aiIndex / 10
			if iDiv%10 != 0 {
				sFirstPart += strconv.Itoa(iDiv%10) + " зона. "
			}
		}

		return sFirstPart + sSecondPart
	}

	return ""
}

func (s *Storage) GetPlaceAsStr(seasonID int, stageIndex int, sportPlace string) (string, error) {

	placeText := ""

	if stageIndex > 0 {
		if sportPlace == "1" {
			placeText = config.GsSportWinner
		} else {
			placeText = StageIndexToText(stageIndex)
		}
	} else {
		if maxPlaceStr, ok := maxPlaceCache[seasonID]; ok {
			placeText = sportPlace + maxPlaceStr
		} else {
			sqlText := fmt.Sprintf(
				"SELECT MAX(%s) AS mx FROM %s WHERE id_season = ?",
				storage.Fld_common_sport_place,
				storage.Tbl_sport_tables,
			)

			var mx sql.NullString
			err := s.db.QueryRow(sqlText, seasonID).Scan(&mx)
			if err != nil && err != sql.ErrNoRows {
				return "", err
			}

			s := ""
			if mx.Valid && mx.String != "" {
				s = " " + config.GsFrom + " " + mx.String
			}

			placeText = sportPlace + s
			maxPlaceCache[seasonID] = s
		}
	}
	return placeText, nil
}

func (s *Storage) GetSportRankAsInt(iRank int) string {
	var result string

	switch {
	case iRank >= storage.KPlayoffMin && iRank <= storage.KPlayoffMax:
		result = strconv.Itoa(iRank - storage.KSeasonsPlayoff_Delta)

	case iRank >= storage.KPlayoffMatchesMin && iRank <= storage.KPlayoffMatchesMax:
		result = strconv.Itoa(iRank - storage.KSeasonsPlayoff_Delta*2)

	case iRank == storage.KSeasonsRank_Friendly,
		iRank == storage.KSeasonsRank_PreSeason,
		iRank == storage.KSeasonsNoRank,
		iRank == storage.KSeasonsRankCupTournament,
		iRank == storage.KRank_Tournament:
		result = ""

	default:
		result = strconv.Itoa(iRank)
	}

	if result == "0" {
		result = ""
	}

	return result
}

func (s *Storage) GetTeamTreeIDs(teamID int) ([]int, error) {
	result := []int{teamID}
	used := map[int]bool{
		teamID: true,
	}

	var getChildren func(id int) error

	getChildren = func(id int) error {
		rows, err := s.db.Query(
			"SELECT "+storage.Fld_common_id+
				" FROM "+storage.Tbl_class_team+
				" WHERE "+storage.Fld_common_id_successor+" = ?",
			id,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var childID int
			if err := rows.Scan(&childID); err != nil {
				return err
			}
			if used[childID] {
				continue
			}
			used[childID] = true
			result = append(result, childID)
			if err := getChildren(childID); err != nil {
				return err
			}
		}
		return rows.Err()
	}

	err := getChildren(teamID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Storage) GetTeamTreeIDsAsString(teamID int) (string, error) {
	ids, err := s.GetTeamTreeIDs(teamID)
	if err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "-1", nil
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.Itoa(id))
	}
	return strings.Join(parts, ","), nil
}

func Sport_IsList(idx int) bool {
	return idx == storage.KSeasonsRankCup ||
		idx == storage.KSeasonsRankCupTournament ||
		(idx >= storage.KPlayoffMin && idx <= storage.KPlayoffMax) ||
		(idx >= storage.KPlayoffMatchesMin && idx <= storage.KPlayoffMatchesMax) ||
		(idx >= storage.KPlayoff_InternationalMin && idx <= storage.KPlayoff_InternationalMax)
}

func Sport_IsTable(idx int) bool {
	return (idx >= storage.KSeasonsRankMin && idx <= storage.KSeasonsRankMax) ||
		(idx >= storage.KSeasonsRank_InternationalMin && idx <= storage.KSeasonsRank_InternationalMax) ||
		idx == storage.KRank_Tournament
}

func GetSportTeamName(team, city string) string {
	result := strings.TrimSpace(team + " " + city)

	teamLower := strings.ToLower(team)

	if strings.HasPrefix(teamLower, "команда города ") {
		return team
	}

	if strings.HasPrefix(teamLower, "команда ") &&
		strings.Contains(teamLower, "области") {
		return team
	}

	return result
}

func SportScoreAsTxt(resultType, score, missed, scoreP, missedP int) string {
	scoreText := strconv.Itoa(score) + ":" + strconv.Itoa(missed)

	extraScore := ""
	if scoreP != -1 && missedP != -1 {
		extraScore = strconv.Itoa(scoreP) + ":" + strconv.Itoa(missedP)
	}

	addET := func(buf string) string {
		if extraScore != "" {
			return scoreText + " (" + strings.ToLower(buf) + " " + extraScore + ")"
		}
		return scoreText + " " + buf
	}

	switch resultType {
	case storage.RtScorePlusMinus:
		return "+:-"
	case storage.RtScoreMinusPlus:
		return "-:+"
	case storage.RtScoreMinusMinus:
		return "-:-"
	case storage.RtScoreWL:
		return config.GsWP
	case storage.RtScoreLW:
		return config.GsPW
	case storage.RtScoreDD:
		return config.GsNN
	case storage.RtScoreQuestion:
		return "?:?"
	case storage.RtScoreOT:
		return addET(config.GsSportShortOvertime)
	case storage.RtScoreEt:
		return addET(config.GsSportShortExtraTime)
	case storage.RtScoreB:
		return addET(config.GsSportShortShootout)
	case storage.RtScoreP:
		return addET(config.GsSportShortPenalty)
	case storage.RtScoreAllExtra:
		return addET("*")
	default:
		return scoreText
	}
}

func SportScoreAsTxtShort(resultType, score, missed, scoreP, missedP int) string {
	scoreText := strconv.Itoa(score) + ":" + strconv.Itoa(missed)
	switch resultType {
	case storage.RtScorePlusMinus:
		return "+:-"
	case storage.RtScoreMinusPlus:
		return "-:+"
	case storage.RtScoreMinusMinus:
		return "-:-"
	case storage.RtScoreWL:
		return config.GsWP
	case storage.RtScoreLW:
		return config.GsPW
	case storage.RtScoreDD:
		return config.GsNN
	case storage.RtScoreQuestion:
		return "?:?"
	case storage.RtScoreOT,
		storage.RtScoreEt,
		storage.RtScoreB,
		storage.RtScoreP,
		storage.RtScoreAllExtra:
		return scoreText + "*"
	default:
		return scoreText
	}
}

func GetSportScoreColor(mask int, defaultColor string, isDarkSkin bool) string {
	if mask >= kExtraMaskBorder {
		mask -= kExtraMaskBorder
	}

	if isDarkSkin {
		switch mask {
		case kWin:
			return kColorDarkWin
		case kLoose:
			return kColorDarkLoose
		case kWinET:
			return kColorWinET
		case kLooseET:
			return kColorLooseET
		case kDraw:
			return defaultColor
		}

		return defaultColor
	}

	switch mask {
	case kWin:
		return kColorWin
	case kLoose:
		return kColorLoose
	case kWinET:
		return kColorWinET
	case kLooseET:
		return kColorLooseET
	case kDraw:
		return defaultColor
	}

	return defaultColor
}

func GetScoresResult(result_type, scored, scored_et, missed, missed_et int) int {
	result := kDraw

	if scored > missed {
		result = kWin
	} else if scored < missed {
		result = kLoose
	} else {
		if result_type == storage.RtScorePlusMinus {
			result = kWin
		} else if result_type == storage.RtScoreMinusPlus {
			result = kLoose
		} else if result_type == storage.RtScoreMinusMinus {
			result = kDraw
		} else if result_type > missed_et {
			result = kWinET
		} else if result_type < missed_et {
			result = kLooseET
		} else if result_type == storage.RtScoreWL {
			result = kWin
		} else if result_type == storage.RtScoreLW {
			result = kLoose
		}
	}

	return result
}

func StrToIntDef(s string, def int) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return def
	}
	return v
}

func ParseSeasonOptions(options1 string) storage.SeasonOptions {
	result := storage.SeasonOptions{}

	s := strings.TrimSpace(options1)
	if s == "" {
		return result
	}

	parts := strings.Split(s, ";")

	for _, part := range parts {
		s := strings.ToUpper(strings.TrimSpace(part))

		if strings.HasPrefix(s, storage.KOptionsPlusMinus+"=") {
			value := strings.TrimSpace(strings.TrimPrefix(s, storage.KOptionsPlusMinus+"="))

			p := strings.SplitN(value, ":", 2)
			if len(p) == 2 {
				result.PlusScored = StrToIntDef(p[0], 0)
				result.MinusScored = StrToIntDef(p[1], 100)
			}

		} else if strings.HasPrefix(s, storage.KOptionsMinusMinus+"=") {
			value := strings.TrimSpace(strings.TrimPrefix(s, storage.KOptionsMinusMinus+"="))
			result.MinusMinusScored = StrToIntDef(value, 0)

		} else if strings.HasPrefix(s, storage.KOptionsV+"=") {
			result.ViewOption =
				strings.TrimSpace(strings.TrimPrefix(s, storage.KOptionsV+"="))

		} else if strings.HasPrefix(s, storage.KOptionsParent+"=") {
			result.ParentSeasonIDs =
				strings.TrimSpace(strings.TrimPrefix(s, storage.KOptionsParent+"="))

		} else if strings.HasPrefix(s, storage.KOptionsJoin+"=") {
			result.JoinSeasonIDs =
				strings.TrimSpace(strings.TrimPrefix(s, storage.KOptionsJoin+"="))

		} else if strings.HasPrefix(s, storage.KOptionsDrawlimit+"=") {
			value := strings.TrimSpace(strings.TrimPrefix(s, storage.KOptionsDrawlimit+"="))
			result.DrawLimit = StrToIntDef(value, 0)

		} else if strings.HasPrefix(s, storage.KOptionsMode+"=") {
			result.Mode =
				strings.TrimSpace(strings.TrimPrefix(s, storage.KOptionsMode+"="))

		} else if strings.HasPrefix(s, storage.KOptionsRoot+"=") {
			value := strings.TrimSpace(strings.TrimPrefix(s, storage.KOptionsRoot+"="))
			result.Root = StrToIntDef(value, -1)

		} else if strings.HasPrefix(s, storage.KOptionsResultsOf+"=") {
			value := strings.TrimSpace(strings.TrimPrefix(s, storage.KOptionsResultsOf+"="))
			result.ResultsOf = StrToIntDef(value, -1)
		}
	}

	return result
}

func parseViewOption(options string) (view string, resultsOf int) {
	resultsOf = -1

	options = strings.TrimSpace(options)
	if options == "" {
		return
	}

	parts := strings.Split(options, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		switch {
		case strings.HasPrefix(part, storage.KOptionsV+"="):
			view = strings.ToUpper(
				strings.TrimSpace(
					strings.TrimPrefix(part, storage.KOptionsV+"="),
				),
			)

		case strings.HasPrefix(strings.ToUpper(part), strings.ToUpper(storage.KOptionsResultsOf+"=")):
			s := strings.TrimSpace(
				part[len(storage.KOptionsResultsOf)+1:],
			)
			if n, err := strconv.Atoi(s); err == nil {
				resultsOf = n
			}
		}
	}

	return
}

func parseIntDef(value string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return def
	}
	return n
}

func parseSeasonPoints(points string) (resTableFormat int, pts storage.Points) {
	resTableFormat = -1
	pts.Wins = 0
	pts.WinsET = 0
	pts.Draws = 0
	pts.LossesET = 0
	pts.Losses = 0
	pts.NA = 0

	points = strings.TrimSpace(points)
	if points == "" {
		return
	}

	options := strings.Split(points, ";")

	for _, option := range options {
		option = strings.ToUpper(strings.TrimSpace(option))
		if option == "" {
			continue
		}

		parts := strings.Split(option, ",")
		resTableFormat = len(parts)

		switch resTableFormat {
		case 3:
			pts.Wins = parseIntDef(parts[0], 0)
			pts.WinsET = 0
			pts.Draws = parseIntDef(parts[1], 0)
			pts.LossesET = 0
			pts.Losses = parseIntDef(parts[2], 0)
		case 4:
			pts.Wins = parseIntDef(parts[0], 0)
			pts.WinsET = 0
			pts.Draws = parseIntDef(parts[1], 0)
			pts.LossesET = 0
			pts.Losses = parseIntDef(parts[2], 0)
			pts.NA = parseIntDef(parts[3], 0) + 1
		case 5:
			pts.Wins = parseIntDef(parts[0], 0)
			pts.WinsET = parseIntDef(parts[1], 0)
			pts.Draws = parseIntDef(parts[2], 0)
			pts.LossesET = parseIntDef(parts[3], 0)
			pts.Losses = parseIntDef(parts[4], 0)
		}
	}

	return
}

func SportDateToText(src string) string {
	switch len(src) {
	case 8:
		return src[0:4] + "." + src[4:6] + "." + src[6:8]

	case 6:
		return src[0:4] + "." + src[4:6]

	case 4:
		return src[0:4]
	}

	return ""
}

func SportGetStageValue(aiStage int, aiBase ...int) int {
	result := 0

	switch aiStage {
	case 0:
		result = 0 // Минимум

	case 1:
		result = 250 // Финал (максимум)
	case 2:
		result = 220 // 1/2 финала
	case 3:
		result = 180 // 1/4 финала
	case 4:
		result = 150 // 1/8 финала
	case 5:
		result = 120 // 1/16 финала
	case 6:
		result = 90 // 1/32 финала
	case 7:
		result = 60 // 1/64 финала
	case 8:
		result = 40 // 1/128 финала
	case 9:
		result = 20 // 1/256 финала
	case 10:
		result = 10 // 1/512 финала
	case 11:
		result = 5 // 1/1024 финала

	case 50:
		result = 140 // Матчи за 5–8 место
	case 51:
		result = 145 // Матч за 7 место
	case 52:
		result = 147 // Матч за 5 место
	case 53:
		result = 210 // Матч за 3 место
	case 54:
		result = 210 // 3-е место
	case 57:
		result = 215 // 2-е место
	case 65:
		result = 250 // 1-е место

	case 66:
		result = 138 // 9-е место
	case 67:
		result = 136 // 11-е место
	case 68:
		result = 134 // 13-е место
	case 69:
		result = 132 // 15-е место
	case 85:
		result = 131 // 16-е место
	case 70:
		result = 130 // 17-е место
	case 71:
		result = 128 // 19-е место
	case 72:
		result = 126 // 21-е место
	case 73:
		result = 124 // 23-е место
	case 74:
		result = 122 // 25-е место
	case 75:
		result = 120 // 27-е место
	case 76:
		result = 118 // 29-е место
	case 77:
		result = 116 // 31-е место
	case 78:
		result = 114 // 33-е место
	case 79:
		result = 112 // 35-е место
	case 80:
		result = 110 // 37-е место
	case 81:
		result = 108 // 39-е место

	case 180:
		result = 10 // Путь регионов. 1-й раунд
	case 181:
		result = 20 // Путь регионов. 2-й раунд
	case 182:
		result = 40 // Путь регионов. 3-й раунд
	case 183:
		result = 60 // Путь регионов. 4-й раунд
	case 184:
		result = 90 // Путь регионов. 5-й раунд
	case 185:
		result = 120 // Путь регионов. 6-й раунд

	case 210, 211, 212, 213:
		result = 120 // Путь РПЛ. Группы A–D
	case 214:
		result = 150 // Путь РПЛ. 1/4 финала
	case 215:
		result = 180 // Путь РПЛ. 1/2 финала
	case 218:
		result = 220 // Путь РПЛ. Финал

	case 225:
		result = 90 // Путь регионов. 1/4 финала. 1-й этап
	case 226:
		result = 120 // Путь регионов. 1/4 финала. 2-й этап
	case 227:
		result = 150 // Путь регионов. 1/2 финала. 1-й этап
	case 228:
		result = 180 // Путь регионов. 1/2 финала. 2-й этап
	case 229:
		result = 220 // Путь регионов. Финал

	case 390:
		result = 180 // Этап победителей

	case 3699:
		result = 91 // Финал + 1

	default:
		switch {
		case aiStage >= 150 && aiStage <= 175:
			result = 140 // Элитный групповой раунд

		case aiStage >= 320 && aiStage <= 327:
			result = 120 // Группа A–H

		case aiStage >= 361 && aiStage <= 370:
			result = 120 // Группа 1–10

		case aiStage >= 400 && aiStage <= 3695:
			// Зональные соревнования
			switch aiStage % 10 {
			case 1:
				result = 90 // Финал
			case 2:
				result = 60 // 1/2 финала
			case 3:
				result = 40 // 1/4 финала
			case 4:
				result = 20 // 1/8 финала
			case 5:
				result = 10 // 1/16 финала
			case 6:
				result = 5 // 1/32 финала
			case 7:
				result = 4 // 1/64 финала
			case 8:
				result = 3 // 1/128 финала
			case 9:
				result = 2 // 1/256 финала
			}

		case aiStage >= 3696 && aiStage <= 3698:
			result = 1 // 1/512 зональных соревнований
		}
	}

	return result
}

func (s *Storage) GetTeamNameByID(teamID int) (string, error) {
	const op = "storage.GetTeamNameByID"

	query := fmt.Sprintf(`SELECT t.%s,IFNULL(c.%s, '')  FROM %s t
	LEFT JOIN %s c 	ON t.%s = c.%s WHERE t.%s = ?`,
		storage.Fld_common_name,
		storage.Fld_common_name,
		storage.Tbl_class_team,
		storage.Tbl_countries,
		storage.Fld_common_id_country,
		storage.Fld_common_id,
		storage.Fld_common_id,
	)

	var teamName string
	var cityName string

	err := s.db.QueryRow(query, teamID).Scan(
		&teamName, &cityName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: team not found: id=%d", op, teamID)
		}
		return "", fmt.Errorf("%s: query team name: %w", op, err)
	}

	return GetSportTeamName(teamName, cityName), nil
}
