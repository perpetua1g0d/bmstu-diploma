package signer

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/perpetua1g0d/bmstu-diploma/src/auth-client/internal/config"
	"github.com/perpetua1g0d/bmstu-diploma/src/auth-client/internal/tokens"
)

type TokenSigner struct {
	cfg *config.Config

	tokenSet *tokens.TokenSet
}

func NewTokenSigner(ctx context.Context, clientID string, scopes []string, initSign bool) (*TokenSigner, error) {
	cfg := &config.Config{
		ClientID:        clientID,
		RequestTimeout:  5 * time.Second,
		ErrTokenBackoff: 10 * time.Second,
		SignAuthEnabled: atomic.Pointer[bool]{},
	}
	cfg.SignAuthEnabled.Store(&initSign)

	s := &TokenSigner{
		cfg: cfg,
	}

	if err := s.fetchIdPEndpoints(ctx); err != nil {
		return nil, fmt.Errorf("failed to get idp endpoints: %w", err)
	}

	tokenSet, err := tokens.NewTokenSet(ctx, cfg, scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to create token set: %w", err)
	}

	s.tokenSet = tokenSet

	return s, nil
}

func (s *TokenSigner) fetchIdPEndpoints(_ context.Context) error {
	idpAddress := fmt.Sprintf("%s:80", config.IdPIssuer)
	s.cfg.TokenEndpointAddress = idpAddress + "/realms/service2infra/protocol/openid-connect/token"
	s.cfg.CertsEndpointAddress = idpAddress + "/realms/service2infra/protocol/openid-connect/certs"
	s.cfg.ConfigEndpointAddress = idpAddress + "/realms/service2infra/.well-known/openid-configuration"

	return nil
}
