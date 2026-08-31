package handlers

const Url_login = "/api/v1/auth/login"
const Url_Me = "/api/v1/auth/me"

const Url_Territories = "/api/v1/territories"
const Url_Territories_ID = "id" // "/api/v1/territories/{id}"
// const Url_Territories_Filter = "id"  // "/api/v1/territories/search?filter={text}"
const Url_Territories_Path = "/path" // "/api/v1/territories/path?id={542}"

const Url_Seasons = "/api/v1/seasons" // "/api/v1/seasons?sport_ids=1,2&season_filter=sf&name_filter=nf"
const Url_Seasons_IDs_Sport = "sport_ids"
const Url_Seasons_Season_Filter = "season_filter"
const Url_Seasons_Name_Filter = "name_filter"
const Url_Seasons_Names = "names"

const Url_Sports = "/api/v1/sports" // "/api/v1/sports"

const Url_Teams = "/api/v1/teams"             // "/api/v1/teams?sport_ids=1,2&territory_id=542"
const Url_Teams_ID_Territory = "territory_id" // "/api/v1/teams?territory_id=542"

const Url_Team_Matches = "/api/v1/team_matches" // "/api/v1/team_matches?team_id=77&season_id=1251"
const Url_Team_Matches_ID_Team = "team_id"
const Url_Team_Matches_ID_Season = "season_id"

const Url_Tournament = "/api/v1/tournament" // "/api/v1/tournament?id=2626"
const Url_Tournament_ID = "id"

const Url_Team = "/api/v1/team" // "/api/v1/team?id=2626"
const Url_Team_ID = "id"

const Url_OpponentOptions = "/api/v1/opponent_options" // "/api/v1/opponent_options?sport_ids=1,2"
const Url_OpponentOptions_SportIDs = "sport_ids"
const Url_Comparison = "/api/v1/opponent_comparison" // "api/v1/opponent_comparison?opponent1Type=territory&opponent1Id=3&opponent2Type=territory&opponent2Id=21&competitionFilter=all&sport_ids=1,2&league_ranks=1,2,3"
const Url_Comparison_opponent1Type = "opponent1Type"
const Url_Comparison_opponent1Id = "opponent1Id"
const Url_Comparison_opponent2Type = "opponent2Type"
const Url_Comparison_opponent2Id = "opponent2Id"
const Url_Comparison_competitionFilter = "competitionFilter"
const Url_Comparison_IDs_Sport = "sport_ids"
const (
	OpponentTypeTerritory = "territory"
	OpponentTypeTeam      = "team"
)
const Url_Comparison_LeagueRanks = "league_ranks"
const Url_ComparisonMatches = "/api/v1/opponent_comparison/matches"
const Url_ComparisonMatches_Team1ID = "team1_id"
const Url_ComparisonMatches_Team2ID = "team2_id"

const Url_SummaryCategories = "/api/v1/summary_tables/categories" // "api/v1/summary_tables/categories"
const Url_SummaryTables = "/api/v1/summary_tables"                //"api/v1/summary_tables?category=&league_ranks=1&year_from=1935&year_to=1965&sport_ids=1"
const Url_SummaryTables_Category = "category"
const Url_SummaryTables_LeagueRanks = "league_ranks"
const Url_SummaryTables_YearFrom = "year_from"
const Url_SummaryTables_YearTo = "year_to"
const Url_SummaryTables_SportIDs = "sport_ids"

const Url_SeasonInfo = "/api/v1/season_info" //  "/api/v1/season_info?id=2626"
const Url_TeamInfo = "/api/v1/team_info"     //  "/api/v1/team_info?id=26"
const Url_Info_ID = "id"

const Url_Health = "/health"
