package httperr

import "net/http"

// NotFound http not found error
var NotFound = NewError(http.StatusNotFound, http.StatusText(http.StatusNotFound))

// InternalServerError http internal server error
var InternalServerError = NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))

// UnprocessableEntity http unprocessable entity
var UnprocessableEntity = NewError(http.StatusUnprocessableEntity, http.StatusText(http.StatusUnprocessableEntity))

// BadRequest http bad request
var BadRequest = NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))

// Unauthorized http unauthorized
var Unauthorized = NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))

// Forbidden htt forbidden
var Forbidden = NewError(http.StatusForbidden, http.StatusText(http.StatusForbidden))

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
