package repository_test

import (
	"context"
	"testing"

	"hw/internal/repository"
	pgMock "hw/pkg/pg/mocks"

	"go.uber.org/mock/gomock"
)

// TestBeginTransaction tests the BeginTransaction method of the Repository.
func TestBeginTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Create mock PgxPool and PgxTx
	mockDB := pgMock.NewMockPgxPool(ctrl)
	mockTx := pgMock.NewMockPgxTx(ctrl)

	// Initialize Repository with the mock database
	repo := repository.NewRepository(mockDB)

	ctx := context.Background()

	// Set expected behavior for Begin transaction
	mockDB.EXPECT().Begin(ctx).Return(mockTx, nil)

	// Invoke BeginTransaction
	tx, err := repo.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if tx != mockTx {
		t.Errorf("Expected tx to be %v, got %v", mockTx, tx)
	}
}
