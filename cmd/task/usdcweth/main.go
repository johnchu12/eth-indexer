package main

import (
	"context"
	"log"

	"hw/internal/repository"
	"hw/internal/service"
	"hw/pkg/bigrat"
	"hw/pkg/logger"
	"hw/pkg/pg"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// TODO: use transaction

func main() {
	db, err := pg.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	repo := repository.NewRepository(db)
	service := service.NewService(repo)

	usdcweth := "0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc"
	totalSharePoolPoints := 10000.00

	userSwapSummary, err := service.GetUserSwapSummaryLast7Days(context.Background(), usdcweth)
	if err != nil {
		log.Fatalf("Failed to retrieve user swap summary: %v", err)
	}

	for _, userSwap := range userSwapSummary {
		user, err := service.GetOrCreateAccount(context.Background(), userSwap.Account)
		if err != nil {
			log.Fatalf("Failed to retrieve user: %v", err)
		}

		completed, err := service.IsOnboardingTaskCompleted(context.Background(), userSwap.Account)
		if err != nil {
			log.Fatalf("Failed to retrieve user points history: %v", err)
		}

		// if not completed, skip awarding points
		if !completed {
			continue
		}

		newPoints := bigrat.NewBigN(totalSharePoolPoints).Mul(userSwap.Percentage).ToTruncateFloat64(3)

		if err := service.AccumulateUserPoints(context.Background(), usdcweth, user.Address, "sharepool_usdcweth_task", newPoints); err != nil {
			log.Fatalf("Failed to create points history: %v", err)
		}
	}
	logger.Infow("task completed")
}
