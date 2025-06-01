package k8s

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
)

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type K8sClient struct {
	readSecrets func(name string) ([]byte, error)

	client  *http.Client
	jwksURL string
}

func (k *K8sClient) setup() error {
	caCert, err := k.readSecrets("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return fmt.Errorf("error reading CA cert: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	k.client = client
	k.jwksURL = "https://kubernetes.default.svc/openid/v1/jwks"

	return nil
}

func (k *K8sClient) GetPublicKey() (*rsa.PublicKey, error) {
	token, err := k.readSecrets("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return nil, fmt.Errorf("error reading token: %w", err)
	}

	req, err := http.NewRequest("GET", k.jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating k8s jwks request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+string(token))

	resp, err := k.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("JWKS request failed: %w", err)
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("JWKS parse error: %w", err)
	}

	if len(jwks.Keys) == 0 {
		return nil, errors.New("no keys in JWKS")
	}

	key := jwks.Keys[0]
	return makeRSAPublicKey(key)
}

func makeRSAPublicKey(key JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("invalid modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("invalid exponent: %w", err)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}, nil
}
