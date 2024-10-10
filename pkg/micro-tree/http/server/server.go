package server

import (
	"hw/pkg/logger"

	"github.com/go-chi/chi/v5"
)

// NewHTTPServer initializes and returns a new Chi router.
func NewHTTPServer() *chi.Mux {
	logger.Infof("Initializing Chi router.")

	router := chi.NewRouter()

	return router
}
