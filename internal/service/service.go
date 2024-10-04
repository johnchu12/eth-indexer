package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hw/internal/model"
	"hw/internal/repository"
	"hw/pkg/ethindexa/utils"
	"hw/pkg/logger"

	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/sync/singleflight"
)

// Service defines the interface for the service layer.
type Service interface {
	AccumulateUserPoints(ctx context.Context, token, user, description string, point float64) error
	IsOnboardingTaskCompleted(ctx context.Context, account string) (bool, error)
	GetOrCreateAccount(ctx context.Context, accountId string) (*model.User, error)
	GetTokenByAddress(ctx context.Context, token string) (*model.Token, error)
	CreateSwapHistory(ctx context.Context, history *model.SwapHistory) error
	GetSwapTotalUsd(ctx context.Context, account, token string) (float64, error)
	GetUserSwapSummary(ctx context.Context, account string) (map[string]float64, error)
	GetUserSwapSummaryLast7Days(ctx context.Context, account string) ([]model.UserSwapPercentage, error)
	CreateToken(ctx context.Context, token *model.Token) error
	GetOrCreateToken(ctx context.Context, client *ethclient.Client, tokenId string, blockNumber int64) (*model.Token, error)
	CreateAccount(ctx context.Context, account *model.User) error
	GetPointsHistory(ctx context.Context, account, token string) ([]model.PointsHistory, error)
	GetLeaderboard(ctx context.Context) ([]model.User, error)
}

type service struct {
	group singleflight.Group
	repo  repository.Repository
}

// NewService creates a new instance of Service.
func NewService(repo repository.Repository) Service {
	return &service{repo: repo, group: singleflight.Group{}}
}

// GetLeaderboard retrieves the leaderboard data and returns it as JSON.
func (s *service) GetLeaderboard(ctx context.Context) ([]model.User, error) {
	return s.repo.GetLeaderboard(s.repo.DB(), ctx)
}

// AccumulateUserPoints adds points to a user's account with a description.
func (s *service) AccumulateUserPoints(ctx context.Context, token, user, description string, point float64) error {
	_, err, _ := s.group.Do(user, func() (interface{}, error) {
		// 開始交易
		tx, err := s.repo.BeginTransaction(ctx)
		if err != nil {
			return nil, err
		}

		// 使用閉包處理提交和回滾
		err = func() error {
			// 創建積分歷史紀錄
			pointsHistory := &model.PointsHistory{
				Token:       token,
				Account:     user,
				Points:      point,
				Description: description,
			}

			if err := s.repo.CreatePointsHistory(tx, ctx, pointsHistory); err != nil {
				return err
			}

			// 如果因為衝突而未創建積分歷史紀錄，則跳過更新使用者點數
			if pointsHistory.ID == 0 {
				return nil
			}

			// 原子性地更新使用者的總點數
			if err := s.repo.UpsertUserPoints(tx, ctx, user, point); err != nil {
				return err
			}

			return nil
		}()
		if err != nil {
			tx.Rollback(ctx)
			return nil, err
		}

		// 提交交易
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

		return nil, nil
	})
	return err
}

// GetOrCreateAccount retrieves an existing user or creates a new one if not found.
func (s *service) GetOrCreateAccount(ctx context.Context, accountId string) (*model.User, error) {
	// singleflight is used to ensure that concurrent requests for the same accountId result in a single database query or creation.
	v, err, _ := s.group.Do(accountId, func() (interface{}, error) {
		// Attempt to get the user first
		user, err := s.repo.GetUserByAddress(s.repo.DB(), ctx, accountId)
		if err == nil {
			return user, nil
		}

		// If user does not exist, create a new user
		if errors.Is(err, model.ErrUserNotFound) {
			newUser, err := s.repo.CreateUser(s.repo.DB(), ctx, accountId)
			if err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			return newUser, nil
		}

		// Return other errors directly
		return nil, fmt.Errorf("failed to get or create user: %w", err)
	})

	if err != nil {
		return nil, err
	}

	return v.(*model.User), nil
}

// GetOrCreateToken retrieves an existing token or creates a new one if not found.
func (s *service) GetOrCreateToken(ctx context.Context, client *ethclient.Client, tokenId string, blockNumber int64) (*model.Token, error) {
	// singleflight is utilized here to prevent multiple concurrent requests from fetching or creating the same token simultaneously.
	v, err, _ := s.group.Do(tokenId, func() (interface{}, error) {
		// Try to get the token from the database
		token, err := s.repo.GetTokenByAddress(s.repo.DB(), ctx, tokenId)
		if err == nil {
			return token, nil
		}
		if !errors.Is(err, model.ErrTokenNotFound) {
			return nil, fmt.Errorf("failed to retrieve token %s from DB: %w", tokenId, err)
		}

		// Fetch token information from external source
		tokenInfo, err := utils.GetTokenInfo(ctx, client, tokenId, blockNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch token %s info: %w", tokenId, err)
		}

		// Construct a new token model
		newToken := &model.Token{
			ID:       tokenId,
			Name:     tokenInfo.Name,
			Symbol:   tokenInfo.Symbol,
			Decimals: tokenInfo.Decimals,
		}

		// Begin transaction to ensure atomic database operations
		tx, err := s.repo.BeginTransaction(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}

		defer func() {
			if p := recover(); p != nil {
				tx.Rollback(ctx)
				panic(p) // Re-throw panic after rollback
			}
		}()

		// Save the new token to the database
		if err := s.repo.CreateToken(tx, ctx, newToken); err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to create token %s in DB: %w", tokenId, err)
		}

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return newToken, nil
	})

	if err != nil {
		return nil, err
	}

	return v.(*model.Token), nil
}

// GetTokenByAddress retrieves a token by its address.
func (s *service) GetTokenByAddress(ctx context.Context, token string) (*model.Token, error) {
	return s.repo.GetTokenByAddress(s.repo.DB(), ctx, token)
}

// CreateSwapHistory records a new swap history entry.
func (s *service) CreateSwapHistory(ctx context.Context, history *model.SwapHistory) error {
	return s.repo.CreateSwapHistory(s.repo.DB(), ctx, history)
}

// IsOnboardingTaskCompleted checks if the onboarding task is completed for an account.
func (s *service) IsOnboardingTaskCompleted(ctx context.Context, account string) (bool, error) {
	return s.repo.IsOnboardingTaskCompleted(s.repo.DB(), ctx, account)
}

// GetSwapTotalUsd calculates the total USD value of swaps for an account and token.
func (s *service) GetSwapTotalUsd(ctx context.Context, account, token string) (float64, error) {
	return s.repo.GetSwapTotalUsd(s.repo.DB(), ctx, account, token)
}

// GetUserSwapSummary provides a summary of user swaps.
func (s *service) GetUserSwapSummary(ctx context.Context, account string) (map[string]float64, error) {
	return s.repo.GetUserSwapSummary(s.repo.DB(), ctx, account)
}

// GetUserSwapSummaryLast7Days retrieves the total USD and percentage of swaps for each user over the past seven days for a specific token.
func (s *service) GetUserSwapSummaryLast7Days(ctx context.Context, account string) ([]model.UserSwapPercentage, error) {
	return s.repo.GetUserSwapSummaryLast7Days(s.repo.DB(), ctx, time.Now(), account)
}

func (s *service) GetPointsHistory(ctx context.Context, account, token string) ([]model.PointsHistory, error) {
	return s.repo.GetPointsHistory(s.repo.DB(), ctx, account, token)
}

// CreateAccount creates a new user account if it does not already exist.
func (s *service) CreateAccount(ctx context.Context, account *model.User) error {
	if _, err := s.repo.GetUserByAddress(s.repo.DB(), ctx, account.Address); err != nil {
		logger.Errorf("GetUserByAddress error: %+v", err)
		if !errors.Is(err, model.ErrUserNotFound) {
			return err
		}
		logger.Infow("User not found, creating user")
	}
	_, err := s.repo.CreateUser(s.repo.DB(), ctx, account.Address)
	return err
}

// CreateToken creates a new token if it does not already exist.
func (s *service) CreateToken(ctx context.Context, token *model.Token) error {
	if _, err := s.repo.GetTokenByAddress(s.repo.DB(), ctx, token.ID); err != nil {
		logger.Errorf("GetTokenByAddress error: %+v", err)
		if !errors.Is(err, model.ErrTokenNotFound) {
			return err
		}
	}
	if err := s.repo.CreateToken(s.repo.DB(), ctx, token); err != nil {
		return err
	}
	return nil
}
