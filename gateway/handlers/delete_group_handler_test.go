package handlers

import (
	db_mock "ermes-central-repo/mocks/db"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteGroupHandler(t *testing.T) {
	log := createNewLogger(t)

	redisMock := db_mock.NewGenericDB(t)
	group := "test_group"
	redisMock.EXPECT().DeleteGroup(group).Return(nil).Once()

	handler := DeleteGroupHandler(redisMock, log)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/group/%s", group), nil)
	req.SetPathValue("groupName", group)

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	_, _ = io.ReadAll(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeleteGroupHandler_FailPathValue(t *testing.T) {
	log := createNewLogger(t)
	redisMock := db_mock.NewGenericDB(t)

	group := "test_group"

	handler := DeleteGroupHandler(redisMock, log)

	// Missing groupName path value
	req := httptest.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/group/%s", group), nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	b, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "Missing groupName in path\n", string(b))
}

func TestDeleteGroupHandler_GroupDeletionFailed(t *testing.T) {
	log := createNewLogger(t)

	redisMock := db_mock.NewGenericDB(t)
	group := "test_group"
	redisMock.EXPECT().DeleteGroup(group).Return(errors.New("some error")).Once()

	handler := DeleteGroupHandler(redisMock, log)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/group/%s", group), nil)
	req.SetPathValue("groupName", group)

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	msg, _ := io.ReadAll(resp.Body)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, "Could not delete group '"+group+"': some error\n", string(msg))
}
