package config

import (
	"sync/atomic"
	"time"
)

const (
	IdPIssuer = "http://idp.idp.svc.cluster.local"
)

type Config struct {
	ClientID string

	TokenEndpointAddress  string
	CertsEndpointAddress  string
	ConfigEndpointAddress string

	SignAuthEnabled   atomic.Pointer[bool]
	VerifyAuthEnabled atomic.Pointer[bool]

	RequestTimeout  time.Duration
	ErrTokenBackoff time.Duration
}
