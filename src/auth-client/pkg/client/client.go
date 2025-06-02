package client

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/perpetua1g0d/bmstu-diploma/auth-client/pkg/config"
	"github.com/perpetua1g0d/bmstu-diploma/auth-client/pkg/tokens"
)

type AuthClient struct {
	cfg *config.Config
	ts  *tokens.TokenSet

	verifier *tokens.Verifier
}

func NewAuthClient(ctx context.Context, mux *http.ServeMux, cfg *config.Config, scopes []string) (*AuthClient, error) {
	ts, err := tokens.NewTokenSet(ctx, cfg, scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenset: %w", err)
	}

	verifier, err := tokens.NewVerifier(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create verifier: %w", err)
	}

	authClient := &AuthClient{
		cfg:      cfg,
		ts:       ts,
		verifier: verifier,
	}

	mux.HandleFunc("/refresh_tokens", authClient.NewRefreshTokensHandler())

	return authClient, nil
}

func (c *AuthClient) Token(scope string) (string, error) {
	return c.ts.Token(scope)
}

func (c *AuthClient) VerifyToken(rawToken string, needRoles []string) error {
	return c.verifier.VerifyToken(rawToken, needRoles)
}

func (c *AuthClient) RefreshTokens() {
	c.ts.RefreshTokens()
}

func (c *AuthClient) NewRefreshTokensHandler() http.HandlerFunc {
	handler := func(_ http.ResponseWriter, _ *http.Request) {
		c.RefreshTokens()
		log.Printf("Triggered tokens refresh.")
	}

	return handler
}
