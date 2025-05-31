package handlers

import (
	"fmt"
	"log"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/config"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/jwks"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/tokens"
)

type IssueResp struct {
	AccessToken string    `json:"access_token"`
	Type        string    `json:"token_type"`
	ExpiresIn   time.Time `json:"expires_in"`
}

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
			"postgres-a": {"postgres-b": {"RO", "RW"}},
			"postgres-b": {"postgres-a": {"RO"}},
		},
	}, nil
}

func (i *Issuer) IssueToken(clientID, scope string) (*IssueResp, error) {
	allowedRoles, ok := i.rolesDB[clientID][scope]
	if !ok {
		return nil, fmt.Errorf("access denied for client %s to scope %s", clientID, scope)
	}

	exp := time.Now().Add(i.config.TokenTTL)
	tokenClaims := tokens.Claims{
		Iss:      i.config.Issuer,
		Sub:      clientID,
		ClientID: clientID,
		Aud:      scope,
		Scope:    scope,
		Roles:    allowedRoles,
		Exp:      exp,
		Iat:      time.Now(),
	}

	log.Printf("claims to issue: %v", tokenClaims)

	accessToken, err := jwks.GenerateJWT(i.signer, tokenClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate jwt: %w", err)
	}

	return &IssueResp{
		AccessToken: accessToken,
		Type:        "Bearer",
		ExpiresIn:   exp,
	}, nil
}
