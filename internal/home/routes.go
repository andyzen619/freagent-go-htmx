// Package home serves the app's landing page.
package home

import (
	"embed"
	"net/http"

	"fantasy_waiver_assistant/internal/web"
)

//go:embed templates/page.html
var templates embed.FS

// page is parsed once at startup, merged over the shared layout. The template
// is compiled into the binary, so the server does not depend on the working
// directory.
var page = web.Parse(templates, "templates/page.html")

// LoginState reports whether a Yahoo session exists. *auth.Client satisfies it,
// which keeps this package independent of the auth package.
type LoginState interface {
	LoggedIn() bool
}

// Handler serves the home page.
type Handler struct {
	login LoginState
}

// New returns a Handler that renders the page for the given login state.
func New(login LoginState) *Handler {
	return &Handler{login: login}
}

// Register adds the home route to mux. The pattern is a catch-all: any path not
// matched by a more specific route renders the home page.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", h.index)
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	web.Render(w, r, page, "content", map[string]any{
		"LoggedIn": h.login.LoggedIn(),
		"Nav":      "home", // marks the active tab in the shared layout
	})
}
