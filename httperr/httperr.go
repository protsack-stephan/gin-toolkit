package httperr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error http error struct
type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// NewError create http response error
func NewError(status int, message string) *Error {
	return &Error{
		status,
		message,
	}
}

// NotFound http not found error
func NotFound(c *gin.Context, error ...string) {
	err := NewError(http.StatusNotFound, http.StatusText(http.StatusNotFound))

	if len(error) > 0 {
		err.Message = error[0]
	}

	c.JSON(err.Status, err)
}

// InternalServerError http internal server error
func InternalServerError(c *gin.Context, error ...string) {
	err := NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))

	if len(error) > 0 {
		err.Message = error[0]
	}

	c.JSON(err.Status, err)
}

// UnprocessableEntity http unprocessable entity
func UnprocessableEntity(c *gin.Context, error ...string) {
	err := NewError(http.StatusUnprocessableEntity, http.StatusText(http.StatusUnprocessableEntity))

	if len(error) > 0 {
		err.Message = error[0]
	}

	c.JSON(err.Status, err)
}

// BadRequest http bad request
func BadRequest(c *gin.Context, error ...string) {
	err := NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))

	if len(error) > 0 {
		err.Message = error[0]
	}

	c.JSON(err.Status, err)
}

// Unauthorized http unauthorized
func Unauthorized(c *gin.Context, error ...string) {
	err := NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))

	if len(error) > 0 {
		err.Message = error[0]
	}

	c.JSON(err.Status, err)
}

// Forbidden htt forbidden
func Forbidden(c *gin.Context, error ...string) {
	err := NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))

	if len(error) > 0 {
		err.Message = error[0]
	}

	c.JSON(err.Status, err)
}

// TooManyRequests http to many requests
func TooManyRequests(c *gin.Context, error ...string) {
	err := NewError(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))

	if len(error) > 0 {
		err.Message = error[0]
	}

	c.JSON(err.Status, err)
}
