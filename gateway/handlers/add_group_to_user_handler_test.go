package handlers

import (
	db_mock "ermes-central-repo/mocks/db"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createNewLogger(t *testing.T) *logger.Logger {
	log, err := logger.NewLogger(logger.Config{
		Level:     zerolog.DebugLevel,
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

//func TestRetrieveFunctionHandlerErrorGetFun(t *testing.T) {
//	log := createNewLogger(t)
//	minioMock := db_mock.NewMockGenericDB(t)
//	minioMock.EXPECT().GetFunction("fun", "1.0.0").Return(nil, nil, fmt.Errorf("error get function")).Once()
//
//	handler := RetrieveFunctionHandler(minioMock, log)
//
//	req := httptest.NewRequest("POST", "http://localhost:8080/retrieve-function/fun/1.0.0", nil)
//	req.SetPathValue("name", "fun")
//	req.SetPathValue("version", "1.0.0")
//
//	req.Header.Set("Content-Type", "application/json")
//	w := httptest.NewRecorder()
//
//	handler(w, req)
//
//	resp := w.Result()
//	bodyResp, _ := io.ReadAll(resp.Body)
//
//	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
//	assert.Contains(t, string(bodyResp), "Internal Server Error: failed to retrieve function", "Expected error message for malformed JSON, got %s", string(bodyResp))
//}
//
//func TestRetrieveFunctionHandlerErrorGetQueryTemplate(t *testing.T) {
//	log := createNewLogger(t)
//	metadata := model.FunctionMetadata{
//		Name:          "fun",
//		Version:       "1.0.0",
//		TemplateNames: []string{"query"},
//	}
//
//	minioMock := db_mock.NewMockGenericDB(t)
//	minioMock.EXPECT().GetFunction("fun", "1.0.0").Return([]byte("\x01\x02\x03"), &metadata, nil).Once()
//	minioMock.EXPECT().GetQueryTemplate("query").Return(nil, fmt.Errorf("error get query")).Once()
//
//	handler := RetrieveFunctionHandler(minioMock, log)
//
//	req := httptest.NewRequest("POST", "http://localhost:8080/retrieve-function/fun/1.0.0", nil)
//	req.SetPathValue("name", "fun")
//	req.SetPathValue("version", "1.0.0")
//
//	req.Header.Set("Content-Type", "application/json")
//	w := httptest.NewRecorder()
//
//	handler(w, req)
//
//	resp := w.Result()
//	bodyResp, _ := io.ReadAll(resp.Body)
//
//	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
//	assert.Contains(t, string(bodyResp), "Internal Server Error: failed to retrieve query template", "Expected error message for malformed JSON, got %s", string(bodyResp))
//}
//
//func TestRetrieveFunctionHandlerErrorUnmarshalQueryTemplate(t *testing.T) {
//	log := createNewLogger(t)
//	metadata := model.FunctionMetadata{
//		Name:          "fun",
//		Version:       "1.0.0",
//		TemplateNames: []string{"query"},
//	}
//
//	minioMock := db_mock.NewMockGenericDB(t)
//	minioMock.EXPECT().GetFunction("fun", "1.0.0").Return([]byte("\x01\x02\x03"), &metadata, nil).Once()
//	query := `{"malformed"SELECT":"ALL","AS":"private_data","WHERE":{"AND":[{"field":"owner","eq":"<owner_id>"},{"field":"name","eq":"DATI_PRIVATE"}]}}`
//	minioMock.EXPECT().GetQueryTemplate("query").Return([]byte(query), nil).Once()
//
//	handler := RetrieveFunctionHandler(minioMock, log)
//
//	req := httptest.NewRequest("POST", "http://localhost:8080/retrieve-function/fun/1.0.0", nil)
//	req.SetPathValue("name", "fun")
//	req.SetPathValue("version", "1.0.0")
//
//	req.Header.Set("Content-Type", "application/json")
//	w := httptest.NewRecorder()
//
//	handler(w, req)
//
//	resp := w.Result()
//	bodyResp, _ := io.ReadAll(resp.Body)
//
//	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
//	assert.Contains(t, string(bodyResp), "Internal Server Error: failed to parse query template", "Expected error message for malformed JSON, got %s", string(bodyResp))
//}
