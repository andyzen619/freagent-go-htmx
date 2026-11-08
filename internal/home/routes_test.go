package home

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

func newMux(validator SessionValidator) *http.ServeMux {
	mux := http.NewServeMux()
	New(validator).RegisterPublic(mux)
	New(validator).Register(mux)
	return mux
}

func get(t *testing.T, status session.Status, path string) (*httptest.ResponseRecorder, *stubSession) {
	t.Helper()

	validator := &stubSession{status: status}
	rec := httptest.NewRecorder()
	newMux(validator).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec, validator
}

func TestIndexLoggedOut(t *testing.T) {
	rec, _ := get(t, session.LoggedOut, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `href="/login"`) {
		t.Error("logged-out page should link to /login")
	}
	if strings.Contains(body, "Verified with Yahoo") {
		t.Error("logged-out page should not show the verified message")
	}
}

func TestIndexVerified(t *testing.T) {
	rec, _ := get(t, session.Verified, "/")

	body := rec.Body.String()
	if !strings.Contains(body, "Verified with Yahoo") {
		t.Error("verified page should show the verified message")
	}
	if strings.Contains(body, `href="/login"`) {
		t.Error("verified page should not link to /login")
	}
}

func TestIndexAccessDenied(t *testing.T) {
	rec, _ := get(t, session.AccessDenied, "/")

	body := rec.Body.String()
	for _, want := range []string{
		"Yahoo API access is not authorized",
		"Fantasy Sports → Read",
		`href="https://developer.yahoo.com/apps/"`,
		"This application is not authorized to perform this action.",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("access-denied page missing %q", want)
		}
	}
	if strings.Contains(body, "Unable to verify right now") {
		t.Error("access-denied page should not show the generic outage message")
	}
}

func TestIndexUnavailable(t *testing.T) {
	rec, _ := get(t, session.Unavailable, "/")

	body := rec.Body.String()
	if !strings.Contains(body, "Unable to verify right now") {
		t.Error("unavailable page should explain that verification failed")
	}
	if strings.Contains(body, `href="/login"`) {
		t.Error("unavailable page should not discard a valid session")
	}
}

func TestIndexValidatesOnce(t *testing.T) {
	_, validator := get(t, session.Verified, "/")

	if validator.calls != 1 {
		t.Errorf("Validate calls = %d, want 1", validator.calls)
	}
}

func TestIndexReusesMiddlewareStatus(t *testing.T) {
	validator := &stubSession{status: session.LoggedOut}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(session.WithStatus(req.Context(), session.Verified))
	rec := httptest.NewRecorder()

	newMux(validator).ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "Verified with Yahoo") {
		t.Error("handler should use the status validated earlier in the request")
	}
	if validator.calls != 0 {
		t.Errorf("Validate calls = %d, want 0", validator.calls)
	}
}

// The shared layout marks the current section, so the nav tab carries
// aria-current and the "on" style.
func TestIndexMarksHomeTab(t *testing.T) {
	rec, _ := get(t, session.Verified, "/")

	body := rec.Body.String()
	if !strings.Contains(body, `aria-current="page"`) || !strings.Contains(body, `class="on"`) {
		t.Errorf("home tab should be marked active, got %q", body)
	}
}

// The fallback route keeps unknown paths on the home page rather than a 404.
func TestIndexMatchesUnknownPath(t *testing.T) {
	rec, _ := get(t, session.Verified, "/does-not-exist")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
