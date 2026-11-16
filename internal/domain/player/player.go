// Package player holds the pure domain types for a fantasy basketball player.
// It depends on nothing outside the standard library, so every layer can share
// it without pulling in HTTP, templates or the Yahoo API.
package player

// MissingStat marks a stat the source could not provide. It renders as an em
// dash and is never parsed as a real value.
const MissingStat = "—"

// Stats is the compact, per-game comparison line shown for each player.
type Stats struct {
	Depth     string
	MPG       string
	FGPercent string
	FTPercent string
	ThreePM   string
	Points    string
	Rebounds  string
	Assists   string
	Steals    string
	Blocks    string
	Turnovers string
}

// Value returns the displayed value for a canonical stat key.
func (s Stats) Value(key string) string {
	switch key {
	case "DPT":
		return s.Depth
	case "MPG":
		return s.MPG
	case "FG%":
		return s.FGPercent
	case "FT%":
		return s.FTPercent
	case "3PM":
		return s.ThreePM
	case "PTS":
		return s.Points
	case "REB":
		return s.Rebounds
	case "AST":
		return s.Assists
	case "STL":
		return s.Steals
	case "BLK":
		return s.Blocks
	case "TO":
		return s.Turnovers
	default:
		return ""
	}
}

// Player is one free agent with the stats the players page compares.
type Player struct {
	Name     string
	Position string
	Team     string
	Stats    Stats
}
