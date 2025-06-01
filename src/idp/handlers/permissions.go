package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type PermissionsRequest struct {
	Client string   `json:"client"`
	Scope  string   `json:"scope"`
	Roles  []string `json:"roles"`
}

type PermissionsResponse struct {
	Roles []string `json:"roles"`
}

func (ctl *Controller) NewUpdatePermissionsHandler(ctx context.Context) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req PermissionsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		if err := ctl.repository.UpdatePermissions(req.Client, req.Scope, req.Roles); err != nil {
			log.Printf("failed to update permissions (%s -> %s: %v): %v", req.Client, req.Scope, req.Roles, err)
			respondError(w, fmt.Sprintf("failed to update permissions: %v", err), http.StatusInternalServerError)
			return
		}
	}

	return baseMetricsMiddleware(handler)
}

func (ctl *Controller) NewGetPermissionsHandler(ctx context.Context) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var req PermissionsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		client := req.Client
		scope := req.Scope

		roles := ctl.repository.GetPermissions(client, scope)
		resp := PermissionsResponse{Roles: roles}

		marhsalled, err := json.Marshal(resp)
		if err != nil {
			respondError(w, fmt.Sprintf("failed to marhsal get permissions response: %v", err), http.StatusInternalServerError)
			return
		} else if _, err := w.Write(marhsalled); err != nil {
			respondError(w, fmt.Sprintf("failed to write get permissions response: %v", err), http.StatusInternalServerError)
			return
		}
	}

	return baseMetricsMiddleware(handler)
}

func respondError(w http.ResponseWriter, message string, code int) {
	if code != http.StatusOK {
		log.Printf("request failed: status: %d, message %s", code, message)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(RespErr{Error: message})
}

type RespErr struct {
	Error string `json:"error"`
}
