// Package nba is a read-only client for the unofficial stats.nba.com JSON
// endpoints — the backend behind NBA.com's own statistics pages. It is not a
// published product: requests need browser-like headers and the host can
// rate-limit or block without notice, so callers should fetch sparingly and
// cache.
package nba

import (
	"net/url"
	"strconv"
)

// PerMode selects how stats are aggregated. PerGame returns per-game averages;
// Totals returns season totals; the remaining values normalize per 36/48/etc.
const (
	PerModeTotals            = "Totals"
	PerModePerGame           = "PerGame"
	PerModePer36             = "Per36"
	PerModePer48             = "Per48"
	PerModePer40             = "Per40"
	PerModePerMinute         = "PerMinute"
	PerModePerPossession     = "PerPossession"
	PerModePer100Possessions = "Per100Possessions"
)

// MeasureType selects the stat family. Base is the traditional box score; the
// others expose advanced, scoring, usage and defensive metrics.
const (
	MeasureTypeBase        = "Base"
	MeasureTypeAdvanced    = "Advanced"
	MeasureTypeMisc        = "Misc"
	MeasureTypeFourFactors = "Four Factors"
	MeasureTypeScoring     = "Scoring"
	MeasureTypeUsage       = "Usage"
	MeasureTypeDefense     = "Defense"
)

// SeasonType selects the part of the season.
const (
	SeasonTypeRegularSeason = "Regular Season"
	SeasonTypePlayoffs      = "Playoffs"
	SeasonTypePreSeason     = "Pre Season"
	SeasonTypeAllStar       = "All Star"
)

// SeasonSegment restricts a season to the stretch before or after the All-Star
// break. The zero value means the whole season.
const (
	SeasonSegmentPreAllStar  = "Pre All-Star"
	SeasonSegmentPostAllStar = "Post All-Star"
)

// DefaultSeason is used when Params.Season is empty. Override it (or set the
// field) for the current season; it is not auto-detected.
const DefaultSeason = "2026-27"

// Params are the query parameters accepted by leaguedashplayerstats. Zero
// values mean "all" or a sensible default, so a caller can set only what it
// cares about.
type Params struct {
	// Season is the NBA season as YYYY-YY, e.g. "2026-27". Empty uses
	// DefaultSeason.
	Season string
	// SeasonType is one of the SeasonType* constants. Empty is Regular Season.
	SeasonType string
	// SeasonSegment is "" (whole season) or one of SeasonSegment*.
	SeasonSegment string
	// PerMode is one of the PerMode* constants. Empty is PerGame.
	PerMode string
	// MeasureType is one of the MeasureType* constants. Empty is Base.
	MeasureType string
	// DateFrom and DateTo bound a custom range as MM/DD/YYYY. Empty means the
	// whole season. They only cover the selected Season.
	DateFrom string
	DateTo   string
	// Month filters to a calendar month 1-12; 0 means all months.
	Month int
	// LastNGames limits to the most recent N games; 0 means all games.
	LastNGames int
}

// Values renders the parameters as the query string the endpoint expects. It
// sends the full canonical parameter set, filling defaults for empty fields,
// so the request works without callers remembering every knob.
func (p Params) Values() url.Values {
	v := url.Values{}
	v.Set("College", "")
	v.Set("Conference", "")
	v.Set("Country", "")
	v.Set("DateFrom", p.DateFrom)
	v.Set("DateTo", p.DateTo)
	v.Set("Division", "")
	v.Set("DraftPick", "")
	v.Set("DraftYear", "")
	v.Set("GameScope", "")
	v.Set("GameSegment", "")
	v.Set("Height", "")
	v.Set("LastNGames", strconv.Itoa(p.LastNGames))
	v.Set("LeagueID", "00")
	v.Set("Location", "")
	v.Set("MeasureType", orDefault(p.MeasureType, MeasureTypeBase))
	v.Set("Month", strconv.Itoa(p.Month))
	v.Set("OpponentTeamID", "0")
	v.Set("Outcome", "")
	v.Set("PORound", "0")
	v.Set("PaceAdjust", "N")
	v.Set("PerMode", orDefault(p.PerMode, PerModePerGame))
	v.Set("Period", "0")
	v.Set("PlayerExperience", "")
	v.Set("PlayerPosition", "")
	v.Set("PlusMinus", "N")
	v.Set("Rank", "N")
	v.Set("Season", orDefault(p.Season, DefaultSeason))
	v.Set("SeasonSegment", p.SeasonSegment)
	v.Set("SeasonType", orDefault(p.SeasonType, SeasonTypeRegularSeason))
	v.Set("ShotClockRange", "")
	v.Set("StarterBench", "")
	v.Set("TeamID", "0")
	v.Set("TwoWay", "0")
	v.Set("VsConference", "")
	v.Set("VsDivision", "")
	v.Set("Weight", "")
	return v
}

// Response is the root JSON document every stats.nba.com endpoint returns.
type Response struct {
	Resource   string      `json:"resource"`
	ResultSets []ResultSet `json:"resultSets"`
}

// ResultSet is one named table of results. Headers names the columns and
// RowSet holds the data rows; each row is a heterogeneous list of values in
// the same order as Headers.
type ResultSet struct {
	Name    string   `json:"name"`
	Headers []string `json:"headers"`
	RowSet  [][]any  `json:"rowSet"`
}

// IndexOf returns the column index for a header name, or -1 when absent.
func (r ResultSet) IndexOf(name string) int {
	for i, h := range r.Headers {
		if h == name {
			return i
		}
	}
	return -1
}

// orDefault returns value, or fallback when value is empty.
func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
