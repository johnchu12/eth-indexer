package repository

import (
	"context"
	"fmt"
	"time"

	"hw/internal/model"
)

// CreateSwapHistory inserts a new swap history record into the database.
func (r *repository) CreateSwapHistory(db DB, ctx context.Context, swapHistory *model.SwapHistory) error {
	const query = `
		INSERT INTO swap_history (token, account, transaction_hash, usd_value, last_updated)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := db.QueryRow(
		ctx,
		query,
		swapHistory.Token,
		swapHistory.Account,
		swapHistory.TransactionHash,
		swapHistory.UsdValue,
		swapHistory.LastUpdated,
	).Scan(&swapHistory.ID, &swapHistory.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create swap history: %w", err)
	}

	return nil
}

// GetSwapTotalUsd retrieves the total USD value of swaps for a given account and token.
func (r *repository) GetSwapTotalUsd(db DB, ctx context.Context, account, token string) (float64, error) {
	const query = `
		SELECT SUM(usd_value)
		FROM swap_history
		WHERE account = $1 AND token = $2
	`

	var totalUsd float64
	err := db.QueryRow(ctx, query, account, token).Scan(&totalUsd)
	if err != nil {
		return 0, fmt.Errorf("failed to get total swap USD: %w", err)
	}

	return totalUsd, nil
}

// GetUserSwapSummary retrieves the sum of USD values grouped by token for a given account.
func (r *repository) GetUserSwapSummary(db DB, ctx context.Context, account string) (map[string]float64, error) {
	const query = `
		SELECT token, SUM(usd_value)
		FROM swap_history
		WHERE account = $1
		GROUP BY token
	`

	rows, err := db.Query(ctx, query, account)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token USD sums: %w", err)
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var token string
		var sumUsd float64
		if err := rows.Scan(&token, &sumUsd); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[token] = sumUsd
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return result, nil
}

// GetUserSwapSummaryLast7Days retrieves the total USD and percentage of swaps for each user over the past seven days for a specific token.
func (r *repository) GetUserSwapSummaryLast7Days(db DB, ctx context.Context, referenceTime time.Time, token string) ([]model.UserSwapPercentage, error) {
	const query = `
		WITH total_usd AS (
			SELECT SUM(usd_value) AS sum_usd_value
			FROM swap_history
			WHERE last_updated BETWEEN $1 AND $2 AND token = $3
		)
		SELECT 
			account,
			SUM(usd_value) AS total_usd,
			(SUM(usd_value) / total_usd.sum_usd_value) AS percentage
		FROM swap_history, total_usd
		WHERE last_updated BETWEEN $1 AND $2 AND token = $3
		GROUP BY account, total_usd.sum_usd_value
		ORDER BY total_usd DESC
	`

	startTime := referenceTime.AddDate(0, 0, -7)
	endTime := referenceTime

	rows, err := db.Query(ctx, query, startTime, endTime, token)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user swap percentages: %w", err)
	}
	defer rows.Close()

	var results []model.UserSwapPercentage
	for rows.Next() {
		var usp model.UserSwapPercentage
		if err := rows.Scan(&usp.Account, &usp.TotalUSD, &usp.Percentage); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, usp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}
