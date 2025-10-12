package handlers

import (
	caches_mock "ermes-central-repo/mocks/caches"
	db_mock "ermes-central-repo/mocks/db"
	wasm_mock "ermes-central-repo/mocks/ermes-common/wasm-pool"
	wasm_mock_executor "ermes-central-repo/mocks/ermes-common/wasm-pool/executor"
	tokenEncoder_mock "ermes-central-repo/mocks/tokenEncoder"
	"github.com/ermes-thesis/ermes-common/model"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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
	wasmPoolMock := wasm_mock.NewExecutorPool(t)
	wasmExecutorMock := wasm_mock_executor.NewExecutor(t)
	cacheMock := caches_mock.NewFunctionCache(t)
	tokenEncoderMock := tokenEncoder_mock.NewTokenEncoder(t)

	redisMock.EXPECT().GetUserGroups("test_user").Return(groupList, nil).Once()
	cacheMock.EXPECT().GetBinary(fn, fnVer).Return([]byte("very good binary"), nil)
	wasmPoolMock.EXPECT().Acquire().Return(wasmExecutorMock)
	wasmExecutorMock.EXPECT().Invoke([]byte("very good binary"), "authEntry", [][]byte{[]byte(authToken)}).Return(model.WasmResponse{
		HttpCode:     200,
		Response:     userId,
		CacheUpdates: nil,
	})
	wasmExecutorMock.EXPECT().Release().Return()
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

//func TestDeleteGropHandler_FailPathValue(t *testing.T) {
//	log := createNewLogger(t)
//	redisMock := db_mock.NewGenericDB(t)
//
//	group := "test_group"
//
//	handler := CreateGroupHandler(redisMock, log)
//
//	// Missing groupName path value
//	req := httptest.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/group/%s", group), nil)
//	req.Header.Set("Content-Type", "application/json")
//	w := httptest.NewRecorder()
//
//	handler(w, req)
//
//	resp := w.Result()
//	b, _ := io.ReadAll(resp.Body)
//	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
//	require.Equal(t, "Missing groupName in path\n", string(b))
//}
