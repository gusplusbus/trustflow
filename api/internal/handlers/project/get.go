package project

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gusplusbus/trustflow/api/internal/clients"
	"github.com/gusplusbus/trustflow/api/internal/handlers"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
)

func HandleGet(w http.ResponseWriter, r *http.Request) {
	uid, ok := handlers.UserIDFromCtx(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	cl := clients.ProjectClient()
	out, err := cl.GetProject(context.Background(), &projectv1.GetProjectRequest{
		UserId: uid,
		Id:     id,
	})
	if err != nil {
		// Data server should return NotFound via status, but we map generically here
		http.Error(w, "gRPC: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out.Project)
}
