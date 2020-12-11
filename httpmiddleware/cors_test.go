package httpmiddleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var corsTestMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodDelete,
	http.MethodPut,
	http.MethodOptions,
}

const corsTestURL = "/cors"
const corsTestStatus = http.StatusNoContent

func corsTestServer() http.Handler {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS())
	router.Handle(http.MethodGet, corsTestURL, func(c *gin.Context) {
		c.Status(corsTestStatus)
	})

	return router
}

func testCORSResponse(t *testing.T, res *http.Response) {
	defer res.Body.Close()
	assert.Equal(t, corsTestStatus, res.StatusCode)
	assert.Equal(t, res.Header.Get("Access-Control-Allow-Origin"), "*")
	assert.Equal(t, res.Header.Get("Access-Control-Allow-Headers"), "*")
	assert.Equal(t, res.Header.Get("Access-Control-Allow-Methods"), strings.Join(corsTestMethods, ","))
	assert.Equal(t, res.Header.Get("Access-Control-Max-Age"), "86400")
}

func TestCORS(t *testing.T) {
	srv := httptest.NewServer(corsTestServer())
	defer srv.Close()

	res, err := http.Get(srv.URL + corsTestURL)
	assert.NoError(t, err)
	testCORSResponse(t, res)

	req, err := http.NewRequest(http.MethodOptions, srv.URL+corsTestURL, nil)
	assert.NoError(t, err)

	res, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	testCORSResponse(t, res)
}
