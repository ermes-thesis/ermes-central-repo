package handlers

import (
	caches_mock "ermes-central-repo/mocks/caches"
	"errors"
	"fmt"
	"github.com/ermes-thesis/ermes-common/model"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetAuthFunctionHandler(t *testing.T) {
	log := createNewLogger(t)

	cacheMock := caches_mock.NewFunctionCache(t)
	fn := "test_func"
	fnVer := "1.0.0"

	cacheMock.EXPECT().GetMetadata(fn, fnVer).Return(&model.FunctionMetadata{
		Name:                  fn,
		Version:               fnVer,
		TemplateNames:         nil,
		DeployedOn:            nil,
		StrongCollWriteAccess: false,
		IsAuth:                true,
	}, nil).Once()
	cacheMock.EXPECT().GetMetadata(fn, "latest").Return(&model.FunctionMetadata{
		Name:                  fn,
		Version:               "1.2.3",
		TemplateNames:         nil,
		DeployedOn:            nil,
		StrongCollWriteAccess: false,
		IsAuth:                true,
	}, nil).Once()

	authFunction := "authFunction"
	handler := SetAuthFunctionHandler(cacheMock, &authFunction, log)

	req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:8080/set-auth-function/%s/%s", fn, fnVer), nil)
	req.SetPathValue("name", fn)
	req.SetPathValue("version", fnVer)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	_, _ = io.ReadAll(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, fn+"@"+fnVer, authFunction)

	// No version path value for latest function
	req = httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:8080/set-auth-function/%s/%s", fn, fnVer), nil)
	req.SetPathValue("name", fn)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler(w, req)

	resp = w.Result()
	_, _ = io.ReadAll(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, fn+"@1.2.3", authFunction)
}

func TestSetAuthFunctionHandler_FailPathValue(t *testing.T) {
	log := createNewLogger(t)

	cacheMock := caches_mock.NewFunctionCache(t)
	fn := "test_func"
	fnVer := "1.0.0"

	authFunction := "authFunction"
	handler := SetAuthFunctionHandler(cacheMock, &authFunction, log)

	// Missing functionName path value
	req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:8080/set-auth-function/%s/%s", fn, fnVer), nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	b, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "Missing function name in path\n", string(b))
}

func TestSetAuthFunctionHandler_MetadataRetrievalFailed(t *testing.T) {
	log := createNewLogger(t)

	cacheMock := caches_mock.NewFunctionCache(t)
	fn := "test_func"
	fnVer := "1.0.0"

	cacheMock.EXPECT().GetMetadata(fn, fnVer).Return(nil, errors.New("some error")).Once()

	authFunction := "authFunction"
	handler := SetAuthFunctionHandler(cacheMock, &authFunction, log)

	req := httptest.NewRequest("PUT", fmt.Sprintf("http://localhost:8080/set-auth-function/%s/%s", fn, fnVer), nil)
	req.SetPathValue("name", fn)
	req.SetPathValue("version", fnVer)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	msg, _ := io.ReadAll(resp.Body)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	require.Equal(t, fmt.Sprintf("Function '%s@%s' not found\n", fn, fnVer), string(msg))
}
