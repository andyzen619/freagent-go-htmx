package home

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubLogin is a LoginState with a fixed answer.
type stubLogin bool

func (s stubLogin) LoggedIn() bool { return bool(s) }

func get(t *testing.T, loggedIn bool, path string) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	New(stubLogin(loggedIn)).Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestIndexLoggedOut(t *testing.T) {
	rec := get(t, false, "/")

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
	if strings.Contains(body, "Logged in with Yahoo") {
		t.Error("logged-out page should not show the logged-in message")
	}
}

func TestIndexLoggedIn(t *testing.T) {
	body := get(t, true, "/").Body.String()

	if !strings.Contains(body, "Logged in with Yahoo") {
		t.Error("logged-in page should show the logged-in message")
	}
	if strings.Contains(body, `href="/login"`) {
		t.Error("logged-in page should not link to /login")
	}
}

// The shared layout marks the current section, so the nav tab carries
// aria-current and the "on" style.
func TestIndexMarksHomeTab(t *testing.T) {
	body := get(t, true, "/").Body.String()

	if !strings.Contains(body, `aria-current="page"`) || !strings.Contains(body, `class="on"`) {
		t.Errorf("home tab should be marked active, got %q", body)
	}
}

// The root route is deliberately a catch-all, so unknown paths render the home
// page rather than a 404.
func TestIndexMatchesUnknownPath(t *testing.T) {
	rec := get(t, false, "/does-not-exist")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
