package handler

import (
	"encoding/json"
	"net/http"
)

// Health reports that the service is up and running.
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"estado": "ok"})
}