// Package middleware provides HTTP wrappers shared by the app's routes.
package middleware

import "net/http"

// LoginState reports whether a Yahoo session exists. *auth.Client satisfies it,
// which keeps this package independent of the auth package.
type LoginState interface {
	LoggedIn() bool
}

// RequireSession lets logged-in requests through. Page visits are redirected to
// /login; htmx requests instead get an HX-Redirect header so the client
// navigates there without swapping a login page into the target.
func RequireSession(login LoginState, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if login.LoggedIn() {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/login")
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	})
}
