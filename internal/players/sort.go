package players

import (
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type sortColumn struct {
	key       string
	name      string
	title     string
	betterLow bool
}

var sortColumns = []sortColumn{
	{key: "DPT", name: "depth", title: "Depth", betterLow: true},
	{key: "MPG", name: "minutes per game", title: "Minutes per game"},
	{key: "FG%", name: "field goal percentage", title: "Field goal percentage"},
	{key: "FT%", name: "free throw percentage", title: "Free throw percentage"},
	{key: "3PM", name: "three-pointers made", title: "Three-pointers made"},
	{key: "PTS", name: "points", title: "Points"},
	{key: "REB", name: "rebounds", title: "Rebounds"},
	{key: "AST", name: "assists", title: "Assists"},
	{key: "STL", name: "steals", title: "Steals"},
	{key: "BLK", name: "blocks", title: "Blocks"},
	{key: "TO", name: "turnovers", title: "Turnovers", betterLow: true},
}

// StatColumn is one displayed stats-grid column.
type StatColumn struct {
	Key       string
	Label     string
	Title     string
	BetterLow bool
}

type sortState struct {
	primary      *sortColumn
	primaryDir   string
	secondary    *sortColumn
	secondaryDir string
}

// parseSort reads and validates primary and secondary sort parameters. A
// secondary category needs a primary category and cannot duplicate it.
func parseSort(values url.Values) sortState {
	state := sortState{
		primary: findSortColumn(values.Get("sort")),
	}
	if state.primary == nil {
		return state
	}
	state.primaryDir = parseDirection(values.Get("dir"))

	state.secondary = findSortColumn(values.Get("then"))
	if state.secondary != nil && state.secondary.key == state.primary.key {
		state.secondary = nil
	}
	if state.secondary != nil {
		state.secondaryDir = parseDirection(values.Get("then_dir"))
	}
	return state
}

func findSortColumn(key string) *sortColumn {
	key = strings.TrimSpace(key)
	for i := range sortColumns {
		if strings.EqualFold(key, sortColumns[i].key) {
			return &sortColumns[i]
		}
	}
	return nil
}

func parseDirection(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}

// displayColumns puts the primary and secondary sort categories first, then
// keeps the standard comparison order for the remaining categories.
func displayColumns(state sortState) []StatColumn {
	columns := make([]StatColumn, 0, len(sortColumns))
	seen := make(map[string]bool, len(sortColumns))
	add := func(column *sortColumn) {
		if column == nil || seen[column.key] {
			return
		}
		seen[column.key] = true
		columns = append(columns, StatColumn{
			Key:       column.key,
			Label:     column.key,
			Title:     column.title,
			BetterLow: column.betterLow,
		})
	}

	add(state.primary)
	add(state.secondary)
	for i := range sortColumns {
		add(&sortColumns[i])
	}
	return columns
}

// sortPlayers orders by the primary category, then the optional secondary
// category. Missing values sort last at every level, and remaining ties keep
// their source order.
func sortPlayers(players []Player, state sortState) {
	if state.primary == nil {
		return
	}

	sort.SliceStable(players, func(i, j int) bool {
		if order := compareStat(
			players[i].Stats.Value(state.primary.key),
			players[j].Stats.Value(state.primary.key),
			state.primaryDir,
		); order != 0 {
			return order < 0
		}
		if state.secondary == nil {
			return false
		}
		return compareStat(
			players[i].Stats.Value(state.secondary.key),
			players[j].Stats.Value(state.secondary.key),
			state.secondaryDir,
		) < 0
	})
}

func compareStat(left, right, direction string) int {
	leftValue, leftOK := statValue(left)
	rightValue, rightOK := statValue(right)
	switch {
	case !leftOK && !rightOK:
		return 0
	case !leftOK:
		return 1
	case !rightOK:
		return -1
	case leftValue == rightValue:
		return 0
	case direction == "asc":
		if leftValue < rightValue {
			return -1
		}
		return 1
	default:
		if leftValue > rightValue {
			return -1
		}
		return 1
	}
}

func sortURL(state sortState) string {
	values := url.Values{}
	if state.primary != nil {
		values.Set("sort", state.primary.key)
		values.Set("dir", state.primaryDir)
	}
	if state.secondary != nil {
		values.Set("then", state.secondary.key)
		values.Set("then_dir", state.secondaryDir)
	}
	if len(values) == 0 {
		return "/players"
	}
	return "/players?" + values.Encode()
}

func statValue(raw string) (float64, bool) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, false
	}
	return value, true
}
