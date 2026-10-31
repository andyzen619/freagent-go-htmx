package yahoo

// Name is a player name.
type Name struct {
	Full       string `xml:"full"`
	First      string `xml:"first"`
	Last       string `xml:"last"`
	ASCIIFirst string `xml:"ascii_first"`
	ASCIILast  string `xml:"ascii_last"`
}

// Headshot is the small player image set Yahoo serves.
type Headshot struct {
	URL  string `xml:"url"`
	Size string `xml:"size"`
}

// League is a league resource. Sub-resources (Settings, Players) are only
// populated when the request asked for them.
type League struct {
	LeagueKey             string `xml:"league_key"`
	LeagueID              int    `xml:"league_id"`
	Name                  string `xml:"name"`
	URL                   string `xml:"url"`
	LogoURL               string `xml:"logo_url"`
	DraftStatus           string `xml:"draft_status"` // predraft | drafting | postdraft
	NumTeams              int    `xml:"num_teams"`
	EditKey               string `xml:"edit_key"`
	WeeklyDeadline        string `xml:"weekly_deadline"`
	LeagueUpdateTimestamp int64  `xml:"league_update_timestamp"`
	ScoringType           string `xml:"scoring_type"` // head | roto | points
	LeagueType            string `xml:"league_type"`  // public | private
	Renew                 string `xml:"renew"`
	Renewed               string `xml:"renewed"`
	IrisGroupChatID       string `xml:"iris_group_chat_id"`
	AllowAddToDLExtraPos  bool   `xml:"allow_add_to_dl_extra_pos"`
	IsProLeague           bool   `xml:"is_pro_league"`
	IsCashLeague          bool   `xml:"is_cash_league"`
	CurrentWeek           int    `xml:"current_week"`
	StartWeek             int    `xml:"start_week"`
	StartDate             string `xml:"start_date"`
	EndWeek               int    `xml:"end_week"`
	EndDate               string `xml:"end_date"`
	IsFinished            bool   `xml:"is_finished"`
	GameCode              string `xml:"game_code"`
	Season                int    `xml:"season"`

	Settings *LeagueSettings `xml:"settings"`
	Players  *Players        `xml:"players"`
}

// LeagueSettings holds the league's scoring, roster and waiver configuration.
type LeagueSettings struct {
	DraftType                  string `xml:"draft_type"`
	IsAuctionDraft             bool   `xml:"is_auction_draft"`
	ScoringType                string `xml:"scoring_type"`
	UsesPlayoff                bool   `xml:"uses_playoff"`
	HasPlayoffConsolationGames bool   `xml:"has_playoff_consolation_games"`
	PlayoffStartWeek           int    `xml:"playoff_start_week"`
	UsesPlayoffReseeding       bool   `xml:"uses_playoff_reseeding"`
	UsesLockEliminatedTeams    bool   `xml:"uses_lock_eliminated_teams"`
	NumPlayoffTeams            int    `xml:"num_playoff_teams"`
	NumPlayoffConsolationTeams int    `xml:"num_playoff_consolation_teams"`
	HasMultiweekChampionship   bool   `xml:"has_multiweek_championship"`
	WaiverType                 string `xml:"waiver_type"`
	WaiverRule                 string `xml:"waiver_rule"`
	UsesFAAB                   bool   `xml:"uses_faab"`
	DraftTime                  int64  `xml:"draft_time"`
	DraftPickTime              int    `xml:"draft_pick_time"`
	PostDraftPlayers           string `xml:"post_draft_players"`
	MaxTeams                   int    `xml:"max_teams"`
	WaiverTime                 int    `xml:"waiver_time"`
	TradeEndDate               string `xml:"trade_end_date"`
	TradeRatifyType            string `xml:"trade_ratify_type"`
	TradeRejectTime            int    `xml:"trade_reject_time"`
	PlayerPool                 string `xml:"player_pool"`
	CantCutList                string `xml:"cant_cut_list"`
	SendbirdChannelURL         string `xml:"sendbird_channel_url"`
	PickemEnabled              bool   `xml:"pickem_enabled"`
	UsesFractionalPoints       bool   `xml:"uses_fractional_points"`
	UsesNegativePoints         bool   `xml:"uses_negative_points"`

	RosterPositions []RosterPosition `xml:"roster_positions>roster_position"`
	StatCategories  []StatCategory   `xml:"stat_categories>stats>stat"`
	StatModifiers   []StatModifier   `xml:"stat_modifiers>stats>stat"`
	Divisions       []Division       `xml:"divisions>division"`
}

// RosterPosition is one slot type in the league's lineup, e.g. QB x1.
type RosterPosition struct {
	Position     string `xml:"position"`
	PositionType string `xml:"position_type"`
	Count        int    `xml:"count"`
}

// StatCategory is a scoring category the league tracks.
type StatCategory struct {
	StatID        int      `xml:"stat_id"`
	Enabled       bool     `xml:"enabled"`
	Name          string   `xml:"name"`
	DisplayName   string   `xml:"display_name"`
	SortOrder     int      `xml:"sort_order"`
	PositionType  string   `xml:"position_type"`
	PositionTypes []string `xml:"stat_position_types>stat_position_type>position_type"`
}

// StatModifier is the points-per-unit weight for a stat category.
type StatModifier struct {
	StatID int     `xml:"stat_id"`
	Value  float64 `xml:"value"`
}

// Division is a league division.
type Division struct {
	DivisionID int    `xml:"division_id"`
	Name       string `xml:"name"`
}

// Players is a players collection (<players count="N">).
type Players struct {
	Count   int      `xml:"count,attr"`
	Players []Player `xml:"player"`
}

// Player is a player resource. League-scoped sub-resources (Ownership,
// PercentOwned, ...) are only populated when the request asked for them.
type Player struct {
	PlayerKey                string   `xml:"player_key"`
	PlayerID                 int      `xml:"player_id"`
	Name                     Name     `xml:"name"`
	URL                      string   `xml:"url"`
	Status                   string   `xml:"status"` // injury code: Q, O, IR, P, ... (empty = healthy)
	StatusFull               string   `xml:"status_full"`
	InjuryNote               string   `xml:"injury_note"`
	EditorialPlayerKey       string   `xml:"editorial_player_key"`
	EditorialTeamKey         string   `xml:"editorial_team_key"`
	EditorialTeamFullName    string   `xml:"editorial_team_full_name"`
	EditorialTeamAbbr        string   `xml:"editorial_team_abbr"`
	EditorialTeamURL         string   `xml:"editorial_team_url"`
	ByeWeeks                 []int    `xml:"bye_weeks>week"`
	UniformNumber            int      `xml:"uniform_number"`
	DisplayPosition          string   `xml:"display_position"`
	Headshot                 Headshot `xml:"headshot"`
	ImageURL                 string   `xml:"image_url"`
	IsUndroppable            bool     `xml:"is_undroppable"`
	PositionType             string   `xml:"position_type"`
	PrimaryPosition          string   `xml:"primary_position"`
	EligiblePositions        []string `xml:"eligible_positions>position"`
	EligiblePositionsToAdd   string   `xml:"eligible_positions_to_add"`
	IsKeeper                 IsKeeper `xml:"is_keeper"`
	HasPlayerNotes           bool     `xml:"has_player_notes"`
	HasRecentPlayerNotes     bool     `xml:"has_recent_player_notes"`
	PlayerNotesLastTimestamp int64    `xml:"player_notes_last_timestamp"`

	SelectedPosition *SelectedPosition `xml:"selected_position"`
	PlayerStats      *PlayerStats      `xml:"player_stats"`
	PlayerPoints     *PlayerPoints     `xml:"player_points"`
	Ownership        *Ownership        `xml:"ownership"`
	PercentOwned     *PercentOwned     `xml:"percent_owned"`
	DraftAnalysis    *DraftAnalysis    `xml:"draft_analysis"`
}

// IsKeeper is the keeper status and cost attached to a player.
type IsKeeper struct {
	Status string `xml:"status"`
	Cost   string `xml:"cost"`
}

// SelectedPosition is a player's lineup slot in a roster or matchup context.
type SelectedPosition struct {
	Position string `xml:"position"`
	IsFlex   bool   `xml:"is_flex"`
}

// PlayerStats is a season, week or date stats block. Values are kept as
// strings because Yahoo emits "-" when a player has no data.
type PlayerStats struct {
	CoverageType string       `xml:"coverage_type"` // season | week | date
	Season       int          `xml:"season"`
	Week         int          `xml:"week"`
	Date         string       `xml:"date"`
	Stats        []PlayerStat `xml:"stats>stat"`
}

// PlayerStat is one stat category value for a player.
type PlayerStat struct {
	StatID int    `xml:"stat_id"`
	Value  string `xml:"value"`
}

// PlayerPoints is the fantasy point total for the covered period.
type PlayerPoints struct {
	CoverageType string  `xml:"coverage_type"`
	Season       int     `xml:"season"`
	Week         int     `xml:"week"`
	Date         string  `xml:"date"`
	Total        float64 `xml:"total"`
}

// Ownership is a player's status within one league.
type Ownership struct {
	OwnershipType string `xml:"ownership_type"` // team | waivers | freeagent
	OwnerTeamKey  string `xml:"owner_team_key"`
	OwnerTeamName string `xml:"owner_team_name"`
}

// PercentOwned is the share of Yahoo leagues that own a player.
type PercentOwned struct {
	CoverageType string `xml:"coverage_type"`
	Week         int    `xml:"week"`
	Value        int    `xml:"value"`
	Delta        int    `xml:"delta"`
}

// DraftAnalysis is Yahoo's cross-league draft data for a player. Yahoo sends
// these as decimal strings, so they are kept as strings.
type DraftAnalysis struct {
	AveragePick    string `xml:"average_pick"`
	AverageRound   string `xml:"average_round"`
	AverageCost    string `xml:"average_cost"`
	PercentDrafted string `xml:"percent_drafted"`
}
