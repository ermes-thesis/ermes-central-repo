package handlers

import (
	"net/http"
	"strings"
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

func CheckHTTPReqAuth(r *http.Request) (int, *HttpError) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, &NoAuthHeaderErr
	}

	if !strings.HasPrefix(authHeader, AuthPrefix) {
		return 0, &InvalidAuthHeaderErr
	}

	userId, err := executeAuthFunction()
	if err != nil {
		return 0, &AuthFailedErr
	}

	return userId, nil
}

func executeAuthFunction() (int, error) {
	return 2, nil
}
