package handlers

const Url_login = "/api/v1/auth/login"

const Url_Territories = "/api/v1/territories"
const Url_Territories_ID = "id" // "/api/v1/territories/{id}"
// const Url_Territories_Filter = "id"  // "/api/v1/territories/search?filter={text}"
const Url_Territories_Path = "/path" // "/api/v1/territories/path?id={542}"

const Url_Seasons = "/api/v1/seasons" // "/api/v1/seasons?sport_ids=1,2&season_filter=sf&name_filter=nf"
const Url_Seasons_IDs_Sport = "sport_ids"
const Url_Seasons_Season_Filter = "season_filter"
const Url_Seasons_Name_Filter = "name_filter"

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
