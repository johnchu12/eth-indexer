package pg_test

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"testing"

// 	"hw/internal/pg"

// 	"github.com/jackc/pgx/v5/pgxpool"
// 	"github.com/pashagolub/pgxmock/v4"
// 	"github.com/stretchr/testify/assert"
// )

// type MockPgxPool struct {
// 	pgxmock.PgxPoolIface
// }

// func TestNewPostgresDB_Success(t *testing.T) {
// 	// 设置环境变量
// 	os.Setenv("DATABASE_URL", "postgresql://postgres:mysecretpassword@localhost:5432/pelith?sslmode=disable")
// 	defer os.Unsetenv("DATABASE_URL")

// 	// 创建 pgxmock 的连接池
// 	mockPoolIface, err := pgxmock.NewPool()
// 	assert.NoError(t, err)
// 	defer mockPoolIface.Close()

// 	// 期待 Ping 行为
// 	mockPoolIface.ExpectPing()

// 	// 调用 NewPostgresDB
// 	db, err := pg.NewPostgresDB()
// 	assert.NoError(t, err)
// 	assert.NotNil(t, db)

// 	// 验证所有的预期行为都被满足
// 	err = mockPoolIface.ExpectationsWereMet()
// 	assert.NoError(t, err)
// }

// func TestNewPostgresDB_NoDatabaseURL(t *testing.T) {
// 	// 确保环境变量未设置
// 	os.Unsetenv("DATABASE_URL")

// 	// 调用 NewPostgresDB
// 	db, err := pg.NewPostgresDB()
// 	assert.Error(t, err)
// 	assert.Nil(t, db)
// 	assert.EqualError(t, err, "DATABASE_URL is not set")
// }

// func TestNewPostgresDB_InvalidConnectionString(t *testing.T) {
// 	// 设置无效的连接字符串
// 	os.Setenv("DATABASE_URL", "invalid_connection_string")
// 	defer os.Unsetenv("DATABASE_URL")

// 	// 调用 NewPostgresDB
// 	db, err := pg.NewPostgresDB()
// 	assert.Error(t, err)
// 	assert.Nil(t, db)
// 	assert.Contains(t, err.Error(), "failed to parse connection string")
// }

// func TestNewPostgresDB_FailedToCreatePool(t *testing.T) {
// 	// 设置有效的环境变量
// 	defer os.Unsetenv("DATABASE_URL")

// 	// 保存原始的 PoolCreator，测试结束后恢复
// 	originalPoolCreator := pg.PoolCreator
// 	defer func() { pg.PoolCreator = originalPoolCreator }()

// 	// 模拟创建连接池失败
// 	pg.PoolCreator = func(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error) {
// 		return nil, fmt.Errorf("failed to create pool")
// 	}

// 	// 调用 NewPostgresDB
// 	db, err := pg.NewPostgresDB()
// 	assert.Error(t, err)
// 	assert.Nil(t, db)
// 	assert.EqualError(t, err, "failed to create connection pool: failed to create pool")
// }

// func TestNewPostgresDB_PingFailed(t *testing.T) {
// 	// 设置环境变量
// 	defer os.Unsetenv("DATABASE_URL")

// 	// 保存原始的 PoolCreator，测试结束后恢复
// 	originalPoolCreator := pg.PoolCreator
// 	defer func() { pg.PoolCreator = originalPoolCreator }()

// 	// 创建 pgxmock 的连接池
// 	mockPoolIface, err := pgxmock.NewPool()
// 	assert.NoError(t, err)
// 	defer mockPoolIface.Close()

// 	// 替换 PoolCreator，返回我们的自定义连接池
// 	pg.PoolCreator = func(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error) {
// 		return pgxpool.NewWithConfig(ctx, config)
// 	}

// 	// 模拟 Ping 失败
// 	mockPoolIface.ExpectPing().WillReturnError(fmt.Errorf("ping failed"))

// 	// 调用 NewPostgresDB
// 	db, err := pg.NewPostgresDB()
// 	assert.Error(t, err)
// 	assert.Nil(t, db)
// 	assert.EqualError(t, err, "connection test failed: ping failed")

// 	// 验证所有的预期行为都被满足
// 	err = mockPoolIface.ExpectationsWereMet()
// 	assert.NoError(t, err)
// }
