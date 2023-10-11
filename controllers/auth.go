package controllers

import (
	"github.com/AguilaMike/lenslocked/templates"
	"github.com/AguilaMike/lenslocked/views"
	"github.com/go-chi/chi/v5"
)

func SignUp(r *chi.Mux) {
	var usersC Users
	usersC.Templates.New = views.Must(
		views.ParseFS(
			templates.FS,
			joinPath("layout", "layout.gohtml"),
			joinPath("pages", "auth", "signup.gohtml")))
	r.Get("/signup", usersC.New)
}
