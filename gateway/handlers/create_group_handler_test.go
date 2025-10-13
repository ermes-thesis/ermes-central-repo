package handlers

import (
	"bytes"
	"encoding/json"
	db_mock "ermes-central-repo/mocks/db"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGroupHandler(t *testing.T) {
	log := createNewLogger(t)

	redisMock := db_mock.NewGenericDB(t)
	group := "test_group"
	participants := []string{"test_user", "another_one"}
	redisMock.EXPECT().CreateGroup(group, participants).Return(nil).Once()

	handler := CreateGroupHandler(redisMock, log)

	bodyToSend, err := json.Marshal(participants)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:8080/group/%s", group), bytes.NewBuffer(bodyToSend))
	req.SetPathValue("groupName", group)

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	_, _ = io.ReadAll(resp.Body)

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestCreateGroupHandler_FailPathValue(t *testing.T) {
	log := createNewLogger(t)
	redisMock := db_mock.NewGenericDB(t)

	group := "test_group"
	participants := []string{"test_user", "another_one"}

	handler := CreateGroupHandler(redisMock, log)

	// Missing groupName path value
	bodyToSend, err := json.Marshal(participants)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:8080/group/%s", group), bytes.NewBuffer(bodyToSend))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	b, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "Missing groupName in path\n", string(b))

	// Bad participant list formatting
	bodyToSend, err = json.Marshal("bad")
	require.NoError(t, err)
	req = httptest.NewRequest("POST", fmt.Sprintf("http://localhost:8080/group/%s", group), bytes.NewBuffer(bodyToSend))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("groupName", group)
	w = httptest.NewRecorder()

	handler(w, req)

	resp = w.Result()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateGroupHandler_GroupCreationFailed(t *testing.T) {
	log := createNewLogger(t)

	redisMock := db_mock.NewGenericDB(t)
	group := "test_group"
	participants := []string{"test_user", "another_one"}
	redisMock.EXPECT().CreateGroup(group, participants).Return(errors.New("some error")).Once()

	handler := CreateGroupHandler(redisMock, log)

	bodyToSend, err := json.Marshal(participants)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", fmt.Sprintf("http://localhost:8080/group/%s", group), bytes.NewBuffer(bodyToSend))
	req.SetPathValue("groupName", group)

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	msg, _ := io.ReadAll(resp.Body)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, fmt.Sprintf("Could not create group '%s': %v", group, "some error\n"), string(msg))
}
