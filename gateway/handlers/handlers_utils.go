package handlers

import (
	"github.com/ermes-thesis/ermes-wasm-test/wasm-pool"
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

func CheckHTTPReqAuth(r *http.Request, wasmPool *wasm.ExecutorPool) (string, *HttpError) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "ERROR_ID", &NoAuthHeaderErr
	}

	if !strings.HasPrefix(authHeader, AuthPrefix) {
		return "ERROR_ID", &InvalidAuthHeaderErr
	}

	authToken := strings.TrimPrefix(authHeader, AuthPrefix)
	executor := wasmPool.Acquire()
	userId, fnErrCode := executor.InvokeAuthFn("auth", authToken)
	executor.Release()

	if fnErrCode != 200 {
		return "ERROR_ID", &AuthFailedErr
	}

	return userId, nil
}
