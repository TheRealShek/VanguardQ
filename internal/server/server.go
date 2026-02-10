package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Start() {
	router := chi.NewRouter()
	registerRoutes(router)
	log.Println("Listening on :8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(err)
	}
}
