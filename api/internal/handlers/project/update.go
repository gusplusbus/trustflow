package project

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/gusplusbus/trustflow/api/internal/handlers"
	"github.com/gusplusbus/trustflow/api/internal/clients"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
)

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
	uid, ok := handlers.UserIDFromCtx(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	var req UpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Deref the optional fields safely
	var (
		title   string
		desc    string
		dur     int32
		team    int32
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
	out, err := cl.UpdateProject(context.Background(), &projectv1.UpdateProjectRequest{
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
	_ = json.NewEncoder(w).Encode(out.Project)
}
