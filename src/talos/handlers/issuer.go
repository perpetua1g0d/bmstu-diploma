package handlers

import (
	"fmt"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/config"
	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/jwks"
)

type Issuer struct {
	config  *config.Config
	keyPair *jwks.KeyPair
	signer  jose.Signer
	rolesDB map[string]map[string][]string
}

func NewIssuer(cfg *config.Config, keys *jwks.KeyPair) (*Issuer, error) {
	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.RS256,
			Key: jose.JSONWebKey{
				Key:       keys.PrivateKey,
				KeyID:     keys.KeyID,
				Algorithm: "RS256",
				Use:       "sig",
			},
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	return &Issuer{
		config:  cfg,
		keyPair: keys,
		signer:  signer,
		rolesDB: map[string]map[string][]string{
			"postgres-a": {"postgres-b": {"RW"}},
			"postgres-b": {"postgres-a": {"RO"}},
		},
	}, nil
}

func (i *Issuer) IssueToken(clientID, scope string) (string, error) {
	allowedRoles, ok := i.rolesDB[clientID][scope]
	if !ok {
		return "", fmt.Errorf("access denied for client %s to scope %s", clientID, scope)
	}

	tokenClaims := map[string]interface{}{
		"iss":   i.config.Issuer,
		"sub":   clientID,
		"aud":   scope,
		"scope": scope,
		"roles": allowedRoles,
		"exp":   time.Now().Add(i.config.TokenTTL).Unix(),
		"iat":   time.Now().Unix(),
	}

	return jwks.GenerateJWT(i.signer, tokenClaims)
}
