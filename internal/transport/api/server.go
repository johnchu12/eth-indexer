package api

import (
	"hw/internal/service"
	"hw/pkg/micro-tree/http/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Server represents the HTTP server with dependencies injected.
type Server struct {
	fx.In

	Logger  *zap.Logger
	Service service.Service
}

// ConfigureHTTPServer sets up the HTTP routes and middleware for the Fiber app.
func ConfigureHTTPServer(app *fiber.App, s Server) {
	// Configure CORS settings
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://0.0.0.0:3000",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	// Apply logging middleware
	middleware.UseLog(app, s.Logger)

	// Apply error handling and recovery middleware
	app.Use(middleware.ErrorHandler(s.Logger))
	app.Use(middleware.Recovery(s.Logger))

	// Define routes
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})
	app.Get("/user/:id", s.GetUser)
	app.Get("/user/:id/history", s.GetHistory)
	app.Get("/leaderboard", s.GetLeaderboard)
}
