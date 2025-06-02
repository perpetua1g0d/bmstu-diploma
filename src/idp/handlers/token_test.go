package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Реализация jwt.Claims для тестов
type testClaims struct {
	Namespace string
}

func (c testClaims) GetExpirationTime() (*jwt.NumericDate, error) { return nil, nil }
func (c testClaims) GetIssuedAt() (*jwt.NumericDate, error)       { return nil, nil }
func (c testClaims) GetNotBefore() (*jwt.NumericDate, error)      { return nil, nil }
func (c testClaims) GetIssuer() (string, error)                   { return "", nil }
func (c testClaims) GetSubject() (string, error)                  { return "", nil }
func (c testClaims) GetAudience() (jwt.ClaimStrings, error)       { return nil, nil }

// Моки
type mockK8sVerifier struct{ mock.Mock }

func (m *mockK8sVerifier) VerifyWithClient(token string) (string, jwt.Claims, error) {
	args := m.Called(token)
	return args.String(0), args.Get(1).(jwt.Claims), args.Error(2)
}

type mockIssuer struct{ mock.Mock }

func (m *mockIssuer) IssueToken(clientID, scope string) (*IssueResp, error) {
	args := m.Called(clientID, scope)
	if resp := args.Get(0); resp != nil {
		return resp.(*IssueResp), args.Error(1)
	}
	return nil, args.Error(1)
}

type mockRepository struct{ mock.Mock }

func (m *mockRepository) UpdatePermissions(client, scope string, roles []string) error {
	args := m.Called(client, scope, roles)
	return args.Error(0)
}

func (m *mockRepository) GetPermissions(client, scope string) []string {
	args := m.Called(client, scope)
	return args.Get(0).([]string)
}

func TestTokenHandler_Success(t *testing.T) {
	k8sVerifier := new(mockK8sVerifier)
	k8sVerifier.On("VerifyWithClient", "valid-token").Return(
		"client1", testClaims{Namespace: "ns1"}, nil,
	)

	issuer := new(mockIssuer)
	issuer.On("IssueToken", "client1", "scope1").Return(
		&IssueResp{AccessToken: "token123"}, nil,
	)

	ctl := &Controller{
		k8sVerifier: k8sVerifier,
		issuer:      issuer,
	}

	form := url.Values{}
	form.Add("grant_type", grantTypeTokenExchange)
	form.Add("subject_token_type", k8sTokenType)
	form.Add("subject_token", "valid-token")
	form.Add("scope", "scope1")

	req := httptest.NewRequest("POST", "/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler, err := ctl.NewTokenHandler(context.Background())
	require.NoError(t, err)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp IssueResp
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&tokenResp))
	assert.Equal(t, "token123", tokenResp.AccessToken)

	k8sVerifier.AssertExpectations(t)
	issuer.AssertExpectations(t)
}

func TestTokenHandler_InvalidForm(t *testing.T) {
	ctl := &Controller{}

	req := httptest.NewRequest("POST", "/token", nil)
	w := httptest.NewRecorder()

	handler, err := ctl.NewTokenHandler(context.Background())
	require.NoError(t, err)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"unsupported_grant_type"}`)
}

func TestTokenHandler_UnsupportedGrantType(t *testing.T) {
	ctl := &Controller{}

	form := url.Values{}
	form.Add("grant_type", "invalid_grant")

	req := httptest.NewRequest("POST", "/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler, err := ctl.NewTokenHandler(context.Background())
	require.NoError(t, err)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"unsupported_grant_type"}`)
}

func TestTokenHandler_UnsupportedTokenType(t *testing.T) {
	ctl := &Controller{}

	form := url.Values{}
	form.Add("grant_type", grantTypeTokenExchange)
	form.Add("subject_token_type", "invalid_type")

	req := httptest.NewRequest("POST", "/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler, err := ctl.NewTokenHandler(context.Background())
	require.NoError(t, err)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"unsupported_subject_token_type"}`)
}

func TestTokenHandler_TokenVerificationFailed(t *testing.T) {
	k8sVerifier := new(mockK8sVerifier)
	k8sVerifier.On("VerifyWithClient", "invalid-token").Return(
		"", testClaims{}, errors.New("verification failed"), // Исправлено здесь
	)

	ctl := &Controller{k8sVerifier: k8sVerifier}

	form := url.Values{}
	form.Add("grant_type", grantTypeTokenExchange)
	form.Add("subject_token_type", k8sTokenType)
	form.Add("subject_token", "invalid-token")

	req := httptest.NewRequest("POST", "/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler, err := ctl.NewTokenHandler(context.Background())
	require.NoError(t, err)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"token_not_verified"}`)
	k8sVerifier.AssertExpectations(t)
}

func TestTokenHandler_IssuerError(t *testing.T) {
	k8sVerifier := new(mockK8sVerifier)
	k8sVerifier.On("VerifyWithClient", "valid-token").Return(
		"client1", testClaims{}, nil,
	)

	issuer := new(mockIssuer)
	issuer.On("IssueToken", "client1", "scope1").Return(
		nil, errors.New("issuer error"),
	)

	ctl := &Controller{
		k8sVerifier: k8sVerifier,
		issuer:      issuer,
	}

	form := url.Values{}
	form.Add("grant_type", grantTypeTokenExchange)
	form.Add("subject_token_type", k8sTokenType)
	form.Add("subject_token", "valid-token")
	form.Add("scope", "scope1")

	req := httptest.NewRequest("POST", "/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler, err := ctl.NewTokenHandler(context.Background())
	require.NoError(t, err)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), `{"error":"access_denied"}`)
	issuer.AssertExpectations(t)
}

// Кастомный ResponseWriter, который возвращает ошибку
type errorResponseWriter struct {
	http.ResponseWriter
	failAfter int // после скольких записей возвращать ошибку
	count     int
}

func (w *errorResponseWriter) Write(p []byte) (int, error) {
	if w.count >= w.failAfter {
		return 0, errors.New("forced write error")
	}
	w.count++
	return w.ResponseWriter.Write(p)
}
