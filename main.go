package main

import (
	"fmt"
	"net/http"

	"github.com/AguilaMike/lenslocked/controllers"
	"github.com/go-chi/chi/v5"
)

func homeHandler(r *chi.Mux) {
	controllers.RegisterGetControllerWithTemplate(r, "/", "templates", "home.gohtml", nil)
}

func contactHandler(r *chi.Mux) {
	controllers.RegisterGetControllerWithTemplate(r, "/contact", "templates", "contact.gohtml", nil)
}

func faqHandler(r *chi.Mux) {
	controllers.RegisterGetControllerWithTemplate(r, "/faq", "templates", "faq.gohtml", nil)
}

func main() {
	r := chi.NewRouter()

	homeHandler(r)
	contactHandler(r)
	faqHandler(r)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("404 Not Found: %s", r.URL.Path), http.StatusNotFound)
	})
	fmt.Println("Starting the server on :8080...")
	http.ListenAndServe(":8080", r)
}
