package main

import (
	"log"

	"hw/internal/repository"
	"hw/internal/service"
	"hw/internal/transport/api"
	"hw/pkg/environment"
	"hw/pkg/logger"
	httpserver "hw/pkg/micro-tree/http/server"
	"hw/pkg/pg"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func init() {
	logger.Init()
}

type ServerConfig struct {
	PORT string `envconfig:"PORT" default:"8080"`
}

var config ServerConfig

func init() {
	// Load API environment configuration
	if err := environment.LoadConfig("server", &config); err != nil {
		log.Fatalf("Failed to load Server configuration: %v", err)
	}
	logger.Infof("Server configuration: %+v", config)
}

// TODO: add cache
func main() {
	app := fx.New(
		// Provide dependency services
		fx.Provide(
			logger.Init,
			pg.NewPostgresDB,
			repository.NewRepository,
			service.NewService,
		),
		// Start HTTP server
		httpserver.Module,

		// Configure HTTP server
		fx.Invoke(
			api.ConfigureHTTPServer,
		),

		// dependency services' logs can be commented
		fx.NopLogger,

		fx.WithLogger(
			func(log *zap.Logger) fxevent.Logger {
				return &logger.CustomZapLogger{Logger: log}
			},
		),
	)

	// Run the application
	app.Run()
}
