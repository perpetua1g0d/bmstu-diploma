package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
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

func (ctl *Controller) NewTokenHandler(ctx context.Context) (http.HandlerFunc, error) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var err error
		var scope, clientID string
		issueStart := time.Now()
		defer func() {
			issueDuration := float64(time.Since(issueStart).Milliseconds())
			tokenResult := "ok"
			if err != nil {
				tokenResult = "error"
			}
			if clientID == "" {
				clientID = "unknown"
			}
			if scope == "" {
				scope = "unknown"
			}

			tokenIssuedTotal.WithLabelValues(tokenResult, clientID, scope).Inc()
			tokenIssueDuration.WithLabelValues(tokenResult, clientID, scope).Observe(issueDuration)
		}()

		if err = r.ParseForm(); err != nil {
			log.Printf("failed to parse form request params: %v", err)
			http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
			return
		}

		// log.Printf("Incoming request: Method=%s, URL=%s, Body=%s", r.Method, r.URL, r.Form)

		req := TokenRequest{
			GrantType:        r.FormValue("grant_type"),
			SubjectTokenType: r.FormValue("subject_token_type"),
			SubjectToken:     r.FormValue("subject_token"),
			Scope:            r.FormValue("scope"),
		}
		scope = req.Scope

		if req.GrantType != grantTypeTokenExchange {
			log.Printf("unexpected grant_type: %s", req.GrantType)
			http.Error(w, `{"error":"unsupported_grant_type"}`, http.StatusBadRequest)
			return
		} else if req.SubjectTokenType != k8sTokenType {
			log.Printf("unexpected subject_token_type: %s", req.GrantType)
			http.Error(w, `{"error":"unsupported_subject_token_type"}`, http.StatusBadRequest)
			return
		}

		clientID, _, err = ctl.k8sVerifier.VerifyWithClient(req.SubjectToken)
		if err != nil {
			log.Printf("failed to verify k8s token: %v", err)
			http.Error(w, `{"error":"token_not_verified"}`, http.StatusBadRequest)
			return
		}

		issueResp, err := ctl.issuer.IssueToken(clientID, scope)
		if err != nil {
			log.Printf("failed to issue idp token: %v", err)
			http.Error(w, `{"error":"access_denied"}`, http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// json.NewEncoder(w).Encode(issueResp)
		if err := json.NewEncoder(w).Encode(issueResp); err != nil {
			log.Printf("failed to write token response: %v", err)
			http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
			return
		}

		log.Printf("token issued, clientID: %s, scope: %s", clientID, scope)
	}

	return baseMetricsMiddleware(handler), nil
}
