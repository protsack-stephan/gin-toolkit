package httpmw

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const cacheTestData = "hello cache"
const cacheTestURL = "/info"
const cacheTestContentType = "application/json; charset=utf-8"
const cacheTestExpire = time.Second * 1

type cacheCmdableMock struct {
	redis.Client
	mock.Mock
}

func (c *cacheCmdableMock) Get(ctx context.Context, key string) *redis.StringCmd {
	args := c.Called(key)
	cmd := redis.NewStringCmd(ctx)
	cmd.SetErr(args.Error(0))
	return cmd
}

func (c *cacheCmdableMock) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := c.Called(key, value, expiration)
	cmd := redis.NewStatusCmd(ctx)
	cmd.SetErr(args.Error(0))
	return cmd
}

func createCacheServer(cache redis.Cmdable, statusCode int) http.Handler {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET(cacheTestURL, Cache(cache, cacheTestExpire, func(c *gin.Context) {
		c.Data(statusCode, cacheTestContentType, []byte(cacheTestData))
	}))

	return router
}

func TestCache(t *testing.T) {
	assert := assert.New(t)

	t.Run("cache write", func(t *testing.T) {
		cmdable := new(cacheCmdableMock)

		srv := httptest.NewServer(createCacheServer(cmdable, http.StatusOK))
		defer srv.Close()

		cmdable.On("Get", cacheTestURL).Return(redis.Nil)
		cmdable.On("Set", cacheTestURL, []byte(cacheTestData), cacheTestExpire).Return(nil)

		res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, cacheTestURL))
		assert.NoError(err)
		defer res.Body.Close()

		data, err := ioutil.ReadAll(res.Body)
		assert.NoError(err)
		assert.Equal(cacheTestData, string(data))
		assert.Equal(http.StatusOK, res.StatusCode)
		cmdable.AssertNumberOfCalls(t, "Get", 1)
		cmdable.AssertNumberOfCalls(t, "Set", 1)
	})

	t.Run("cache get", func(t *testing.T) {
		cmdable := new(cacheCmdableMock)

		srv := httptest.NewServer(createCacheServer(cmdable, http.StatusOK))
		defer srv.Close()

		cmdable.On("Get", cacheTestURL).Return(nil)

		res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, cacheTestURL))
		assert.NoError(err)
		defer res.Body.Close()

		assert.NoError(err)
		assert.Equal(http.StatusOK, res.StatusCode)
		cmdable.AssertNumberOfCalls(t, "Get", 1)
		cmdable.AssertNumberOfCalls(t, "Set", 0)
	})

	t.Run("cache error", func(t *testing.T) {
		cmdable := new(cacheCmdableMock)

		srv := httptest.NewServer(createCacheServer(cmdable, http.StatusInternalServerError))
		defer srv.Close()

		cmdable.On("Get", cacheTestURL).Return(redis.Nil)

		res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, cacheTestURL))
		assert.NoError(err)
		defer res.Body.Close()

		assert.NoError(err)
		assert.Equal(http.StatusInternalServerError, res.StatusCode)
		cmdable.AssertNumberOfCalls(t, "Get", 1)
		cmdable.AssertNumberOfCalls(t, "Set", 0)
	})
}
