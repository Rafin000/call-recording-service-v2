package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Redis key constants
const (
	PortaOneSessionKey = "portaone_session_id"
	SessionTimeout     = 25 * time.Minute
)

// RedisUtils provides methods for interacting with Redis
type RedisUtils struct {
	client *redis.Client
}

// NewRedisUtils creates a new RedisUtils instance
func NewRedisUtils(client *redis.Client) *RedisUtils {
	return &RedisUtils{client: client}
}

// SaveSessionID stores the session ID in Redis
func (r *RedisUtils) SaveSessionID(ctx context.Context, sessionID string) error {
	return r.client.Set(ctx, PortaOneSessionKey, sessionID, SessionTimeout).Err()
}

// GetSessionID retrieves the session ID from Redis
func (r *RedisUtils) GetSessionID(ctx context.Context) (string, error) {
	return r.client.Get(ctx, PortaOneSessionKey).Result()
}
