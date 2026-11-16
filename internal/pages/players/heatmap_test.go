package players

import (
	"testing"

	"fantasy_waiver_assistant/internal/domain/player"
)

func columnFor(key string) StatColumn {
	for _, column := range displayColumns(sortState{}) {
		if column.Key == key {
			return column
		}
	}
	panic("unknown column " + key)
}

func TestHeatClassesHigherIsStronger(t *testing.T) {
	players := []player.Player{
		{Stats: player.Stats{Points: "10"}},
		{Stats: player.Stats{Points: "20"}},
		{Stats: player.Stats{Points: "30"}},
		{Stats: player.Stats{Points: "40"}},
		{Stats: player.Stats{Points: "50"}},
	}

	got := heatClasses(players, columnFor("PTS"))
	want := []string{"heat-1", "heat-2", "heat-3", "heat-4", "heat-5"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("class[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestHeatClassesInvertsDepthAndTurnovers(t *testing.T) {
	players := []player.Player{
		{Stats: player.Stats{Depth: "1", Turnovers: "1"}},
		{Stats: player.Stats{Depth: "2", Turnovers: "2"}},
		{Stats: player.Stats{Depth: "3", Turnovers: "3"}},
		{Stats: player.Stats{Depth: "4", Turnovers: "4"}},
		{Stats: player.Stats{Depth: "5", Turnovers: "5"}},
	}

	depth := heatClasses(players, columnFor("DPT"))
	turnovers := heatClasses(players, columnFor("TO"))
	want := []string{"heat-5", "heat-4", "heat-3", "heat-2", "heat-1"}
	for i := range want {
		if depth[i] != want[i] {
			t.Errorf("depth class[%d] = %q, want %q", i, depth[i], want[i])
		}
		if turnovers[i] != want[i] {
			t.Errorf("turnover class[%d] = %q, want %q", i, turnovers[i], want[i])
		}
	}
}

func TestHeatClassesMissingAndEqualValues(t *testing.T) {
	missing := []player.Player{
		{Stats: player.Stats{Points: "10"}},
		{Stats: player.Stats{Points: player.MissingStat}},
		{Stats: player.Stats{Points: "30"}},
	}
	got := heatClasses(missing, columnFor("PTS"))
	if got[0] != "heat-1" || got[1] != "heat-missing" || got[2] != "heat-5" {
		t.Errorf("missing classes = %v", got)
	}

	equal := []player.Player{
		{Stats: player.Stats{Points: "20"}},
		{Stats: player.Stats{Points: "20"}},
		{Stats: player.Stats{Points: player.MissingStat}},
	}
	got = heatClasses(equal, columnFor("PTS"))
	if got[0] != "heat-3" || got[1] != "heat-3" || got[2] != "heat-missing" {
		t.Errorf("equal classes = %v", got)
	}

	single := []player.Player{
		{Stats: player.Stats{Points: "20"}},
		{Stats: player.Stats{Points: player.MissingStat}},
	}
	got = heatClasses(single, columnFor("PTS"))
	if got[0] != "heat-3" || got[1] != "heat-missing" {
		t.Errorf("single-value classes = %v", got)
	}
}

func TestPlayerViewsFollowActiveColumnOrder(t *testing.T) {
	players := []player.Player{
		{Name: "First", Stats: player.Stats{Depth: "1", Points: "10"}},
		{Name: "Second", Stats: player.Stats{Depth: "2", Points: "20"}},
	}
	active := sortState{primary: &sortColumns[5]} // PTS
	views := playerViews(players, displayColumns(active))

	if got := views[0].Cells[0]; got.Label != "PTS" || got.Value != "10" || got.Heat != "heat-1" {
		t.Errorf("first cell = %+v", got)
	}
	if got := views[0].Cells[1]; got.Label != "DPT" || got.Value != "1" || got.Heat != "heat-5" {
		t.Errorf("depth cell = %+v", got)
	}
	if views[0].Name != "First" {
		t.Errorf("player identity lost: %+v", views[0])
	}
}
