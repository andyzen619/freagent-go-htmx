package players

// SortChoice is one option in a primary or secondary sort menu.
type SortChoice struct {
	Value    string
	Label    string
	Selected bool
	Disabled bool
}

// SortControl is the template state for one sort level.
type SortControl struct {
	Key         string
	Direction   string
	Arrow       string
	ToggleURL   string
	ToggleLabel string
	Options     []SortChoice
}

func sortControls(state sortState) (primary, secondary SortControl) {
	primary = SortControl{
		Direction: state.primaryDir,
		Options:   sortChoices(state, false),
	}
	if primary.Direction == "" {
		primary.Direction = "desc"
	}
	primary.Arrow = directionArrow(primary.Direction)
	if state.primary == nil {
		primary.ToggleLabel = "No primary sort selected"
	} else {
		next := flipDirection(state.primaryDir)
		toggled := state
		toggled.primaryDir = next
		primary.Key = state.primary.key
		primary.ToggleURL = sortURL(toggled)
		primary.ToggleLabel = "Primary sort by " + state.primary.name + " " + directionName(state.primaryDir) +
			"; activate to sort " + directionName(next)
	}

	secondary = SortControl{
		Key:       keyOf(state.secondary),
		Direction: state.secondaryDir,
		Options:   sortChoices(state, true),
	}
	if secondary.Direction == "" {
		secondary.Direction = "desc"
	}
	secondary.Arrow = directionArrow(secondary.Direction)
	if state.secondary == nil {
		secondary.ToggleLabel = "No secondary sort selected"
	} else {
		next := flipDirection(state.secondaryDir)
		toggled := state
		toggled.secondaryDir = next
		secondary.ToggleURL = sortURL(toggled)
		secondary.ToggleLabel = "Secondary sort by " + state.secondary.name + " " + directionName(state.secondaryDir) +
			"; activate to sort " + directionName(next)
	}
	return primary, secondary
}

func sortChoices(state sortState, secondary bool) []SortChoice {
	choices := make([]SortChoice, 0, len(sortColumns)+1)
	noneSelected := state.primary == nil
	if secondary {
		noneSelected = state.secondary == nil
	}
	choices = append(choices, SortChoice{Value: "", Label: "None", Selected: noneSelected})

	for i := range sortColumns {
		column := &sortColumns[i]
		choice := SortChoice{Value: column.key, Label: column.key}
		if secondary {
			choice.Selected = state.secondary != nil && state.secondary.key == column.key
			choice.Disabled = state.primary != nil && state.primary.key == column.key
		} else {
			choice.Selected = state.primary != nil && state.primary.key == column.key
		}
		choices = append(choices, choice)
	}
	return choices
}

func directionArrow(direction string) string {
	if direction == "asc" {
		return "↑"
	}
	return "↓"
}

func directionName(direction string) string {
	if direction == "asc" {
		return "ascending"
	}
	return "descending"
}

func flipDirection(direction string) string {
	if direction == "asc" {
		return "desc"
	}
	return "asc"
}

func keyOf(column *sortColumn) string {
	if column == nil {
		return ""
	}
	return column.key
}
