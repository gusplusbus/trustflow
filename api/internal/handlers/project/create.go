package project

import (
	"encoding/json"
	"net/http"

	"github.com/gusplusbus/trustflow/api/internal/clients"
	"github.com/gusplusbus/trustflow/api/internal/middleware"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
)

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromCtx(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	cl := clients.ProjectClient()
	out, err := cl.CreateProject(r.Context(), &projectv1.CreateProjectRequest{
		UserId:               uid,
		Title:                req.Title,
		Description:          req.Description,
		DurationEstimate:     int32(req.DurationEstimate),
		TeamSize:             int32(req.TeamSize),
		ApplicationCloseTime: req.ApplicationCloseTime, // keep as-is per your DS
	})
	if err != nil {
		http.Error(w, "gRPC: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(out.GetProject())
}
