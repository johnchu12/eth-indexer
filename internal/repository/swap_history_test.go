package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"hw/internal/model"
	"hw/internal/repository"
	pgMock "hw/pkg/pg/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestCreateSwapHistory_Success tests the successful creation of SwapHistory.
func TestCreateSwapHistory_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	swapHistory := &model.SwapHistory{
		Token:           "tokenABC",
		Account:         "accountXYZ",
		TransactionHash: "tx123456",
		UsdValue:        250.75,
		LastUpdated:     time.Now(),
	}

	const query = `
		INSERT INTO swap_history (token, account, transaction_hash, usd_value, last_updated)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	mockDB.EXPECT().QueryRow(
		ctx,
		query,
		swapHistory.Token,
		swapHistory.Account,
		swapHistory.TransactionHash,
		swapHistory.UsdValue,
		swapHistory.LastUpdated,
	).Return(mockRow)

	expectedID := 1
	expectedCreatedAt := time.Now()
	mockRow.EXPECT().Scan(
		gomock.AssignableToTypeOf(&swapHistory.ID),
		gomock.AssignableToTypeOf(&swapHistory.CreatedAt),
	).DoAndReturn(func(dest ...any) error {
		*(dest[0].(*int)) = expectedID
		*(dest[1].(*time.Time)) = expectedCreatedAt
		return nil
	})

	err := repo.CreateSwapHistory(ctx, swapHistory)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, swapHistory.ID)
	assert.WithinDuration(t, expectedCreatedAt, swapHistory.CreatedAt, time.Second)
}

// TestCreateSwapHistory_Failure tests the failure scenario when creating SwapHistory.
func TestCreateSwapHistory_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	swapHistory := &model.SwapHistory{
		Token:           "tokenABC",
		Account:         "accountXYZ",
		TransactionHash: "tx123456",
		UsdValue:        250.75,
		LastUpdated:     time.Now(),
	}

	const query = `
		INSERT INTO swap_history (token, account, transaction_hash, usd_value, last_updated)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	mockDB.EXPECT().QueryRow(
		ctx,
		query,
		swapHistory.Token,
		swapHistory.Account,
		swapHistory.TransactionHash,
		swapHistory.UsdValue,
		swapHistory.LastUpdated,
	).Return(nil).DoAndReturn(func(ctx context.Context, query string, args ...interface{}) *pgMock.MockPgxRows {
		mockRow := pgMock.NewMockPgxRows(ctrl)
		mockRow.EXPECT().Scan(&swapHistory.ID, &swapHistory.CreatedAt).Return(errors.New("insert error"))
		return mockRow
	})

	err := repo.CreateSwapHistory(ctx, swapHistory)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create swap history")
}

// TestGetSwapTotalUsd_Success tests the successful retrieval of total USD value.
func TestGetSwapTotalUsd_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	account := "accountXYZ"
	token := "tokenABC"
	expectedTotalUsd := 1000.50

	const query = `
		SELECT SUM(usd_value)
		FROM swap_history
		WHERE account = $1 AND token = $2
	`

	mockDB.EXPECT().QueryRow(ctx, query, account, token).Return(mockRow)

	mockRow.EXPECT().Scan(gomock.AssignableToTypeOf(&expectedTotalUsd)).DoAndReturn(func(dest ...any) error {
		*(dest[0].(*float64)) = expectedTotalUsd
		return nil
	})

	totalUsd, err := repo.GetSwapTotalUsd(ctx, account, token)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotalUsd, totalUsd)
}

// TestGetSwapTotalUsd_Failure tests the failure scenario when retrieving total USD value.
func TestGetSwapTotalUsd_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	account := "accountXYZ"
	token := "tokenABC"

	const query = `
		SELECT SUM(usd_value)
		FROM swap_history
		WHERE account = $1 AND token = $2
	`

	mockDB.EXPECT().QueryRow(ctx, query, account, token).Return(mockRow)

	mockRow.EXPECT().Scan(gomock.Any()).Return(errors.New("scan error"))

	totalUsd, err := repo.GetSwapTotalUsd(ctx, account, token)

	assert.Error(t, err)
	assert.Equal(t, float64(0), totalUsd)
	assert.Contains(t, err.Error(), "failed to get total swap USD")
}

// TestGetUserSwapSummary_Success tests the successful retrieval of user swap summary.
func TestGetUserSwapSummary_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRows := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	account := "accountXYZ"

	const query = `
		SELECT token, SUM(usd_value)
		FROM swap_history
		WHERE account = $1
		GROUP BY token
	`

	mockDB.EXPECT().Query(ctx, query, account).Return(mockRows, nil)

	mockRows.EXPECT().Next().Return(true)
	mockRows.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(dest ...interface{}) error {
		*(dest[0].(*string)) = "tokenABC"
		*(dest[1].(*float64)) = 1000.50
		return nil
	})
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	summary, err := repo.GetUserSwapSummary(ctx, account)

	assert.NoError(t, err)
	assert.Len(t, summary, 1)
	assert.Equal(t, 1000.50, summary["tokenABC"])
}

// TestGetUserSwapSummary_Failure tests the failure scenario when retrieving user swap summary.
func TestGetUserSwapSummary_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	account := "accountXYZ"

	const query = `
		SELECT token, SUM(usd_value)
		FROM swap_history
		WHERE account = $1
		GROUP BY token
	`

	mockDB.EXPECT().Query(ctx, query, account).Return(nil, errors.New("query error"))

	summary, err := repo.GetUserSwapSummary(ctx, account)

	assert.Error(t, err)
	assert.Nil(t, summary)
	assert.Contains(t, err.Error(), "failed to retrieve token USD sums")
}

// TestGetUserSwapSummaryLast7Days_Success tests the successful retrieval of user swap summary for the last 7 days.
func TestGetUserSwapSummaryLast7Days_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRows := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	referenceTime := time.Now()
	token := "tokenABC"

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

	mockDB.EXPECT().Query(ctx, query, startTime, endTime, token).Return(mockRows, nil)

	mockRows.EXPECT().Next().Return(true)
	mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(dest ...interface{}) error {
		*(dest[0].(*string)) = "accountXYZ"
		*(dest[1].(*float64)) = 1000.50
		*(dest[2].(*float64)) = 0.75
		return nil
	})
	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	summary, err := repo.GetUserSwapSummaryLast7Days(ctx, referenceTime, token)

	assert.NoError(t, err)
	assert.Len(t, summary, 1)
	assert.Equal(t, "accountXYZ", summary[0].Account)
	assert.Equal(t, 1000.50, summary[0].TotalUSD)
	assert.Equal(t, 0.75, summary[0].Percentage)
}

// TestGetUserSwapSummaryLast7Days_Failure tests the failure scenario when retrieving user swap summary for the last 7 days.
func TestGetUserSwapSummaryLast7Days_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	referenceTime := time.Now()
	token := "tokenABC"

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

	mockDB.EXPECT().Query(ctx, query, startTime, endTime, token).Return(nil, errors.New("query error"))

	summary, err := repo.GetUserSwapSummaryLast7Days(ctx, referenceTime, token)

	assert.Error(t, err)
	assert.Nil(t, summary)
	assert.Contains(t, err.Error(), "failed to retrieve user swap percentages")
}
