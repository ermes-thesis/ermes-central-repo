package handlers

import (
	db_mock "ermes-central-repo/mocks/db"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createNewLogger(t *testing.T) logger.Logger {
	log, err := logger.NewLogger(logger.Config{
		Level:     logger.DebugLevel,
		AddSource: false,
	})
	require.NoError(t, err)
	return log
}

func TestAddToGroupHandler(t *testing.T) {
	log := createNewLogger(t)

	redisMock := db_mock.NewGenericDB(t)
	group := "test_group"
	userId := "test_user"
	redisMock.EXPECT().AddGroupToUser(group, userId).Return(nil).Once()

	handler := AddToGroupHandler(redisMock, log)

	req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:8080/group/%s/%s", group, userId), nil)
	req.SetPathValue("groupName", group)
	req.SetPathValue("userId", userId)

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	_, _ = io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAddToGroupHandler_FailPathValue(t *testing.T) {
	log := createNewLogger(t)
	redisMock := db_mock.NewGenericDB(t)

	group := "test_group"
	userId := "test_user"

	handler := AddToGroupHandler(redisMock, log)

	// Missing groupName path value
	req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:8080/group/%s/%s", group, userId), nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	b, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "Missing groupName in path\n", string(b))

	// Missing userId path value
	req = httptest.NewRequest("POST", fmt.Sprintf("http://localhost:8080/group/%s/%s", group, userId), nil)
	req.SetPathValue("groupName", group)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler(w, req)

	resp = w.Result()
	b, _ = io.ReadAll(resp.Body)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "Missing userId in path\n", string(b))
}

func TestAddToGroupHandler_AddGroupToUserFailed(t *testing.T) {
	log := createNewLogger(t)

	redisMock := db_mock.NewGenericDB(t)
	group := "test_group"
	userId := "test_user"
	redisMock.EXPECT().AddGroupToUser(group, userId).Return(errors.New("some error")).Once()

	handler := AddToGroupHandler(redisMock, log)

	req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:8080/group/%s/%s", group, userId), nil)
	req.SetPathValue("groupName", group)
	req.SetPathValue("userId", userId)

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	msg, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("Internal Server Error: could not add user %s to group %s: %v", userId, group, "some error\n"), string(msg))
}
