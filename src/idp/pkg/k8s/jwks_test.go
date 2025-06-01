package k8s

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	defaultTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultCAPath    = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

func TestMakeRSAPublicKey(t *testing.T) {
	tests := []struct {
		name    string
		jwk     JWK
		wantErr bool
	}{
		{
			name: "valid jwk",
			jwk: JWK{
				Kty: "RSA",
				Kid: "test",
				Use: "sig",
				Alg: "RS256",
				N:   base64.RawURLEncoding.EncodeToString(big.NewInt(12345).Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(big.NewInt(65537).Bytes()),
			},
			wantErr: false,
		},
		{
			name: "invalid base64 in N",
			jwk: JWK{
				N: "invalid base64",
				E: base64.RawURLEncoding.EncodeToString(big.NewInt(65537).Bytes()),
			},
			wantErr: true,
		},
		{
			name: "invalid base64 in E",
			jwk: JWK{
				N: base64.RawURLEncoding.EncodeToString(big.NewInt(12345).Bytes()),
				E: "invalid base64",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := makeRSAPublicKey(tt.jwk)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetPublicKey(t *testing.T) {
	// mock Kubernetes API server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"keys":[{"kty":"RSA","kid":"test","use":"sig","alg":"RS256","n":"test","e":"AQAB"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	oldURL := os.Getenv("KUBERNETES_SERVICE_HOST")
	os.Setenv("KUBERNETES_SERVICE_HOST", ts.URL[7:]) // strip "http://"
	defer os.Setenv("KUBERNETES_SERVICE_HOST", oldURL)

	t.Run("successful key retrieval", func(t *testing.T) {
		tokenFile, err := os.CreateTemp("", "token")
		require.NoError(t, err)
		defer os.Remove(tokenFile.Name())
		_, err = tokenFile.WriteString("test-token")
		require.NoError(t, err)
		tokenFile.Close()

		caFile, err := os.CreateTemp("", "ca.crt")
		require.NoError(t, err)
		defer os.Remove(caFile.Name())
		_, err = caFile.WriteString("-----BEGIN CERTIFICATE-----\nTEST\n-----END CERTIFICATE-----")
		require.NoError(t, err)
		caFile.Close()

		readSecrets := func(name string) ([]byte, error) {
			switch name {
			case "/var/run/secrets/kubernetes.io/serviceaccount/token":
				return os.ReadFile(tokenFile.Name())
			case "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt":
				return os.ReadFile(caFile.Name())
			default:
				return nil, fmt.Errorf("file not found: %s", name)
			}
		}

		k8sClient := &K8sClient{
			readSecrets: readSecrets,
			client:      ts.Client(),
		}

		err = k8sClient.setup()
		require.NoError(t, err)

		k8sClient.jwksURL = ts.URL

		key, err := k8sClient.GetPublicKey()
		assert.NoError(t, err)
		assert.IsType(t, &rsa.PublicKey{}, key)
	})

	t.Run("missing token file", func(t *testing.T) {
		k8sClient := &K8sClient{
			readSecrets: func(name string) ([]byte, error) {
				if name == "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt" {
					return []byte("norm"), nil
				}
				return nil, fmt.Errorf("file not found")
			},
		}

		// Инициализируем клиент
		err := k8sClient.setup()
		require.NoError(t, err)

		// URL не важен, так как запрос не дойдет до него
		k8sClient.jwksURL = "http://invalid"

		_, err = k8sClient.GetPublicKey()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error reading token")
	})
}

func TestVerifier_VerifyWithClient(t *testing.T) {
	// Generate a test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	verifier := &Verifier{publicKey: publicKey}

	now := time.Now()
	exp := now.Add(time.Hour)
	iat := now
	nbf := now

	validClaims := privateClaims{
		Exp: jwt.NewNumericDate(exp),
		Iat: jwt.NewNumericDate(iat),
		Nbf: jwt.NewNumericDate(nbf),
		Iss: "https://kubernetes.default.svc",
		Sub: "system:serviceaccount:postgres-a:default",
		Aud: jwt.ClaimStrings{"https://kubernetes.default.svc.cluster.local", "k3s"},
		Kubernetes: kubernetesClaims{
			Namespace: "postgres-a",
			Pod: ref{
				Name: "postgres-a-6794fcb5f7-qb9zm",
				UID:  "911749e7-551b-463e-bcfd-7f124aae815e",
			},
		},
	}

	tests := []struct {
		name       string
		token      string
		claims     privateClaims
		wantErr    bool
		errMessage string
	}{
		{
			name:    "valid token",
			token:   generateTestToken(t, privateKey, validClaims),
			claims:  validClaims,
			wantErr: false,
		},
		{
			name:       "invalid signature",
			token:      "invalid.token.signature",
			wantErr:    true,
			errMessage: "parsing jwt",
		},
		{
			name: "expired token",
			token: generateTestToken(t, privateKey, privateClaims{
				Exp: jwt.NewNumericDate(now.Add(-time.Hour)),
				Iat: jwt.NewNumericDate(now.Add(-2 * time.Hour)),
				Nbf: jwt.NewNumericDate(now.Add(-2 * time.Hour)),
				Kubernetes: kubernetesClaims{
					Namespace: "postgres-a",
					Pod: ref{
						Name: "postgres-a-6794fcb5f7-qb9zm",
					},
				},
			}),
			wantErr:    true,
			errMessage: "token is expired",
		},
		{
			name: "missing namespace",
			token: generateTestToken(t, privateKey, privateClaims{
				Exp: jwt.NewNumericDate(exp),
				Iat: jwt.NewNumericDate(iat),
				Nbf: jwt.NewNumericDate(nbf),
				Kubernetes: kubernetesClaims{
					Pod: ref{
						Name: "postgres-a-6794fcb5f7-qb9zm",
					},
				},
			}),
			wantErr:    true,
			errMessage: "invalid k8s token claims",
		},
		{
			name: "namespace and pod name mismatch",
			token: generateTestToken(t, privateKey, privateClaims{
				Exp: jwt.NewNumericDate(exp),
				Iat: jwt.NewNumericDate(iat),
				Nbf: jwt.NewNumericDate(nbf),
				Kubernetes: kubernetesClaims{
					Namespace: "postgres-a",
					Pod: ref{
						Name: "other-pod-123",
					},
				},
			}),
			wantErr:    true,
			errMessage: "pod name and namespace must both start with service name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientID, claims, err := verifier.VerifyWithClient(tt.token)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.claims.Kubernetes.Namespace, clientID)
				assert.Equal(t, tt.claims.Kubernetes.Namespace, claims.(privateClaims).Kubernetes.Namespace)
			}
		})
	}
}

func TestPrivateClaimsMethods(t *testing.T) {
	now := time.Now()
	claims := privateClaims{
		Exp: jwt.NewNumericDate(now),
		Iat: jwt.NewNumericDate(now),
		Nbf: jwt.NewNumericDate(now),
		Iss: "test-issuer",
		Sub: "test-subject",
		Aud: jwt.ClaimStrings{"test-audience"},
	}

	t.Run("GetExpirationTime", func(t *testing.T) {
		exp, err := claims.GetExpirationTime()
		assert.NoError(t, err)
		assert.True(t, math.Abs(float64(now.Second())-float64(exp.Time.Second())) < 1)
	})

	t.Run("GetIssuedAt", func(t *testing.T) {
		iat, err := claims.GetIssuedAt()
		assert.NoError(t, err)
		assert.True(t, math.Abs(float64(now.Second())-float64(iat.Time.Second())) < 1)
	})

	t.Run("GetNotBefore", func(t *testing.T) {
		nbf, err := claims.GetNotBefore()
		assert.NoError(t, err)
		assert.True(t, math.Abs(float64(now.Second())-float64(nbf.Time.Second())) < 1)
	})

	t.Run("GetIssuer", func(t *testing.T) {
		iss, err := claims.GetIssuer()
		assert.NoError(t, err)
		assert.Equal(t, "test-issuer", iss)
	})

	t.Run("GetSubject", func(t *testing.T) {
		sub, err := claims.GetSubject()
		assert.NoError(t, err)
		assert.Equal(t, "test-subject", sub)
	})

	t.Run("GetAudience", func(t *testing.T) {
		aud, err := claims.GetAudience()
		assert.NoError(t, err)
		assert.Equal(t, jwt.ClaimStrings{"test-audience"}, aud)
	})

	t.Run("empty claims", func(t *testing.T) {
		empty := privateClaims{}
		_, err := empty.GetExpirationTime()
		assert.NoError(t, err) // nil is valid for these methods
	})
}

func generateTestToken(t *testing.T, key *rsa.PrivateKey, claims privateClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(key)
	require.NoError(t, err)
	return tokenString
}
