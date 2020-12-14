package httperr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const httperrTestStatus = http.StatusNotFound
const httperrTestMessage = "not found"

const httperrTestNotFoundURL = "/not-found"
const httperrInternalServerErrorURL = "/internal-error"
const httperrUnprocessableEntityURL = "/unprocessable-entity"
const httperrBadRequestURL = "/bad-request"
const httperrUnauthorizedURL = "/unauthorized"
const httperrForbiddenURL = "/forbidden"

func creatErrorTestServer() http.Handler {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Handle(http.MethodGet, httperrTestNotFoundURL, func(c *gin.Context) {
		NotFound(c)
	})

	router.Handle(http.MethodGet, httperrInternalServerErrorURL, func(c *gin.Context) {
		InternalServerError(c)
	})

	router.Handle(http.MethodGet, httperrUnprocessableEntityURL, func(c *gin.Context) {
		UnprocessableEntity(c)
	})

	router.Handle(http.MethodGet, httperrBadRequestURL, func(c *gin.Context) {
		BadRequest(c)
	})

	router.Handle(http.MethodGet, httperrUnauthorizedURL, func(c *gin.Context) {
		Unauthorized(c)
	})

	router.Handle(http.MethodGet, httperrForbiddenURL, func(c *gin.Context) {
		Forbidden(c)
	})

	return router
}

func TestError(t *testing.T) {
	err := NewError(httperrTestStatus, httperrTestMessage)
	assert.NotNil(t, err)
	assert.Equal(t, httperrTestStatus, err.Status)
	assert.Equal(t, httperrTestMessage, err.Message)
}

func TestErrors(t *testing.T) {
	srv := httptest.NewServer(creatErrorTestServer())
	defer srv.Close()

	for _, test := range []struct {
		URL   string
		Error *Error
	}{
		{
			httperrTestNotFoundURL,
			NewError(http.StatusNotFound, http.StatusText(http.StatusNotFound)),
		},
		{
			httperrInternalServerErrorURL,
			NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)),
		},
		{
			httperrUnprocessableEntityURL,
			NewError(http.StatusUnprocessableEntity, http.StatusText(http.StatusUnprocessableEntity)),
		},
		{
			httperrBadRequestURL,
			NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest)),
		},
		{
			httperrUnauthorizedURL,
			NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)),
		},
		{
			httperrForbiddenURL,
			NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden)),
		},
	} {
		res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, test.URL))
		assert.NoError(t, err)
		defer res.Body.Close()

		data, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err)

		error := new(Error)
		err = json.Unmarshal(data, error)
		assert.NoError(t, err)

		assert.Equal(t, test.Error.Status, error.Status)
		assert.Equal(t, test.Error.Message, error.Message)
		assert.Equal(t, test.Error.Status, res.StatusCode)
	}
}
