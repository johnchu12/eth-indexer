package repository

import (
	"context"
	"time"

	"hw/internal/model"
	"hw/pkg/pg"
)

// mockgen -source=internal/repository/repository.go -destination=internal/repository/mocks/repository_mock.go -package=mocks

// Repository defines the interface for repository operations.
type Repository interface {
	// BeginTransaction starts a new transaction.
	BeginTransaction(ctx context.Context) (pg.PgxTx, error)
	// CreatePointsHistory inserts a new PointsHistory record into the database.
	CreatePointsHistory(ctx context.Context, pointsHistory *model.PointsHistory) error
	// IsOnboardingTaskCompleted checks if the onboarding task is completed for the specified account.
	IsOnboardingTaskCompleted(ctx context.Context, account string) (bool, error)
	// GetPointsHistory retrieves the points history for the specified account and token.
	GetPointsHistory(ctx context.Context, account, token string) ([]model.PointsHistory, error)
	// CreateSwapHistory inserts a new swap history record into the database.
	CreateSwapHistory(ctx context.Context, swapHistory *model.SwapHistory) error
	// GetSwapTotalUsd retrieves the total USD value of swaps for a given account and token.
	GetSwapTotalUsd(ctx context.Context, account, token string) (float64, error)
	// GetUserSwapSummary retrieves the sum of USD values grouped by token for a given account.
	GetUserSwapSummary(ctx context.Context, account string) (map[string]float64, error)
	// GetUserSwapSummaryLast7Days retrieves the total USD and percentage of swaps for each user over the past seven days for a specific token.
	GetUserSwapSummaryLast7Days(ctx context.Context, referenceTime time.Time, token string) ([]model.UserSwapPercentage, error)
	// GetTokenByAddress retrieves a token by its address from the database.
	GetTokenByAddress(ctx context.Context, address string) (*model.Token, error)
	// CreateToken inserts a new token into the database.
	CreateToken(ctx context.Context, token *model.Token) error
	// CreateUser inserts a new user into the users table.
	CreateUser(ctx context.Context, userId string) (*model.User, error)
	// GetUserByAddress retrieves a user by their address.
	GetUserByAddress(ctx context.Context, address string) (*model.User, error)
	// UpsertUserPoints atomically updates a user's total points.
	UpsertUserPoints(ctx context.Context, address string, point float64) error
	// GetLeaderboard retrieves the leaderboard.
	GetLeaderboard(ctx context.Context) ([]model.User, error)
}

// repository manages database operations for users.
type repository struct {
	db pg.PgxPool
}

// BeginTransaction starts a new transaction.
func (r *repository) BeginTransaction(ctx context.Context) (pg.PgxTx, error) {
	return r.db.Begin(ctx)
}

// NewRepository creates a new Repository with the provided PostgresDB.
func NewRepository(pgdb pg.PgxPool) Repository {
	return &repository{
		db: pgdb,
	}
}
