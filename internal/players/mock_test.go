package players

import (
	"context"
	"testing"
)

func TestMockSourceReturnsAllPlayers(t *testing.T) {
	found, err := NewMockSource().FreeAgents(context.Background(), "")
	if err != nil {
		t.Fatalf("FreeAgents: %v", err)
	}
	if len(found) != len(mockPlayers) {
		t.Errorf("players = %d, want %d", len(found), len(mockPlayers))
	}
}

func TestMockSourceFiltersByPosition(t *testing.T) {
	found, err := NewMockSource().FreeAgents(context.Background(), "C")
	if err != nil {
		t.Fatalf("FreeAgents: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("centers = %d, want 2", len(found))
	}
	for _, player := range found {
		if player.Position != "C" {
			t.Errorf("%s has position %q, want C", player.Name, player.Position)
		}
	}
}

func TestMockSourceUnknownPosition(t *testing.T) {
	found, err := NewMockSource().FreeAgents(context.Background(), "XX")
	if err != nil {
		t.Fatalf("FreeAgents: %v", err)
	}
	if len(found) != 0 {
		t.Errorf("players = %d, want 0", len(found))
	}
}

func TestMockSourceReturnsACopy(t *testing.T) {
	source := NewMockSource()
	first, _ := source.FreeAgents(context.Background(), "")
	first[0].Name = "Changed"

	second, _ := source.FreeAgents(context.Background(), "")
	if second[0].Name == "Changed" {
		t.Error("callers should not be able to mutate the sample data")
	}
}
