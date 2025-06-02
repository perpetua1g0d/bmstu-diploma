package signer

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *TokenSigner) NewRealodHandler() http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received config reload request")

		var data struct {
			Sign   bool `json:"sign"`
			Verify bool `json:"verify"`
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			log.Printf("JSON decode error: %v", err)
			return
		}

		// атомарно
		s.cfg.SignAuthEnabled.Store(&data.Sign)
		// s.cfg.VerifyAuthEnabled.Store(&data.Verify)

		log.Printf("Signer auth settings updated via HTTP: SIGN=%v", data.Sign)
		// log.Printf("Signer auth settings updated via HTTP: SIGN=%v, VERIFY=%v", data.Sign, data.Verify)
		w.WriteHeader(http.StatusOK)
	}

	return handler
}

func (s *TokenSigner) refreshTokens() {
	s.tokenSet.RefreshTokens()
}

func (s *TokenSigner) NewRefreshTokensHandler() http.HandlerFunc {
	handler := func(_ http.ResponseWriter, _ *http.Request) {
		s.refreshTokens()
		log.Printf("Triggered tokens refresh.")
	}

	return handler
}
