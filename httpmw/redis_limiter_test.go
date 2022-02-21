package httpmw

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

const limiterTestIdentifier = "john.doe"
const limiterTestEntity = "limiter_test"

func TestLimiter(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s:count", limiterTestEntity, limiterTestIdentifier)
	exp := time.Second * 1

	t.Run("test seen", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(err)
		defer mr.Close()

		cmd := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})
		lmr := NewRedisLimiter(cmd, limiterTestEntity, 10, exp)

		assert.NoError(lmr.Seen(ctx, limiterTestIdentifier))
		assert.Equal("1", cmd.Get(ctx, key).Val())

		assert.NoError(lmr.Seen(ctx, limiterTestIdentifier))
		assert.Equal("2", cmd.Get(ctx, key).Val())

		mr.FastForward(exp)

		assert.NoError(lmr.Seen(ctx, limiterTestIdentifier))
		assert.Equal("1", cmd.Get(ctx, key).Val())
	})

	t.Run("test allow", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(err)
		defer mr.Close()

		cmd := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})
		lmr := NewRedisLimiter(cmd, limiterTestEntity, 2, exp)

		assert.NoError(lmr.Seen(ctx, limiterTestIdentifier))
		assert.Equal("1", cmd.Get(ctx, key).Val())

		assert.NoError(lmr.Seen(ctx, limiterTestIdentifier))
		assert.Equal("2", cmd.Get(ctx, key).Val())

		ok, err := lmr.Allow(ctx, limiterTestIdentifier)
		assert.NoError(err)
		assert.False(ok)

		mr.FastForward(exp)

		ok, err = lmr.Allow(ctx, limiterTestIdentifier)
		assert.NoError(err)
		assert.True(ok)
	})
}
