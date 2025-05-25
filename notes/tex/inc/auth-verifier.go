package tokens

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/auth/config"
	"github.com/samber/lo"
	// "github.com/perpetua1g0d/bmstu-diploma/talos/pkg/tokens"
)

const (
	talosIssuer = "http://talos.talos.svc.cluster.local"
)

type tokenClaims struct {
	Exp      time.Time `json:"exp"`
	Iat      time.Time `json:"iat"`
	Iss      string    `json:"iss"`
	Sub      string    `json:"sub"`
	Aud      string    `json:"aud"`
	Scope    string    `json:"scope"`
	Roles    []string  `json:"roles"`
	ClientID string    `json:"clientID"`
}

type Verifier struct {
	cfg *config.Config

	certs *jose.JSONWebKeySet
}

func NewVerifier(ctx context.Context, cfg *config.Config) (*Verifier, error) {
	v := &Verifier{
		cfg: cfg,
	}

	if cfg.VerifyEnabled {
		certs, err := v.fetchJWKs(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get talos certificates: %w", err)
		}
		v.certs = certs
	}

	return v, nil
}

func (v *Verifier) VerifyToken(rawToken string, needRoles []string) error {
	claims, err := verifyToken(rawToken, v.certs)
	if err != nil {
		return err
	}

	if err = v.verifyClaims(claims, needRoles); err != nil {
		log.Printf("claims error, claims: %v", claims)
		return fmt.Errorf("verify claims error: %w", err)
	}

	return nil
}

func (v *Verifier) verifyClaims(claims *tokenClaims, needRoles []string) error {
	if claims.Scope != claims.Aud || claims.Scope != v.cfg.ClientID {
		return fmt.Errorf("scope or aud is unexpected, service: %s, scope: %s, aud: %s", v.cfg.ClientID, claims.Scope, claims.Aud)
	} else if claims.Iss != talosIssuer {
		return fmt.Errorf("unexpected issuer, expected: %s, got: %s", talosIssuer, claims.Iss)
	} else if expired := claims.Exp.Before(time.Now()); expired {
		return fmt.Errorf("token is expired, exp: %s, now: %s", claims.Exp, time.Now())
	} else if rolesOk := lo.Every(claims.Roles, needRoles); !rolesOk {
		return fmt.Errorf("roles mismatched, want: %v, got: %v", needRoles, claims.Roles)
	}

	return nil
}

func verifyToken(rawToken string, certs *jose.JSONWebKeySet) (*tokenClaims, error) {
	token, err := jwt.ParseSigned(rawToken)
	if err != nil {
		log.Printf("failed to parse token: %v", err)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	var claims tokenClaims
	for _, header := range token.Headers {
		keys := certs.Key(header.KeyID)
		if len(keys) == 0 {
			continue
		}

		for _, key := range keys {
			if err := token.Claims(key.Public(), &claims); err == nil {
				return &claims, nil
			}
		}
	}

	log.Printf("no certificate found to parse token. certs: %v, tokenHeaders: %v", certs, token.Headers)
	return nil, fmt.Errorf("no certificate found to parse token")
}

func (v *Verifier) fetchJWKs(ctx context.Context) (*jose.JSONWebKeySet, error) {
	talosCertEndpoint := v.cfg.CertsEndpointAddress
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, talosCertEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create talos certs request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: v.cfg.RequestTimeout}
	resp, err := client.Do(req)

	var respBytes []byte
	if resp != nil && resp.Body != nil {
		respBytes, _ = io.ReadAll(resp.Body)
	}
	if err != nil {
		log.Printf("failed to get talos certs: %v; respBody: %s", err, string(respBytes))
		return nil, fmt.Errorf("failed to get talos certs: %w", err)
	}
	defer resp.Body.Close()

	var jwks jose.JSONWebKeySet
	if marshalErr := json.Unmarshal(respBytes, &jwks); marshalErr != nil {
		log.Printf("failed to unmarshal certs: %v; body: %s", marshalErr, string(respBytes))
		return nil, fmt.Errorf("failed to unmarshal certs response: %w", err)
	}

	return &jwks, nil
}
