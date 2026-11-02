// Package web is the shared presentation layer: the HTML layout shell and the
// page-versus-fragment decision htmx relies on.
package web

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

//go:embed templates/layout.html
var layoutFS embed.FS

// Parse merges the shared layout with a feature's templates. The feature must
// define "content" and every fragment name passed to Render.
func Parse(feature fs.FS, patterns ...string) *template.Template {
	t := template.Must(template.ParseFS(layoutFS, "templates/layout.html"))
	return template.Must(t.ParseFS(feature, patterns...))
}

// Render executes the layout on direct visits, or just the fragment when the
// request comes from htmx. Vary stops a cached fragment from being served as a
// full page on reload.
func Render(w http.ResponseWriter, r *http.Request, t *template.Template, fragment string, data any) {
	name := "layout"
	if r.Header.Get("HX-Request") == "true" {
		name = fragment
	}
	w.Header().Add("Vary", "HX-Request")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		// Headers are already sent, so log instead of writing an error page.
		log.Printf("web: render %q: %v", name, err)
	}
}
