package handlers

import (
	caches_mock "ermes-central-repo/mocks/caches"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	db_mock "ermes-central-repo/mocks/db"
	wasm_mock "ermes-central-repo/mocks/ermes-common/github.com/ermes-thesis/ermes-common/wasm-pool"
	tokenEncoder_mock "ermes-central-repo/mocks/tokenEncoder"

	"github.com/ermes-thesis/ermes-common/model"
	"github.com/stretchr/testify/require"
)

func TestGroupTokenHandler(t *testing.T) {
	log := createNewLogger(t)

	fn := "test_func"
	fnVer := "1.2.3"
	groupList := []string{"group1", "group2", "group3"}
	userId := "test_user"
	authFunction := fn + "@" + fnVer
	authToken := "authToken"

	redisMock := db_mock.NewGenericDB(t)
	wasmPoolMock := wasm_mock.NewWASP(t)
	cacheMock := caches_mock.NewFunctionCache(t)
	tokenEncoderMock := tokenEncoder_mock.NewTokenEncoder(t)

	redisMock.EXPECT().GetUserGroups("test_user").Return(groupList, nil).Once()
	cacheMock.EXPECT().GetBinary(fn, fnVer).Return([]byte("very good binary"), nil)
	wasmPoolMock.EXPECT().InvokeFunction(authFunction, []byte("very good binary"), "authEntry", [][]byte{[]byte(authToken)}).Return(model.WasmResponse{
		HttpCode:     200,
		Response:     userId,
		CacheUpdates: nil,
	})
	tokenEncoderMock.EXPECT().EncodeGroupToken(groupList, userId).Return("groupToken", nil)

	handler := RetrieveGroupToken(redisMock, wasmPoolMock, tokenEncoderMock, &authFunction, log, cacheMock)

	req := httptest.NewRequest("GET", "http://localhost:8080/token", nil)
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	_, _ = io.ReadAll(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroupTokenHandler_MissingAuthorizationHeader(t *testing.T) {
	log := createNewLogger(t)

	fn := "test_func"
	fnVer := "1.2.3"
	authFunction := fn + "@" + fnVer

	redisMock := db_mock.NewGenericDB(t)
	wasmPoolMock := wasm_mock.NewWASP(t)
	cacheMock := caches_mock.NewFunctionCache(t)
	tokenEncoderMock := tokenEncoder_mock.NewTokenEncoder(t)

	handler := RetrieveGroupToken(redisMock, wasmPoolMock, tokenEncoderMock, &authFunction, log, cacheMock)

	req := httptest.NewRequest("GET", "http://localhost:8080/token", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	msg, _ := io.ReadAll(resp.Body)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "No authorization header\n", string(msg))
}

func TestGroupTokenHandler_MissingAuthPrefix(t *testing.T) {
	log := createNewLogger(t)

	fn := "test_func"
	fnVer := "1.2.3"
	authFunction := fn + "@" + fnVer

	redisMock := db_mock.NewGenericDB(t)
	wasmPoolMock := wasm_mock.NewWASP(t)
	cacheMock := caches_mock.NewFunctionCache(t)
	tokenEncoderMock := tokenEncoder_mock.NewTokenEncoder(t)

	handler := RetrieveGroupToken(redisMock, wasmPoolMock, tokenEncoderMock, &authFunction, log, cacheMock)

	req := httptest.NewRequest("GET", "http://localhost:8080/token", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "No bearer prefix")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	msg, _ := io.ReadAll(resp.Body)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "Invalid authorization header format\n", string(msg))
}

func TestGroupTokenHandler_InvalidAuthFunction(t *testing.T) {
	log := createNewLogger(t)

	authFunction := ""

	redisMock := db_mock.NewGenericDB(t)
	wasmPoolMock := wasm_mock.NewWASP(t)
	cacheMock := caches_mock.NewFunctionCache(t)
	tokenEncoderMock := tokenEncoder_mock.NewTokenEncoder(t)

	handler := RetrieveGroupToken(redisMock, wasmPoolMock, tokenEncoderMock, &authFunction, log, cacheMock)

	req := httptest.NewRequest("GET", "http://localhost:8080/token", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	msg, _ := io.ReadAll(resp.Body)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, "Invalid authentication function\n", string(msg))
}

func TestGroupTokenHandler_AuthenticationFailed(t *testing.T) {
	log := createNewLogger(t)

	fn := "test_func"
	fnVer := "1.2.3"
	userId := "test_user"
	authFunction := fn + "@" + fnVer
	authToken := "authToken"

	redisMock := db_mock.NewGenericDB(t)
	wasmPoolMock := wasm_mock.NewWASP(t)
	cacheMock := caches_mock.NewFunctionCache(t)
	tokenEncoderMock := tokenEncoder_mock.NewTokenEncoder(t)

	cacheMock.EXPECT().GetBinary(fn, fnVer).Return([]byte("very good binary"), nil)
	wasmPoolMock.EXPECT().InvokeFunction(authFunction, []byte("very good binary"), "authEntry", [][]byte{[]byte(authToken)}).Return(model.WasmResponse{
		HttpCode:     500,
		Response:     userId,
		CacheUpdates: nil,
	})

	handler := RetrieveGroupToken(redisMock, wasmPoolMock, tokenEncoderMock, &authFunction, log, cacheMock)

	req := httptest.NewRequest("GET", "http://localhost:8080/token", nil)
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	msg, _ := io.ReadAll(resp.Body)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Equal(t, "Authentication failed\n", string(msg))
}

func TestGroupTokenHandler_UserGroupsRetrievalFailed(t *testing.T) {
	log := createNewLogger(t)

	fn := "test_func"
	fnVer := "1.2.3"
	userId := "test_user"
	authFunction := fn + "@" + fnVer
	authToken := "authToken"

	redisMock := db_mock.NewGenericDB(t)
	wasmPoolMock := wasm_mock.NewWASP(t)
	cacheMock := caches_mock.NewFunctionCache(t)
	tokenEncoderMock := tokenEncoder_mock.NewTokenEncoder(t)

	cacheMock.EXPECT().GetBinary(fn, fnVer).Return([]byte("very good binary"), nil)
	wasmPoolMock.EXPECT().InvokeFunction(authFunction, []byte("very good binary"), "authEntry", [][]byte{[]byte(authToken)}).Return(model.WasmResponse{
		HttpCode:     200,
		Response:     userId,
		CacheUpdates: nil,
	})
	redisMock.EXPECT().GetUserGroups("test_user").Return(nil, errors.New("some error")).Once()

	handler := RetrieveGroupToken(redisMock, wasmPoolMock, tokenEncoderMock, &authFunction, log, cacheMock)

	req := httptest.NewRequest("GET", "http://localhost:8080/token", nil)
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	msg, _ := io.ReadAll(resp.Body)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, "Could not retrieve groups for user '"+userId+"': some error\n", string(msg))
}

func TestGroupTokenHandler_EncodeGroupTokenFailed(t *testing.T) {
	log := createNewLogger(t)

	fn := "test_func"
	fnVer := "1.2.3"
	groupList := []string{"group1", "group2", "group3"}
	userId := "test_user"
	authFunction := fn + "@" + fnVer
	authToken := "authToken"

	redisMock := db_mock.NewGenericDB(t)
	wasmPoolMock := wasm_mock.NewWASP(t)
	cacheMock := caches_mock.NewFunctionCache(t)
	tokenEncoderMock := tokenEncoder_mock.NewTokenEncoder(t)

	cacheMock.EXPECT().GetBinary(fn, fnVer).Return([]byte("very good binary"), nil)
	wasmPoolMock.EXPECT().InvokeFunction(authFunction, []byte("very good binary"), "authEntry", [][]byte{[]byte(authToken)}).Return(model.WasmResponse{
		HttpCode:     200,
		Response:     userId,
		CacheUpdates: nil,
	})
	redisMock.EXPECT().GetUserGroups("test_user").Return(groupList, nil).Once()
	tokenEncoderMock.EXPECT().EncodeGroupToken(groupList, userId).Return("", errors.New("some error")).Once()

	handler := RetrieveGroupToken(redisMock, wasmPoolMock, tokenEncoderMock, &authFunction, log, cacheMock)

	req := httptest.NewRequest("GET", "http://localhost:8080/token", nil)
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	msg, _ := io.ReadAll(resp.Body)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, "Failed to generate group token: some error\n", string(msg))
}
