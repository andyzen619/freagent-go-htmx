package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"fantasy_waiver_assistant/internal/session"
)

// stubSession is a SessionValidator with a fixed answer.
type stubSession struct {
	status session.Status
	calls  int
}

func (s *stubSession) Validate(context.Context) session.Status {
	s.calls++
	return s.status
}

func TestRequireSession(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	t.Run("verified passes through", func(t *testing.T) {
		validator := &stubSession{status: session.Verified}
		rec := httptest.NewRecorder()
		RequireSession(validator, ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want 204", rec.Code)
		}
		if validator.calls != 1 {
			t.Errorf("Validate calls = %d, want 1", validator.calls)
		}
	})

	t.Run("logged out page redirects to login", func(t *testing.T) {
		validator := &stubSession{status: session.LoggedOut}
		rec := httptest.NewRecorder()
		RequireSession(validator, ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/login" {
			t.Errorf("status = %d, Location = %q, want 302 /login", rec.Code, rec.Header().Get("Location"))
		}
	})

	t.Run("logged out htmx request gets HX-Redirect", func(t *testing.T) {
		validator := &stubSession{status: session.LoggedOut}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()
		RequireSession(validator, ok).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || rec.Header().Get("HX-Redirect") != "/login" {
			t.Errorf("status = %d, HX-Redirect = %q, want 200 /login", rec.Code, rec.Header().Get("HX-Redirect"))
		}
	})

	t.Run("access denied returns to home", func(t *testing.T) {
		validator := &stubSession{status: session.AccessDenied}
		rec := httptest.NewRecorder()
		RequireSession(validator, ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/players", nil))
		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
			t.Errorf("status = %d, Location = %q, want 303 /", rec.Code, rec.Header().Get("Location"))
		}
	})

	t.Run("access denied htmx request gets HX-Redirect home", func(t *testing.T) {
		validator := &stubSession{status: session.AccessDenied}
		req := httptest.NewRequest(http.MethodGet, "/hx/players/panel", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()
		RequireSession(validator, ok).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || rec.Header().Get("HX-Redirect") != "/" {
			t.Errorf("status = %d, HX-Redirect = %q, want 200 /", rec.Code, rec.Header().Get("HX-Redirect"))
		}
	})

	t.Run("unavailable returns 503", func(t *testing.T) {
		validator := &stubSession{status: session.Unavailable}
		rec := httptest.NewRecorder()
		RequireSession(validator, ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want 503", rec.Code)
		}
	})
}
