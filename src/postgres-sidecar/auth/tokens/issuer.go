package tokens

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/auth/config"
)

const (
	saSecretTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount"
)

type Issuer struct {
	cfg *config.Config
}

func NewIssuer(cfg *config.Config) *Issuer {
	return &Issuer{cfg: cfg}
}

func (i *Issuer) IssueToken(ctx context.Context, scope string) (*TokenResp, error) {
	k8sToken, err := getK8SToken()
	if err != nil {
		return nil, fmt.Errorf("getting k8s token: %w", err)
	}

	v := url.Values{}
	v.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	v.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt:kubernetes")
	v.Set("subject_token", k8sToken)
	v.Set("scope", scope)
	body := v.Encode()

	req, err := http.NewRequest(http.MethodPost, i.cfg.TokenEndpointAddress, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create talos token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: i.cfg.RequestTimeout}
	resp, err := client.Do(req)

	var respBytes []byte
	if resp != nil && resp.Body != nil {
		respBytes, _ = io.ReadAll(resp.Body)
	}
	if err != nil {
		log.Printf("failed to get talos token: %v; respBody: %s", err, string(respBytes))
		return nil, fmt.Errorf("failed to get talos token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp TokenResp
	if marshalErr := json.Unmarshal(respBytes, &tokenResp); marshalErr != nil {
		log.Printf("failed to unmarshal token: %v; body: %s", marshalErr, string(respBytes))
		return nil, fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	return &tokenResp, nil
}

func getK8SToken() (string, error) {
	tokenPath := filepath.Join(saSecretTokenPath, "token")
	token, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("failed to read k8s token: %v", err)
	}
	return string(token), nil
}
