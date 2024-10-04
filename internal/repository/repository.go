package repository

import (
	"context"
	"time"

	"hw/internal/model"
	"hw/pkg/pg"

	"github.com/jackc/pgx/v5"
)

// DB defines the interface for database operations.
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Repository defines the interface for repository operations.
type Repository interface {
	DB() DB
	BeginTransaction(ctx context.Context) (pgx.Tx, error)
	CreatePointsHistory(db DB, ctx context.Context, pointsHistory *model.PointsHistory) error
	IsOnboardingTaskCompleted(db DB, ctx context.Context, account string) (bool, error)
	GetPointsHistory(db DB, ctx context.Context, account, token string) ([]model.PointsHistory, error)
	CreateSwapHistory(db DB, ctx context.Context, swapHistory *model.SwapHistory) error
	GetSwapTotalUsd(db DB, ctx context.Context, account, token string) (float64, error)
	GetUserSwapSummary(db DB, ctx context.Context, account string) (map[string]float64, error)
	GetUserSwapSummaryLast7Days(db DB, ctx context.Context, referenceTime time.Time, token string) ([]model.UserSwapPercentage, error)
	GetTokenByAddress(db DB, ctx context.Context, address string) (*model.Token, error)
	CreateToken(db DB, ctx context.Context, token *model.Token) error
	CreateUser(db DB, ctx context.Context, userId string) (*model.User, error)
	GetUserByAddress(db DB, ctx context.Context, address string) (*model.User, error)
	UpsertUserPoints(db DB, ctx context.Context, address string, point float64) error
	GetLeaderboard(db DB, ctx context.Context) ([]model.User, error)
}

// repository manages database operations for users.
type repository struct {
	db DB
}

// BeginTransaction starts a new transaction.
func (r *repository) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

// DB returns the database interface.
func (r *repository) DB() DB {
	return r.db
}

// NewRepository creates a new Repository with the provided PostgresDB.
func NewRepository(pgdb *pg.PostgresDB) Repository {
	return &repository{
		db: pgdb.GetPool(),
	}
}
