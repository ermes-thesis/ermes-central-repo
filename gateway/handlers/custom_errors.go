package handlers

import (
	"net/http"
)

const AuthPrefix = "Bearer "

type HttpError struct {
	message    string
	statusCode int
}

func (e *HttpError) Error() string {
	return e.message
}

func (e *HttpError) StatusCode() int {
	return e.statusCode
}

var (
	NoAuthHeaderErr = HttpError{
		message:    "No authorization header",
		statusCode: http.StatusBadRequest,
	}

	InvalidAuthHeaderErr = HttpError{
		message:    "Invalid authorization header format",
		statusCode: http.StatusBadRequest,
	}

	AuthFailedErr = HttpError{
		message:    "Authentication failed",
		statusCode: http.StatusUnauthorized,
	}
)
