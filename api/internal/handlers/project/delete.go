package project

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gusplusbus/trustflow/api/internal/clients"
	"github.com/gusplusbus/trustflow/api/internal/handlers"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
)

func HandleDelete(w http.ResponseWriter, r *http.Request) {
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
	out, err := cl.DeleteProject(context.Background(), &projectv1.DeleteProjectRequest{
		UserId: uid,
		Id:     id,
	})
	if err != nil {
		http.Error(w, "gRPC: "+err.Error(), http.StatusBadGateway)
		return
	}
	if !out.GetDeleted() {
		// still return 204; client doesn’t need the reason
	}

	w.WriteHeader(http.StatusNoContent)
}
