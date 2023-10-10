package main

import (
	"fmt"
	"net/http"

	"github.com/AguilaMike/lenslocked/controllers"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	controllers.Home(r)
	controllers.Contact(r)
	controllers.FAQ(r)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("404 Not Found: %s", r.URL.Path), http.StatusNotFound)
	})
	fmt.Println("Starting the server on :8080...")
	http.ListenAndServe(":8080", r)
}
