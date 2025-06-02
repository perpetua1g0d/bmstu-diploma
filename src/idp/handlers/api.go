package handlers

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/config"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/jwks"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/k8s"
)

type K8sVerifier interface {
	VerifyWithClient(k8sToken string) (string, jwt.Claims, error)
}

type Issuer interface {
	IssueToken(clientID, scope string) (*IssueResp, error)
}

type Repository interface {
	UpdatePermissions(client, scope string, roles []string) error
	GetPermissions(client, scope string) []string
}

type ControllerOpts struct {
	Cfg  *config.Config
	Keys *jwks.KeyPair

	Repository Repository
}

type Controller struct {
	k8sVerifier K8sVerifier
	repository  Repository
	issuer      Issuer

	cfg  *config.Config
	keys *jwks.KeyPair
}

func NewController(ctx context.Context, opts *ControllerOpts) (*Controller, error) {
	cfg := opts.Cfg
	keys := opts.Keys
	repository := opts.Repository

	issuer, err := NewIssuer(cfg, keys, repository)
	if err != nil {
		return nil, fmt.Errorf("failed to create issued: %w", err)
	}

	k8sVerifier, err := k8s.NewVerifier(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s verifier: %w", err)
	}

	return &Controller{
		cfg:  cfg,
		keys: keys,

		k8sVerifier: k8sVerifier,
		repository:  repository,
		issuer:      issuer,
	}, nil
}
