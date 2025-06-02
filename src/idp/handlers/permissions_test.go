package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdatePermissionsHandler_Success(t *testing.T) {
	repo := new(mockRepository)
	repo.On("UpdatePermissions", "client1", "scope1", []string{"admin"}).Return(nil)

	ctl := &Controller{repository: repo}

	body := `{"client":"client1","scope":"scope1","roles":["admin"]}`
	req := httptest.NewRequest("POST", "/permissions", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()

	handler := ctl.NewUpdatePermissionsHandler(context.Background())
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

func TestUpdatePermissionsHandler_InvalidJSON(t *testing.T) {
	ctl := &Controller{}

	body := `invalid_json`
	req := httptest.NewRequest("POST", "/permissions", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()

	handler := ctl.NewUpdatePermissionsHandler(context.Background())
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")
}

func TestUpdatePermissionsHandler_RepoError(t *testing.T) {
	repo := new(mockRepository)
	repo.On("UpdatePermissions", "client1", "scope1", mock.Anything).Return(errors.New("db error"))

	ctl := &Controller{repository: repo}

	body := `{"client":"client1","scope":"scope1","roles":["admin"]}`
	req := httptest.NewRequest("POST", "/permissions", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()

	handler := ctl.NewUpdatePermissionsHandler(context.Background())
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to update permissions")
	repo.AssertExpectations(t)
}

func TestGetPermissionsHandler_Success(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetPermissions", "client1", "scope1").Return([]string{"admin", "user"})

	ctl := &Controller{repository: repo}

	body := `{"client":"client1","scope":"scope1"}`
	req := httptest.NewRequest("POST", "/permissions", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()

	handler := ctl.NewGetPermissionsHandler(context.Background())
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp PermissionsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, []string{"admin", "user"}, resp.Roles)
	repo.AssertExpectations(t)
}

func TestGetPermissionsHandler_NoPermissions(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetPermissions", "client1", "scope1").Return([]string{})

	ctl := &Controller{repository: repo}

	body := `{"client":"client1","scope":"scope1"}`
	req := httptest.NewRequest("POST", "/permissions", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()

	handler := ctl.NewGetPermissionsHandler(context.Background())
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp PermissionsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Empty(t, resp.Roles)
	repo.AssertExpectations(t)
}

func TestGetPermissionsHandler_InvalidJSON(t *testing.T) {
	ctl := &Controller{}

	body := `invalid_json`
	req := httptest.NewRequest("POST", "/permissions", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()

	handler := ctl.NewGetPermissionsHandler(context.Background())
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")
}

func TestGetPermissionsHandler_MarshalError(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetPermissions", "client1", "scope1").Return([]string{"admin"})

	ctl := &Controller{repository: repo}

	body := `{"client":"client1","scope":"scope1"}`
	req := httptest.NewRequest("POST", "/permissions", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()
	customRW := &errorResponseWriter{ResponseWriter: w, failAfter: 0}

	handler := ctl.NewGetPermissionsHandler(context.Background())
	handler.ServeHTTP(customRW, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code) // Только проверка статуса
}
