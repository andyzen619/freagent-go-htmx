package players

import (
	"context"
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubSource is a Source with a fixed answer.
type stubSource struct {
	players []Player
	err     error
}

func (s stubSource) FreeAgents(context.Context) ([]Player, error) {
	return s.players, s.err
}

func get(t *testing.T, source Source, path string, htmx bool) string {
	t.Helper()
	return getWith(t, New(source), path, htmx)
}

func getMock(t *testing.T, path string, htmx bool) string {
	t.Helper()
	return getWith(t, NewMock(NewMockSource()), path, htmx)
}

func getWith(t *testing.T, handler *Handler, path string, htmx bool) string {
	t.Helper()

	mux := http.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(http.MethodGet, path, nil)
	if htmx {
		req.Header.Set("HX-Request", "true")
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec.Body.String()
}

var sample = []Player{{
	Name: "Jayson Tatum", Position: "SF", Team: "BOS",
	Stats: Stats{
		Depth: "1", MPG: "36.1", FGPercent: "47.1", FTPercent: "84.0",
		ThreePM: "3.1", Points: "26.8", Rebounds: "8.2", Assists: "4.5",
		Steals: "1.0", Blocks: "0.6", Turnovers: "2.4",
	},
}}

func TestPage(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/players", false)
	if !strings.Contains(body, "<html") || !strings.Contains(body, "Jayson Tatum") {
		t.Errorf("page should render the layout with players, got %q", body)
	}
}

// The layout marks the current section from pageData.Nav.
func TestPageMarksPlayersTab(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/players", false)
	if !strings.Contains(body, `aria-current="page"`) || !strings.Contains(body, `href="/players"`) {
		t.Errorf("players tab should be marked active, got %q", body)
	}
}

func TestPageAsHtmxReturnsPanel(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/players", true)
	if strings.Contains(body, "<html") || !strings.Contains(body, "Jayson Tatum") {
		t.Errorf("htmx response should be the panel only, got %q", body)
	}
}

func TestPanelFragment(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel", true)
	if strings.Contains(body, "<html") || !strings.Contains(body, "Jayson Tatum") {
		t.Errorf("fragment should render the panel only, got %q", body)
	}
}

func TestPanelDirectVisitStaysFragment(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel", false)
	if strings.Contains(body, "<html") {
		t.Errorf("component route should never render the layout, got %q", body)
	}
}

func TestPanelShowsIdentityAboveStats(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel", true)

	identity := strings.Index(body, "Jayson Tatum")
	stats := strings.Index(body, `<dl class="player-stats"`)
	if identity < 0 || stats < 0 || identity > stats {
		t.Errorf("identity should render before the stats section, got %q", body)
	}
	if !strings.Contains(body, `<p class="player-meta">BOS · SF</p>`) {
		t.Errorf("identity should include team and position, got %q", body)
	}
}

func TestPanelShowsAllComparisonColumns(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel", true)

	for _, label := range []string{"DPT", "MPG", "FG%", "FT%", "3PM", "PTS", "REB", "AST", "STL", "BLK", "TO"} {
		if !strings.Contains(body, ">"+label+"<") {
			t.Errorf("missing %s column", label)
		}
	}
	if !strings.Contains(body, `title="Depth"`) {
		t.Error("DPT should retain its full accessible label")
	}
	if strings.Count(body, `class="player-row"`) != 1 {
		t.Errorf("one player should render one row, got %q", body)
	}
}

func TestPanelShowsMissingStatsAsDash(t *testing.T) {
	body := get(t, stubSource{players: []Player{{Name: "No Stats"}}}, "/hx/players/panel", true)

	if strings.Count(body, "—") != 11 {
		t.Errorf("missing values should render 11 dashes, got %q", body)
	}
	if strings.Count(body, "heat-missing") < 11 {
		t.Errorf("missing values should have a neutral heat class, got %q", body)
	}
}

func TestPanelShowsHeatLegend(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel", true)

	for _, want := range []string{"heat-legend", "heat-1", "heat-5", "Gray = no stats"} {
		if !strings.Contains(body, want) {
			t.Errorf("heat legend missing %q", want)
		}
	}
}

func TestPanelHasNoPositionFilterOrHorizontalScroll(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel", true)

	for _, unwanted := range []string{"?pos=", `class="chips"`, "overflow-x"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("players panel should not contain %q, got %q", unwanted, body)
		}
	}
}

func TestSortControls(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel?sort=MPG&dir=desc&then=REB&then_dir=asc", true)
	unescaped := html.UnescapeString(body)

	for _, want := range []string{
		`method="get" action="/players"`,
		`hx-get="/players" hx-trigger="change"`,
		`hx-push-url="true"`,
		`name="dir" value="desc"`,
		`name="then_dir" value="asc"`,
		`<option value="MPG" selected >MPG</option>`,
		`<option value="REB" selected >REB</option>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("sort controls missing %q", want)
		}
	}
	if !strings.Contains(unescaped, `href="/players?dir=asc&sort=MPG&then=REB&then_dir=asc"`) {
		t.Errorf("primary direction toggle URL wrong: %q", body)
	}
	if !strings.Contains(unescaped, `href="/players?dir=desc&sort=MPG&then=REB&then_dir=desc"`) {
		t.Errorf("secondary direction toggle URL wrong: %q", body)
	}
	if !strings.Contains(body, `aria-label="Primary sort by minutes per game descending; activate to sort ascending"`) {
		t.Errorf("primary toggle label missing: %q", body)
	}
	if !strings.Contains(body, `aria-label="Secondary sort by rebounds ascending; activate to sort descending"`) {
		t.Errorf("secondary toggle label missing: %q", body)
	}
}

func TestSortControlsNoPrimary(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel", true)

	if !strings.Contains(body, `<select class="sort-select" id="sort-then" name="then" disabled>`) {
		t.Errorf("secondary selector should be disabled without a primary sort: %q", body)
	}
	if !strings.Contains(body, `<option value="" selected >None</option>`) {
		t.Errorf("primary None option should be selected: %q", body)
	}
	if !strings.Contains(body, `class="sort-dir disabled"`) {
		t.Errorf("direction toggles should be disabled without a sort: %q", body)
	}
	if strings.Contains(body, `hx-push-url="/players?`) {
		t.Errorf("no direction link should be rendered without a sort: %q", body)
	}
}

func TestSecondarySelectEnabledWithPrimaryOnly(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel?sort=PTS&dir=desc", true)

	if strings.Contains(body, `id="sort-then" name="then" disabled`) {
		t.Errorf("secondary selector should be enabled once a primary is chosen: %q", body)
	}
	if !strings.Contains(body, `<option value="" selected >None</option>`) {
		t.Errorf("secondary None option should be selected: %q", body)
	}
}

func TestSortAscendingAndDescending(t *testing.T) {
	players := []Player{
		{Name: "Ten", Stats: Stats{MPG: "10.0"}},
		{Name: "Two", Stats: Stats{MPG: "2.0"}},
		{Name: "Nine", Stats: Stats{MPG: "9.0"}},
		{Name: "Missing", Stats: Stats{MPG: missingStat}},
	}
	source := stubSource{players: players}

	ascending := get(t, source, "/hx/players/panel?sort=MPG&dir=asc", true)
	if !inOrder(ascending, "Two", "Nine", "Ten", "Missing") {
		t.Errorf("ascending order wrong: %q", ascending)
	}

	descending := get(t, source, "/hx/players/panel?sort=MPG&dir=desc", true)
	if !inOrder(descending, "Ten", "Nine", "Two", "Missing") {
		t.Errorf("descending order wrong: %q", descending)
	}
}

func TestSecondarySortBreaksTies(t *testing.T) {
	players := []Player{
		{Name: "LowRebound", Stats: Stats{Points: "20", Rebounds: "4"}},
		{Name: "HighRebound", Stats: Stats{Points: "20", Rebounds: "9"}},
		{Name: "TopScore", Stats: Stats{Points: "30", Rebounds: "2"}},
		{Name: "MidRebound", Stats: Stats{Points: "20", Rebounds: "6"}},
	}
	body := get(t, stubSource{players: players}, "/hx/players/panel?sort=PTS&dir=desc&then=REB&then_dir=desc", true)

	if !inOrder(body, "TopScore", "HighRebound", "MidRebound", "LowRebound") {
		t.Errorf("secondary tie-break order wrong: %q", body)
	}
}

func TestSecondarySortColumnMovesSecond(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel?sort=PTS&dir=desc&then=REB&then_dir=asc", true)

	start := strings.Index(body, `<dl class="player-stats"`)
	if start < 0 {
		t.Fatalf("no stats section in %q", body)
	}
	stats := body[start:]
	points := strings.Index(stats, `<abbr title="Points">PTS</abbr>`)
	rebounds := strings.Index(stats, `<abbr title="Rebounds">REB</abbr>`)
	depth := strings.Index(stats, `<abbr title="Depth">DPT</abbr>`)
	if points < 0 || rebounds < 0 || depth < 0 || !(points < rebounds && rebounds < depth) {
		t.Errorf("columns should be PTS, REB, then DPT, got %q", stats)
	}
}

func TestActiveSortColumnMovesFirst(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel?sort=PTS&dir=desc", true)

	start := strings.Index(body, `<dl class="player-stats"`)
	if start < 0 {
		t.Fatalf("no stats section in %q", body)
	}
	stats := body[start:]
	points := strings.Index(stats, `<abbr title="Points">PTS</abbr>`)
	depth := strings.Index(stats, `<abbr title="Depth">DPT</abbr>`)
	if points < 0 || depth < 0 || points > depth {
		t.Errorf("active PTS column should render before DPT, got %q", stats)
	}
	firstValue := stats[strings.Index(stats, "<dd>"):]
	firstValue = strings.TrimPrefix(firstValue, "<dd>")
	firstValue = strings.SplitN(firstValue, "</dd>", 2)[0]
	if firstValue != "26.8" {
		t.Errorf("first value = %q, want the PTS value 26.8", firstValue)
	}
}

func TestNoSortKeepsStandardColumnOrder(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel", true)

	start := strings.Index(body, `<dl class="player-stats"`)
	stats := body[start:]
	depth := strings.Index(stats, `<abbr title="Depth">DPT</abbr>`)
	mpg := strings.Index(stats, `<abbr title="Minutes per game">MPG</abbr>`)
	if depth < 0 || mpg < 0 || depth > mpg {
		t.Errorf("unsorted columns should start with DPT then MPG, got %q", stats)
	}
}

func TestInvalidSortKeepsSourceOrder(t *testing.T) {
	players := []Player{
		{Name: "Second Input", Stats: Stats{MPG: "2.0"}},
		{Name: "First Input", Stats: Stats{MPG: "10.0"}},
	}
	body := get(t, stubSource{players: players}, "/hx/players/panel?sort=BOGUS&dir=asc", true)

	if !inOrder(body, "Second Input", "First Input") {
		t.Errorf("invalid sort should keep the source order, got %q", body)
	}
	if strings.Contains(body, `aria-current="true"`) {
		t.Errorf("invalid sort should not activate a chip, got %q", body)
	}
}

func inOrder(body string, names ...string) bool {
	previous := -1
	for _, name := range names {
		index := strings.Index(body, name)
		if index < 0 || index <= previous {
			return false
		}
		previous = index
	}
	return true
}

func TestMissingLeagueKey(t *testing.T) {
	body := get(t, stubSource{err: ErrNoLeague}, "/players", false)
	if !strings.Contains(body, "YAHOO_LEAGUE_KEY") {
		t.Errorf("error should mention the missing league key, got %q", body)
	}
}

func TestNoPlayersMessage(t *testing.T) {
	body := get(t, stubSource{}, "/players", false)
	if !strings.Contains(body, "No free agents found.") {
		t.Errorf("empty result should show a message, got %q", body)
	}
}

func TestMockPageShowsLabelAndPlayers(t *testing.T) {
	body := getMock(t, "/players", false)
	for _, want := range []string{"Mock data — not from Yahoo", "Malik Carter", "31.4", "DPT"} {
		if !strings.Contains(body, want) {
			t.Errorf("mock page missing %q", want)
		}
	}
}

func TestMockFragmentShowsAllPlayers(t *testing.T) {
	body := getMock(t, "/hx/players/panel", true)
	if !strings.Contains(body, "Noah Bennett") || !strings.Contains(body, "Malik Carter") {
		t.Errorf("mock fragment should include players from all positions, got %q", body)
	}
}

func TestLivePageHasNoMockLabel(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/players", false)
	if strings.Contains(body, "Mock data") {
		t.Errorf("live page should not be labelled as mock data, got %q", body)
	}
}
