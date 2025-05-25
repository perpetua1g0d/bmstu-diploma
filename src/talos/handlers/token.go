package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/config"
	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/jwks"
	"github.com/perpetua1g0d/bmstu-diploma/talos/pkg/k8s"
)

const (
	grantTypeTokenExchange = "urn:ietf:params:oauth:grant-type:token-exchange" // RFC 8693
	k8sTokenType           = "urn:ietf:params:oauth:token-type:jwt:kubernetes"
)

type TokenRequest struct {
	GrantType        string `form:"grant_type"`
	SubjectTokenType string `form:"subject_token_type"`
	SubjectToken     string `form:"subject_token"`
	Scope            string `form:"scope"`
}

func NewTokenHandler(ctx context.Context, cfg *config.Config, keys *jwks.KeyPair) (http.HandlerFunc, error) {
	issuer, err := NewIssuer(cfg, keys)
	if err != nil {
		return nil, fmt.Errorf("failed to create issued: %w", err)
	}

	k8sVerifier, err := k8s.NewVerifier(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s verifier: %w", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("failed to parse form request params: %v", err)
			http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
			return
		}

		log.Printf("Incoming request: Method=%s, URL=%s, Body=%s", r.Method, r.URL, r.Form)

		req := TokenRequest{
			GrantType:        r.FormValue("grant_type"),
			SubjectTokenType: r.FormValue("subject_token_type"),
			SubjectToken:     r.FormValue("subject_token"),
			Scope:            r.FormValue("scope"),
		}

		if req.GrantType != grantTypeTokenExchange {
			log.Printf("unexpected grant_type: %s", req.GrantType)
			http.Error(w, `{"error":"unsupported_grant_type"}`, http.StatusBadRequest)
			return
		} else if req.SubjectTokenType != k8sTokenType {
			log.Printf("unexpected subject_token_type: %s", req.GrantType)
			http.Error(w, `{"error":"unsupported_subject_token_type"}`, http.StatusBadRequest)
			return
		}

		clientID, _, err := k8sVerifier.VerifyWithClient(req.SubjectToken)
		if err != nil {
			log.Printf("failed to verify k8s token: %v", err)
			http.Error(w, `{"error":"token_not_verified"}`, http.StatusBadRequest)
			return
		}

		issueResp, err := issuer.IssueToken(clientID, req.Scope)
		if err != nil {
			log.Printf("failed to issue talos token: %v", err)
			http.Error(w, `{"error":"access_denied"}`, http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issueResp)
		// json.NewEncoder(w).Encode(map[string]string{
		// 	"access_token": token,
		// 	"token_type":   "Bearer",
		// 	"expires_in":   issuer.config.TokenTTL.String(),
		// })

		log.Printf("token issued, clientID: %s, scope: %s", clientID, req.Scope)
	}, nil
}
