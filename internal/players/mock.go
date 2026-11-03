package players

import "context"

// mockPlayers is deterministic sample data for UI development without Yahoo
// API access. Names are fictional, but positions and teams match the layout.
var mockPlayers = []Player{
	{Name: "Malik Carter", Position: "PG", Team: "ATL"},
	{Name: "Darius Reed", Position: "PG", Team: "MEM"},
	{Name: "Jalen Price", Position: "PG", Team: "CLE"},
	{Name: "Jordan Ellis", Position: "SG", Team: "BOS"},
	{Name: "Trey Lawson", Position: "SG", Team: "DAL"},
	{Name: "Owen Foster", Position: "SG", Team: "SAC"},
	{Name: "Cameron Hughes", Position: "SF", Team: "LAL"},
	{Name: "Devin Brooks", Position: "SF", Team: "MIA"},
	{Name: "Marcus Grant", Position: "PF", Team: "DEN"},
	{Name: "Elijah Stone", Position: "PF", Team: "PHX"},
	{Name: "Noah Bennett", Position: "C", Team: "NYK"},
	{Name: "Andre Coleman", Position: "C", Team: "MIN"},
}

// MockSource returns built-in free agents without calling Yahoo.
type MockSource struct{}

// NewMockSource returns a deterministic Source for local UI development.
func NewMockSource() *MockSource {
	return &MockSource{}
}

// FreeAgents returns the sample players, optionally narrowed to one position.
func (s *MockSource) FreeAgents(_ context.Context, pos string) ([]Player, error) {
	if pos == "" {
		out := make([]Player, len(mockPlayers))
		copy(out, mockPlayers)
		return out, nil
	}

	out := make([]Player, 0, len(mockPlayers))
	for _, player := range mockPlayers {
		if player.Position == pos {
			out = append(out, player)
		}
	}
	return out, nil
}
