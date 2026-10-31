package yahoo

import (
	"fmt"
	"strconv"
	"strings"
)

// LeagueKey builds a league key from a game key ("470", "nfl") and league id.
func LeagueKey(gameKey string, leagueID int) string {
	return fmt.Sprintf("%s.l.%d", gameKey, leagueID)
}

// PlayerKey builds a player key from a game key ("470", "nfl") and player id.
func PlayerKey(gameKey string, playerID int) string {
	return fmt.Sprintf("%s.p.%d", gameKey, playerID)
}

// PlayerStatus filters a league's players by availability.
type PlayerStatus string

const (
	// StatusAll returns all players, available or taken (A).
	StatusAll PlayerStatus = "A"
	// StatusFreeAgents returns players that can be added immediately (FA).
	StatusFreeAgents PlayerStatus = "FA"
	// StatusWaivers returns players on waivers (W).
	StatusWaivers PlayerStatus = "W"
	// StatusTaken returns players rostered by a team (T).
	StatusTaken PlayerStatus = "T"
	// StatusKeepers returns keeper players (K).
	StatusKeepers PlayerStatus = "K"
)

// PlayerFilter narrows a league players collection. Zero values are omitted and
// filters only apply in a league context.
type PlayerFilter struct {
	// Position is a position code such as "QB" or "W/R/T".
	Position string
	// Status selects available, waived or taken players.
	Status PlayerStatus
	// Search matches a player name.
	Search string
	// Sort is a stat id, NAME, OR (overall rank), AR (actual rank) or PTS.
	Sort string
	// SortType is season, week, date, lastweek or lastmonth.
	SortType string
	// SortSeason is the season year used with SortType "season".
	SortSeason int
	// SortWeek is the week used with SortType "week" (football).
	SortWeek int
	// SortDate is the YYYY-MM-DD date used with SortType "date" (non-football).
	SortDate string
	// Start is the pagination offset.
	Start int
	// Count is the page size. Yahoo caps pages at 25 players per request.
	Count int
	// Out lists sub-resources to include, e.g. "ownership", "percent_owned".
	Out []string
}

// encode renders the filter as Yahoo's semicolon-delimited path parameters.
func (f PlayerFilter) encode() string {
	var b strings.Builder
	add := func(key, value string) {
		if value == "" {
			return
		}
		b.WriteByte(';')
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(escapeSegment(value))
	}

	add("position", f.Position)
	add("status", string(f.Status))
	add("search", f.Search)
	add("sort", f.Sort)
	add("sort_type", f.SortType)
	if f.SortSeason > 0 {
		add("sort_season", strconv.Itoa(f.SortSeason))
	}
	if f.SortWeek > 0 {
		add("sort_week", strconv.Itoa(f.SortWeek))
	}
	add("sort_date", f.SortDate)
	if f.Start > 0 {
		add("start", strconv.Itoa(f.Start))
	}
	if f.Count > 0 {
		add("count", strconv.Itoa(f.Count))
	}
	add("out", strings.Join(f.Out, ","))

	return b.String()
}

func leaguePath(leagueKey string, out ...string) string {
	return "/league/" + escapeSegment(leagueKey) + outParam(out)
}

func leagueSettingsPath(leagueKey string) string {
	return "/league/" + escapeSegment(leagueKey) + "/settings"
}

func leaguePlayersPath(leagueKey string, f PlayerFilter) string {
	return "/league/" + escapeSegment(leagueKey) + "/players" + f.encode()
}

func playersPath(playerKeys []string) string {
	keys := make([]string, len(playerKeys))
	for i, key := range playerKeys {
		keys[i] = escapeSegment(key)
	}
	return "/players;player_keys=" + strings.Join(keys, ",")
}

func playerPath(playerKey string, out ...string) string {
	return "/player/" + escapeSegment(playerKey) + outParam(out)
}

func outParam(out []string) string {
	if len(out) == 0 {
		return ""
	}
	return ";out=" + strings.Join(out, ",")
}

// escapeSegment protects a value used inside a path segment. Yahoo position
// values contain '/', so full path escaping would corrupt them; only the
// characters that would end or split the URL are encoded.
func escapeSegment(s string) string {
	return strings.NewReplacer("%", "%25", " ", "%20", "?", "%3F", "#", "%23").Replace(s)
}
