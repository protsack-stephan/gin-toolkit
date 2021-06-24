package httpmw

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const limitTestURL = "/some-url"
const limitTestIPRanges = "192.168.10.1-192.168.10.10,192.168.20.1-192.168.20.10"
const limitTestCount = 3

func createLimitServer(limits int, ipRanges string) http.Handler {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(Limit(limits, ipRanges))
	router.GET(limitTestURL, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return router
}

func TestLimit(t *testing.T) {
	assert := assert.New(t)

	t.Run("without IP restriction", func(t *testing.T) {
		srv := httptest.NewServer(createLimitServer(limitTestCount, ""))
		defer srv.Close()

		for i := 0; i < limitTestCount+1; i++ {
			res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, limitTestURL))
			assert.NoError(err)

			if i < limitTestCount {
				assert.Equal(http.StatusOK, res.StatusCode)
			} else {
				assert.Equal(http.StatusTooManyRequests, res.StatusCode)
			}
		}
	})

	t.Run("with IP restriction", func(t *testing.T) {
		srv := httptest.NewServer(createLimitServer(limitTestCount, limitTestIPRanges))
		defer srv.Close()

		for i := 0; i < limitTestCount+1; i++ {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", srv.URL, limitTestURL), nil)
			assert.NoError(err)
			req.Header.Add("X-Forwarded-For", "192.168.10.2")

			res, err := http.DefaultClient.Do(req)
			assert.NoError(err)

			if i < limitTestCount {
				assert.Equal(http.StatusOK, res.StatusCode)
			} else {
				assert.Equal(http.StatusTooManyRequests, res.StatusCode)
			}
		}

		for i := 0; i < limitTestCount+100; i++ {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", srv.URL, limitTestURL), nil)
			assert.NoError(err)

			res, err := http.DefaultClient.Do(req)
			assert.NoError(err)
			assert.Equal(http.StatusOK, res.StatusCode)
		}
	})
}
