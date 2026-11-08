package players

import (
	"net/url"
	"testing"
)

func mustValues(t *testing.T, query string) url.Values {
	t.Helper()
	values, err := url.ParseQuery(query)
	if err != nil {
		t.Fatalf("parse query %q: %v", query, err)
	}
	return values
}

func sortTestPlayers() []Player {
	return []Player{
		{Name: "Ten", Stats: Stats{MPG: "10.0", Points: "30.0", Rebounds: "5.0"}},
		{Name: "Two", Stats: Stats{MPG: "2.0", Points: "40.0", Rebounds: "4.0"}},
		{Name: "Nine", Stats: Stats{MPG: "9.0", Points: "20.0", Rebounds: "8.0"}},
		{Name: "Missing", Stats: Stats{MPG: missingStat, Points: missingStat, Rebounds: missingStat}},
	}
}

func names(players []Player) []string {
	out := make([]string, len(players))
	for i, player := range players {
		out[i] = player.Name
	}
	return out
}

func equalNames(got []Player, want ...string) bool {
	names := names(got)
	if len(names) != len(want) {
		return false
	}
	for i := range names {
		if names[i] != want[i] {
			return false
		}
	}
	return true
}

func primarySort(key, direction string) sortState {
	return sortState{primary: findSortColumn(key), primaryDir: direction}
}

func TestSortPlayersNumeric(t *testing.T) {
	players := sortTestPlayers()
	sortPlayers(players, primarySort("MPG", "asc"))
	if !equalNames(players, "Two", "Nine", "Ten", "Missing") {
		t.Errorf("ascending = %v", names(players))
	}

	players = sortTestPlayers()
	sortPlayers(players, primarySort("MPG", "desc"))
	if !equalNames(players, "Ten", "Nine", "Two", "Missing") {
		t.Errorf("descending = %v", names(players))
	}
}

func TestSortPlayersSecondaryBreaksTies(t *testing.T) {
	players := []Player{
		{Name: "LowRebound", Stats: Stats{Points: "20", Rebounds: "4"}},
		{Name: "HighRebound", Stats: Stats{Points: "20", Rebounds: "9"}},
		{Name: "TopScore", Stats: Stats{Points: "30", Rebounds: "2"}},
		{Name: "MidRebound", Stats: Stats{Points: "20", Rebounds: "6"}},
	}
	state := sortState{
		primary:      findSortColumn("PTS"),
		primaryDir:   "desc",
		secondary:    findSortColumn("REB"),
		secondaryDir: "desc",
	}

	sortPlayers(players, state)

	if !equalNames(players, "TopScore", "HighRebound", "MidRebound", "LowRebound") {
		t.Errorf("secondary order = %v", names(players))
	}
}

func TestSortPlayersSecondaryDirectionIsIndependent(t *testing.T) {
	players := []Player{
		{Name: "LowRebound", Stats: Stats{Points: "20", Rebounds: "4"}},
		{Name: "HighRebound", Stats: Stats{Points: "20", Rebounds: "9"}},
		{Name: "MidRebound", Stats: Stats{Points: "20", Rebounds: "6"}},
	}
	state := sortState{
		primary:      findSortColumn("PTS"),
		primaryDir:   "asc",
		secondary:    findSortColumn("REB"),
		secondaryDir: "asc",
	}

	sortPlayers(players, state)

	if !equalNames(players, "LowRebound", "MidRebound", "HighRebound") {
		t.Errorf("secondary ascending = %v", names(players))
	}
}

func TestSortPlayersKeepsMissingLast(t *testing.T) {
	for _, direction := range []string{"asc", "desc"} {
		players := sortTestPlayers()
		sortPlayers(players, primarySort("PTS", direction))
		if got := names(players)[3]; got != "Missing" {
			t.Errorf("%s missing value should sort last, got %v", direction, names(players))
		}
	}
}

func TestSortPlayersIsStable(t *testing.T) {
	players := []Player{
		{Name: "First", Stats: Stats{MPG: "20.0"}},
		{Name: "Second", Stats: Stats{MPG: "20.0"}},
		{Name: "Third", Stats: Stats{MPG: "20.0"}},
	}

	sortPlayers(players, primarySort("MPG", "desc"))

	if !equalNames(players, "First", "Second", "Third") {
		t.Errorf("ties should keep source order, got %v", names(players))
	}
}

func TestDisplayColumnsPutsPrimaryAndSecondaryFirst(t *testing.T) {
	state := sortState{
		primary:    findSortColumn("PTS"),
		secondary:  findSortColumn("REB"),
		primaryDir: "desc",
	}
	columns := displayColumns(state)

	if len(columns) != 11 {
		t.Fatalf("columns = %d, want 11", len(columns))
	}
	if columns[0].Key != "PTS" || columns[1].Key != "REB" || columns[2].Key != "DPT" {
		t.Errorf("columns = %+v, want PTS, REB, then DPT", columns)
	}
	seen := make(map[string]bool)
	for _, column := range columns {
		if seen[column.Key] {
			t.Errorf("duplicate column %q", column.Key)
		}
		seen[column.Key] = true
	}
}

func TestParseSort(t *testing.T) {
	tests := []struct {
		query   string
		key     string
		dir     string
		thenKey string
		thenDir string
	}{
		{query: ""},
		{query: "sort=MPG&dir=asc", key: "MPG", dir: "asc"},
		{query: "sort=mpg", key: "MPG", dir: "desc"},
		{query: "sort=PTS&dir=sideways", key: "PTS", dir: "desc"},
		{query: "sort=NOPE&dir=asc"},
		{query: "sort=PTS&dir=desc&then=REB&then_dir=asc", key: "PTS", dir: "desc", thenKey: "REB", thenDir: "asc"},
		{query: "sort=pts&then=reb", key: "PTS", dir: "desc", thenKey: "REB", thenDir: "desc"},
		{query: "sort=PTS&then=PTS&then_dir=asc", key: "PTS", dir: "desc"},
		{query: "then=REB&then_dir=asc"},
	}
	for _, tt := range tests {
		state := parseSort(mustValues(t, tt.query))
		if got := keyOf(state.primary); got != tt.key {
			t.Errorf("%q primary = %q, want %q", tt.query, got, tt.key)
		}
		if state.primaryDir != tt.dir {
			t.Errorf("%q primary dir = %q, want %q", tt.query, state.primaryDir, tt.dir)
		}
		if got := keyOf(state.secondary); got != tt.thenKey {
			t.Errorf("%q secondary = %q, want %q", tt.query, got, tt.thenKey)
		}
		if state.secondaryDir != tt.thenDir {
			t.Errorf("%q secondary dir = %q, want %q", tt.query, state.secondaryDir, tt.thenDir)
		}
	}
}

func TestSortURLKeepsBothLevels(t *testing.T) {
	state := sortState{
		primary:      findSortColumn("PTS"),
		primaryDir:   "desc",
		secondary:    findSortColumn("REB"),
		secondaryDir: "asc",
	}
	if got := sortURL(state); got != "/players?dir=desc&sort=PTS&then=REB&then_dir=asc" {
		t.Errorf("sortURL = %q", got)
	}
}
