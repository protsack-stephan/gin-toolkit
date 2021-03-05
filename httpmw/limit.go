package httpmw

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/protsack-stephan/gin-toolkit/httperr"
	"golang.org/x/time/rate"
)

type visitor struct {
	reqTime time.Time
	limit   *rate.Limiter
}

func (v *visitor) allow() bool {
	v.reqTime = time.Now()
	return v.limit.Allow()
}

func (v *visitor) seen() time.Duration {
	return time.Since(v.reqTime)
}

type limiter struct {
	visitors *sync.Map
	limit    int
}

func (l *limiter) visitor(ip string) *visitor {
	entity, _ := l.visitors.LoadOrStore(ip, &visitor{
		time.Now(),
		rate.NewLimiter(1, l.limit),
	})

	return entity.(*visitor)
}

func (l *limiter) cleanup() {
	l.visitors.Range(func(key interface{}, val interface{}) bool {
		if val.(*visitor).seen() > time.Minute*1 {
			l.visitors.Delete(key)
		}

		return true
	})
}

// Limit middleware to limit number of request per second
func Limit(limit int) gin.HandlerFunc {
	limiter := &limiter{
		&sync.Map{},
		limit,
	}

	go func() {
		for {
			time.Sleep(time.Minute * 1)
			limiter.cleanup()
		}
	}()

	return func(c *gin.Context) {
		if !limiter.visitor(c.ClientIP()).allow() {
			httperr.TooManyRequests(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
