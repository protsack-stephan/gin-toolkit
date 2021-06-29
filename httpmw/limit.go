package httpmw

import (
	"bytes"
	"net"
	"strings"
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
// if you pass "inRanges" parameter the limit will be applied only to those IP addresses
// * IP ranges verification in format "192.168.10.1-192.168.10.10,192.168.90.1-192.168.90.10"
func Limit(limit int, ipRanges ...string) gin.HandlerFunc {
	ranges := [][2]net.IP{}

	for _, ips := range ipRanges {
		for _, ips := range strings.Split(ips, ",") {
			ip := strings.Split(ips, "-")

			if len(ip) == 2 {
				ranges = append(ranges, [2]net.IP{net.ParseIP(ip[0]), net.ParseIP(ip[1])})
			}
		}
	}

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
		ipAddr := c.ClientIP()
		visitor := limiter.visitor(ipAddr)

		if len(ranges) > 0 {
			ip := net.ParseIP(ipAddr)

			for _, ips := range ranges {
				if bytes.Compare(ip, ips[0]) >= 0 && bytes.Compare(ip, ips[1]) <= 0 {
					if !visitor.allow() {
						httperr.TooManyRequests(c)
						c.Abort()
						return
					}
					break
				}
			}
		} else if !visitor.allow() {
			httperr.TooManyRequests(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
