package repository

import (
	"context"
	"fmt"

	"hw/internal/model"

	"github.com/jackc/pgx/v5"
)

// CreateUser inserts a new user into the users table.
func (r *repository) CreateUser(ctx context.Context, userId string) (*model.User, error) {
	const query = `
		INSERT INTO users (address)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`

	user := &model.User{
		Address: userId,
	}

	err := r.db.QueryRow(ctx, query, user.Address).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByAddress retrieves a user by their address.
func (r *repository) GetUserByAddress(ctx context.Context, address string) (*model.User, error) {
	const query = `
		SELECT id, address, total_points, created_at, updated_at
		FROM users
		WHERE address = $1
		LIMIT 1
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, address).Scan(
		&user.ID,
		&user.Address,
		&user.TotalPoints,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpsertUserPoints atomically updates a user's total points.
func (r *repository) UpsertUserPoints(ctx context.Context, address string, point float64) error {
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

	err := r.db.QueryRow(ctx, query, user.Address, user.TotalPoints).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert user points: %w", err)
	}

	return nil
}

// GetLeaderboard retrieves the leaderboard.
func (r *repository) GetLeaderboard(ctx context.Context) ([]model.User, error) {
	const query = `
		SELECT id, address, total_points, created_at, updated_at
		FROM users
		ORDER BY total_points DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.Address,
			&user.TotalPoints,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}
