package httphandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const statusTestURL = "/status"
const statusTestCache = "redis"
const statusTestDB = "db"

func statusTestServer(services map[string]StatusCheck) http.Handler {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Handle(http.MethodGet, statusTestURL, Status(services))

	return router
}

func TestStatus(t *testing.T) {
	assert := assert.New(t)
	errCache := errors.New("cache is offline")
	srv := httptest.NewServer(statusTestServer(map[string]StatusCheck{
		statusTestDB: func(_ context.Context) error {
			return nil
		},
		statusTestCache: func(ctx context.Context) error {
			return errCache
		},
	}))
	defer srv.Close()

	// make sure uptime at least 1 second
	time.Sleep(time.Second * 1)

	res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, statusTestURL))
	assert.NoError(err)
	assert.Equal(http.StatusOK, res.StatusCode)
	defer res.Body.Close()

	data := new(StatusResponse)
	assert.NoError(json.NewDecoder(res.Body).Decode(data))

	assert.NotZero(data.Uptime)
	assert.Equal(data.Errors[statusTestCache], errCache.Error())
	assert.Equal(data.Errors[statusTestDB], "")
	assert.Equal(data.Online[statusTestDB], true)
	assert.Equal(data.Online[statusTestCache], false)
}
