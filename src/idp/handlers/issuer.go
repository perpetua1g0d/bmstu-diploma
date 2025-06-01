package handlers

import (
	"fmt"
	"log"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/config"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/db"
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

	repository *db.Repository
}

func NewIssuer(cfg *config.Config, keys *jwks.KeyPair, repository *db.Repository) (*Issuer, error) {
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
		config:     cfg,
		keyPair:    keys,
		signer:     signer,
		repository: repository,
	}, nil
}

func (i *Issuer) IssueToken(clientID, scope string) (*IssueResp, error) {
	allowedRoles := i.repository.GetPermissions(clientID, scope)
	// if !ok {
	// 	return nil, fmt.Errorf("access denied for client %s to scope %s", clientID, scope)
	// }

	timeNow := time.Now()
	exp := timeNow.Add(i.config.TokenTTL)
	tokenClaims := tokens.Claims{
		Iss:      i.config.Issuer,
		Sub:      clientID,
		ClientID: clientID,
		Aud:      scope,
		Scope:    scope,
		Roles:    allowedRoles,
		Exp:      exp,
		Iat:      timeNow,
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
