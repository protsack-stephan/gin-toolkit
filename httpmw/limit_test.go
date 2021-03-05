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
const limitTestCount = 3

func createLimitServer(limits int) http.Handler {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(Limit(limits))
	router.GET(limitTestURL, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return router
}

func TestLimit(t *testing.T) {
	assert := assert.New(t)
	srv := httptest.NewServer(createLimitServer(limitTestCount))
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
}
