package config

import "time"

type Config struct {
	ClientID string

	TokenEndpointAddress  string
	CertsEndpointAddress  string
	ConfigEndpointAddress string

	RequestTimeout time.Duration

	ErrTokenBackoff time.Duration
}
