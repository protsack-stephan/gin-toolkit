package httpmiddleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const authTestURL = "/auth"
const basicAuthUser = "test"
const basicAuthPassword = "user"

func basicAutTestServer() http.Handler {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(BasicAuth(fmt.Sprintf("%s:%s", basicAuthUser, basicAuthPassword)))
	router.Handle(http.MethodGet, authTestURL, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return router
}

func TestBasicAuth(t *testing.T) {
	srv := httptest.NewServer(basicAutTestServer())
	url := srv.URL + authTestURL
	defer srv.Close()

	res, err := http.Get(url)
	assert.NoError(t, err)
	res.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	req.SetBasicAuth(basicAuthUser, basicAuthPassword)

	res, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
