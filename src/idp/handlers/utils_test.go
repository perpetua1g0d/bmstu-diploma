package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()
	respondError(w, "test error", http.StatusBadRequest)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errResp RespErr
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test error", errResp.Error)
}

func TestBaseMetricsMiddleware(t *testing.T) {
	// Тест для проверки, что middleware не ломает обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := baseMetricsMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
