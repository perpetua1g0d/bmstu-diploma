package handlers

import (
	"errors"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSigner struct{ mock.Mock }

func (m *mockSigner) Sign(payload []byte) (*jose.JSONWebSignature, error) {
	args := m.Called(payload)

	jws := &jose.JSONWebSignature{
		Signatures: []jose.Signature{
			{
				Protected: jose.Header{},
				Signature: []byte("signature"),
			},
		},
	}

	// Подменяем CompactSerialize через embedding
	type mockJWS struct {
		*jose.JSONWebSignature
		compact string
	}

	// mocked := &mockJWS{
	// 	JSONWebSignature: jws,
	// 	compact:          args.Get(0).(string),
	// }

	// jws.CompactSerialize = func() (string, error) {
	// 	return mocked.compact, nil
	// }

	return jws, args.Error(1)
}

func (m *mockSigner) Options() jose.SignerOptions {
	return jose.SignerOptions{}
}

// func TestTokenIssuer_IssueToken_Success(t *testing.T) {
// 	repo := new(mockRepository)
// 	repo.On("GetPermissions", "client1", "scope1").Return([]string{"admin"})

// 	signer := new(mockSigner)
// 	signer.On("Sign", mock.Anything).Return("mock.jwt.token", nil)

// 	issuer := &TokenIssuer{
// 		repository: repo,
// 		signer:     signer,
// 		config: &config.Config{
// 			Issuer:   "test-issuer",
// 			TokenTTL: 10 * time.Minute,
// 		},
// 	}

// 	resp, err := issuer.IssueToken("client1", "scope1")
// 	require.NoError(t, err)
// 	assert.Equal(t, "mock.jwt.token", resp.AccessToken)
// 	repo.AssertExpectations(t)
// 	signer.AssertExpectations(t)
// }

// func TestTokenIssuer_IssueToken_NoPermissions(t *testing.T) {
// 	repo := new(mockRepository)
// 	repo.On("GetPermissions", "client1", "scope1").Return([]string{})

// 	signer := new(mockSigner)
// 	signer.On("Sign", mock.Anything).Return("mock.jwt.token", nil)

// 	issuer := &TokenIssuer{
// 		repository: repo,
// 		signer:     signer,
// 		config: &config.Config{
// 			Issuer:   "test-issuer",
// 			TokenTTL: 10 * time.Minute,
// 		},
// 	}

// 	resp, err := issuer.IssueToken("client1", "scope1")
// 	require.NoError(t, err)
// 	assert.Equal(t, "mock.jwt.token", resp.AccessToken)
// 	repo.AssertExpectations(t)
// 	signer.AssertExpectations(t)
// }

func TestTokenIssuer_IssueToken_SignerError(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetPermissions", "client1", "scope1").Return([]string{"admin"})

	signer := new(mockSigner)
	signer.On("Sign", mock.Anything).Return("", errors.New("sign error"))

	issuer := &TokenIssuer{
		repository: repo,
		signer:     signer,
		config: &config.Config{
			Issuer:   "test-issuer",
			TokenTTL: 10 * time.Minute,
		},
	}

	_, err := issuer.IssueToken("client1", "scope1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate jwt")
	repo.AssertExpectations(t)
	signer.AssertExpectations(t)
}

// mockJSONWebSignature реализует jose.JSONWebSignature с поддержкой CompactSerialize
type mockJSONWebSignature struct {
	compact string
}

func (m *mockJSONWebSignature) Verify(key interface{}) ([]byte, error) {
	return nil, nil
}

func (m *mockJSONWebSignature) VerifyMulti(key ...interface{}) ([]byte, error) {
	return nil, nil
}

func (m *mockJSONWebSignature) DetachedVerify(payload []byte, key interface{}) error {
	return nil
}

func (m *mockJSONWebSignature) DetachedVerifyMulti(payload []byte, key ...interface{}) error {
	return nil
}

func (m *mockJSONWebSignature) FullSerialize() string {
	return ""
}

func (m *mockJSONWebSignature) CompactSerialize() (string, error) {
	return m.compact, nil
}

func (m *mockJSONWebSignature) Signatures() []jose.Signature {
	return []jose.Signature{}
}
