package controllers

import "github.com/go-chi/chi/v5"

func SignUp(r *chi.Mux) {
	registerGetControllerDefaultFs(r, "/signup", "layout.gohtml", "pages", "auth", "signup.gohtml")
}
