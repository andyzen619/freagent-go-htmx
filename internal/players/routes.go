// Package players serves the free-agent browser: a full page plus the rows
// fragment htmx swaps into it.
package players

import (
	"context"
	"errors"
	"log"
	"net/http"

	"fantasy_waiver_assistant/internal/web"
)

// page is parsed once at startup, merged over the shared layout.
var page = web.Parse(templates, "templates/*.html")

// Source loads the free agents shown on the players page.
type Source interface {
	FreeAgents(ctx context.Context, pos string) ([]Player, error)
}

// Handler serves the players page and its rows fragment.
type Handler struct {
	source Source
}

// New returns a Handler that reads players from source.
func New(source Source) *Handler {
	return &Handler{source: source}
}

// Register adds the page and the fragment route to mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /players", h.index)
	mux.HandleFunc("GET /hx/players/rows", h.rows)
}

// pageData is the template input for both the page and the rows fragment.
type pageData struct {
	Pos     string
	Players []Player
	Error   string
}

// load reads the ?pos filter and resolves the matching players.
func (h *Handler) load(r *http.Request) pageData {
	pos := r.URL.Query().Get("pos")
	found, err := h.source.FreeAgents(r.Context(), pos)
	if err != nil {
		log.Printf("players: load %q: %v", pos, err)
		msg := "Could not load free agents."
		if errors.Is(err, ErrNoLeague) {
			msg = err.Error()
		}
		return pageData{Pos: pos, Error: msg}
	}
	return pageData{Pos: pos, Players: found}
}

// index serves /players. Direct visits get the layout; htmx gets the rows.
func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	web.Render(w, r, page, "rows", h.load(r))
}

// rows serves /hx/players/rows, which only ever returns the fragment.
func (h *Handler) rows(w http.ResponseWriter, r *http.Request) {
	web.Fragment(w, page, "rows", h.load(r))
}
