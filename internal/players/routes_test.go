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

var sample = []Player{{Name: "Bijan Robinson", Position: "RB", Team: "ATL"}}

func TestPage(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/players", false)
	if !strings.Contains(body, "<html") || !strings.Contains(body, "Bijan Robinson") {
		t.Errorf("page should render the layout with players, got %q", body)
	}
}

func TestPageAsHtmxReturnsRows(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/players", true)
	if strings.Contains(body, "<html") || !strings.Contains(body, "Bijan Robinson") {
		t.Errorf("htmx response should be rows only, got %q", body)
	}
}

func TestRowsFragment(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/rows?pos=RB", true)
	if strings.Contains(body, "<html") || !strings.Contains(body, "Bijan Robinson") {
		t.Errorf("fragment should render rows only, got %q", body)
	}
}

func TestRowsDirectVisitStaysFragment(t *testing.T) {
	body := get(t, stubSource{players: sample}, "/hx/players/rows", false)
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
