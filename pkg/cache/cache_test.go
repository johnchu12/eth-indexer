package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/cache/v9"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/sync/singleflight"
)

// mockCache is a mock implementation of the Cache interface
type mockCache struct {
	mock.Mock
}

// Get retrieves an item from the cache.
func (m *mockCache) Get(ctx context.Context, key string, object interface{}) error {
	args := m.Called(ctx, key, object)
	return args.Error(0)
}

// Set adds an item to the cache with a specified TTL.
func (m *mockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

// GetFunc retrieves an item from the cache or executes a function to obtain it.
func (m *mockCache) GetFunc(ctx context.Context, key string, obj interface{}, ttl time.Duration, fn func(ctx context.Context) (interface{}, error)) error {
	args := m.Called(ctx, key, obj, ttl, fn)
	return args.Error(0)
}

// FormatKey formats the cache key with given arguments.
func (m *mockCache) FormatKey(args ...interface{}) string {
	callArgs := m.Called(args)
	return callArgs.String(0)
}

// Del removes an item from the cache.
func (m *mockCache) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// TestNewLocalCache verifies the creation of a new local cache instance.
func TestNewLocalCache(t *testing.T) {
	c := NewLocalCache()
	assert.NotNil(t, c)
	assert.IsType(t, &cacheImpl{}, c)
	impl := c.(*cacheImpl)
	assert.NotNil(t, impl.cache)
	assert.NotEmpty(t, impl.defaultTTL)
}

// TestNewRedisCache verifies the creation of a new Redis cache instance.
func TestNewRedisCache(t *testing.T) {
	c := NewRedisCache()
	assert.NotNil(t, c)
	assert.IsType(t, &cacheImpl{}, c)
	impl := c.(*cacheImpl)
	assert.NotNil(t, impl.cache)
	assert.NotEmpty(t, impl.defaultTTL)
}

// TestNewHybridCache verifies the creation of a new hybrid cache instance.
func TestNewHybridCache(t *testing.T) {
	c := NewHybridCache()
	assert.NotNil(t, c)
	assert.IsType(t, &cacheImpl{}, c)
	impl := c.(*cacheImpl)
	assert.NotNil(t, impl.cache)
	assert.NotEmpty(t, impl.defaultTTL)
}

// TestGet tests the Get method of the cache implementation.
func TestGet(t *testing.T) {
	db, mock := redismock.NewClientMock()
	c := &cacheImpl{
		cache:  cache.New(&cache.Options{Redis: db}),
		prefix: "test",
	}

	ctx := context.Background()
	key := "testKey"
	var value string

	t.Run("Successful Get", func(t *testing.T) {
		mock.ExpectGet(c.FormatKey(key)).SetVal("testValue")
		err := c.Get(ctx, key, &value)
		assert.NoError(t, err)
		assert.Equal(t, "testValue", value)
	})

	t.Run("Key Not Found", func(t *testing.T) {
		mock.ExpectGet(c.FormatKey(key)).RedisNil()
		err := c.Get(ctx, key, &value)
		assert.Error(t, err)
		assert.Equal(t, "cache: key is missing", err.Error())
	})

	t.Run("Redis Error", func(t *testing.T) {
		mock.ExpectGet(c.FormatKey(key)).SetErr(errors.New("redis error"))
		err := c.Get(ctx, key, &value)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis error")
	})

	t.Run("Invalid Value Type", func(t *testing.T) {
		mock.ExpectGet(c.FormatKey(key)).SetVal("123")
		var intValue int
		err := c.Get(ctx, key, &intValue)
		assert.Error(t, err)
	})
}

// TestSet tests the Set method of the cache implementation.
func TestSet(t *testing.T) {
	db, mock := redismock.NewClientMock()
	c := &cacheImpl{
		cache:      cache.New(&cache.Options{Redis: db}),
		prefix:     "test",
		defaultTTL: time.Minute,
	}

	ctx := context.Background()
	key := "testKey"
	value := "testValue"

	t.Run("Successful Set", func(t *testing.T) {
		mock.ExpectSet(c.FormatKey(key), []byte(value), time.Minute).SetVal("OK")
		err := c.Set(ctx, key, value, time.Minute)
		assert.NoError(t, err)
	})

	t.Run("Set with Default TTL", func(t *testing.T) {
		mock.ExpectSet(c.FormatKey(key), []byte(value), time.Minute).SetVal("OK")
		err := c.Set(ctx, key, value, 0)
		assert.NoError(t, err)
	})

	t.Run("Set with Negative TTL", func(t *testing.T) {
		err := c.Set(ctx, key, value, -time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TTL cannot be negative")
	})

	t.Run("Set with Nil Value", func(t *testing.T) {
		err := c.Set(ctx, key, nil, time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value cannot be nil")
	})

	t.Run("Redis Error", func(t *testing.T) {
		mock.ExpectSet(c.FormatKey(key), []byte(value), time.Minute).SetErr(errors.New("redis error"))
		err := c.Set(ctx, key, value, time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis error")
	})
}

// TestGetFunc tests the GetFunc method of the cache implementation.
func TestGetFunc(t *testing.T) {
	t.Run("Success - String Value", func(t *testing.T) {
		db, mock := redismock.NewClientMock()

		c := &cacheImpl{
			cache:      cache.New(&cache.Options{Redis: db}),
			prefix:     "test",
			defaultTTL: time.Minute,
			sf:         &singleflight.Group{},
		}

		ctx := context.Background()
		key := "test_key"
		formattedKey := c.FormatKey(key)
		expectedValue := "test_value"

		mock.ExpectSet(formattedKey, []byte(expectedValue), time.Minute).SetVal("OK")

		fn := func(ctx context.Context) (interface{}, error) {
			return expectedValue, nil
		}

		var result string
		err := c.GetFunc(ctx, key, &result, time.Minute, fn)

		assert.NoError(t, err)
		assert.Equal(t, expectedValue, result)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestFormatKey tests the FormatKey method of the cache implementation.
func TestFormatKey(t *testing.T) {
	c := &cacheImpl{prefix: "test"}

	t.Run("With Prefix", func(t *testing.T) {
		key := c.FormatKey("key1", "key2")
		assert.Equal(t, "test:key1:key2", key)
	})

	t.Run("Without Prefix", func(t *testing.T) {
		c.prefix = ""
		key := c.FormatKey("key1", "key2")
		assert.Equal(t, "key1:key2", key)
	})

	t.Run("With Multiple Types", func(t *testing.T) {
		key := c.FormatKey("key1", 123, true)
		assert.Equal(t, "key1:123:true", key)
	})

	t.Run("With Empty Args", func(t *testing.T) {
		key := c.FormatKey()
		assert.Equal(t, "", key)
	})
}

// TestDel tests the Del method of the cache implementation.
func TestDel(t *testing.T) {
	db, mock := redismock.NewClientMock()
	c := &cacheImpl{
		cache:  cache.New(&cache.Options{Redis: db}),
		prefix: "test",
	}

	ctx := context.Background()
	key := "testKey"

	t.Run("Successful Delete", func(t *testing.T) {
		mock.ExpectDel(c.FormatKey(key)).SetVal(1)
		err := c.Del(ctx, key)
		assert.NoError(t, err)
	})

	t.Run("Delete Non-Existent Key", func(t *testing.T) {
		mock.ExpectDel(c.FormatKey(key)).SetVal(0)
		err := c.Del(ctx, key)
		assert.NoError(t, err)
	})

	t.Run("Redis Error", func(t *testing.T) {
		mock.ExpectDel(c.FormatKey(key)).SetErr(errors.New("redis error"))
		err := c.Del(ctx, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis error")
	})
}

// TestBuildKeys tests the BuildKeys function.
func TestBuildKeys(t *testing.T) {
	t.Run("With Parameters", func(t *testing.T) {
		keys := BuildKeys("base", "param1", "param2")
		assert.Equal(t, []interface{}{"param1", "param2", "base"}, keys)
	})

	t.Run("Without Parameters", func(t *testing.T) {
		keys := BuildKeys("base")
		assert.Equal(t, []interface{}{"base"}, keys)
	})

	t.Run("With Empty Parameters", func(t *testing.T) {
		keys := BuildKeys("base", "", "param2")
		assert.Equal(t, []interface{}{"param2", "base"}, keys)
	})

	t.Run("With All Empty Parameters", func(t *testing.T) {
		keys := BuildKeys("")
		assert.Equal(t, []interface{}{""}, keys)
	})
}

// TestJoin tests the join function.
func TestJoin(t *testing.T) {
	t.Run("Various Types", func(t *testing.T) {
		result := join("string", 123, true, 3.14)
		assert.Equal(t, "string:123:true:3.14", result)
	})

	t.Run("Pointer Types", func(t *testing.T) {
		str := "pointer"
		result := join(&str, (*string)(nil))
		assert.Equal(t, "pointer:", result)
	})

	t.Run("Empty Args", func(t *testing.T) {
		result := join()
		assert.Equal(t, "", result)
	})

	t.Run("Nil Values", func(t *testing.T) {
		var nilPtr *string
		result := join(nil, nilPtr, "valid")
		assert.Equal(t, "::valid", result)
	})
}
