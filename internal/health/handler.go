// Package health serves the service liveness endpoint.
package health

import (
	"encoding/json"
	"net/http"
)

// Handler responds to GET /health with 200 and {"status":"ok"}.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
