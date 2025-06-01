package handlers

import (
	"encoding/json"
	"net/http"
)

func (ctl *Controller) CertsHandler() http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		jwks := ctl.keys.JWKS()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}

	return baseMetricsMiddleware(handler)
}
