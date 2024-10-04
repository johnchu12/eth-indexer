package repository

import (
	"context"
	"fmt"

	"hw/internal/model"

	"github.com/jackc/pgx/v5"
)

// CreateUser inserts a new user into the users table.
func (r *repository) CreateUser(db DB, ctx context.Context, userId string) (*model.User, error) {
	const query = `
		INSERT INTO users (address)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`

	user := &model.User{
		Address: userId,
	}

	err := db.QueryRow(ctx, query, user.Address).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByAddress retrieves a user by their address.
func (r *repository) GetUserByAddress(db DB, ctx context.Context, address string) (*model.User, error) {
	const query = `
		SELECT id, address, total_points, created_at, updated_at
		FROM users
		WHERE address = $1
	`

	rows, err := db.Query(ctx, query, address)
	if err != nil {
		return &model.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[model.User])
	if err != nil {
		if err == pgx.ErrNoRows {
			return &model.User{}, model.ErrUserNotFound
		}
		return &model.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpsertUserPoints atomically updates a user's total points.
func (r *repository) UpsertUserPoints(db DB, ctx context.Context, address string, point float64) error {
	const query = `
		INSERT INTO users (address, total_points)
		VALUES ($1, $2)
		ON CONFLICT (address) DO UPDATE SET 
			total_points = users.total_points + EXCLUDED.total_points,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`

	user := &model.User{
		Address:     address,
		TotalPoints: point,
	}

	err := db.QueryRow(ctx, query, user.Address, user.TotalPoints).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert user points: %w", err)
	}

	return nil
}

// GetLeaderboard retrieves the leaderboard.
func (r *repository) GetLeaderboard(db DB, ctx context.Context) ([]model.User, error) {
	const query = `
		SELECT id, address, total_points, created_at, updated_at
		FROM users
		ORDER BY total_points DESC
	`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[model.User])
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return users, nil
}
