package yahoo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type capturedRequest struct {
	auth string
	uri  string
}

// newTestServer serves fixture for every request and records the last request.
func newTestServer(t *testing.T, fixture string) (*httptest.Server, *capturedRequest) {
	t.Helper()
	got := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.auth = r.Header.Get("Authorization")
		got.uri = r.URL.RequestURI()
		data, err := os.ReadFile("testdata/" + fixture)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)
	return srv, got
}

func testClient(baseURL string, token TokenSource) *Client {
	c := NewClient(token)
	c.BaseURL = baseURL
	return c
}

func TestLeagueRequest(t *testing.T) {
	srv, got := newTestServer(t, "league.xml")
	c := testClient(srv.URL, StaticToken("tok"))

	lg, err := c.League(context.Background(), "390.l.1000", "settings")
	if err != nil {
		t.Fatalf("League: %v", err)
	}
	if got.auth != "Bearer tok" {
		t.Errorf("Authorization = %q", got.auth)
	}
	if got.uri != "/league/390.l.1000;out=settings" {
		t.Errorf("URI = %q", got.uri)
	}
	if lg.Name != "Yahoo Public 1000" {
		t.Errorf("Name = %q", lg.Name)
	}
}

func TestLeagueRequestWithoutToken(t *testing.T) {
	srv, got := newTestServer(t, "league.xml")
	c := testClient(srv.URL, nil)

	if _, err := c.League(context.Background(), "390.l.1000"); err != nil {
		t.Fatalf("League: %v", err)
	}
	if got.auth != "" {
		t.Errorf("Authorization = %q, want empty", got.auth)
	}
	if got.uri != "/league/390.l.1000" {
		t.Errorf("URI = %q", got.uri)
	}
}

func TestLeaguePlayersRequest(t *testing.T) {
	srv, got := newTestServer(t, "league_players.xml")
	c := testClient(srv.URL, StaticToken("tok"))

	ps, err := c.LeaguePlayers(context.Background(), LeagueKey("470", 1000), PlayerFilter{
		Position: "RB",
		Status:   StatusFreeAgents,
		Sort:     "PTS",
		SortType: "season",
		Start:    25,
		Count:    25,
		Out:      []string{"ownership", "percent_owned"},
	})
	if err != nil {
		t.Fatalf("LeaguePlayers: %v", err)
	}

	want := "/league/470.l.1000/players;position=RB;status=FA;sort=PTS;sort_type=season;start=25;count=25;out=ownership,percent_owned"
	if got.uri != want {
		t.Errorf("URI = %q, want %q", got.uri, want)
	}
	if ps.Count != 2 || len(ps.Players) != 2 {
		t.Errorf("players = %d/%d", ps.Count, len(ps.Players))
	}
}

func TestPlayersRequest(t *testing.T) {
	srv, got := newTestServer(t, "players_top.xml")
	c := testClient(srv.URL, StaticToken("tok"))

	ps, err := c.Players(context.Background(), "470.p.30121", "470.p.40059")
	if err != nil {
		t.Fatalf("Players: %v", err)
	}
	if got.uri != "/players;player_keys=470.p.30121,470.p.40059" {
		t.Errorf("URI = %q", got.uri)
	}
	if len(ps.Players) != 2 {
		t.Errorf("players = %d", len(ps.Players))
	}
}

func TestPlayersRequiresKeys(t *testing.T) {
	c := testClient("http://unused", nil)
	if _, err := c.Players(context.Background()); err == nil {
		t.Fatal("Players with no keys should fail")
	}
}

func TestPlayerRequest(t *testing.T) {
	srv, got := newTestServer(t, "player.xml")
	c := testClient(srv.URL, StaticToken("tok"))

	p, err := c.Player(context.Background(), "470.p.30121", "ownership", "percent_owned")
	if err != nil {
		t.Fatalf("Player: %v", err)
	}
	if got.uri != "/player/470.p.30121;out=ownership,percent_owned" {
		t.Errorf("URI = %q", got.uri)
	}
	if p.Name.Full != "Christian McCaffrey" {
		t.Errorf("Name = %q", p.Name.Full)
	}
}

func TestLeagueSettingsRequest(t *testing.T) {
	srv, got := newTestServer(t, "league_settings.xml")
	c := testClient(srv.URL, StaticToken("tok"))

	s, err := c.LeagueSettings(context.Background(), "390.l.1000")
	if err != nil {
		t.Fatalf("LeagueSettings: %v", err)
	}
	if got.uri != "/league/390.l.1000/settings" {
		t.Errorf("URI = %q", got.uri)
	}
	if s.WaiverRule != "gametime" {
		t.Errorf("WaiverRule = %q", s.WaiverRule)
	}
}

func TestUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	c := testClient(srv.URL, StaticToken("expired"))
	if _, err := c.League(context.Background(), "390.l.1000"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestErrorDocumentWithStatusOK(t *testing.T) {
	data, err := os.ReadFile("testdata/error.xml")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)

	c := testClient(srv.URL, StaticToken("tok"))
	_, err = c.League(context.Background(), "470.l.1000")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if !strings.Contains(apiErr.Error(), "must be logged in") {
		t.Errorf("error = %q", apiErr.Error())
	}
}

func TestErrorDocumentWithHTTPError(t *testing.T) {
	data, err := os.ReadFile("testdata/error.xml")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)

	c := testClient(srv.URL, StaticToken("tok"))
	_, err = c.League(context.Background(), "470.l.1000")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
}

func TestPlainHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	c := testClient(srv.URL, StaticToken("tok"))
	_, err := c.League(context.Background(), "470.l.1000")
	if err == nil || !strings.Contains(err.Error(), "500") || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v", err)
	}
}
