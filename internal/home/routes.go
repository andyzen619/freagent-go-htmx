// Package home serves the app's landing page and login status.
package home

import (
	"context"
	"embed"
	"net/http"

	"fantasy_waiver_assistant/internal/session"
	"fantasy_waiver_assistant/internal/web"
)

//go:embed templates/page.html
var templates embed.FS

// page is parsed once at startup, merged over the shared layout. The template
// is compiled into the binary, so the server does not depend on the working
// directory.
var page = web.Parse(templates, "templates/page.html")

// SessionValidator checks whether the Yahoo token is still accepted.
type SessionValidator interface {
	Validate(ctx context.Context) session.Status
}

// Handler serves the home page.
type Handler struct {
	validator SessionValidator
}

// New returns a Handler that renders the verified login state.
func New(validator SessionValidator) *Handler {
	return &Handler{validator: validator}
}

// RegisterPublic adds the unauthenticated home route.
func (h *Handler) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.index)
}

// Register adds a fallback for paths not matched by a more specific route.
// Those requests are still gated by middleware.RequireSession.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", h.index)
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	status, ok := session.FromContext(r.Context())
	if !ok {
		status = h.validator.Validate(r.Context())
	}

	web.Render(w, r, page, "content", map[string]any{
		"Status": status.String(),
		"Nav":    "home", // marks the active tab in the shared layout
	})
}
