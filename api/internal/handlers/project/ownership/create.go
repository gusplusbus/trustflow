package ownership

import (
	"encoding/json"
	"net/http"

	"github.com/gusplusbus/trustflow/api/internal/clients"
	"github.com/gusplusbus/trustflow/api/internal/middleware"
	"github.com/gusplusbus/trustflow/api/internal/providers"
	ghprov "github.com/gusplusbus/trustflow/api/internal/providers/github"
	ownershipv1 "github.com/gusplusbus/trustflow/data_server/gen/ownershipv1"
)

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromCtx(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pc, ok := middleware.ProjectCtx(r)
	if !ok || pc == nil || pc.Project == nil {
		http.Error(w, "project context missing", http.StatusInternalServerError)
		return
	}
	projectID := pc.Project.GetId()

	var req CreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if req.Organization == "" || req.Repository == "" {
		http.Error(w, "organization and repository are required", http.StatusBadRequest)
		return
	}

	// OPTIONAL: for now, allow only one ownership per project
	if len(pc.Ownerships) > 0 {
		http.Error(w, "project already has ownership", http.StatusConflict)
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
		http.Error(w, "provider access check failed: "+err.Error(), http.StatusForbidden)
		return
	}

	// ---- save ownership (using your existing flattened DS request shape)
	cl := clients.OwnershipClient()
	out, err := cl.CreateOwnership(r.Context(), &ownershipv1.CreateOwnershipRequest{
		UserId:       uid,
		ProjectId:    projectID,
		Organization: req.Organization,
		Repository:   req.Repository,
		Provider:     coalesce(req.Provider, "github"),
		WebUrl:       req.WebURL,
	})
	if err != nil {
		http.Error(w, "gRPC: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(out.GetOwnership())
}

func coalesce(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
