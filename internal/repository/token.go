package repository

import (
	"context"
	"fmt"

	"hw/internal/model"

	"github.com/jackc/pgx/v5"
)

// GetTokenByAddress retrieves a token by its address from the database.
func (r *repository) GetTokenByAddress(ctx context.Context, address string) (*model.Token, error) {
	const query = `
		SELECT id, name, symbol, decimals, created_at
		FROM tokens
		WHERE id = $1
	`

	token := &model.Token{}
	err := r.db.QueryRow(ctx, query, address).Scan(
		&token.ID,
		&token.Name,
		&token.Symbol,
		&token.Decimals,
		&token.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to retrieve token: %w", err)
	}

	return token, nil
}

// CreateToken inserts a new token into the database.
func (r *repository) CreateToken(ctx context.Context, token *model.Token) error {
	const query = `
		INSERT INTO tokens (id, name, symbol, decimals)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		token.ID,
		token.Name,
		token.Symbol,
		token.Decimals,
	).Scan(
		&token.ID,
		&token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create token: %s %w", token.ID, err)
	}

	return nil
}
