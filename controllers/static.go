package controllers

import (
	"net/http"
	"path/filepath"

	"github.com/AguilaMike/lenslocked/views"
	"github.com/go-chi/chi/v5"
)

type Static struct {
	Template views.Template
}

func RegisterGetControllerWithTemplate(r *chi.Mux, path, templateFolder, templateFile string, data interface{}) {
	tpl, err := views.Parse(filepath.Join(templateFolder, templateFile))
	if err != nil {
		panic(err)
	}
	r.Get(path, staticHandler(tpl, data))
}

func staticHandler(tpl views.Template, data interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, data)
	}
}
