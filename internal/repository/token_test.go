package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"hw/internal/model"
	"hw/internal/repository"
	pgMock "hw/pkg/pg/mocks"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestGetTokenByAddress_Success tests successfully retrieving a token by address.
func TestGetTokenByAddress_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	address := "0x1234567890123456789012345678901234567890"

	expectedToken := &model.Token{
		ID:        address,
		Name:      "Test Token",
		Symbol:    "TST",
		Decimals:  18,
		CreatedAt: time.Now(),
	}

	const query = `
		SELECT id, name, symbol, decimals, created_at
		FROM tokens
		WHERE id = $1
	`

	mockDB.EXPECT().QueryRow(ctx, query, address).Return(mockRow)

	mockRow.EXPECT().Scan(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).DoAndReturn(func(dest ...interface{}) error {
		*(dest[0].(*string)) = expectedToken.ID
		*(dest[1].(*string)) = expectedToken.Name
		*(dest[2].(*string)) = expectedToken.Symbol
		*(dest[3].(*int64)) = expectedToken.Decimals
		*(dest[4].(*time.Time)) = expectedToken.CreatedAt
		return nil
	})

	token, err := repo.GetTokenByAddress(ctx, address)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}

// TestGetTokenByAddress_NotFound tests the scenario where a token is not found.
func TestGetTokenByAddress_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	address := "0x1234567890123456789012345678901234567890"

	const query = `
		SELECT id, name, symbol, decimals, created_at
		FROM tokens
		WHERE id = $1
	`

	mockDB.EXPECT().QueryRow(ctx, query, address).Return(mockRow)

	mockRow.EXPECT().Scan(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(pgx.ErrNoRows)

	token, err := repo.GetTokenByAddress(ctx, address)

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Equal(t, model.ErrTokenNotFound, err)
}

// TestCreateToken_Success tests successfully creating a token.
func TestCreateToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	token := &model.Token{
		ID:       "0x1234567890123456789012345678901234567890",
		Name:     "Test Token",
		Symbol:   "TST",
		Decimals: 18,
	}

	const query = `
		INSERT INTO tokens (id, name, symbol, decimals)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	mockDB.EXPECT().QueryRow(
		ctx,
		query,
		token.ID,
		token.Name,
		token.Symbol,
		token.Decimals,
	).Return(mockRow)

	expectedCreatedAt := time.Now()
	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(dest ...interface{}) error {
		*(dest[0].(*string)) = token.ID
		*(dest[1].(*time.Time)) = expectedCreatedAt
		return nil
	})

	err := repo.CreateToken(ctx, token)

	assert.NoError(t, err)
	assert.Equal(t, expectedCreatedAt, token.CreatedAt)
}

// TestCreateToken_Failure tests the failure scenario of creating a token.
func TestCreateToken_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	token := &model.Token{
		ID:       "0x1234567890123456789012345678901234567890",
		Name:     "Test Token",
		Symbol:   "TST",
		Decimals: 18,
	}

	const query = `
		INSERT INTO tokens (id, name, symbol, decimals)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	mockDB.EXPECT().QueryRow(
		ctx,
		query,
		token.ID,
		token.Name,
		token.Symbol,
		token.Decimals,
	).Return(mockRow)

	expectedError := errors.New("insertion error")
	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(expectedError)

	err := repo.CreateToken(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create token")
	assert.Contains(t, err.Error(), expectedError.Error())
}
