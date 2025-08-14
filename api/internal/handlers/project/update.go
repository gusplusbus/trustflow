package project

import (
	"encoding/json"
	"net/http"

	"github.com/gusplusbus/trustflow/api/internal/clients"
	"github.com/gusplusbus/trustflow/api/internal/middleware"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
)

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
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
	id := pc.Project.GetId()

	var req UpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Safely deref optional fields
	var (
		title    string
		desc     string
		dur      int32
		team     int32
		appClose string
	)
	if req.Title != nil {
		title = *req.Title
	}
	if req.Description != nil {
		desc = *req.Description
	}
	if req.DurationEstimate != nil {
		dur = int32(*req.DurationEstimate)
	}
	if req.TeamSize != nil {
		team = int32(*req.TeamSize)
	}
	if req.ApplicationCloseTime != nil {
		appClose = *req.ApplicationCloseTime
	}

	cl := clients.ProjectClient()
	out, err := cl.UpdateProject(r.Context(), &projectv1.UpdateProjectRequest{
		Id:                   id,
		UserId:               uid,
		Title:                title,
		Description:          desc,
		DurationEstimate:     dur,
		TeamSize:             team,
		ApplicationCloseTime: appClose,
	})
	if err != nil {
		http.Error(w, "gRPC: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out.GetProject())
}
