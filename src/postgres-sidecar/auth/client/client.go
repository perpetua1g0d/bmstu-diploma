package auth_client

import (
	"context"
	"fmt"

	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/auth/config"
	"github.com/perpetua1g0d/bmstu-diploma/postgres-sidecar/auth/tokens"
)

type AuthClient struct {
	cfg *config.Config
	ts  *tokens.TokenSet

	verifier *tokens.Verifier
}

func NewAuthClient(ctx context.Context, cfg *config.Config, scopes []string) (*AuthClient, error) {
	ts, err := tokens.NewTokenSet(ctx, cfg, scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenset: %w", err)
	}

	verifier, err := tokens.NewVerifier(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create verifier: %w", err)
	}

	return &AuthClient{
		cfg:      cfg,
		ts:       ts,
		verifier: verifier,
	}, nil
}

func (c *AuthClient) Token(scope string) (string, error) {
	return c.ts.Token(scope)
}

func (c *AuthClient) VerifyToken(rawToken string, needRoles []string) error {
	return c.verifier.VerifyToken(rawToken, needRoles)
}
