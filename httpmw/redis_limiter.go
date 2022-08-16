package httpmw

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// NewRedisLimiter creates new redis limiter
// Note: if expire is set to 0 it means the key has no expiration time
func NewRedisLimiter(cmdable redis.Cmdable, entity string, limit int, expire time.Duration) *RedisLimiter {
	return &RedisLimiter{
		cmdable: cmdable,
		expire:  expire,
		limit:   limit,
		entity:  entity,
	}
}

// RedisLimiter struct is used for storing limit, expiration and redis client
type RedisLimiter struct {
	limit   int
	entity  string
	expire  time.Duration
	cmdable redis.Cmdable
}

// buildKey returns built key for redis storage
func (rl *RedisLimiter) getKey(identifier string) string {
	return fmt.Sprintf("%s:%s:count", rl.entity, identifier)
}

// Allow checks identifier is allowed to continue depending on limit
func (rl *RedisLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	key := rl.getKey(identifier)

	if rl.expire != 0 {
		ttl := rl.cmdable.PTTL(ctx, key).Val()
		if ttl.Milliseconds() < 10 {
			rl.cmdable.Del(ctx, key)
			return true, nil
		}
	}

	count, err := rl.cmdable.Get(ctx, key).Int()

	if err == redis.Nil {
		return true, nil
	}

	if err != nil {
		return true, err
	}

	return count < rl.limit, nil
}

// Seen increments number of actions performed by this particular identifier
func (rl *RedisLimiter) Seen(ctx context.Context, identifier string) error {
	key := rl.getKey(identifier)

	if rl.cmdable.Exists(ctx, key).Val() == 0 {
		return rl.cmdable.Set(ctx, key, 1, rl.expire).Err()
	}

	return rl.cmdable.Incr(ctx, key).Err()
}
