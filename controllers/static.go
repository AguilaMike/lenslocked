package controllers

import (
	"net/http"
	"path/filepath"

	"github.com/AguilaMike/lenslocked/templates"
	"github.com/AguilaMike/lenslocked/views"
	"github.com/go-chi/chi/v5"
)

type Static struct {
	Template views.Template
}

func registerGetControllerDefaultFs(r *chi.Mux, path, layout string, pages ...string) {
	registerGetControllerWithTemplateFs(r, nil, path, joinPath("layout", layout), joinPath(pages...))
}

func registerGetControllerWithTemplateFs(r *chi.Mux, data interface{}, path string, templateFile ...string) {
	tpl := views.Must(views.ParseFS(templates.FS, templateFile...))
	r.Get(path, staticHandler(tpl, data))
}

func registerGetControllerWithTemplate(r *chi.Mux, path, templateFolder, templateFile string, data interface{}) {
	tpl := views.Must(views.Parse(filepath.Join(templateFolder, templateFile)))
	r.Get(path, staticHandler(tpl, data))
}

func staticHandler(tpl *views.Template, data interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, data)
	}
}

func joinPath(path ...string) string {
	return filepath.Join(path...)
}

func Home(r *chi.Mux) {
	registerGetControllerDefaultFs(r, "/", "layout.gohtml", "pages", "home.gohtml")
}

func Contact(r *chi.Mux) {
	registerGetControllerDefaultFs(r, "/contact", "layout.gohtml", "pages", "contact.gohtml")
}
