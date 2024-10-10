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

// TestCreatePointsHistory_Success tests the successful creation of points history.
func TestCreatePointsHistory_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	pointsHistory := &model.PointsHistory{
		Token:       "token123",
		Account:     "account123",
		Points:      100.5,
		Description: "Test description",
	}

	mockDB.EXPECT().QueryRow(
		ctx,
		gomock.Any(),
		pointsHistory.Token,
		pointsHistory.Account,
		pointsHistory.Points,
		pointsHistory.Description,
	).Return(mockRow)

	expectedID := 1
	expectedCreatedAt := time.Now()
	mockRow.EXPECT().Scan(
		gomock.AssignableToTypeOf(&pointsHistory.ID),
		gomock.AssignableToTypeOf(&pointsHistory.CreatedAt),
	).DoAndReturn(func(dest ...any) error {
		*(dest[0].(*int)) = expectedID
		*(dest[1].(*time.Time)) = expectedCreatedAt
		return nil
	})

	err := repo.CreatePointsHistory(ctx, pointsHistory)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, pointsHistory.ID)
	assert.Equal(t, expectedCreatedAt, pointsHistory.CreatedAt)
}

// TestIsOnboardingTaskCompleted_Success tests the scenario where the onboarding task is completed.
func TestIsOnboardingTaskCompleted_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	account := "account123"
	description := "onboarding_task"
	mockDB.EXPECT().QueryRow(
		ctx,
		gomock.Any(),
		account,
		description,
	).Return(mockRow)

	var count int
	mockRow.EXPECT().Scan(
		gomock.AssignableToTypeOf(&count),
	).DoAndReturn(func(dest ...any) error {
		*(dest[0].(*int)) = 1
		return nil
	})

	completed, err := repo.IsOnboardingTaskCompleted(ctx, account)

	assert.NoError(t, err)
	assert.True(t, completed)
}

// TestIsOnboardingTaskCompleted_NotFound tests the scenario where the onboarding task is not found.
func TestIsOnboardingTaskCompleted_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	account := "account123"
	description := "onboarding_task"
	mockDB.EXPECT().QueryRow(
		ctx,
		gomock.Any(),
		account,
		description,
	).Return(mockRow)

	var count int
	mockRow.EXPECT().Scan(
		gomock.AssignableToTypeOf(&count),
	).DoAndReturn(func(dest ...any) error {
		*(dest[0].(*int)) = 0
		return nil
	})

	completed, err := repo.IsOnboardingTaskCompleted(ctx, account)

	assert.NoError(t, err)
	assert.False(t, completed)
}

// TestIsOnboardingTaskCompleted_QueryError tests the scenario where there is a query error.
func TestIsOnboardingTaskCompleted_QueryError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	account := "account123"

	// Define a custom error message
	expectedErr := errors.New("custom query error message")

	// Mock QueryRow method to return error
	mockDB.EXPECT().QueryRow(
		ctx,
		gomock.Any(),
		account,
		"onboarding_task",
	).Return(nil).DoAndReturn(func(ctx context.Context, query string, args ...interface{}) *pgMock.MockPgxRows {
		// When QueryRow returns an error, Scan method should return the error
		mockRow := pgMock.NewMockPgxRows(ctrl)
		mockRow.EXPECT().Scan(gomock.Any()).Return(expectedErr)
		return mockRow
	})

	// Call the method under test
	completed, err := repo.IsOnboardingTaskCompleted(ctx, account)

	// Assert the results
	assert.Error(t, err)
	assert.False(t, completed)
	assert.Contains(t, err.Error(), "failed to retrieve points history records:")
	assert.Contains(t, err.Error(), expectedErr.Error())
}

// TestGetPointsHistory_Success tests the successful retrieval of points history.
func TestGetPointsHistory_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRows := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	account := "account123"
	token := "token123"

	// Set expected database behavior
	mockDB.EXPECT().Query(ctx, gomock.Any(), account, token).Return(mockRows, nil)

	// Simulate row data
	firstCall := mockRows.EXPECT().Next().Return(true)
	secondCall := mockRows.EXPECT().Next().Return(false)
	gomock.InOrder(firstCall, secondCall)

	expectedPH := model.PointsHistory{
		ID:          1,
		Token:       token,
		Account:     account,
		Points:      100.5,
		Description: "Test description",
		CreatedAt:   time.Now(),
	}

	mockRows.EXPECT().Scan(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).DoAndReturn(func(dest ...any) error {
		*(dest[0].(*int)) = expectedPH.ID
		*(dest[1].(*string)) = expectedPH.Token
		*(dest[2].(*string)) = expectedPH.Account
		*(dest[3].(*float64)) = expectedPH.Points
		*(dest[4].(*string)) = expectedPH.Description
		*(dest[5].(*time.Time)) = expectedPH.CreatedAt
		return nil
	})

	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	// Call the method under test
	histories, err := repo.GetPointsHistory(ctx, account, token)

	// Assert the results
	assert.NoError(t, err)
	assert.Len(t, histories, 1)
	assert.Equal(t, expectedPH, histories[0])
}

// TestGetPointsHistory_QueryError tests the scenario where there is a query error.
func TestGetPointsHistory_QueryError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	account := "account123"
	token := "token123"
	expectedErr := errors.New("query error")

	// Mock Query method to return error
	mockDB.EXPECT().Query(ctx, gomock.Any(), account, token).Return(nil, expectedErr)

	// Call the method under test
	histories, err := repo.GetPointsHistory(ctx, account, token)

	// Assert the results
	assert.Error(t, err)
	assert.Nil(t, histories)
	assert.Contains(t, err.Error(), "failed to query points history")
	assert.Contains(t, err.Error(), expectedErr.Error())
}

// TestGetPointsHistory_ScanError tests the scenario where there is a scan error.
func TestGetPointsHistory_ScanError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRows := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	account := "account123"
	token := "token123"
	expectedErr := errors.New("scan error")

	mockDB.EXPECT().Query(ctx, gomock.Any(), account, token).Return(mockRows, nil)

	mockRows.EXPECT().Next().Return(true)
	mockRows.EXPECT().Scan(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(expectedErr)
	mockRows.EXPECT().Close()

	// Call the method under test
	histories, err := repo.GetPointsHistory(ctx, account, token)

	// Assert the results
	assert.Error(t, err)
	assert.Nil(t, histories)
	assert.Contains(t, err.Error(), "failed to scan points history row")
	assert.Contains(t, err.Error(), expectedErr.Error())
}

// TestGetPointsHistory_RowsError tests the scenario where there is a rows error.
func TestGetPointsHistory_RowsError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRows := pgMock.NewMockPgxRows(ctrl)

	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	account := "account123"
	token := "token123"
	expectedErr := errors.New("rows error")

	mockDB.EXPECT().Query(ctx, gomock.Any(), account, token).Return(mockRows, nil)

	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(expectedErr)
	mockRows.EXPECT().Close()

	// Call the method under test
	histories, err := repo.GetPointsHistory(ctx, account, token)

	// Assert the results
	assert.Error(t, err)
	assert.Nil(t, histories)
	assert.Contains(t, err.Error(), "failed to iterate through points history rows")
	assert.Contains(t, err.Error(), expectedErr.Error())
}
