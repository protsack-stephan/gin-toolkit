package httpmw

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// NewRedisLimiter creates new redis limiter.
// Note:
// if expire is set to 0 it means the key has no expiration time.
func NewRedisLimiter(cmdable redis.Cmdable, limit int, expire time.Duration) *RedisLimiter {
	return &RedisLimiter{
		cmdable: cmdable,
		expire:  expire,
		limit:   limit,
	}
}

// RedisLimiter struct is used for storing limit, expiration and redis client.
type RedisLimiter struct {
	limit   int
	expire  time.Duration
	cmdable redis.Cmdable
}

// buildKey returns built key for redis storage.
func (rl *RedisLimiter) buildKey(identifier string, entity string) string {
	return fmt.Sprintf("%s:%s:count", entity, identifier)
}

// Allow checks identifier is alowed to continue depending on limit.
func (rl *RedisLimiter) Allow(ctx context.Context, identifier string, entity string) (bool, error) {
	count, err := rl.cmdable.Get(ctx, rl.buildKey(identifier, entity)).Int()

	if err == redis.Nil {
		return true, nil
	}

	if err != nil {
		return true, err
	}

	return count < rl.limit, nil
}

// Seen increments number of actions performed by this particular identifier.
func (rl *RedisLimiter) Seen(ctx context.Context, identifier string, entity string) error {
	if rl.cmdable.Exists(ctx, rl.buildKey(identifier, entity)).Val() == 0 {
		return rl.cmdable.Set(ctx, rl.buildKey(identifier, entity), 1, rl.expire).Err()
	}

	return rl.cmdable.Incr(ctx, rl.buildKey(identifier, entity)).Err()
}
