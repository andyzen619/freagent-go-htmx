package yahoo

import "testing"

func TestKeyBuilders(t *testing.T) {
	if got := LeagueKey("470", 1000); got != "470.l.1000" {
		t.Errorf("LeagueKey = %q", got)
	}
	if got := LeagueKey("nfl", 1000); got != "nfl.l.1000" {
		t.Errorf("LeagueKey = %q", got)
	}
	if got := PlayerKey("470", 30121); got != "470.p.30121" {
		t.Errorf("PlayerKey = %q", got)
	}
}

func TestPlayerFilterEncoding(t *testing.T) {
	if got := (PlayerFilter{}).encode(); got != "" {
		t.Errorf("empty filter = %q, want empty", got)
	}

	got := PlayerFilter{
		Position:   "W/R/T",
		Status:     StatusWaivers,
		Search:     "smith jones",
		Sort:       "PTS",
		SortType:   "week",
		SortSeason: 2026,
		SortWeek:   8,
		SortDate:   "2026-10-31",
		Start:      10,
		Count:      25,
		Out:        []string{"ownership", "percent_owned"},
	}.encode()

	want := ";position=W/R/T;status=W;search=smith%20jones;sort=PTS;sort_type=week;sort_season=2026;sort_week=8;sort_date=2026-10-31;start=10;count=25;out=ownership,percent_owned"
	if got != want {
		t.Errorf("filter = %q\nwant     %q", got, want)
	}
}

func TestLeaguePath(t *testing.T) {
	if got := leaguePath("470.l.1000"); got != "/league/470.l.1000" {
		t.Errorf("path = %q", got)
	}
	if got := leaguePath("470.l.1000", "settings", "teams"); got != "/league/470.l.1000;out=settings,teams" {
		t.Errorf("path with out = %q", got)
	}
}

func TestPlayerPath(t *testing.T) {
	if got := playerPath("470.p.30121"); got != "/player/470.p.30121" {
		t.Errorf("path = %q", got)
	}
	if got := playerPath("470.p.30121", "stats"); got != "/player/470.p.30121;out=stats" {
		t.Errorf("path with out = %q", got)
	}
}

func TestEscapeSegment(t *testing.T) {
	cases := map[string]string{
		"470.l.1000": "470.l.1000",
		"W/R/T":      "W/R/T",
		"a b":        "a%20b",
		"a?b#c%d":    "a%3Fb%23c%25d",
	}
	for in, want := range cases {
		if got := escapeSegment(in); got != want {
			t.Errorf("escapeSegment(%q) = %q, want %q", in, got, want)
		}
	}
}
