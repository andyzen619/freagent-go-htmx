package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// stubLogin is a LoginState with a fixed answer.
type stubLogin bool

func (s stubLogin) LoggedIn() bool { return bool(s) }

func TestRequireSession(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	t.Run("logged in passes through", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RequireSession(stubLogin(true), ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want 204", rec.Code)
		}
	})

	t.Run("page redirects to login", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RequireSession(stubLogin(false), ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/login" {
			t.Errorf("status = %d, Location = %q, want 302 /login", rec.Code, rec.Header().Get("Location"))
		}
	})

	t.Run("htmx request gets HX-Redirect", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()
		RequireSession(stubLogin(false), ok).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || rec.Header().Get("HX-Redirect") != "/login" {
			t.Errorf("status = %d, HX-Redirect = %q, want 200 /login", rec.Code, rec.Header().Get("HX-Redirect"))
		}
	})
}
