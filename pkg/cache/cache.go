package cache

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"hw/pkg/common"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
)

// TODO: SetXX, SetNX

// Cache interface defines the methods for interacting with the cache
type Cache interface {
	Get(ctx context.Context, key string, object interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	GetFunc(ctx context.Context, key string, obj interface{}, ttl time.Duration, fn func(ctx context.Context) (interface{}, error)) error
	FormatKey(args ...interface{}) string
	Del(ctx context.Context, key string) error
}

// cacheImpl is the implementation of the Cache interface
type cacheImpl struct {
	prefix     string
	cache      *cache.Cache
	redisCache *redis.Client
	defaultTTL time.Duration
}

// NewLocalCache creates a new local cache instance
func NewLocalCache() Cache {
	prefix := common.GetEnv("CACHE_PREFIX", "")
	defaultTTL := common.MustParseDuration(common.GetEnv("CACHE_DEFAULT_TTL", "1m"))
	return &cacheImpl{
		prefix: prefix,
		cache: cache.New(&cache.Options{
			LocalCache: cache.NewTinyLFU(1000, defaultTTL),
		}),
		defaultTTL: defaultTTL,
	}
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache() Cache {
	prefix := common.GetEnv("CACHE_PREFIX", "")
	redisAddr := common.GetEnv("CACHE_REDIS_ADDR", "localhost:6379")
	redisPassword := common.GetEnv("CACHE_REDIS_PASSWORD", "")
	redisDB, _ := strconv.Atoi(common.GetEnv("CACHE_REDIS_DB", "0"))
	defaultTTL := common.MustParseDuration(common.GetEnv("CACHE_DEFAULT_TTL", "1m"))

	redisOpts := &redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	}

	redisClient := redis.NewClient(redisOpts)
	return &cacheImpl{
		prefix:     prefix,
		cache:      cache.New(&cache.Options{Redis: redisClient}),
		redisCache: redisClient,
		defaultTTL: defaultTTL,
	}
}

// NewHybridCache creates a new hybrid cache instance (local + Redis)
func NewHybridCache() Cache {
	prefix := common.GetEnv("CACHE_PREFIX", "")
	redisAddr := common.GetEnv("CACHE_REDIS_ADDR", "localhost:6379")
	redisPassword := common.GetEnv("CACHE_REDIS_PASSWORD", "")
	redisDB, _ := strconv.Atoi(common.GetEnv("CACHE_REDIS_DB", "0"))
	defaultTTL := common.MustParseDuration(common.GetEnv("CACHE_DEFAULT_TTL", "1m"))

	redisOpts := &redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	}

	redisClient := redis.NewClient(redisOpts)
	return &cacheImpl{
		prefix: prefix,
		cache: cache.New(&cache.Options{
			LocalCache: cache.NewTinyLFU(1000, defaultTTL),
			Redis:      redisClient,
		}),
		redisCache: redisClient,
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value from the cache
func (c *cacheImpl) Get(ctx context.Context, key string, object interface{}) error {
	return c.cache.Get(ctx, c.FormatKey(key), object)
}

// Set stores a value in the cache with the specified TTL (or default TTL if not provided)
func (c *cacheImpl) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl < 0 {
		return fmt.Errorf("TTL cannot be negative")
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	return c.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   c.FormatKey(key),
		Value: value,
		TTL:   ttl,
	})
}

// ErrDataNotFound is returned when the requested data is not found in the cache
var ErrDataNotFound = errors.New("data not found")

// GetFunc retrieves a value from the cache or computes it using the provided function
// This method uses the Once functionality to ensure that only one goroutine computes the value
// while others wait for the result
func (c *cacheImpl) GetFunc(ctx context.Context, key string, obj interface{}, ttl time.Duration, fn func(ctx context.Context) (interface{}, error)) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	// The Once method ensures that only one goroutine executes the function to compute the value
	// If multiple goroutines try to access the same key simultaneously, only one will execute the function
	// while others wait for the result. This helps prevent the "thundering herd" problem.
	err := c.cache.Once(&cache.Item{
		Ctx:   ctx,
		Key:   c.FormatKey(key),
		Value: obj,
		TTL:   ttl,
		Do: func(item *cache.Item) (interface{}, error) {
			// This function is executed only if the value is not found in the cache
			result, err := fn(ctx)
			if err != nil {
				return nil, err
			}
			if result == nil {
				return nil, ErrDataNotFound
			}
			return result, nil
		},
	})

	if err == ErrDataNotFound {
		return ErrDataNotFound
	}

	return err
}

// FormatKey generates a formatted cache key with an optional prefix
func (c *cacheImpl) FormatKey(args ...interface{}) string {
	if c.prefix != "" {
		return join(c.prefix, join(args...))
	}
	return join(args...)
}

// Del removes a value from the cache
func (c *cacheImpl) Del(ctx context.Context, key string) error {
	return c.cache.Delete(ctx, c.FormatKey(key))
}

// BuildKeys constructs a slice of interface{} from a base string and optional string parameters.
// Parameters must not be empty strings.
func BuildKeys(base string, params ...string) []interface{} {
	keys := make([]interface{}, 0, len(params)+1)
	for _, param := range params {
		keys = append(keys, param)
	}
	keys = append(keys, base)
	return keys
}

// join concatenates multiple arguments into a single string, separated by colons
func join(args ...interface{}) string {
	s := make([]string, len(args))
	for i, v := range args {
		switch v := v.(type) {
		case string:
			s[i] = v
		case *string:
			if v != nil {
				s[i] = *v
			} else {
				s[i] = ""
			}
		case int64, uint64, float64, bool, *big.Int:
			s[i] = fmt.Sprintf("%v", v)
		default:
			s[i] = fmt.Sprintf("%+v", v)
		}
	}
	return strings.Join(s, ":")
}
