// Package middleware provides HTTP wrappers shared by the app's routes.
package middleware

import (
	"context"
	"net/http"

	"fantasy_waiver_assistant/internal/session"
)

// SessionValidator reports whether the Yahoo session is actually usable.
type SessionValidator interface {
	Validate(ctx context.Context) session.Status
}

// RequireSession lets verified requests through. Page visits are redirected to
// /login when logged out; htmx requests instead get an HX-Redirect header so the
// client navigates there without swapping a login page into the target.
// Missing Yahoo API permissions also return to the home page, which explains
// the problem. A Yahoo outage returns 503 and leaves the token alone.
func RequireSession(validator SessionValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := validator.Validate(r.Context())
		switch status {
		case session.Verified:
			next.ServeHTTP(w, r.WithContext(session.WithStatus(r.Context(), status)))
		case session.LoggedOut:
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
		case session.AccessDenied:
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/")
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		default:
			http.Error(w, "Could not verify your Yahoo login. Try again shortly.", http.StatusServiceUnavailable)
		}
	})
}
