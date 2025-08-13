package ownership

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gusplusbus/trustflow/api/internal/clients"
	"github.com/gusplusbus/trustflow/api/internal/handlers"
	"github.com/gusplusbus/trustflow/api/internal/providers"
	ghprov "github.com/gusplusbus/trustflow/api/internal/providers/github"
	ownershipv1 "github.com/gusplusbus/trustflow/data_server/gen/ownershipv1"
)

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	uid, ok := handlers.UserIDFromCtx(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	projectID := mux.Vars(r)["id"]
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

	// ---- verify access with provider before saving
	var verifier providers.RepoAccessVerifier
	switch req.Provider {
	case "", "github":
		v, err := ghprov.NewVerifierFromEnv()
		if err != nil {
			http.Error(w, "github verifier not configured: "+err.Error(), http.StatusBadGateway)
			return
		}
		verifier = v
	default:
		http.Error(w, "unsupported provider: "+req.Provider, http.StatusBadRequest)
		return
	}

	if err := verifier.VerifyAccess(r.Context(), req.Organization, req.Repository); err != nil {
		// Surface a clear message (user action: install app / give repo access)
		http.Error(w, "provider access check failed: "+err.Error(), http.StatusForbidden)
		return
	}

	// ---- save ownership
	cl := clients.OwnershipClient()
	out, err := cl.CreateOwnership(
		r.Context(),
		&ownershipv1.CreateOwnershipRequest{
			UserId:       uid,
			ProjectId:    projectID,
			Organization: req.Organization,
			Repository:   req.Repository,
			Provider:     coalesce(req.Provider, "github"),
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

func coalesce(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
