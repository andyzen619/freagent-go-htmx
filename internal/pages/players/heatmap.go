package players

import (
	"fmt"
	"math"

	"fantasy_waiver_assistant/internal/domain/player"
)

// StatCell is one value in a player's comparison row.
type StatCell struct {
	Label string
	Title string
	Value string
	Heat  string
}

// PlayerView is a player plus the ordered, heat-mapped cells to render.
type PlayerView struct {
	player.Player
	Cells []StatCell
}

// playerViews applies the per-column heat map after sorting and determines the
// visible column order.
func playerViews(players []player.Player, columns []StatColumn) []PlayerView {
	classes := make([][]string, len(columns)) // classes[column][player]
	for c, column := range columns {
		classes[c] = heatClasses(players, column)
	}

	views := make([]PlayerView, len(players))
	for i, p := range players {
		cells := make([]StatCell, len(columns))
		for c, column := range columns {
			cells[c] = statCell(p, column, classes[c][i])
		}
		views[i] = PlayerView{Player: p, Cells: cells}
	}
	return views
}

// statCell renders one player's value for a column, falling back to the missing
// marker and a neutral heat class when the value is absent.
func statCell(p player.Player, column StatColumn, class string) StatCell {
	value := p.Stats.Value(column.Key)
	if value == "" {
		value = player.MissingStat
		class = "heat-missing"
	}
	return StatCell{Label: column.Label, Title: column.Title, Value: value, Heat: class}
}

// heatClasses ranks each valid value within one column. Higher is stronger
// except for explicitly inverted categories such as depth and turnovers.
func heatClasses(players []player.Player, column StatColumn) []string {
	classes := make([]string, len(players))
	values := make([]float64, len(players))
	valid := make([]bool, len(players))
	validCount := 0
	var min, max float64

	for i, p := range players {
		value, ok := statValue(p.Stats.Value(column.Key))
		if !ok {
			classes[i] = "heat-missing"
			continue
		}
		values[i] = value
		valid[i] = true
		if validCount == 0 || value < min {
			min = value
		}
		if validCount == 0 || value > max {
			max = value
		}
		validCount++
	}

	if validCount == 1 || (validCount > 1 && min == max) {
		for i := range classes {
			if valid[i] {
				classes[i] = "heat-3"
			}
		}
		return classes
	}

	for i := range players {
		if !valid[i] {
			continue
		}
		worse := 0
		for j := range players {
			if !valid[j] || i == j {
				continue
			}
			if (column.BetterLow && values[j] > values[i]) ||
				(!column.BetterLow && values[j] < values[i]) {
				worse++
			}
		}
		score := float64(worse) / float64(validCount-1)
		level := 1 + int(math.Round(score*4))
		classes[i] = fmt.Sprintf("heat-%d", level)
	}
	return classes
}
