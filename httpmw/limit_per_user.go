package httpmw

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/protsack-stephan/gin-toolkit/httperr"
)

// LimitPerUser middleware is used to limit number of request per second for user.
// Note: if expiration is set to 0 it means the key has no expiration time.
func LimitPerUser(cmdable redis.Cmdable, limit int, key string, expiration time.Duration, groups ...string) gin.HandlerFunc {
	limiter := NewRedisLimiter(cmdable, fmt.Sprintf("limit:%s", key), limit, expiration)

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

		allowed, err := limiter.Allow(c.Request.Context(), user.Username)

		if err != nil {
			httperr.InternalServerError(c, err.Error())
			c.Abort()
			return
		}

		if !allowed {
			httperr.TooManyRequests(c)
			c.Abort()
			return
		}

		if err := limiter.Seen(c.Request.Context(), user.Username); err != nil {
			httperr.InternalServerError(c, err.Error())
			c.Abort()
			return
		}

		c.Next()
	}
}
