package storage

const kMaxP = 10
const kCatalogsCount = 5
const kPicturesCount = 6
const kShortPicturesCount = 4
const kAuctionPicturesCount = 3
const KSubIDsCount = 10

const KSeasonsPlayoff_Delta = 10

const KSeasonsNoRank = -1000
const KSeasonsRankCup = 0
const KSeasonsRankCupTournament = 250

const KOptionsJoin = "JOIN"
const KOptionsResultsOf = "RESULTSOF"
const KOptionsMode = "MODE"
const KOptionsViewModeResults = "RESULTS"

const (
	RtScoreNormal     = 0  //  1:0, 0:3
	RtScoreOT         = 1  //  3:2 ОТ, scored_et=-1 missed_et=-1
	RtScoreB          = 2  //  3:2 Б, scored_et=-1 missed_et=-1
	RtScoreP          = 3  //  3:2 П, scored_et=-1 or not missed_et=-1 or not
	RtScoreEt         = 4  //  3:2 ОТ, scored_et=-1 missed_et=-1
	RtScoreAllExtra   = 5  //
	RtScorePlusMinus  = 10 // +:-
	RtScoreMinusPlus  = 11 // -:+
	RtScoreMinusMinus = 12 // -:-
	RtScoreWL         = 14 // w:l / в:п
	RtScoreLW         = 15 // l:w / п:в
	RtScoreDD         = 16 // d:d / н:н
	RtScoreQuestion   = 17 // ?:?
)

const KSeasonsRank_1 = 1
const KSeasonsRank_2 = 2
const KSeasonsRank_3 = 3
const KSeasonsRank_4 = 4
const KSeasonsRank_5 = 5
const KSeasonsRank_6 = 6
const KSeasonsRank_7 = 7
const KSeasonsRank_8 = 8
const KSeasonsRank_9 = 9
const KSeasonsRankMin = KSeasonsRank_1
const KSeasonsRankMax = KSeasonsRank_9

const KSeasonsRank_International_1 = 51
const KSeasonsRank_International_2 = 52
const KSeasonsRank_International_3 = 53
const KSeasonsRank_International_4 = 54
const KSeasonsRank_International_5 = 55
const KSeasonsRank_International_6 = 56
const KSeasonsRank_International_7 = 57
const KSeasonsRank_International_8 = 58
const KSeasonsRank_International_9 = 59
const KSeasonsRank_InternationalMin = KSeasonsRank_International_1
const KSeasonsRank_InternationalMax = KSeasonsRank_International_9

const KPlayoff_Rank_1 = KSeasonsRank_1 + KSeasonsPlayoff_Delta
const KPlayoff_Rank_2 = KSeasonsRank_2 + KSeasonsPlayoff_Delta
const KPlayoff_Rank_3 = KSeasonsRank_3 + KSeasonsPlayoff_Delta
const KPlayoff_Rank_4 = KSeasonsRank_4 + KSeasonsPlayoff_Delta
const KPlayoff_Rank_5 = KSeasonsRank_5 + KSeasonsPlayoff_Delta
const KPlayoff_Rank_6 = KSeasonsRank_6 + KSeasonsPlayoff_Delta
const KPlayoff_Rank_7 = KSeasonsRank_7 + KSeasonsPlayoff_Delta
const KPlayoff_Rank_8 = KSeasonsRank_8 + KSeasonsPlayoff_Delta
const KPlayoff_Rank_9 = KSeasonsRank_9 + KSeasonsPlayoff_Delta
const KPlayoffMin = KPlayoff_Rank_1
const KPlayoffMax = KPlayoff_Rank_9

const KPlayoffMatches_Rank_1 = KSeasonsRank_1 + KSeasonsPlayoff_Delta*2
const KPlayoffMatches_Rank_2 = KSeasonsRank_2 + KSeasonsPlayoff_Delta*2
const KPlayoffMatches_Rank_3 = KSeasonsRank_3 + KSeasonsPlayoff_Delta*2
const KPlayoffMatches_Rank_4 = KSeasonsRank_4 + KSeasonsPlayoff_Delta*2
const KPlayoffMatches_Rank_5 = KSeasonsRank_5 + KSeasonsPlayoff_Delta*2
const KPlayoffMatches_Rank_6 = KSeasonsRank_6 + KSeasonsPlayoff_Delta*2
const KPlayoffMatches_Rank_7 = KSeasonsRank_7 + KSeasonsPlayoff_Delta*2
const KPlayoffMatches_Rank_8 = KSeasonsRank_8 + KSeasonsPlayoff_Delta*2
const KPlayoffMatches_Rank_9 = KSeasonsRank_9 + KSeasonsPlayoff_Delta*2
const KPlayoffMatchesMin = KPlayoffMatches_Rank_1
const KPlayoffMatchesMax = KPlayoffMatches_Rank_9

const KPlayoff_International_1 = KSeasonsRank_International_1 + KSeasonsPlayoff_Delta
const KPlayoff_International_2 = KSeasonsRank_International_2 + KSeasonsPlayoff_Delta
const KPlayoff_International_3 = KSeasonsRank_International_3 + KSeasonsPlayoff_Delta
const KPlayoff_International_4 = KSeasonsRank_International_4 + KSeasonsPlayoff_Delta
const KPlayoff_International_5 = KSeasonsRank_International_5 + KSeasonsPlayoff_Delta
const KPlayoff_International_6 = KSeasonsRank_International_6 + KSeasonsPlayoff_Delta
const KPlayoff_International_7 = KSeasonsRank_International_7 + KSeasonsPlayoff_Delta
const KPlayoff_International_8 = KSeasonsRank_International_8 + KSeasonsPlayoff_Delta
const KPlayoff_International_9 = KSeasonsRank_International_9 + KSeasonsPlayoff_Delta
const KPlayoff_InternationalMin = KPlayoff_International_1
const KPlayoff_InternationalMax = KPlayoff_International_9

const KSeasonsRank_Friendly = 1000
const KSeasonsRank_PreSeason = 1001
const KRank_Tournament = 1100

const size_str_fld_extralarge = "5000"
const size_str_fld_max = "1000"
const size_str_fld_255 = "255"
const size_str_fld_150 = "150"
const size_str_fld_50 = "50"
const size_str_fld_25 = "25"

const size_str_fld_system = "5"
const size_str_fld_denomination = "50"
const size_str_fld_date = "30"
const size_str_fld_any_date = "50"

const Tbl_backup_infos = "backup_infos"
const Tbl_banknotes = "banknotes"
const Tbl_coins = "coins"
const Tbl_countries = "countries"
const Tbl_icons = "icons"
const Tbl_class_base = "class_base"
const Tbl_class_location = "class_location"
const Tbl_class_cashier = "class_cashier"
const Tbl_class_issue_type = "class_issue_type"
const Tbl_class_condition = "class_condition"
const Tbl_class_calendar = "class_calendar"
const Tbl_class_dependent_territory = "class_dependent_territory"
const Tbl_class_designer = "class_designer"
const Tbl_class_dynasty = "class_dynasty"
const Tbl_class_disposition = "class_disposition"
const Tbl_class_extra_a = "class_extra_a"
const Tbl_class_extra_b = "class_extra_b"
const Tbl_class_shape = "class_shape"
const Tbl_class_material = "class_material"
const Tbl_class_mint = "class_mint"
const Tbl_class_national_bank_head = "class_national_bank_head"
const Tbl_class_currency = "class_currency"
const Tbl_class_fineness = "class_fineness"
const Tbl_class_edge = "class_edge"
const Tbl_class_serie = "class_serie"
const Tbl_class_source = "class_source"
const Tbl_class_title = "class_title"
const Tbl_class_tag = "class_tag"
const Tbl_class_people = "class_people"
const Tbl_map_view_items = "map_view_items"
const Tbl_multiple_records = "multiple_records"
const Tbl_system = "system"
const Tbl_tree = "tree"
const Tbl_tags = "tags"
const Tbl_arms = "arms"
const Tbl_flags = "flags"
const Tbl_auctions = "auctions"
const Tbl_anthems = "anthems"
const Tbl_custom_searches = "custom_searches"
const Tbl_custom_searches_details = "custom_searches_details"
const Tbl_files = "files"
const Tbl_history = "history"
const Tbl_hyperlinks = "hyperlinks"
const Tbl_graphics = "graphics"
const Tbl_links = "links"
const Tbl_maps = "maps"
const Tbl_notes = "notes"
const Tbl_resources_numista_countries = "resources_numista_countries"
const Tbl_resources_numista_catalogs = "resources_numista_catalogs"
const Tbl_rulers = "rulers"
const Tbl_spread_sheets = "spread_sheets"
const Tbl_statistic_maps = "statistic_maps"
const Tbl_statistic_maps_details = "statistic_maps_details"
const Tbl_web_pages = "web_pages"
const Tbl_pdf = "pdf"
const Tbl_class_team = "class_team"
const Tbl_class_season = "class_season"
const Tbl_sport_results = "sport_results"
const Tbl_sport_tables = "sport_tables"

const Fld_common_id = "id"
const Fld_common_sub_id = "sub_id"
const Fld_common_id_country = "id_country"
const Fld_common_denomination = "denomination"
const Fld_common_id_currency = "id_currency"
const Fld_common_date = "date"
const Fld_common_weight = "weight"
const Fld_common_id_material = "id_material"
const Fld_common_thickness = "thickness"
const Fld_common_id_shape = "id_shape"
const Fld_common_id_edge = "id_edge"
const Fld_common_id_condition_1 = "id_condition_1"
const Fld_common_id_condition_2 = "id_condition_2"
const Fld_common_id_disposition = "id_disposition"
const Fld_common_id_serie = "id_serie"
const Fld_common_id_fineness = "id_fineness"
const Fld_common_id_extra_a = "id_extra_a"
const Fld_common_id_extra_b = "id_extra_b"
const Fld_common_id_issue_type = "id_issue_type"
const Fld_common_id_obverse_designer = "id_obverse_designer"
const Fld_common_id_reverse_designer = "id_reverse_designer"
const Fld_common_id_mint = "id_mint"
const Fld_common_id_dependent_territory = "id_dependent_territory"
const Fld_common_id_location = "id_location"
const Fld_common_id_calendar = "id_calendar"
const Fld_common_item_date = "item_date"
const Fld_common_numista_number = "numista_number"
const Fld_common_numista_comment = "numista_comment"
const Fld_common_ruler = "ruler"
const Fld_common_period = "period"
const Fld_common_issue_years = "issue_years"
const Fld_common_issue_reason = "issue_reason"
const Fld_common_sertified = "sertified"
const Fld_common_type = "type"
const Fld_common_mint_mark = "mint_mark"
const Fld_common_mintage = "mintage"
const Fld_common_list_number = "list_number"
const Fld_common_place = "place"
const Fld_common_multiple = "multiple"
const Fld_common_count = "count"
const Fld_common_hyperlink_caption = "hyperlink_caption"
const Fld_common_font_color = "font_color"
const Fld_common_background_color = "background_color"
const Fld_common_present = "present"
const Fld_common_demonetized = "demonetized"
const Fld_common_extra_s_1 = "extra_s_1"
const Fld_common_extra_s_2 = "extra_s_2"
const Fld_common_extra_n = "extra_n"
const Fld_common_marker = "marker"
const Fld_common_standard_weight = "standard_weight"
const Fld_common_standard_thickness = "standard_thickness"
const Fld_common_text_issuer = "text_issuer"
const Fld_common_text_numeric_value = "text_numeric_value"
const Fld_common_text_currency = "text_currency"
const Fld_common_text_tags = "text_tags"
const Fld_common_obverse_text = "obverse_text"
const Fld_common_obverse_description = "obverse_description"
const Fld_common_obverse_lettering_script = "obverse_lettering_script"
const Fld_common_obverse_unabridged_legend = "obverse_unabridged_legend"
const Fld_common_obverse_translation_legend = "obverse_translation_legend"
const Fld_common_revers_text = "revers_text"
const Fld_common_revers_description = "revers_description"
const Fld_common_revers_lettering_script = "revers_lettering_script"
const Fld_common_revers_unabridged_legend = "revers_unabridged_legend"
const Fld_common_revers_translation_legend = "revers_translation_legend"
const Fld_common_edge_text = "edge_text"
const Fld_common_edge_description = "edge_description"
const Fld_common_info = "info"
const Fld_common_label = "label"
const Fld_common_price = "price"
const Fld_common_id_price_currensy = "id_price_currensy"
const Fld_common_purchased_date = "purchased_date"
const Fld_common_purchased_id_source = "purchased_id_source"
const Fld_common_purchased_info = "purchased_info"
const Fld_common_id_seller = "id_seller"
const Fld_common_sell_price = "sell_price"
const Fld_common_id_sell_price_currensy = "id_sell_price_currensy"
const Fld_common_sell_date = "sell_date"
const Fld_common_id_buyer = "id_buyer"
const Fld_common_id_data = "id_data"

const Fld_common_id_base = "id_base"
const Fld_common_catalog = "catalog"
const Fld_common_name = "name"
const Fld_common_description = "description"
const Fld_common_picture = "pictrure"
const Fld_common_field_level = "field_level"
const Fld_common_field_image_index_level = "image_index_level"
const Fld_common_field_memo = "memo"
const Fld_common_sortmain = "sort_main"
const Fld_common_sortdate = "sort_date"
const Fld_common_last_timestamp = "last_timestamp"
const Fld_common_start = "start"
const Fld_common_end = "end"
const Fld_common_text = "text"
const Fld_common_category = "category"
const Fld_common_pdf = "pdf"

const Fld_common_group_id = "group_id"
const Fld_common_sort_order = "sort_order"

const Fld_common_id_team = "id_team"
const Fld_common_id_team_1 = "id_team_1"
const Fld_common_id_team_2 = "id_team_2"
const Fld_common_id_season = "id_season"
const Fld_common_sport_place = "place"

const Fld_common_favorite = "favorite"
const Fld_common_id_successor = "id_successor"
const Fld_common_founded_date = "founded_date"
const Fld_common_disbanded_date = "disbanded_date"
const Fld_common_other_names = "other_names"
const Fld_common_winner_id = "winner_id"
const Fld_common_private = "private_flag"
const Fld_common_prefix = "prefix"

const Fld_calc_country_full_name = "country_full_name"
const Fld_calc_tags = "tags"
const Fld_live_parent = "parent"
const Fld_live_id = "id"

const db_tbl_old_prefix = "_old"
const db_index_prefix = "idx_"

const Fld_backup_infos_backup_icon_type = "backup_icon_type"
const Fld_backup_infos_backup_file = "backup_file"
const Fld_backup_infos_backup_date = "backup_date"
const Fld_backup_infos_backup_description = "backup_description"

const Fld_coins_size = "size"
const Fld_coins_id_technique = "id_technique"
const Fld_coins_standard_size = "standard_size"
const Fld_coins_capsule = "capsule"
const Fld_coins_edge_lettering_script = "edge_lettering_script"
const Fld_coins_edge_unabridged_legend = "edge_unabridged_legend"
const Fld_coins_edge_translation_legend = "edge_translation_legend"

const Fld_banknotes_height = "height"
const Fld_banknotes_width = "width"
const Fld_banknotes_id_national_bank_head = "id_national_bank_head"
const Fld_banknotes_id_cashier = "id_cashier"
const Fld_banknotes_serial_number = "serial_number"
const Fld_banknotes_watermark = "watermark"
const Fld_banknotes_watermark_description = "watermark_description"
const Fld_banknotes_standard_height = "standard_height"
const Fld_banknotes_standard_width = "standard_width"
const Fld_banknotes_printers = "printers"

const Fld_countries_id_parent = "id_parent"
const Fld_countries_image_index = "image_index"
const Fld_countries_issues = "issues"
const Fld_countries_issues_start_date = "issues_start_date"
const Fld_countries_issues_end_date = "issues_end_date"
const Fld_countries_issues_grade = "issues_rarity"
const Fld_countries_info = "info"
const Fld_countries_original_graphic_name = "original_name"
const Fld_countries_original_text_name = "original_text_name"
const Fld_countries_default_id_base = "default_id_base"
const Fld_countries_path_divider = "path_divider"
const Fld_countries_use_path = "use_path"
const Fld_countries_sort_order = "sort_order"
const Fld_countries_counts_all_all = "counts_all_all"
const Fld_countries_counts_main_all = "counts_main_all"
const Fld_countries_counts_all_selected = "counts_all_selected"
const Fld_countries_counts_main_selected = "counts_main_selected"

const Fld_icons_icon = "icon"

const Fld_class_base_visible = "visible"
const Fld_class_base_default_color = "default_color"
const Fld_class_base_default_pictures_count = "default_pictures_count"
const Fld_class_base_options = "base_options"
const Fld_class_base_available = "available"

const Fld_class_condition_grade = "grade"

const Fld_class_dynasty_flag = "flag"

const Fld_class_currency_use_in_price = "use_in_price"
const Fld_class_currency_declension_11 = "declension_11"
const Fld_class_currency_declension_12 = "declension_12"
const Fld_class_currency_declension_21 = "declension_21"
const Fld_class_currency_declension_22 = "declension_22"
const Fld_class_currency_declension_31 = "declension_31"
const Fld_class_currency_declension_32 = "declension_32"
const Fld_class_currency_declension_41 = "declension_41"
const Fld_class_currency_declension_42 = "declension_42"

const Fld_class_tag_tag_group = "tag_group"
const Fld_class_tag_name_2 = "${fld_common_name}_2"

const Fld_class_people_alias = "alias"
const Fld_class_people_source = "source"
const Fld_class_people_address = "address"
const Fld_class_people_email = "email"
const Fld_class_people_birth_date = "birth_date"
const Fld_class_people_telephone = "telephone"
const Fld_class_people_web = "web"

const Fld_map_view_items_id_map_view = "id_map_view"
const Fld_map_view_items_map_key = "map_key"
const Fld_map_view_items_items = "items"

const Fld_multiple_records_record_type = "record_type"
const Fld_multiple_records_id_record = "id_record"

const Fld_system_param_1 = "param_1"

const Fld_tree_id_root_node = "id_root_node"
const Fld_tree_active = "active"
const Fld_tree_type_1 = "type_1"
const Fld_tree_type_2 = "type_2"
const Fld_tree_type_3 = "type_3"

const Fld_tags_id_item = "id_item"
const Fld_tags_id_tag = "id_tag"
const Fld_tags_id_side = "id_side"
const Fld_tags_mode = "mode"

const Fld_auctions_date = "date"
const Fld_auctions_remark = "remark"

const Fld_anthems_date = "date"
const Fld_anthems_midi = "midi"

const Fld_custom_searches_details_id_custom_search = "id_custom_search"
const Fld_custom_searches_details_id_field = "id_field"
const Fld_custom_searches_details_operation = "operation"
const Fld_custom_searches_details_iparam_1 = "iparam_1"
const Fld_custom_searches_details_iparam_2 = "iparam_2"
const Fld_custom_searches_details_sparam_1 = "sparam_1"
const Fld_custom_searches_details_sparam_2 = "sparam_2"

const Fld_files_file_name = "file_name"
const Fld_files_original_location = "original_location"
const Fld_files_short_description = "short_description"
const Fld_files_file = "file"
const Fld_files_file_time_stamp = "file_time_stamp"
const Fld_files_file_icon = "file_icon"

const Fld_hyperlinks_id_item = "id_item"
const Fld_hyperlinks_link_type = "link_type"
const Fld_hyperlinks_link = "link"
const Fld_hyperlinks_mode = "mode"

const Fld_links_link = "link"

const Fld_maps_parent = "parent"

const Fld_notes_notes = "notes"

const Fld_resources_numista_countries_code = "code"
const Fld_resources_numista_countries_id_wikidata = "id_wikidata"
const Fld_resources_numista_countries_ids_cabinet = "ids_cabinet"
const Fld_resources_numista_countries_parent_code = "parent_code"
const Fld_resources_numista_countries_parent_name = "parent_name"

const Fld_resources_numista_catalogs_id_numista = "id_numista"
const Fld_resources_numista_catalogs_code = "code"
const Fld_resources_numista_catalogs_title = "title"
const Fld_resources_numista_catalogs_author = "author"
const Fld_resources_numista_catalogs_publisher = "publisher"
const Fld_resources_numista_catalogs_isbn13 = "isbn13"

const Fld_rulers_id_dinasty = "id_dinasty"
const Fld_rulers_id_title = "id_title"
const Fld_rulers_flag = "flag"
const Fld_rulers_info = "info"

const Fld_spread_sheets_parent = "parent"
const Fld_spread_sheets_sheet = "sheet"

const Fld_statistic_maps_start_date = "start_date"
const Fld_statistic_maps_end_date = "end_date"
const Fld_statistic_maps_map = "map"

const Fld_statistic_maps_details_id_map = "id_map"
const Fld_statistic_maps_details_color = "color"

const Fld_web_pages_date = "date"
const Fld_web_pages_html = "html"

const Fld_pdf = "pdf"

const Fld_class_season_season = "season"
const Fld_class_season_start_date = "start_date"
const Fld_class_season_end_date = "end_date"
const Fld_class_season_points = "points"
const Fld_class_season_league_rank = "league_rank"
const Fld_class_season_options_1 = "options_1"
const Fld_class_season_options_2 = "options_2"
const Fld_class_season_plain_text = "plain_text"
const Fld_class_season_group = "season_group"

const Fld_sport_games_played = "games_played"
const Fld_sport_wins = "wins"
const Fld_sport_wins_et = "wins_et"
const Fld_sport_draws = "draws"
const Fld_sport_losses_et = "losses_et"
const Fld_sport_losses = "losses"
const Fld_sport_goals_for = "goals_for"
const Fld_sport_goals_against = "goals_against"
const Fld_sport_points = "points"

const Fld_sport_home_games_played = "home_games_played"
const Fld_sport_home_wins = "home_wins"
const Fld_sport_home_wins_et = "home_wins_et"
const Fld_sport_home_draws = "home_draws"
const Fld_sport_home_losses_et = "home_losses_et"
const Fld_sport_home_losses = "home_losses"
const Fld_sport_home_goals_for = "home_goals_for"
const Fld_sport_home_goals_against = "home_goals_against"
const Fld_sport_home_points = "home_points"

const Fld_sport_away_games_played = "away_games_played"
const Fld_sport_away_wins = "away_wins"
const Fld_sport_away_wins_et = "away_wins_et"
const Fld_sport_away_draws = "away_draws"
const Fld_sport_away_losses_et = "away_losses_et"
const Fld_sport_away_losses = "away_losses"
const Fld_sport_away_goals_for = "away_goals_for"
const Fld_sport_away_goals_against = "away_goals_against"
const Fld_sport_away_points = "away_points"
const Fld_sport_result_index = "result_index"
const Fld_sport_points_adjustment = "points_adjustment"

const Fld_sport_scored = "scored"
const Fld_sport_missed = "missed"
const Fld_sport_scored_et = "scored_et"
const Fld_sport_missed_et = "missed_et"
const Fld_sport_result_type = "result_type" // 1= ОТ, 2=П(Б), 3: +/-
const Fld_sport_match_type = "match_type"   // 1= home, 2=unknown
const Fld_sport_stage_index = "stage_index"
const Fld_sport_tour = "tour"

const Fld_tmp_name = "nm"

type TreeID struct {
	ID             int
	ParentID       int
	SubIDs         [KSubIDsCount]int
	Issues         int
	SecondParentID int
}

type TblCountry struct {
	ID          int
	Name        string
	HasChildren bool
	SortOrder   string
}

type TblSeason struct {
	ID         int
	Season     string
	Prefix     string
	Name       string
	SportID    int
	GroupID    int
	LeagueRank int
	SortOrder  int
	Points     string
	Options1   string
	Options2   string
}

type TblSport struct {
	ID         int
	Name       string
	BaseOption string
}

type Stat struct {
	Games         int
	Wins          int
	WinsET        int
	Draws         int
	LossesET      int
	Losses        int
	Goals_For     int
	Goals_Against int
}

type TblTeams struct {
	ID            int
	SportID       int
	Season        string
	SeasonName    string
	TeamName      string
	TeamTerritory string
	TeamID        int
	GroupID       int
	LeagueRank    string
	Place         string
	StageIndex    int
	Result        string
	Stat
	WinnerID int
	Options  string
}

type TblTeamMatches struct {
	TeamName1 string `json:"TeamName1"`
	TeamID1   int    `json:"TeamID1"`
	TeamName2 string `json:"TeamName2"`
	TeamID2   int    `json:"TeamID2"`
	Score     string `json:"Score"`
	Date      string `json:"Date"`
	Color     string `json:"Color,omitempty"`
}
