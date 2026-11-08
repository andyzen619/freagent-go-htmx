package players

import (
	"testing"

	"fantasy_waiver_assistant/internal/yahoo"
)

func yahooPlayer(stats map[int]string) yahoo.Player {
	p := yahoo.Player{
		Name:              yahoo.Name{Full: "Test Player"},
		DisplayPosition:   "SG",
		EditorialTeamAbbr: "BOS",
		PlayerStats:       &yahoo.PlayerStats{CoverageType: "season"},
	}
	for id, value := range stats {
		p.PlayerStats.Stats = append(p.PlayerStats.Stats, yahoo.PlayerStat{StatID: id, Value: value})
	}
	return p
}

func TestPlayerFromYahooConvertsSeasonTotalsToPerGame(t *testing.T) {
	p := yahooPlayer(map[int]string{
		statGamesPlayed: "10",
		statMinutes:     "300",
		statFGPercent:   "0.500",
		statFTPercent:   "0.800",
		statThreePM:     "20",
		statPoints:      "250",
		statRebounds:    "80",
		statAssists:     "70",
		statSteals:      "10",
		statBlocks:      "5",
		statTurnovers:   "30",
	})

	got := playerFromYahoo(p)

	if got.Name != "Test Player" || got.Team != "BOS" || got.Position != "SG" {
		t.Errorf("identity = %+v", got)
	}
	want := Stats{
		Depth: missingStat, MPG: "30.0", FGPercent: "50.0", FTPercent: "80.0",
		ThreePM: "2.0", Points: "25.0", Rebounds: "8.0", Assists: "7.0",
		Steals: "1.0", Blocks: "0.5", Turnovers: "3.0",
	}
	if got.Stats != want {
		t.Errorf("stats = %+v\nwant   %+v", got.Stats, want)
	}
}

func TestPlayerFromYahooMissingStats(t *testing.T) {
	got := playerFromYahoo(yahoo.Player{Name: yahoo.Name{Full: "No Stats"}})

	want := Stats{
		Depth: missingStat, MPG: missingStat, FGPercent: missingStat,
		FTPercent: missingStat, ThreePM: missingStat, Points: missingStat,
		Rebounds: missingStat, Assists: missingStat, Steals: missingStat,
		Blocks: missingStat, Turnovers: missingStat,
	}
	if got.Stats != want {
		t.Errorf("stats = %+v\nwant   %+v", got.Stats, want)
	}
}

func TestPlayerFromYahooNeedsGamesPlayedForCountingStats(t *testing.T) {
	p := yahooPlayer(map[int]string{
		statFGPercent: "0.450",
		statPoints:    "200",
	})

	got := playerFromYahoo(p)

	if got.Stats.Points != missingStat || got.Stats.MPG != missingStat {
		t.Errorf("counting stats should be missing without games played: %+v", got.Stats)
	}
	if got.Stats.FGPercent != "45.0" {
		t.Errorf("FGPercent = %q, want 45.0", got.Stats.FGPercent)
	}
}

func TestPlayerFromYahooHandlesMissingStatValues(t *testing.T) {
	p := yahooPlayer(map[int]string{
		statGamesPlayed: "10",
		statPoints:      "-",
		statFGPercent:   "",
	})

	got := playerFromYahoo(p)

	if got.Stats.Points != missingStat || got.Stats.FGPercent != missingStat {
		t.Errorf("invalid values should be missing: %+v", got.Stats)
	}
}
