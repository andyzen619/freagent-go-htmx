// Package players serves the free-agent browser: a full page plus the rows
// fragment htmx swaps into it.
package players

import (
	"errors"
	"log"
	"net/http"

	"fantasy_waiver_assistant/internal/service/freeagents"
	"fantasy_waiver_assistant/internal/web"
)

// page is parsed once at startup, merged over the shared layout.
var page = web.Parse(templates, "templates/*.html")

// Handler serves the players page and its rows fragment.
type Handler struct {
	source freeagents.Source
	mock   bool
}

// New returns a Handler that reads players from source.
func New(source freeagents.Source) *Handler {
	return &Handler{source: source}
}

// NewMock returns a Handler that labels the page as mock data.
func NewMock(source freeagents.Source) *Handler {
	return &Handler{source: source, mock: true}
}

// Register adds the page route to mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /players", h.index)
}

// pageData is the template input for the page and its rows fragment.
type pageData struct {
	Nav           string // active tab in the shared layout; ignored by the fragment
	Players       []PlayerView
	Error         string
	Mock          bool
	PrimarySort   SortControl
	SecondarySort SortControl
}

// load resolves the free agents shown on the page.
func (h *Handler) load(r *http.Request) pageData {
	state := parseSort(r.URL.Query())
	primary, secondary := sortControls(state)
	data := pageData{
		Nav:           "players",
		Mock:          h.mock,
		PrimarySort:   primary,
		SecondarySort: secondary,
	}

	found, err := h.source.FreeAgents(r.Context())
	if err != nil {
		log.Printf("players: load: %v", err)
		msg := "Could not load free agents."
		if errors.Is(err, freeagents.ErrNoLeague) {
			msg = err.Error()
		}
		data.Error = msg
		return data
	}

	sortPlayers(found, state)
	data.Players = playerViews(found, displayColumns(state))
	return data
}

// index serves /players. Direct visits get the layout; htmx gets the panel.
func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	web.Render(w, r, page, "panel", h.load(r))
}
