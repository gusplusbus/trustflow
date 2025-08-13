package ownership

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gusplusbus/trustflow/api/internal/clients"
	"github.com/gusplusbus/trustflow/api/internal/handlers"
	ownershipv1 "github.com/gusplusbus/trustflow/data_server/gen/ownershipv1"
)

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	uid, ok := handlers.UserIDFromCtx(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	projectID := vars["id"]
	if projectID == "" {
		http.Error(w, "Missing project id in path", http.StatusBadRequest)
		return
	}

	var req CreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if req.Organization == "" || req.Repository == "" {
		http.Error(w, "organization and repository are required", http.StatusBadRequest)
		return
	}

	// Use the OWNERSHIP client and CreateOwnership method
	cl := clients.OwnershipClient()
	out, err := cl.CreateOwnership(
		r.Context(),
		&ownershipv1.CreateOwnershipRequest{
			UserId:       uid,
			ProjectId:    projectID,
			Organization: req.Organization,
			Repository:   req.Repository,
			Provider:     req.Provider,
			WebUrl:       req.WebURL,
		},
	)
	if err != nil {
		http.Error(w, "gRPC: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out.Ownership)
}
