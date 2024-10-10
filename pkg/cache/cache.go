package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	"hw/pkg/common"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
)

// Cache defines the methods for interacting with the cache.
type Cache interface {
	Get(ctx context.Context, key string, object interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	GetFunc(ctx context.Context, key string, obj interface{}, ttl time.Duration, fn func(ctx context.Context) (interface{}, error)) error
	FormatKey(args ...interface{}) string
	Del(ctx context.Context, key string) error
}

// cacheImpl implements the Cache interface.
type cacheImpl struct {
	prefix     string
	cache      *cache.Cache
	defaultTTL time.Duration
	sf         *singleflight.Group
}

// NewLocalCache creates a new local cache instance.
func NewLocalCache() Cache {
	prefix := common.GetEnv("CACHE_PREFIX", "")
	defaultTTL := common.MustParseDuration(common.GetEnv("CACHE_DEFAULT_TTL", "1m"))
	return &cacheImpl{
		prefix: prefix,
		cache: cache.New(&cache.Options{
			LocalCache: cache.NewTinyLFU(1000, defaultTTL),
		}),
		defaultTTL: defaultTTL,
		sf:         &singleflight.Group{},
	}
}

// NewRedisCache creates a new Redis cache instance.
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
		defaultTTL: defaultTTL,
		sf:         &singleflight.Group{},
	}
}

// NewHybridCache creates a new hybrid cache instance combining local and Redis caches.
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
		defaultTTL: defaultTTL,
		sf:         &singleflight.Group{},
	}
}

// Get retrieves a value from the cache.
func (c *cacheImpl) Get(ctx context.Context, key string, object interface{}) error {
	return c.cache.Get(ctx, c.FormatKey(key), object)
}

// Set stores a value in the cache with the specified TTL.
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

// ErrDataNotFound is returned when the requested data is not found in the cache.
var ErrDataNotFound = errors.New("data not found")

type NullObject struct{}

// GetFunc retrieves a value from the cache or computes it using the provided function.
// This method ensures that only one goroutine computes the value while others wait for the result.
func (c *cacheImpl) GetFunc(ctx context.Context, key string, obj interface{}, ttl time.Duration, fn func(ctx context.Context) (interface{}, error)) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	v, err, _ := c.sf.Do(key, func() (interface{}, error) {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context cancelled before execution: %w", err)
		}

		result, err := fn(ctx)
		if err != nil {
			return nil, fmt.Errorf("error executing function: %w", err)
		}

		if result == nil {
			if err := c.Set(ctx, key, NullObject{}, ttl); err != nil {
				return nil, fmt.Errorf("error setting null object in cache: %w", err)
			}
			return nil, ErrDataNotFound
		}

		var cacheValue interface{}
		switch val := result.(type) {
		case []byte:
			cacheValue = val
		case string:
			cacheValue = []byte(val)
		default:
			jsonData, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("error marshaling result to JSON: %w", err)
			}
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("context cancelled after JSON marshaling: %w", err)
			}
			cacheValue = jsonData
		}

		if err := c.Set(ctx, key, cacheValue, ttl); err != nil {
			return result, fmt.Errorf("error setting value in cache (result still returned): %w", err)
		}

		return result, nil
	})

	if err != nil {
		if errors.Is(err, ErrDataNotFound) {
			return ErrDataNotFound
		}
		return err
	}

	if v == nil {
		reflect.ValueOf(obj).Elem().Set(reflect.Zero(reflect.TypeOf(obj).Elem()))
	} else {
		switch val := v.(type) {
		case []byte:
			reflect.ValueOf(obj).Elem().SetBytes(val)
		case string:
			reflect.ValueOf(obj).Elem().SetString(val)
		default:
			reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(v))
		}
	}

	return nil
}

// FormatKey generates a formatted cache key with an optional prefix.
func (c *cacheImpl) FormatKey(args ...interface{}) string {
	if c.prefix != "" {
		return join(c.prefix, join(args...))
	}
	return join(args...)
}

// Del removes a value from the cache.
func (c *cacheImpl) Del(ctx context.Context, key string) error {
	return c.cache.Delete(ctx, c.FormatKey(key))
}

// BuildKeys constructs a slice of interface{} from a base string and optional string parameters.
// Parameters must not be empty strings.
func BuildKeys(base string, params ...string) []interface{} {
	keys := make([]interface{}, 0, len(params)+1)
	for _, param := range params {
		if param != "" {
			keys = append(keys, param)
		}
	}
	keys = append(keys, base)
	return keys
}

// join concatenates multiple arguments into a single string, separated by colons.
func join(args ...interface{}) string {
	s := make([]string, 0, len(args))
	for _, v := range args {
		switch val := v.(type) {
		case string:
			s = append(s, val)
		case *string:
			if val != nil {
				s = append(s, *val)
			} else {
				s = append(s, "")
			}
		case nil:
			s = append(s, "")
		case int64, uint64, float64, bool, *big.Int:
			s = append(s, fmt.Sprintf("%v", val))
		default:
			s = append(s, fmt.Sprintf("%+v", val))
		}
	}
	return strings.Join(s, ":")
}
