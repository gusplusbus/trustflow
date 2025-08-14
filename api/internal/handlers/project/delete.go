package project

import (
	"net/http"

	"github.com/gusplusbus/trustflow/api/internal/clients"
	"github.com/gusplusbus/trustflow/api/internal/middleware"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
)

func HandleDelete(w http.ResponseWriter, r *http.Request) {
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

	cl := clients.ProjectClient()
	out, err := cl.DeleteProject(r.Context(), &projectv1.DeleteProjectRequest{
		UserId: uid,
		Id:     id,
	})
	if err != nil {
		http.Error(w, "gRPC: "+err.Error(), http.StatusBadGateway)
		return
	}
	// We don't surface "not deleted" reasons; still return 204.
	_ = out

	w.WriteHeader(http.StatusNoContent)
}
