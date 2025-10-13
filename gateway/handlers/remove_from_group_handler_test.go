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

func TestRemoveFromGroupHandler(t *testing.T) {
	log := createNewLogger(t)

	redisMock := db_mock.NewGenericDB(t)
	group := "test_group"
	userId := "test_user"
	redisMock.EXPECT().RemoveGroupFromUser(group, userId).Return(nil).Once()

	handler := RemoveFromGroupHandler(redisMock, log)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/group/%s/%s", group, userId), nil)
	req.SetPathValue("groupName", group)
	req.SetPathValue("userId", userId)

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	_, _ = io.ReadAll(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveFromGroupHandler_FailPathValue(t *testing.T) {
	log := createNewLogger(t)
	redisMock := db_mock.NewGenericDB(t)

	group := "test_group"
	userId := "test_user"

	handler := RemoveFromGroupHandler(redisMock, log)

	// Missing groupName path value
	req := httptest.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/group/%s/%s", group, userId), nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	b, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "Missing groupName in path\n", string(b))

	// Missing userId path value
	req = httptest.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/group/%s/%s", group, userId), nil)
	req.SetPathValue("groupName", group)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler(w, req)

	resp = w.Result()
	b, _ = io.ReadAll(resp.Body)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "Missing userId in path\n", string(b))
}

func TestRemoveFromGroupHandler_GroupRemovalFailed(t *testing.T) {
	log := createNewLogger(t)

	redisMock := db_mock.NewGenericDB(t)
	group := "test_group"
	userId := "test_user"
	redisMock.EXPECT().RemoveGroupFromUser(group, userId).Return(errors.New("some error")).Once()

	handler := RemoveFromGroupHandler(redisMock, log)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/group/%s/%s", group, userId), nil)
	req.SetPathValue("groupName", group)
	req.SetPathValue("userId", userId)

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	msg, _ := io.ReadAll(resp.Body)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, fmt.Sprintf("Could not remove user '%s' from group '%s': %v", userId, group, "some error\n"), string(msg))
}
