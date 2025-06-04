package verifier

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *Verifier) NewRealodHandler() http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received config reload request")

		var data struct {
			Verify bool `json:"verify"`
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			log.Printf("JSON decode error: %v", err)
			return
		}

		// атомарно
		s.cfg.VerifyAuthEnabled.Store(&data.Verify)

		log.Printf("Verifier auth settings updated via HTTP: VERIFY=%v", data.Verify)
		w.WriteHeader(http.StatusOK)
	}

	return handler
}
