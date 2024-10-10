package main

import (
	"fmt"
	"log"
	"os"

	"hw/internal/indexer/handlers"
	"hw/internal/repository"
	"hw/internal/service"
	"hw/pkg/ethindexa"
	"hw/pkg/logger"
	"hw/pkg/pg"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func migrateDB() {
	// TODO: Configure according to production environment settings
	connString := os.Getenv("DATABASE_URL")

	m, err := migrate.New(
		"file://migrations",
		connString,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Execute Down migration to remove all tables
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration Down failed: %v", err)
	}

	// Execute Up migration to recreate tables
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
}

// setupIndexer initializes the indexer with the necessary handlers and starts the event listeners.
func setupIndexer(db *pg.PostgresDB, svc service.Service) error {
	// Define all event handlers to be registered
	// key come from contract {name}:{network}:{event} in config file
	handlersMap := map[string]ethindexa.EventHandler{
		"UniswapV2:mainnet:Swap": handlers.HandleUSDCWETHSwap,

		// If you need to handle other events, add them here
		"USDC:mainnet:Transfer": handlers.HandleTransfer,
		"USDC:base:Approval":    handlers.HandleApproval,
		"AAVE:mainnet:Approval": handlers.HandleApproval,
	}

	// Create indexer with registered events only
	_, err := ethindexa.NewIndexer(db, svc, handlersMap)
	if err != nil {
		return fmt.Errorf("failed to create indexer: %w", err)
	}

	// Start all event listeners
	// indexer.StartAllEventListeners()

	return nil
}

func main() {
	// Initialize logger
	logger.Init()

	// Initialize PostgresDB
	db, err := pg.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to PostgresDB: %v", err)
	}
	defer db.Close()

	// Initialize Repository
	repo := repository.NewRepository(db)

	// Initialize Service
	svc := service.NewService(repo)

	// Perform database migrations
	migrateDB()

	// Setup Indexer
	if err := setupIndexer(db, svc); err != nil {
		log.Fatalf("Failed to setup indexer: %v", err)
	}

	// Block main goroutine
	select {}
}
