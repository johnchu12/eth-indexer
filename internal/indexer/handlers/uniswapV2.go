package handlers

import (
	"context"
	"math/big"
	"strings"
	"time"

	"hw/internal/model"
	"hw/pkg/bigrat"
	"hw/pkg/ethindexa"
	"hw/pkg/logger"

	"github.com/google/uuid"
)

const (
	USDCWETHPool = "0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc"
	USDC         = "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
)

// HandleUSDCWETHSwap processes a USDC-WETH swap event.
func HandleUSDCWETHSwap(idx *ethindexa.IndexerService, event ethindexa.Event) {
	// token0 = USDC
	// token1 = WETH

	block, err := idx.GetBlockByHash(event.BlockHash)
	if err != nil {
		logger.Errorf("Error retrieving block details: %v", err)
		return
	}

	tx, err := idx.GetTransactionByHash(event.TransactionHash)
	if err != nil {
		logger.Errorf("Error retrieving transaction details: %v", err)
		return
	}

	// Retrieve user account ID
	accountID := strings.ToLower(tx.From.Hex())

	// create request id for tracing
	reqID := uuid.New().String()[:8]

	// print processed message
	logger.Infof("#%s:%s:%s %s %s at %d", event.NetworkName, event.ContractName, event.EventName, event.ContractAddress, event.TransactionHash.Hex(), block.Time())

	ctx := context.WithValue(event.Ctx, "reqID", reqID)

	// Retrieve or create USDC token information
	usdcToken, err := idx.Service.GetOrCreateToken(ctx, idx.Client, USDC, block.Number().Int64())
	if err != nil {
		logger.Errorw("Error retrieving USDC token:", err)
		return
	}

	// Calculate USDC value
	usdValue := bigrat.NewBigN(event.Args["amount0In"].(*big.Int))
	if event.Args["amount0Out"].(*big.Int).Cmp(big.NewInt(0)) != 0 {
		usdValue = bigrat.NewBigN(event.Args["amount0Out"].(*big.Int))
	}

	// Create swap history record
	swapHistory := &model.SwapHistory{
		Token:           USDCWETHPool, // USDC-WETH pool address
		Account:         accountID,
		TransactionHash: event.TransactionHash.Hex(),
		UsdValue:        usdValue.Div(bigrat.NewBigN(10).Pow(usdcToken.Decimals)).ToTruncateFloat64(6),
		LastUpdated:     time.Unix(int64(block.Time()), 0),
	}

	if err := idx.Service.CreateSwapHistory(event.Ctx, swapHistory); err != nil {
		logger.Errorw("Error creating swap history:", err)
		return
	}

	// Check if onboarding task is completed
	completed, err := idx.Service.IsOnboardingTaskCompleted(event.Ctx, accountID)
	if err != nil {
		logger.Errorw("Error checking onboarding task status:", err)
		return
	}

	// If not completed, verify if onboarding criteria are met
	if !completed {
		totalUSD, err := idx.Service.GetSwapTotalUsd(event.Ctx, accountID, USDCWETHPool)
		if err != nil {
			logger.Errorw("Error retrieving total swap USD:", err)
			return
		}
		if totalUSD >= 1000 {
			if err := idx.Service.AccumulateUserPoints(event.Ctx, USDCWETHPool, accountID, "onboarding_task", 100); err != nil {
				logger.Errorw("Error accumulating user points:", err)
			}
		}
	}
}
