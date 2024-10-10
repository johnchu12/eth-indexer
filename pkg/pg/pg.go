package pg

import (
	"context"
	"fmt"

	"hw/pkg/common"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// mockgen -source=pkg/pg/pg.go -destination=pkg/pg/mocks/pg_mock.go -package=mocks

// PgxPool defines the methods required by pgxpool.Pool.
type PgxPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// PgxTx defines the methods required by pgx.Tx.
type PgxTx interface {
	pgx.Tx
}

type PgxRows interface {
	pgx.Rows
}

// PostgresDB encapsulates a pgx connection pool.
type PostgresDB struct {
	pool PgxPool
}

func (db *PostgresDB) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

func (db *PostgresDB) Close() {
	db.pool.Close()
}

func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db *PostgresDB) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, sql, arguments...)
}

func (db *PostgresDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, args...)
}

func (db *PostgresDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return db.pool.QueryRow(ctx, sql, args...)
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
