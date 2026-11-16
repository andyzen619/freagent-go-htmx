package freeagents

import (
	"testing"

	"fantasy_waiver_assistant/internal/domain/player"
	"fantasy_waiver_assistant/internal/infra/yahoo"
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
	want := player.Stats{
		Depth: player.MissingStat, MPG: "30.0", FGPercent: "50.0", FTPercent: "80.0",
		ThreePM: "2.0", Points: "25.0", Rebounds: "8.0", Assists: "7.0",
		Steals: "1.0", Blocks: "0.5", Turnovers: "3.0",
	}
	if got.Stats != want {
		t.Errorf("stats = %+v\nwant   %+v", got.Stats, want)
	}
}

func TestPlayerFromYahooMissingStats(t *testing.T) {
	got := playerFromYahoo(yahoo.Player{Name: yahoo.Name{Full: "No Stats"}})

	want := player.Stats{
		Depth: player.MissingStat, MPG: player.MissingStat, FGPercent: player.MissingStat,
		FTPercent: player.MissingStat, ThreePM: player.MissingStat, Points: player.MissingStat,
		Rebounds: player.MissingStat, Assists: player.MissingStat, Steals: player.MissingStat,
		Blocks: player.MissingStat, Turnovers: player.MissingStat,
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

	if got.Stats.Points != player.MissingStat || got.Stats.MPG != player.MissingStat {
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

	if got.Stats.Points != player.MissingStat || got.Stats.FGPercent != player.MissingStat {
		t.Errorf("invalid values should be missing: %+v", got.Stats)
	}
}
