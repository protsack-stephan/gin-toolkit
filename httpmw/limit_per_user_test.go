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

func createRedisLimitServer(cmdable redis.Cmdable, groups ...string) http.Handler {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	limitTestUserGroups := make(map[string]struct{})
	limitTestUserGroups[limitTestUserGroup] = struct{}{}

	limitTestUser := &CognitoUser{
		Username: limitTestUserName,
		Groups:   limitTestUserGroups,
	}

	router.Use(func(c *gin.Context) {
		c.Set("user", limitTestUser)
	})
	if len(groups) > 0 {
		router.Use(LimitPerUser(cmdable, limitTestLimit, time.Second*1, groups...))
	} else {
		router.Use(LimitPerUser(cmdable, limitTestLimit, time.Second*1))
	}
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
	limitGroups := []string{limitTestUserGroup}

	srv := httptest.NewServer(createRedisLimitServer(cmdable, limitGroups...))
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
func TestLimitPerUserInGroup(t *testing.T) {
	assert := assert.New(t)

	mr, err := miniredis.Run()
	assert.NoError(err)
	defer mr.Close()

	cmdable := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	limitGroups := []string{limitTestUserGroup}

	srv := httptest.NewServer(createRedisLimitServer(cmdable, limitGroups...))
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
func TestLimitPerUserInAnotherGroup(t *testing.T) {
	assert := assert.New(t)

	mr, err := miniredis.Run()
	assert.NoError(err)
	defer mr.Close()

	cmdable := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	limitGroups := []string{limitTestUserAnotherGroup}

	srv := httptest.NewServer(createRedisLimitServer(cmdable, limitGroups...))
	defer srv.Close()

	for i := 0; i < limitTestCount+1; i++ {
		res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, limitTestUrl))

		assert.NoError(err)
		assert.Equal(http.StatusOK, res.StatusCode)
	}
}
