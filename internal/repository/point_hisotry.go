package repository

import (
	"context"
	"fmt"

	"hw/internal/model"
)

// CreatePointsHistory inserts a new PointsHistory record into the database.
func (r *repository) CreatePointsHistory(ctx context.Context, pointsHistory *model.PointsHistory) error {
	const query = `
		INSERT INTO points_history (token, account, points, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		pointsHistory.Token,
		pointsHistory.Account,
		pointsHistory.Points,
		pointsHistory.Description,
	).Scan(&pointsHistory.ID, &pointsHistory.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create points history record: %w", err)
	}

	return nil
}

// IsOnboardingTaskCompleted checks if the onboarding task is completed for the specified account.
func (r *repository) IsOnboardingTaskCompleted(ctx context.Context, account string) (bool, error) {
	const (
		description = "onboarding_task"
		query       = `
			SELECT COUNT(*)
			FROM points_history
			WHERE account = $1 AND description = $2
		`
	)

	var count int
	if err := r.db.QueryRow(ctx, query, account, description).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to retrieve points history records: %w", err)
	}

	return count > 0, nil
}

// GetPointsHistory retrieves the points history for the specified account and token.
func (r *repository) GetPointsHistory(ctx context.Context, account, token string) ([]model.PointsHistory, error) {
	const query = `
		SELECT id, token, account, points, description, created_at
		FROM points_history
		WHERE account = $1 AND token = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, account, token)
	if err != nil {
		return nil, fmt.Errorf("failed to query points history: %w", err)
	}
	defer rows.Close()

	var histories []model.PointsHistory
	for rows.Next() {
		var ph model.PointsHistory
		if err := rows.Scan(
			&ph.ID,
			&ph.Token,
			&ph.Account,
			&ph.Points,
			&ph.Description,
			&ph.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan points history row: %w", err)
		}
		histories = append(histories, ph)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate through points history rows: %w", err)
	}

	return histories, nil
}
