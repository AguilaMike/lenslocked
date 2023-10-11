package main

import (
	"fmt"
	"net/http"

	"github.com/AguilaMike/lenslocked/pkg/app/router"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	router.Router(r)
	fmt.Println("Starting the server on :8080...")
	fmt.Println(http.ListenAndServe(":8080", r))
}
