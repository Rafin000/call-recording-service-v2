package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/go-redis/redis/v8"
)

// RedisClient defines the interface for Redis operations
type RedisClient interface {
	GetClient() *redis.Client
	Close() error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
}

// redisClient handles Redis operations
type redisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new RedisClient instance
func NewRedisClient(ctx context.Context, redisConfig common.RedisConfig) (RedisClient, error) {
	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	}

	client := redis.NewClient(options)

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisClient{client: client}, nil
}

// GetClient returns the underlying Redis client
func (r *redisClient) GetClient() *redis.Client {
	return r.client
}

// Close closes the Redis connection
func (r *redisClient) Close() error {
	return r.client.Close()
}

// Helper methods for common operations
func (r *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *redisClient) Exists(ctx context.Context, keys ...string) (bool, error) {
	n, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *redisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

func (r *redisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}
