package modules

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const initTestModule = "test"
const initTestURL = "/hello"
const initTestStatus = http.StatusOK
const initTestResponse = "hello world"

var initTestModules = []Module{
	{
		Path: initTestModule,
		Routes: []Route{
			{
				Path:   initTestURL,
				Method: http.MethodGet,
				Handler: func(c *gin.Context) {
					c.String(initTestStatus, initTestResponse)
				},
			},
		},
	},
}

func TestInit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	err := Init(router, initTestModules)
	assert.NoError(t, err)

	srv := httptest.NewServer(router)
	defer srv.Close()

	res, err := http.Get(fmt.Sprintf("%s/%s%s", srv.URL, initTestModule, initTestURL))
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, initTestResponse, string(data))
}
