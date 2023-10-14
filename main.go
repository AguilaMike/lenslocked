package main

import (
	"fmt"
	"net/http"

	"github.com/AguilaMike/lenslocked/pkg/app/models"
	"github.com/AguilaMike/lenslocked/pkg/app/router"
	"github.com/go-chi/chi/v5"
)

func main() {
	// Setup a database connection
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Setup our model services
	userService := models.UserService{
		DB: db,
	}

	r := chi.NewRouter()
	router.Router(r, userService)
	fmt.Println("Starting the server on :8080...")
	fmt.Println(http.ListenAndServe(":8080", r))
}
