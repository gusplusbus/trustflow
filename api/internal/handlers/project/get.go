package project

import (
	"encoding/json"
	"net/http"
	"strconv"

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

	// read include_ownerships=true|false
	include := false
	if v := r.URL.Query().Get("include_ownerships"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			include = b
		}
	}

	cl := clients.ProjectClient()
	out, err := cl.GetProject(r.Context(), &projectv1.GetProjectRequest{
		UserId:            uid,
		Id:                id,
		IncludeOwnerships: include,
	})
	if err != nil {
		http.Error(w, "gRPC: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out.Project)
}
