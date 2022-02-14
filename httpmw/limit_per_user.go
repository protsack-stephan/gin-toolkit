package httpmw

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/protsack-stephan/gin-toolkit/httperr"
)

// LimitPerUser middleware is used to limit number of request per second for user.
// Note:
// if expiration is set to 0 it means the key has no expiration time.
func LimitPerUser(cmdable redis.Cmdable, limit int, expiration time.Duration, groups ...string) gin.HandlerFunc {
	limiter := NewRedisLimiter(cmdable, limit, expiration)

	return func(c *gin.Context) {
		var user *CognitoUser

		if model, ok := c.Get("user"); ok {
			switch model := model.(type) {
			case *CognitoUser:
				user = model
			}
		}

		if user == nil {
			httperr.Unauthorized(c)
			c.Abort()
			return
		}

		if len(groups) > 0 && !user.IsInGroup(groups) {
			c.Next()
			return
		}

		allowed, err := limiter.Allow(c, user.Username, "limit")

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

		if err := limiter.Seen(c, user.Username, "limit"); err != nil {
			httperr.InternalServerError(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
