package players

import "context"

// mockPlayers is deterministic sample data for UI development without Yahoo
// API access. Names are fictional; depth is position-specific.
var mockPlayers = []Player{
	{
		Name: "Malik Carter", Position: "PG", Team: "ATL",
		Stats: Stats{Depth: "2", MPG: "31.4", FGPercent: "48.7", FTPercent: "81.2", ThreePM: "2.1", Points: "24.3", Rebounds: "4.1", Assists: "7.2", Steals: "1.4", Blocks: "0.2", Turnovers: "2.8"},
	},
	{
		Name: "Darius Reed", Position: "PG", Team: "MEM",
		Stats: Stats{Depth: "1", MPG: "29.8", FGPercent: "46.1", FTPercent: "84.0", ThreePM: "1.7", Points: "18.6", Rebounds: "3.5", Assists: "6.1", Steals: "1.1", Blocks: "0.1", Turnovers: "2.1"},
	},
	{
		Name: "Jalen Price", Position: "PG", Team: "CLE",
		Stats: Stats{Depth: "3", MPG: "22.6", FGPercent: "44.8", FTPercent: "79.5", ThreePM: "1.2", Points: "11.4", Rebounds: "2.8", Assists: "4.0", Steals: "0.8", Blocks: "0.2", Turnovers: "1.7"},
	},
	{
		Name: "Jordan Ellis", Position: "SG", Team: "BOS",
		Stats: Stats{Depth: "1", MPG: "33.1", FGPercent: "47.5", FTPercent: "88.1", ThreePM: "2.9", Points: "25.8", Rebounds: "4.4", Assists: "4.8", Steals: "1.3", Blocks: "0.3", Turnovers: "2.4"},
	},
	{
		Name: "Trey Lawson", Position: "SG", Team: "DAL",
		Stats: Stats{Depth: "2", MPG: "27.9", FGPercent: "45.2", FTPercent: "82.7", ThreePM: "2.4", Points: "17.2", Rebounds: "3.7", Assists: "3.6", Steals: "1.0", Blocks: "0.2", Turnovers: "1.9"},
	},
	{
		Name: "Owen Foster", Position: "SG", Team: "SAC",
		Stats: Stats{Depth: "3", MPG: "20.8", FGPercent: "43.9", FTPercent: "76.4", ThreePM: "1.8", Points: "10.9", Rebounds: "2.5", Assists: "2.7", Steals: "0.7", Blocks: "0.1", Turnovers: "1.4"},
	},
	{
		Name: "Cameron Hughes", Position: "SF", Team: "LAL",
		Stats: Stats{Depth: "1", MPG: "34.8", FGPercent: "49.4", FTPercent: "77.8", ThreePM: "2.8", Points: "21.7", Rebounds: "6.3", Assists: "3.9", Steals: "1.0", Blocks: "0.4", Turnovers: "2.5"},
	},
	{
		Name: "Devin Brooks", Position: "SF", Team: "MIA",
		Stats: Stats{Depth: "2", MPG: "28.4", FGPercent: "46.7", FTPercent: "80.1", ThreePM: "1.9", Points: "14.8", Rebounds: "5.1", Assists: "2.9", Steals: "0.9", Blocks: "0.3", Turnovers: "1.8"},
	},
	{
		Name: "Marcus Grant", Position: "PF", Team: "DEN",
		Stats: Stats{Depth: "1", MPG: "32.2", FGPercent: "52.1", FTPercent: "74.6", ThreePM: "1.1", Points: "19.5", Rebounds: "9.2", Assists: "3.2", Steals: "0.8", Blocks: "0.9", Turnovers: "2.2"},
	},
	{
		Name: "Elijah Stone", Position: "PF", Team: "PHX",
		Stats: Stats{Depth: "2", MPG: "25.7", FGPercent: "47.8", FTPercent: "71.3", ThreePM: "0.8", Points: "12.1", Rebounds: "7.6", Assists: "2.0", Steals: "0.6", Blocks: "0.7", Turnovers: "1.6"},
	},
	{
		Name: "Noah Bennett", Position: "C", Team: "NYK",
		Stats: Stats{Depth: "1", MPG: "30.6", FGPercent: "55.3", FTPercent: "69.8", ThreePM: "0.3", Points: "16.4", Rebounds: "11.8", Assists: "2.4", Steals: "0.7", Blocks: "1.9", Turnovers: "2.0"},
	},
	{
		Name: "Andre Coleman", Position: "C", Team: "MIN",
		Stats: Stats{Depth: "2", MPG: "23.9", FGPercent: "51.0", FTPercent: "66.5", ThreePM: "0.1", Points: "9.7", Rebounds: "8.9", Assists: "1.5", Steals: "0.4", Blocks: "1.3", Turnovers: "1.4"},
	},
}

// MockSource returns built-in free agents without calling Yahoo.
type MockSource struct{}

// NewMockSource returns a deterministic Source for local UI development.
func NewMockSource() *MockSource {
	return &MockSource{}
}

// FreeAgents returns a copy of the sample players.
func (s *MockSource) FreeAgents(context.Context) ([]Player, error) {
	out := make([]Player, len(mockPlayers))
	copy(out, mockPlayers)
	return out, nil
}
