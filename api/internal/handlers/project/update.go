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

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	cl := clients.ProjectClient()
	out, err := cl.UpdateProject(context.Background(), &projectv1.UpdateProjectRequest{
		UserId:               uid,
		Id:                   id,
		Title:                req.Title,
		Description:          req.Description,
		DurationEstimate:     int32(req.DurationEstimate),
		TeamSize:             int32(req.TeamSize),
		ApplicationCloseTime: req.ApplicationCloseTime,
	})
	if err != nil {
		http.Error(w, "gRPC: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out.Project)
}
