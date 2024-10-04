package cache

import (
	"context"
	"math/big"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheImplementations(t *testing.T) {
	os.Setenv("CACHE_DEFAULT_TTL", "1s")
	// Define different Cache instances
	caches := map[string]Cache{
		"LocalCache":  NewLocalCache(),
		"RedisCache":  NewRedisCache(),
		"HybridCache": NewHybridCache(),
	}

	for name, cacheImpl := range caches {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			// Test Set method
			t.Run("Set", func(t *testing.T) {
				key := "test_key_set_" + name
				value := "test_value_set"

				err := cacheImpl.Set(ctx, key, value, 500*time.Millisecond)
				assert.NoError(t, err, "Set should not return an error")
			})

			// Test Get method
			t.Run("Get", func(t *testing.T) {
				key := "test_key_get_" + name
				value := "test_value_get"

				// Pre-set value
				err := cacheImpl.Set(ctx, key, value, 500*time.Millisecond)
				assert.NoError(t, err, "Set should not return an error")

				var result string
				err = cacheImpl.Get(ctx, key, &result)
				assert.NoError(t, err, "Get should not return an error")
				assert.Equal(t, value, result, "Retrieved value should match set value")
			})

			// Test GetFunc method
			t.Run("GetFunc", func(t *testing.T) {
				funcKey := "func_key_" + name
				expectedValue := "computed_value"

				fn := func(ctx context.Context) (interface{}, error) {
					return expectedValue, nil
				}

				var result string
				err := cacheImpl.GetFunc(ctx, funcKey, &result, 500*time.Millisecond, fn)
				assert.NoError(t, err, "GetFunc should not return an error")
				assert.Equal(t, expectedValue, result, "Computed value should match")
			})

			// Test Del method
			t.Run("Del", func(t *testing.T) {
				key := "test_key_del_" + name
				value := "test_value_del"

				// Pre-set value
				err := cacheImpl.Set(ctx, key, value, 500*time.Millisecond)
				assert.NoError(t, err, "Set should not return an error")

				// Delete key
				err = cacheImpl.Del(ctx, key)
				assert.NoError(t, err, "Del should not return an error")

				// Attempt to get deleted key
				var result string
				err = cacheImpl.Get(ctx, key, &result)
				if name != "LocalCache" { // LocalCache does not support Get after Del
					assert.Error(t, err, "Get should return an error after Del")
				}
			})

			// Add test case: SetLargeValue
			t.Run("SetLargeValue", func(t *testing.T) {
				key := "test_key_large_" + name
				// Assume large value is a 1MB string
				largeValue := strings.Repeat("a", 1024*1024)

				err := cacheImpl.Set(ctx, key, largeValue, 1*time.Second)
				assert.NoError(t, err, "Set with large value should not return an error")

				var result string
				err = cacheImpl.Get(ctx, key, &result)
				assert.NoError(t, err, "Get should not return an error for large value")
				assert.Equal(t, largeValue, result, "Retrieved large value should match set value")
			})

			// Add test case: OverwriteExistingKey
			t.Run("OverwriteExistingKey", func(t *testing.T) {
				key := "test_key_overwrite_" + name
				initialValue := "initial_value"
				newValue := "new_value"

				err := cacheImpl.Set(ctx, key, initialValue, 1*time.Second)
				assert.NoError(t, err, "Initial Set should not return an error")

				err = cacheImpl.Set(ctx, key, newValue, 1*time.Second)
				assert.NoError(t, err, "Overwrite Set should not return an error")

				var result string
				err = cacheImpl.Get(ctx, key, &result)
				assert.NoError(t, err, "Get should not return an error after overwrite")
				assert.Equal(t, newValue, result, "Retrieved value should match the new value")
			})

			// Add test case: SetZeroTTL
			t.Run("SetZeroTTL", func(t *testing.T) {
				key := "test_key_zero_ttl_" + name
				value := "test_value_zero_ttl"

				err := cacheImpl.Set(ctx, key, value, 0)
				assert.NoError(t, err, "Set with zero TTL should not return an error")

				var result string
				err = cacheImpl.Get(ctx, key, &result)
				assert.NoError(t, err, "Get should not return an error for key with zero TTL")
				assert.Equal(t, value, result, "Retrieved value should match set value with zero TTL")
			})

			// Add test case: SetNegativeTTL
			t.Run("SetNegativeTTL", func(t *testing.T) {
				key := "test_key_negative_ttl_" + name
				value := "test_value_negative_ttl"

				err := cacheImpl.Set(ctx, key, value, -1*time.Second)
				assert.Error(t, err, "Set with negative TTL should return an error")
			})

			// Add test case: GetNonExistentKey
			t.Run("GetNonExistentKey", func(t *testing.T) {
				key := "nonexistent_key_" + name

				var result string
				err := cacheImpl.Get(ctx, key, &result)
				if name != "LocalCache" { // Assume LocalCache does not support getting non-existent keys and returns an error
					assert.Error(t, err, "Get should return an error for non-existent key")
				} else {
					assert.Equal(t, "", result, "Result should be empty string for non-existent key in LocalCache")
				}
			})

			// Add test case: DelNonExistentKey
			t.Run("DelNonExistentKey", func(t *testing.T) {
				key := "del_nonexistent_key_" + name

				err := cacheImpl.Del(ctx, key)
				assert.NoError(t, err, "Del on non-existent key should not return an error")
			})
		})
	}
}

func TestCacheExpire(t *testing.T) {
	os.Setenv("CACHE_DEFAULT_TTL", "1s")

	// Define different Cache instances
	caches := map[string]Cache{
		"LocalCache":  NewLocalCache(),
		"RedisCache":  NewRedisCache(),
		"HybridCache": NewHybridCache(),
	}

	// Use different keys to test TTL expiration for each Cache
	expireTests := map[string]string{
		"LocalCache":  "expire_key_local",
		"RedisCache":  "expire_key_redis",
		"HybridCache": "expire_key_hybrid",
	}

	t.Run("Expire", func(t *testing.T) {
		var wg sync.WaitGroup
		for name, cacheImpl := range caches {
			wg.Add(1)
			go func(name string, cacheImpl Cache) {
				defer wg.Done()
				ctx := context.Background()
				expireKey := expireTests[name]
				value := "expire_value_" + name

				// Set value and define TTL
				err := cacheImpl.Set(ctx, expireKey, value, time.Second)
				assert.NoError(t, err, "Set should not return an error")

				// Wait for TTL to expire
				time.Sleep(2 * time.Second)

				var result string
				err = cacheImpl.Get(ctx, expireKey, &result)
				assert.Error(t, err, "Get should return an error after TTL expires")
				assert.Equal(t, "", result, "Result should be empty after TTL expires")
			}(name, cacheImpl)
		}
		wg.Wait()
	})
}

func TestBuildKeys(t *testing.T) {
	testCases := []struct {
		name     string
		base     string
		params   []string
		expected []interface{}
	}{
		{
			name:     "No parameters",
			base:     "base",
			params:   []string{},
			expected: []interface{}{"base"},
		},
		{
			name:     "Single parameter",
			base:     "base",
			params:   []string{"param1"},
			expected: []interface{}{"param1", "base"},
		},
		{
			name:     "Multiple parameters",
			base:     "base",
			params:   []string{"param1", "param2"},
			expected: []interface{}{"param1", "param2", "base"},
		},
		{
			name:     "Includes empty string parameter",
			base:     "base",
			params:   []string{"param1", "", "param3"},
			expected: []interface{}{"param1", "", "param3", "base"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := BuildKeys(tc.base, tc.params...)
			assert.Equal(t, tc.expected, result, "BuildKeys should correctly construct the key slice")
		})
	}
}

// TestJoin tests the join function with various input scenarios
func TestJoin(t *testing.T) {
	testCases := []struct {
		name     string
		args     []interface{}
		expected string
	}{
		{
			name:     "All strings",
			args:     []interface{}{"user", "123", "profile"},
			expected: "user:123:profile",
		},
		{
			name:     "Mixed string and *string",
			args:     []interface{}{"user", ptr("123"), "profile"},
			expected: "user:123:profile",
		},
		{
			name:     "Includes nil *string",
			args:     []interface{}{"user", (*string)(nil), "profile"},
			expected: "user::profile",
		},
		{
			name:     "Includes int64",
			args:     []interface{}{"order", int64(456), "details"},
			expected: "order:456:details",
		},
		{
			name:     "Includes uint64",
			args:     []interface{}{"item", uint64(789), "info"},
			expected: "item:789:info",
		},
		{
			name:     "Includes float64",
			args:     []interface{}{"price", 99.99, "USD"},
			expected: "price:99.99:USD",
		},
		{
			name:     "Includes bool",
			args:     []interface{}{"feature", true, "enabled"},
			expected: "feature:true:enabled",
		},
		{
			name:     "Includes *big.Int",
			args:     []interface{}{"transaction", big.NewInt(1000), "complete"},
			expected: "transaction:1000:complete",
		},
		{
			name:     "Includes unsupported type",
			args:     []interface{}{"data", []int{1, 2, 3}},
			expected: "data:[1 2 3]",
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable to avoid concurrent issues
		t.Run(tc.name, func(t *testing.T) {
			result := join(tc.args...)
			assert.Equal(t, tc.expected, result, "join should correctly concatenate strings")
		})
	}
}

// ptr is a helper function to create a string pointer
func ptr(s string) *string {
	return &s
}
