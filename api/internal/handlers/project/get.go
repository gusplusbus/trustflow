package project

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gusplusbus/trustflow/api/internal/middleware"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
)

func HandleGet(w http.ResponseWriter, r *http.Request) {
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

	// read include_ownerships=true|false (default false)
	include := false
	if v := r.URL.Query().Get("include_ownerships"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			include = b
		}
	}

	// Serve from context to avoid another DS call.
	// If include=false, send a shallow copy with Ownerships nil.
	var resp *projectv1.Project = pc.Project
	if !include {
		cp := *pc.Project
		cp.Ownerships = nil
		resp = &cp
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
