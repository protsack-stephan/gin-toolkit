package httpmw

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/protsack-stephan/gin-toolkit/httperr"
)

// LimitPerUser middleware to limit number of request per second for user
func LimitPerUser(cmdable redis.Cmdable, limit int) gin.HandlerFunc {
	redisLimiter := NewRedisLimiter(cmdable, limit, time.Second*1)

	return func(c *gin.Context) {
		username, exists := c.Get("username")

		if !exists {
			httperr.Unauthorized(c)
			c.Abort()
			return
		}

		allowed, err := redisLimiter.Allow(c, username.(string), "limit")

		if err != nil {
			httperr.InternalServerError(c)
			c.Abort()
			return
		}

		if !allowed {
			httperr.TooManyRequests(c)
			c.Abort()
			return
		}

		err = redisLimiter.Seen(c, username.(string), "limit")

		if err != nil {
			httperr.InternalServerError(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
