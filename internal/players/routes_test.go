package players

import (
	"context"
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

func (s stubSource) FreeAgents(context.Context, string) ([]Player, error) {
	return s.players, s.err
}

func get(t *testing.T, source Source, path string, htmx bool) string {
	t.Helper()

	mux := http.NewServeMux()
	New(source).Register(mux)

	req := httptest.NewRequest(http.MethodGet, path, nil)
	if htmx {
		req.Header.Set("HX-Request", "true")
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec.Body.String()
}

var sample = []Player{{Name: "Jayson Tatum", Position: "SF", Team: "BOS"}}

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
	body := get(t, stubSource{players: sample}, "/hx/players/panel?pos=SF", true)
	if strings.Contains(body, "<html") || !strings.Contains(body, "Jayson Tatum") {
		t.Errorf("fragment should render the panel only, got %q", body)
	}
}

func TestPanelActiveChip(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel?pos=SF", true)
	if !strings.Contains(body, `href="/players?pos=SF"`) || !strings.Contains(body, `class="on"`) {
		t.Errorf("active position chip should be marked, got %q", body)
	}
}

func TestPanelDirectVisitStaysFragment(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/panel", false)
	if strings.Contains(body, "<html") {
		t.Errorf("component route should never render the layout, got %q", body)
	}
}

func TestMissingLeagueKey(t *testing.T) {
	body := get(t, stubSource{err: ErrNoLeague}, "/players", false)
	if !strings.Contains(body, "YAHOO_LEAGUE_KEY") {
		t.Errorf("error should mention the missing league key, got %q", body)
	}
}
