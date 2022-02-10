package httpmw

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

const limitTestUser = "alex_test"
const limitTestUrl = "/some-url"
const limitTestLimit = 3

func createRedisLimitServer(cmdable redis.Cmdable) http.Handler {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("username", limitTestUser)
	})
	router.Use(LimitPerUser(cmdable, limitTestLimit))
	router.GET(limitTestUrl, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return router
}

func TestLimitPerUser(t *testing.T) {
	assert := assert.New(t)

	mr, err := miniredis.Run()
	assert.NoError(err)
	defer mr.Close()

	cmdable := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	srv := httptest.NewServer(createRedisLimitServer(cmdable))
	defer srv.Close()

	for i := 0; i < limitTestCount+1; i++ {
		res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, limitTestUrl))

		assert.NoError(err)

		if i < limitTestCount {
			assert.Equal(http.StatusOK, res.StatusCode)
		} else {
			assert.Equal(http.StatusTooManyRequests, res.StatusCode)
		}
	}
}
