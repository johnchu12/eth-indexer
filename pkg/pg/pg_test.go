package pg

import (
	"context"
	"os"
	"testing"

	"hw/pkg/pg/mocks"

	"go.uber.org/mock/gomock"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

// TestPostgresDB_Ping tests the Ping method of PostgresDB.
func TestPostgresDB_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPool := mocks.NewMockPgxPool(ctrl)
	db := &PostgresDB{pool: mockPool}

	ctx := context.Background()

	// Set expectation: Ping is called with ctx and returns nil.
	mockPool.EXPECT().Ping(ctx).Return(nil)

	if err := db.Ping(ctx); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestPostgresDB_Close tests the Close method of PostgresDB.
func TestPostgresDB_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPool := mocks.NewMockPgxPool(ctrl)
	db := &PostgresDB{pool: mockPool}

	// Set expectation: Close is called once.
	mockPool.EXPECT().Close()

	db.Close()
}

// TestPostgresDB_Exec tests the Exec method of PostgresDB.
func TestPostgresDB_Exec(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPool := mocks.NewMockPgxPool(ctrl)
	db := &PostgresDB{pool: mockPool}

	ctx := context.Background()
	sql := "INSERT INTO users (name) VALUES ($1)"
	args := []any{"John Doe"}
	commandTag := pgconn.NewCommandTag("INSERT 1")
	var expectedErr error = nil

	// Set expectation: Exec is called with ctx, sql, and args, returns commandTag and nil.
	mockPool.EXPECT().Exec(ctx, sql, args...).Return(commandTag, expectedErr)

	tag, err := db.Exec(ctx, sql, args...)
	if tag != commandTag {
		t.Errorf("Expected CommandTag %v, got %v", commandTag, tag)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestPostgresDB_Query tests the Query method of PostgresDB.
func TestPostgresDB_Query(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPool := mocks.NewMockPgxPool(ctrl)
	db := &PostgresDB{pool: mockPool}

	ctx := context.Background()
	sql := "SELECT id, name FROM users WHERE id = $1"
	args := []any{1}
	mockRows := mocks.NewMockPgxRows(ctrl)
	var expectedErr error = nil

	// Set expectation: Query is called with ctx, sql, and args, returns mockRows and nil.
	mockPool.EXPECT().Query(ctx, sql, args...).Return(mockRows, expectedErr)

	rows, err := db.Query(ctx, sql, args...)
	if rows != mockRows {
		t.Errorf("Expected rows %v, got %v", mockRows, rows)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestPostgresDB_QueryRow tests the QueryRow method of PostgresDB.
func TestPostgresDB_QueryRow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPool := mocks.NewMockPgxPool(ctrl)
	db := &PostgresDB{pool: mockPool}

	ctx := context.Background()
	sql := "SELECT id, name FROM users WHERE id = $1"
	args := []any{1}

	// Set expectation: QueryRow is called with ctx, sql, and args, returns mockRow.
	mockPool.EXPECT().QueryRow(ctx, sql, args...).Return(nil) // Adjust as needed.

	row := db.QueryRow(ctx, sql, args...)
	if row != nil {
		t.Errorf("Expected row to be nil, got %v", row)
	}
}

// TestPostgresDB_Begin tests the Begin method of PostgresDB.
func TestPostgresDB_Begin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPool := mocks.NewMockPgxPool(ctrl)
	db := &PostgresDB{pool: mockPool}

	ctx := context.Background()
	mockTx := mocks.NewMockPgxTx(ctrl)
	var expectedErr error = nil

	// Set expectation: Begin is called with ctx, returns mockTx and nil.
	mockPool.EXPECT().Begin(ctx).Return(mockTx, expectedErr)

	tx, err := db.Begin(ctx)
	if tx != mockTx {
		t.Errorf("Expected tx %v, got %v", mockTx, tx)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestNewPostgresDB_DatabaseURLNotSet tests the NewPostgresDB function.
func TestNewPostgresDB_DatabaseURLNotSet(t *testing.T) {
	// 保存原始的 getEnv 函数
	originalGetEnv := os.Getenv("DATABASE_URL")
	defer func() {
		os.Setenv("DATABASE_URL", originalGetEnv)
	}()

	// 模拟 getEnv 返回空字符串
	os.Setenv("DATABASE_URL", "")

	db, err := NewPostgresDB()

	assert.Nil(t, db)
	assert.EqualError(t, err, "DATABASE_URL is not set")
}

// TestNewPostgresDB_ParseConfigError tests the NewPostgresDB error case.
func TestNewPostgresDB_ParseConfigError(t *testing.T) {
	os.Setenv("DATABASE_URL", "invalid-connection-string")
	db, err := NewPostgresDB()

	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to parse connection string")
}

// TestNewPostgresDB_Success tests the NewPostgresDB success case.
func TestNewPostgresDB_Success(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgresql://postgres:mysecretpassword@localhost:15050/pelith")
	db, err := NewPostgresDB()

	assert.NotNil(t, db)
	assert.NoError(t, err)
}

// TestNewPostgresDB_ConnectionRefused tests the NewPostgresDB connection refused case.
func TestNewPostgresDB_ConnectionRefused(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgresql://postgres:wrongpassword@localhost:14050/pelith")
	db, err := NewPostgresDB()

	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "connection refused")
}
