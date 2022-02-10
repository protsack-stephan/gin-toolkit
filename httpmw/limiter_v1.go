package httpmw

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// NewRedisLimiter creates redisLimiter.
func NewRedisLimiter(cmdable redis.Cmdable, limit int, expire time.Duration) *RedisLimiter {
	return &RedisLimiter{
		cmdable: cmdable,
		expire:  expire,
		limit:   limit,
	}
}

type RedisLimiter struct {
	limit   int
	expire  time.Duration
	cmdable redis.Cmdable
}

// BuildRedisKey returns built key for redis storage.
func (rl *RedisLimiter) BuildRedisKey(identifier string, entity string) string {
	return fmt.Sprintf("%s:%s:count", entity, identifier)
}

// Allow checks identifier is alowed to continue depending on limit.
func (rl *RedisLimiter) Allow(ctx context.Context, identifier string, entity string) (bool, error) {
	count, err := rl.cmdable.Get(ctx, rl.BuildRedisKey(identifier, entity)).Int()

	if err == redis.Nil {
		return true, nil
	} else if err != nil {
		return true, err
	}

	return count < rl.limit, nil
}

// Seen adds or increments the counter for identifier in the redisLimiter.
func (rl *RedisLimiter) Seen(ctx context.Context, identifier string, entity string) error {
	if rl.cmdable.Exists(ctx, rl.BuildRedisKey(identifier, entity)).Val() == 0 {
		return rl.cmdable.Set(ctx, rl.BuildRedisKey(identifier, entity), 1, rl.expire).Err()
	}

	return rl.cmdable.Incr(ctx, rl.BuildRedisKey(identifier, entity)).Err()
}
