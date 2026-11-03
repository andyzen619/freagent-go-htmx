package players

import (
	"context"
	"testing"
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

	for _, player := range found {
		stats := player.Stats
		if player.Name == "" || player.Position == "" || player.Team == "" {
			t.Errorf("identity incomplete: %+v", player)
		}
		if stats.Depth == "" || stats.Depth == missingStat {
			t.Errorf("%s has no mock depth ranking", player.Name)
		}
		for label, value := range map[string]string{
			"MPG": stats.MPG, "FG%": stats.FGPercent, "FT%": stats.FTPercent,
			"3PM": stats.ThreePM, "PTS": stats.Points, "REB": stats.Rebounds,
			"AST": stats.Assists, "STL": stats.Steals, "BLK": stats.Blocks, "TO": stats.Turnovers,
		} {
			if value == "" || value == missingStat {
				t.Errorf("%s has no %s value", player.Name, label)
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
