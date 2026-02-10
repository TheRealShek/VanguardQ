package server

import (
	"github.com/TheRealShek/VanguardQ/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func registerRoutes(r *chi.Mux) {
	r.Get("/health", handlers.Health)

	r.Route("/jobs", func(r chi.Router) {
		r.Post("/", handlers.EnqueueImmediateJob)
		r.Post("/delayed", handlers.EnqueueDelayedJob)
		r.Get("/{id}", handlers.GetJob)
		r.Delete("/{id}", handlers.DeleteJob)
	})

	r.Route("/queues", func(r chi.Router) {
		r.Get("/{name}/stats", handlers.QueueStats)
	})

}
