package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("error encoding health check response: %v", err)
	}
}
