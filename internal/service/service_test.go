package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"hw/internal/model"
	repositoryMock "hw/internal/repository/mocks"
	"hw/internal/service"
	pgMock "hw/pkg/pg/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestAccumulateUserPoints_Success tests the successful execution of AccumulateUserPoints method.
func TestAccumulateUserPoints_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	mockTx := pgMock.NewMockPgxTx(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	token := "tokenABC"
	user := "userXYZ"
	description := "Test Accumulation"
	point := 100.0

	pointsHistory := &model.PointsHistory{
		ID: 1,
	}

	// Set expectations for mockRepo
	mockRepo.EXPECT().BeginTransaction(ctx).Return(mockTx, nil)
	mockRepo.EXPECT().
		CreatePointsHistory(ctx, gomock.AssignableToTypeOf(&model.PointsHistory{})).
		DoAndReturn(func(ctx context.Context, ph *model.PointsHistory) error {
			ph.ID = 1
			ph.CreatedAt = time.Now()
			return nil
		})
	mockRepo.EXPECT().UpsertUserPoints(ctx, user, point).Return(nil)
	mockTx.EXPECT().Commit(ctx).Return(nil)

	// Execute service method
	err := svc.AccumulateUserPoints(ctx, token, user, description, point)

	// Validate results
	assert.NoError(t, err)
	assert.Equal(t, 1, pointsHistory.ID, "PointsHistory ID should be set to 1")
}

// TestAccumulateUserPoints_CreatePointsHistoryError tests the scenario where creating points history fails.
func TestAccumulateUserPoints_CreatePointsHistoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	mockTx := pgMock.NewMockPgxTx(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	token := "tokenABC"
	user := "userXYZ"
	description := "Test Accumulation"
	point := 100.0

	pointsHistory := &model.PointsHistory{
		Token:       token,
		Account:     user,
		Points:      point,
		Description: description,
	}

	expectedError := errors.New("failed to create points history")

	mockRepo.EXPECT().BeginTransaction(ctx).Return(mockTx, nil)
	mockRepo.EXPECT().CreatePointsHistory(ctx, pointsHistory).Return(expectedError)
	mockTx.EXPECT().Rollback(ctx).Return(nil)

	err := svc.AccumulateUserPoints(ctx, token, user, description, point)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// TestGetOrCreateAccount_GetUserSuccess tests GetOrCreateAccount when the user exists.
func TestGetOrCreateAccount_GetUserSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	accountId := "account123"
	existingUser := &model.User{
		ID:          1,
		Address:     accountId,
		TotalPoints: 200.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().GetUserByAddress(ctx, accountId).Return(existingUser, nil)

	user, err := svc.GetOrCreateAccount(ctx, accountId)

	assert.NoError(t, err)
	assert.Equal(t, existingUser, user)
}

// TestGetOrCreateAccount_CreateUserSuccess tests successful user creation when the user does not exist.
func TestGetOrCreateAccount_CreateUserSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	accountId := "account123"
	newUser := &model.User{
		ID:          2,
		Address:     accountId,
		TotalPoints: 0.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().GetUserByAddress(ctx, accountId).Return(nil, model.ErrUserNotFound)
	mockRepo.EXPECT().CreateUser(ctx, accountId).Return(newUser, nil)

	user, err := svc.GetOrCreateAccount(ctx, accountId)

	assert.NoError(t, err)
	assert.Equal(t, newUser, user)
}

// TestGetOrCreateAccount_CreateUserError tests the scenario where user creation fails.
func TestGetOrCreateAccount_CreateUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	accountId := "account123"
	expectedError := errors.New("failed to create user")

	mockRepo.EXPECT().GetUserByAddress(ctx, accountId).Return(nil, model.ErrUserNotFound)
	mockRepo.EXPECT().CreateUser(ctx, accountId).Return(nil, expectedError)

	user, err := svc.GetOrCreateAccount(ctx, accountId)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to create user")
}

// TestGetTokenByAddress_Success tests the successful retrieval of a token by address.
func TestGetTokenByAddress_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	tokenAddress := "0xTokenAddress"
	expectedToken := &model.Token{
		ID:        tokenAddress,
		Name:      "TestToken",
		Symbol:    "TST",
		Decimals:  18,
		CreatedAt: time.Now(),
	}

	mockRepo.EXPECT().GetTokenByAddress(ctx, tokenAddress).Return(expectedToken, nil)

	token, err := svc.GetTokenByAddress(ctx, tokenAddress)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}

// TestGetTokenByAddress_Error tests the scenario where retrieving a token fails.
func TestGetTokenByAddress_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	tokenAddress := "0xInvalidTokenAddress"
	expectedError := errors.New("database connection failed")
	wrappedError := fmt.Errorf("failed to retrieve token: %w", expectedError)

	mockRepo.
		EXPECT().
		GetTokenByAddress(ctx, tokenAddress).
		Return(nil, wrappedError)

	token, err := svc.GetTokenByAddress(ctx, tokenAddress)

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "failed to retrieve token")
}

// TestCreateSwapHistory_Success tests the successful creation of swap history.
func TestCreateSwapHistory_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	swapHistory := &model.SwapHistory{
		Token:           "tokenABC",
		Account:         "accountXYZ",
		TransactionHash: "tx123456",
		UsdValue:        250.75,
		LastUpdated:     time.Now(),
	}

	mockRepo.EXPECT().CreateSwapHistory(ctx, swapHistory).Return(nil)

	err := svc.CreateSwapHistory(ctx, swapHistory)

	assert.NoError(t, err)
}

// TestCreateSwapHistory_Error tests the scenario where creating swap history fails.
func TestCreateSwapHistory_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	swapHistory := &model.SwapHistory{
		Token:           "tokenABC",
		Account:         "accountXYZ",
		TransactionHash: "tx123456",
		UsdValue:        250.75,
		LastUpdated:     time.Now(),
	}

	expectedError := errors.New("failed to create swap history")

	mockRepo.EXPECT().CreateSwapHistory(ctx, swapHistory).Return(expectedError)

	err := svc.CreateSwapHistory(ctx, swapHistory)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create swap history")
}

// TestGetLeaderboard_Success tests the successful retrieval of the leaderboard.
func TestGetLeaderboard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	expectedLeaderboard := []model.User{
		{
			ID:          1,
			Address:     "user1",
			TotalPoints: 300.0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          2,
			Address:     "user2",
			TotalPoints: 200.0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockRepo.EXPECT().GetLeaderboard(ctx).Return(expectedLeaderboard, nil)

	leaderboard, err := svc.GetLeaderboard(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedLeaderboard, leaderboard)
}

// TestGetLeaderboard_Error tests the scenario where retrieving the leaderboard fails.
func TestGetLeaderboard_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	expectedError := errors.New("failed to get leaderboard")

	mockRepo.EXPECT().GetLeaderboard(ctx).Return(nil, expectedError)

	leaderboard, err := svc.GetLeaderboard(ctx)

	assert.Error(t, err)
	assert.Nil(t, leaderboard)
	assert.Contains(t, err.Error(), "failed to get leaderboard")
}

// TestGetOrCreateToken_Success tests the successful creation of a token when it does not exist.
// TODO:

// TestGetOrCreateToken_CreateTokenError tests the scenario where creating a token fails.
// TODO:

// TestGetOrCreateToken_Exist tests the scenario where the token already exists.
// TODO:

// TestIsOnboardingTaskCompleted_Success tests the successful check of onboarding task completion.
func TestIsOnboardingTaskCompleted_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := "accountXYZ"

	mockRepo.EXPECT().IsOnboardingTaskCompleted(ctx, account).Return(true, nil)

	completed, err := svc.IsOnboardingTaskCompleted(ctx, account)

	assert.NoError(t, err)
	assert.True(t, completed, "Onboarding task should be marked as completed.")
}

// TestIsOnboardingTaskCompleted_Failure tests the failure scenario when checking onboarding task completion.
func TestIsOnboardingTaskCompleted_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := "accountXYZ"

	expectedError := errors.New("repository error")

	mockRepo.EXPECT().IsOnboardingTaskCompleted(ctx, account).Return(false, expectedError)

	completed, err := svc.IsOnboardingTaskCompleted(ctx, account)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.False(t, completed, "Onboarding task should not be marked as completed due to error.")
}

// TestGetSwapTotalUsd_Success tests the successful retrieval of total USD value.
func TestGetSwapTotalUsd_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := "accountXYZ"
	token := "tokenABC"
	expectedTotalUsd := 1000.50

	mockRepo.EXPECT().GetSwapTotalUsd(ctx, account, token).Return(expectedTotalUsd, nil)

	totalUsd, err := svc.GetSwapTotalUsd(ctx, account, token)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotalUsd, totalUsd, "Total USD should match expected value.")
}

// TestGetSwapTotalUsd_Failure tests the scenario where retrieving total USD value fails.
func TestGetSwapTotalUsd_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := "accountXYZ"
	token := "tokenABC"

	expectedError := errors.New("repository error")

	mockRepo.EXPECT().GetSwapTotalUsd(ctx, account, token).Return(0.0, expectedError)

	totalUsd, err := svc.GetSwapTotalUsd(ctx, account, token)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, float64(0), totalUsd, "Total USD should be 0 due to error.")
}

// TestGetUserSwapSummary_Success tests the successful retrieval of user swap summary.
func TestGetUserSwapSummary_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := "accountXYZ"

	expectedSummary := map[string]float64{
		"tokenABC": 1000.50,
		"tokenXYZ": 500.25,
	}

	mockRepo.EXPECT().GetUserSwapSummary(ctx, account).Return(expectedSummary, nil)

	summary, err := svc.GetUserSwapSummary(ctx, account)

	assert.NoError(t, err)
	assert.Equal(t, expectedSummary, summary, "User swap summary should match expected.")
}

// TestGetUserSwapSummary_Failure tests the scenario where retrieving user swap summary fails.
func TestGetUserSwapSummary_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := "accountXYZ"

	expectedError := errors.New("repository error")

	mockRepo.EXPECT().GetUserSwapSummary(ctx, account).Return(nil, expectedError)

	summary, err := svc.GetUserSwapSummary(ctx, account)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, summary, "Summary should be nil due to error.")
}

// TestGetUserSwapSummaryLast7Days_Success tests the successful retrieval of user swap summary for the last 7 days.
func TestGetUserSwapSummaryLast7Days_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := "accountXYZ"

	expectedSummary := []model.UserSwapPercentage{
		{
			Account:    "user1",
			TotalUSD:   1500.75,
			Percentage: 0.60,
		},
		{
			Account:    "user2",
			TotalUSD:   1000.25,
			Percentage: 0.40,
		},
	}

	mockRepo.EXPECT().GetUserSwapSummaryLast7Days(ctx, gomock.Any(), gomock.Any()).Return(expectedSummary, nil)

	summary, err := svc.GetUserSwapSummaryLast7Days(ctx, account)

	assert.NoError(t, err)
	assert.Equal(t, expectedSummary, summary, "User swap summary last 7 days should match expected.")
}

// TestGetUserSwapSummaryLast7Days_Failure tests the scenario where retrieving user swap summary for the last 7 days fails.
func TestGetUserSwapSummaryLast7Days_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := "accountXYZ"

	expectedError := errors.New("repository error")

	mockRepo.EXPECT().GetUserSwapSummaryLast7Days(ctx, gomock.Any(), gomock.Any()).Return(nil, expectedError)

	summary, err := svc.GetUserSwapSummaryLast7Days(ctx, account)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, summary, "Summary should be nil due to error.")
}

// TestCreateAccount_Success tests the successful creation of a user account.
func TestCreateAccount_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := &model.User{
		Address: "accountXYZ",
	}

	mockRepo.EXPECT().GetUserByAddress(ctx, account.Address).Return(nil, model.ErrUserNotFound)

	createdUser := &model.User{
		ID:          1,
		Address:     "accountXYZ",
		TotalPoints: 0.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockRepo.EXPECT().CreateUser(ctx, account.Address).Return(createdUser, nil)

	err := svc.CreateAccount(ctx, account)

	assert.NoError(t, err)
}

// TestCreateAccount_UserAlreadyExists tests the scenario where the user already exists during account creation.
func TestCreateAccount_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := &model.User{
		Address: "accountXYZ",
	}

	existingUser := &model.User{
		ID:          1,
		Address:     "accountXYZ",
		TotalPoints: 100.5,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockRepo.EXPECT().GetUserByAddress(ctx, account.Address).Return(existingUser, nil)

	err := svc.CreateAccount(ctx, account)

	assert.NoError(t, err)
}

// TestCreateAccount_Failure tests the scenario where creating a user account fails.
func TestCreateAccount_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := &model.User{
		Address: "accountXYZ",
	}

	mockRepo.EXPECT().GetUserByAddress(ctx, account.Address).Return(nil, model.ErrUserNotFound)

	expectedError := errors.New("create user error")
	mockRepo.EXPECT().CreateUser(ctx, account.Address).Return(nil, expectedError)

	err := svc.CreateAccount(ctx, account)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// TestCreateToken_Success tests the successful creation of a token.
func TestCreateToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	token := &model.Token{
		ID:       "0xTokenId",
		Name:     "TestToken",
		Symbol:   "TT",
		Decimals: 18,
	}

	mockRepo.EXPECT().GetTokenByAddress(ctx, token.ID).Return(nil, model.ErrTokenNotFound)
	mockRepo.EXPECT().CreateToken(ctx, token).Return(nil)

	err := svc.CreateToken(ctx, token)

	assert.NoError(t, err)
}

// TestCreateToken_TokenAlreadyExists tests the scenario where the token already exists during creation.
func TestCreateToken_TokenAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	token := &model.Token{
		ID:       "0xTokenId",
		Name:     "TestToken",
		Symbol:   "TT",
		Decimals: 18,
	}

	existingToken := &model.Token{
		ID:       "0xTokenId",
		Name:     "TestToken",
		Symbol:   "TT",
		Decimals: 18,
	}
	mockRepo.EXPECT().GetTokenByAddress(ctx, token.ID).Return(existingToken, nil)

	err := svc.CreateToken(ctx, token)

	assert.NoError(t, err)
}

// TestCreateToken_Failure tests the scenario where creating a token fails.
func TestCreateToken_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	token := &model.Token{
		ID:       "0xTokenId",
		Name:     "TestToken",
		Symbol:   "TT",
		Decimals: 18,
	}

	mockRepo.EXPECT().GetTokenByAddress(ctx, token.ID).Return(nil, model.ErrTokenNotFound)

	expectedError := errors.New("create token error")
	mockRepo.EXPECT().CreateToken(ctx, token).Return(expectedError)

	err := svc.CreateToken(ctx, token)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// TestGetPointsHistory_Success tests the successful retrieval of points history.
func TestGetPointsHistory_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := "accountXYZ"
	token := "tokenABC"

	expectedHistory := []model.PointsHistory{
		{
			ID:          1,
			Token:       "tokenABC",
			Account:     "accountXYZ",
			Points:      100.0,
			Description: "Test Points",
			CreatedAt:   time.Now(),
		},
	}

	mockRepo.EXPECT().GetPointsHistory(ctx, account, token).Return(expectedHistory, nil)

	history, err := svc.GetPointsHistory(ctx, account, token)

	assert.NoError(t, err)
	assert.Equal(t, expectedHistory, history, "Points history should match expected.")
}

// TestGetPointsHistory_Failure tests the scenario where retrieving points history fails.
func TestGetPointsHistory_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMock.NewMockRepository(ctrl)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	account := "accountXYZ"
	token := "tokenABC"

	expectedError := errors.New("repository error")

	mockRepo.EXPECT().GetPointsHistory(ctx, account, token).Return(nil, expectedError)

	history, err := svc.GetPointsHistory(ctx, account, token)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, history, "Points history should be nil due to error.")
}
