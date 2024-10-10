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

// TestCreateUser_Success tests the successful creation of a user.
func TestCreateUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	userID := "0x1234567890123456789012345678901234567890"

	const query = `
		INSERT INTO users (address)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`

	mockDB.EXPECT().QueryRow(ctx, query, userID).Return(mockRow)

	expectedUser := &model.User{
		ID:        1,
		Address:   userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRow.EXPECT().Scan(
		gomock.AssignableToTypeOf(&expectedUser.ID),
		gomock.AssignableToTypeOf(&expectedUser.CreatedAt),
		gomock.AssignableToTypeOf(&expectedUser.UpdatedAt),
	).DoAndReturn(func(dest ...interface{}) error {
		*(dest[0].(*int)) = expectedUser.ID
		*(dest[1].(*time.Time)) = expectedUser.CreatedAt
		*(dest[2].(*time.Time)) = expectedUser.UpdatedAt
		return nil
	})

	user, err := repo.CreateUser(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

// TestCreateUser_Error tests the failure to create a user.
func TestCreateUser_Error(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	userId := "0x1234567890123456789012345678901234567890"

	const query = `
		INSERT INTO users (address)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`

	mockDB.EXPECT().QueryRow(ctx, query, userId).Return(mockRow)

	expectedError := errors.New("database error")
	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedError)

	user, err := repo.CreateUser(ctx, userId)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to create user")
}

// TestGetUserByAddress_Success verifies successful retrieval of a user by address.
func TestGetUserByAddress_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	address := "0x1234567890123456789012345678901234567890"

	const query = `
		SELECT id, address, total_points, created_at, updated_at
		FROM users
		WHERE address = $1
		LIMIT 1
	`

	mockDB.EXPECT().QueryRow(ctx, query, address).Return(mockRow)

	expectedUser := &model.User{
		ID:          1,
		Address:     address,
		TotalPoints: 100.5,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRow.EXPECT().Scan(
		gomock.AssignableToTypeOf(&expectedUser.ID),
		gomock.AssignableToTypeOf(&expectedUser.Address),
		gomock.AssignableToTypeOf(&expectedUser.TotalPoints),
		gomock.AssignableToTypeOf(&expectedUser.CreatedAt),
		gomock.AssignableToTypeOf(&expectedUser.UpdatedAt),
	).DoAndReturn(func(dest ...interface{}) error {
		*(dest[0].(*int)) = expectedUser.ID
		*(dest[1].(*string)) = expectedUser.Address
		*(dest[2].(*float64)) = expectedUser.TotalPoints
		*(dest[3].(*time.Time)) = expectedUser.CreatedAt
		*(dest[4].(*time.Time)) = expectedUser.UpdatedAt
		return nil
	})

	user, err := repo.GetUserByAddress(ctx, address)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

// TestGetUserByAddress_NotFound verifies behavior when a user is not found by address.
func TestGetUserByAddress_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	address := "0x1234567890123456789012345678901234567890"

	const query = `
		SELECT id, address, total_points, created_at, updated_at
		FROM users
		WHERE address = $1
		LIMIT 1
	`

	mockDB.EXPECT().QueryRow(ctx, query, address).Return(mockRow)
	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(pgx.ErrNoRows)

	user, err := repo.GetUserByAddress(ctx, address)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, model.ErrUserNotFound, err)
}

// TestUpsertUserPoints_Success verifies successful upsertion of user points.
func TestUpsertUserPoints_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRow := pgMock.NewMockPgxRows(ctrl)
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()
	address := "0x1234567890123456789012345678901234567890"
	points := 50.5

	const query = `
		INSERT INTO users (address, total_points)
		VALUES ($1, $2)
		ON CONFLICT (address) DO UPDATE SET 
			total_points = users.total_points + EXCLUDED.total_points,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`

	mockDB.EXPECT().QueryRow(ctx, query, address, points).Return(mockRow)

	mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	err := repo.UpsertUserPoints(ctx, address, points)

	assert.NoError(t, err)
}

// TestGetLeaderboard_Success verifies successful retrieval of the leaderboard.
func TestGetLeaderboard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRows := pgMock.NewMockPgxRows(ctrl)
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()

	expectedQuery := `
		SELECT id, address, total_points, created_at, updated_at
		FROM users
		ORDER BY total_points DESC
	`

	mockDB.EXPECT().
		Query(ctx, expectedQuery).
		Return(mockRows, nil)

	usersData := []model.User{
		{
			ID:          1,
			Address:     "address1",
			TotalPoints: 100.0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          2,
			Address:     "address2",
			TotalPoints: 90.0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	gomock.InOrder(
		mockRows.EXPECT().Next().Return(true),
		mockRows.EXPECT().Scan(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).DoAndReturn(func(dest ...any) error {
			*(dest[0].(*int)) = usersData[0].ID
			*(dest[1].(*string)) = usersData[0].Address
			*(dest[2].(*float64)) = usersData[0].TotalPoints
			*(dest[3].(*time.Time)) = usersData[0].CreatedAt
			*(dest[4].(*time.Time)) = usersData[0].UpdatedAt
			return nil
		}),

		mockRows.EXPECT().Next().Return(true),
		mockRows.EXPECT().Scan(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).DoAndReturn(func(dest ...any) error {
			*(dest[0].(*int)) = usersData[1].ID
			*(dest[1].(*string)) = usersData[1].Address
			*(dest[2].(*float64)) = usersData[1].TotalPoints
			*(dest[3].(*time.Time)) = usersData[1].CreatedAt
			*(dest[4].(*time.Time)) = usersData[1].UpdatedAt
			return nil
		}),

		mockRows.EXPECT().Next().Return(false),
		mockRows.EXPECT().Err().Return(nil),
		mockRows.EXPECT().Close(),
	)

	result, err := repo.GetLeaderboard(ctx)

	assert.NoError(t, err)
	assert.Equal(t, usersData, result)
}

// TestGetLeaderboard_EmptyResult verifies behavior when the leaderboard is empty.
func TestGetLeaderboard_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRows := pgMock.NewMockPgxRows(ctrl)
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()

	expectedQuery := `
		SELECT id, address, total_points, created_at, updated_at
		FROM users
		ORDER BY total_points DESC
	`

	mockDB.EXPECT().Query(ctx, expectedQuery).Return(mockRows, nil)

	mockRows.EXPECT().Next().Return(false)
	mockRows.EXPECT().Err().Return(nil)
	mockRows.EXPECT().Close()

	result, err := repo.GetLeaderboard(ctx)

	assert.NoError(t, err)
	assert.Empty(t, result, "Result should be an empty slice")
}

// TestGetLeaderboard_QueryError verifies behavior when the query fails.
func TestGetLeaderboard_QueryError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()

	expectedQuery := `
		SELECT id, address, total_points, created_at, updated_at
		FROM users
		ORDER BY total_points DESC
	`

	expectedError := errors.New("database query error")
	mockDB.EXPECT().Query(ctx, expectedQuery).Return(nil, expectedError)

	result, err := repo.GetLeaderboard(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get leaderboard")
}

// TestGetLeaderboard_ScanError verifies behavior when scanning a row fails.
func TestGetLeaderboard_ScanError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRows := pgMock.NewMockPgxRows(ctrl)
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()

	expectedQuery := `
		SELECT id, address, total_points, created_at, updated_at
		FROM users
		ORDER BY total_points DESC
	`

	mockDB.EXPECT().Query(ctx, expectedQuery).Return(mockRows, nil)

	mockRows.EXPECT().Next().Return(true)
	scanError := errors.New("scan error")
	mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(scanError)
	mockRows.EXPECT().Close()

	result, err := repo.GetLeaderboard(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to scan user")
}

// TestGetLeaderboard_RowsError verifies behavior when an error occurs during row iteration.
func TestGetLeaderboard_RowsError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockRows := pgMock.NewMockPgxRows(ctrl)
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()

	expectedQuery := `
		SELECT id, address, total_points, created_at, updated_at
		FROM users
		ORDER BY total_points DESC
	`

	mockDB.EXPECT().Query(ctx, expectedQuery).Return(mockRows, nil)

	mockRows.EXPECT().Next().Return(false)
	rowsError := errors.New("rows error")
	mockRows.EXPECT().Err().Return(rowsError)
	mockRows.EXPECT().Close()

	result, err := repo.GetLeaderboard(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error iterating rows")
}
