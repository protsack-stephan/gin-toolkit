package httpmw

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

const limitTestUserName = "alex_test"
const limitTestUserGroup = "admin"
const limitTestUserAnotherGroup = "usergroup"
const limitTestUrl = "/some-url"
const limitTestLimit = 3
const limitTestKey = "/v1/test"

func createRedisLimitServer(cmd redis.Cmdable, exp time.Duration, group string, groups ...string) http.Handler {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		user := new(CognitoUser)
		user.SetUsername(limitTestUserName)
		user.SetGroups([]string{group})
		c.Set("user", user)
	})

	router.Use(LimitPerUser(cmd, limitTestLimit, limitTestKey, exp, groups...))

	router.GET(limitTestUrl, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return router
}

func TestLimitPerUser(t *testing.T) {
	assert := assert.New(t)
	exp := time.Second * 1

	t.Run("test limit", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(err)
		defer mr.Close()

		cmd := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		srv := httptest.NewServer(createRedisLimitServer(cmd, exp, limitTestUserGroup, []string{limitTestUserGroup}...))
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

		mr.FastForward(exp)
		res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, limitTestUrl))
		assert.NoError(err)
		assert.Equal(http.StatusOK, res.StatusCode)
	})

	t.Run("test no limit", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(err)
		defer mr.Close()

		cmd := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		srv := httptest.NewServer(createRedisLimitServer(cmd, exp, limitTestUserAnotherGroup, []string{limitTestUserGroup}...))
		defer srv.Close()

		for i := 0; i < limitTestCount+1; i++ {
			res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, limitTestUrl))
			assert.NoError(err)
			assert.Equal(http.StatusOK, res.StatusCode)
		}
	})
}
