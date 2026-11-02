package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestRender(t *testing.T) {
	tmpl := Parse(fstest.MapFS{
		"page.html": {Data: []byte(`{{define "content"}}page{{end}}{{define "rows"}}row{{end}}`)},
	}, "page.html")

	// A direct visit gets the layout around the page content.
	rec := httptest.NewRecorder()
	Render(rec, httptest.NewRequest(http.MethodGet, "/", nil), tmpl, "rows", nil)
	if body := rec.Body.String(); !strings.Contains(body, "<html") || !strings.Contains(body, "page") {
		t.Errorf("direct visit: got %q, want layout with content", body)
	}
	if got := rec.Header().Get("Vary"); got != "HX-Request" {
		t.Errorf("Vary = %q, want HX-Request", got)
	}

	// An htmx request gets only the fragment.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Request", "true")
	rec = httptest.NewRecorder()
	Render(rec, req, tmpl, "rows", nil)
	if body := rec.Body.String(); strings.Contains(body, "<html") || !strings.Contains(body, "row") {
		t.Errorf("htmx request: got %q, want only the fragment", body)
	}
}
