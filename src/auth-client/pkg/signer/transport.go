package signer

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/perpetua1g0d/bmstu-diploma/auth-client/internal/config"
	"github.com/perpetua1g0d/bmstu-diploma/auth-client/internal/metrics"
)

type SignerTransport struct {
	signer *TokenSigner
	scope  string

	defaultRT http.RoundTripper
}

func NewAuthTransport(signer *TokenSigner, scope string) *SignerTransport {
	return &SignerTransport{
		signer:    signer,
		scope:     scope,
		defaultRT: http.DefaultTransport,
	}
}

func (t *SignerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var signEnabled bool
	var signResult = "ok"

	signEnabled = getSignAuth(t.signer.cfg)
	signStart := time.Now()
	if signEnabled {
		token, err := t.signer.tokenSet.Token(t.scope)
		if err != nil {
			log.Printf("failed to issue token in auth client on scope %s: %v", t.scope, err)
			signResult = "error"
		} else {
			r.Header.Set("X-I2I-Token", token)
		}
	}
	signDuration := float64(time.Since(signStart).Milliseconds())

	metrics.TokenSignedTotal.WithLabelValues(t.scope, signResult, strconv.FormatBool(signEnabled), t.scope).Inc()
	metrics.TokenSignDuration.WithLabelValues(t.scope, signResult, strconv.FormatBool(signEnabled), t.scope).Observe(signDuration)

	return t.defaultRT.RoundTrip(r)
}

func getSignAuth(cfg *config.Config) bool {
	loaded := cfg.SignAuthEnabled.Load()
	if loaded == nil {
		log.Printf("config pointer[SignAuthEnabled] is empty! Sign is enabled as fallback")
		return true
	}

	return *loaded
}
