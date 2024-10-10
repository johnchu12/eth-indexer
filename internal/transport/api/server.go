package api

import (
	"errors"
	"net/http"

	"hw/internal/service"
	"hw/pkg/micro-tree/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// Server represents the HTTP server with its dependencies.
type Server struct {
	Logger  *zap.Logger
	Service service.Service
}

// errorResponse defines the error response structure
type errorResponse struct {
	Error          string `json:"error"`
	HTTPStatusCode int    `json:"-"` // http response status code
}

// Render implements the render.Renderer interface
func (e *errorResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	if e.HTTPStatusCode == 0 {
		e.HTTPStatusCode = http.StatusInternalServerError
	}
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ConfigureHTTPServer sets up the HTTP routes and middleware for the Chi router.
func ConfigureHTTPServer(router *chi.Mux, srv Server) {
	// Configure CORS settings
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://0.0.0.0:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	// 添加 Request ID 中間件
	router.Use(middleware.RequestIDMiddleware())

	// Apply logging middleware
	router.Use(middleware.LoggingMiddleware(srv.Logger))

	// Apply error handling and recovery middleware
	// router.Use(middleware.ErrorHandler(s.Logger))
	router.Use(middleware.RecoveryMiddleware(srv.Logger))

	router.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("panic test")
	})

	router.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		middleware.HTTPErrorLogging(w, r, errors.New("errtest"))
		render.Render(w, r, &errorResponse{Error: "errtest"})
	})

	// Define routes
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	router.Get("/user/{id}", srv.GetUser)
	router.Get("/user/{id}/history", srv.GetHistory)
	router.Get("/leaderboard", srv.GetLeaderboard)
}
