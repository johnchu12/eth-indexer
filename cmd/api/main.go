package main

import (
	"log"
	"net/http"

	"hw/internal/repository"
	"hw/internal/service"
	"hw/internal/transport/api"
	"hw/pkg/environment"
	"hw/pkg/logger"
	"hw/pkg/micro-tree/http/server"
	"hw/pkg/pg"

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
	// Initialize the database
	db, err := pg.NewPostgresDB()
	if err != nil {
		log.Fatal("Failed to initialize the database", zap.Error(err))
	}

	// Initialize the repository
	repo := repository.NewRepository(db)

	// Initialize the service
	svc := service.NewService(repo)

	l := logger.Init()

	app := server.NewHTTPServer()

	server := api.Server{
		Logger:  l,
		Service: svc,
	}
	// Configure HTTP server
	api.ConfigureHTTPServer(app, server)

	// Start HTTP server
	logger.Infof("Start server on port: %s", config.PORT)
	if err := http.ListenAndServe(":"+config.PORT, app); err != nil {
		log.Fatal("Failed to start the server", zap.Error(err))
	}
}
