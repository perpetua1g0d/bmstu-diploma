package handlers

import (
	"context"
	"fmt"

	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/config"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/db"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/jwks"
	"github.com/perpetua1g0d/bmstu-diploma/idp/pkg/k8s"
)

type ControllerOpts struct {
	Cfg  *config.Config
	Keys *jwks.KeyPair

	Repository *db.Repository
}

type Controller struct {
	k8sVerifier *k8s.Verifier
	repository  *db.Repository
	issuer      *Issuer

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
