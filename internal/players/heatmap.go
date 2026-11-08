package players

import (
	"fmt"
	"math"
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
	Player
	Cells []StatCell
}

// playerViews applies the per-column heat map after sorting and determines the
// visible column order.
func playerViews(players []Player, columns []StatColumn) []PlayerView {
	heat := make([][]string, len(players))
	for i := range heat {
		heat[i] = make([]string, len(columns))
	}
	for columnIndex, column := range columns {
		for playerIndex, class := range heatClasses(players, column) {
			heat[playerIndex][columnIndex] = class
		}
	}

	views := make([]PlayerView, len(players))
	for playerIndex, player := range players {
		cells := make([]StatCell, len(columns))
		for columnIndex, column := range columns {
			value := player.Stats.Value(column.Key)
			class := heat[playerIndex][columnIndex]
			if value == "" {
				value = missingStat
				class = "heat-missing"
			}
			cells[columnIndex] = StatCell{
				Label: column.Label,
				Title: column.Title,
				Value: value,
				Heat:  class,
			}
		}
		views[playerIndex] = PlayerView{Player: player, Cells: cells}
	}
	return views
}

// heatClasses ranks each valid value within one column. Higher is stronger
// except for explicitly inverted categories such as depth and turnovers.
func heatClasses(players []Player, column StatColumn) []string {
	classes := make([]string, len(players))
	values := make([]float64, len(players))
	valid := make([]bool, len(players))
	validCount := 0
	var min, max float64

	for i, player := range players {
		value, ok := statValue(player.Stats.Value(column.Key))
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
