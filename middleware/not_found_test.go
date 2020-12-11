package middleware

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const notFoundTestBody = `{"status":%d,"message":"%s"}`
const notFoundTestURL = "/not-found"

func notFoundTestServer() http.Handler {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NotFound())

	return router
}

func TestNotFound(t *testing.T) {
	srv := httptest.NewServer(notFoundTestServer())
	defer srv.Close()

	res, err := http.Get(srv.URL + notFoundTestURL)
	assert.NoError(t, err)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Equal(t, fmt.Sprintf(notFoundTestBody, http.StatusNotFound, http.StatusText(http.StatusNotFound)), string(body))
}
