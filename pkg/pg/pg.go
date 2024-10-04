package pg

import (
	"context"
	"fmt"

	"hw/pkg/common"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB encapsulates a pgx connection pool.
type PostgresDB struct {
	pool *pgxpool.Pool
}

// PoolCreator is a function that creates a new pgxpool.Pool.
var PoolCreator = func(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error) {
	return pgxpool.NewWithConfig(ctx, config)
}

// NewPostgresDB creates and initializes a new instance of PostgresDB.
func NewPostgresDB() (*PostgresDB, error) {
	connString := common.GetEnv("DATABASE_URL", "")
	if connString == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	// Set up the connection pool configuration.
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Create the connection pool.
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection.
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	return &PostgresDB{pool: pool}, nil
}

// Close closes the connection pool.
func (db *PostgresDB) Close() {
	db.pool.Close()
}

// GetPool returns the underlying pgxpool.Pool.
func (db *PostgresDB) GetPool() *pgxpool.Pool {
	return db.pool
}
