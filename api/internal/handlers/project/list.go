package project

import (
	"encoding/json"
	"net/http"

  "github.com/gusplusbus/trustflow/api/internal/handlers"
)

func HandleList(w http.ResponseWriter, r *http.Request) {
  uid, ok := handlers.UserIDFromCtx(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: call data_server.ListProjects(uid)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "not implemented"})
}
