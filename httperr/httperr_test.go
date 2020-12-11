package httperr

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const errTestStatus = http.StatusNotFound
const errTestMessage = "not found"

func TestError(t *testing.T) {
	err := NewError(errTestStatus, errTestMessage)

	assert.NotNil(t, err)
	assert.Equal(t, errTestStatus, err.Status)
	assert.Equal(t, errTestMessage, err.Message)

	assert.Equal(t, http.StatusInternalServerError, InternalServerError.Status)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), InternalServerError.Message)

	assert.Equal(t, http.StatusNotFound, NotFound.Status)
	assert.Equal(t, http.StatusText(http.StatusNotFound), NotFound.Message)

	assert.Equal(t, http.StatusUnprocessableEntity, UnprocessableEntity.Status)
	assert.Equal(t, http.StatusText(http.StatusUnprocessableEntity), UnprocessableEntity.Message)

	assert.Equal(t, http.StatusBadRequest, BadRequest.Status)
	assert.Equal(t, http.StatusText(http.StatusBadRequest), BadRequest.Message)

	assert.Equal(t, http.StatusUnauthorized, Unauthorized.Status)
	assert.Equal(t, http.StatusText(http.StatusUnauthorized), Unauthorized.Message)

	assert.Equal(t, http.StatusForbidden, Forbidden.Status)
	assert.Equal(t, http.StatusText(http.StatusForbidden), Forbidden.Message)
}
