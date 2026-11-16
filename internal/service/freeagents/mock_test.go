package freeagents

import (
	"context"
	"testing"

	"fantasy_waiver_assistant/internal/domain/player"
)

func TestMockSourceReturnsAllPlayers(t *testing.T) {
	found, err := NewMockSource().FreeAgents(context.Background())
	if err != nil {
		t.Fatalf("FreeAgents: %v", err)
	}
	if len(found) != len(mockPlayers) {
		t.Errorf("players = %d, want %d", len(found), len(mockPlayers))
	}
}

func TestMockSourceIncludesComparisonStats(t *testing.T) {
	found, _ := NewMockSource().FreeAgents(context.Background())

	for _, p := range found {
		stats := p.Stats
		if p.Name == "" || p.Position == "" || p.Team == "" {
			t.Errorf("identity incomplete: %+v", p)
		}
		if stats.Depth == "" || stats.Depth == player.MissingStat {
			t.Errorf("%s has no mock depth ranking", p.Name)
		}
		for label, value := range map[string]string{
			"MPG": stats.MPG, "FG%": stats.FGPercent, "FT%": stats.FTPercent,
			"3PM": stats.ThreePM, "PTS": stats.Points, "REB": stats.Rebounds,
			"AST": stats.Assists, "STL": stats.Steals, "BLK": stats.Blocks, "TO": stats.Turnovers,
		} {
			if value == "" || value == player.MissingStat {
				t.Errorf("%s has no %s value", p.Name, label)
			}
		}
	}
}

func TestMockSourceReturnsACopy(t *testing.T) {
	source := NewMockSource()
	first, _ := source.FreeAgents(context.Background())
	first[0].Name = "Changed"

	second, _ := source.FreeAgents(context.Background())
	if second[0].Name == "Changed" {
		t.Error("callers should not be able to mutate the sample data")
	}
}
