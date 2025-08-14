package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
	ownershipv1 "github.com/gusplusbus/trustflow/data_server/gen/ownershipv1"

	"github.com/gusplusbus/trustflow/api/internal/clients"
)

type ctxKeyProject string

const projectCtxKey ctxKeyProject = "projectCtx"

type ProjectContext struct {
	Project    *projectv1.Project
	Ownerships []*ownershipv1.Ownership // <-- NOTE: ownershipv1, per DS
}

// WithProjectContext loads project + ownerships for routes with {id} and stashes into context.
// Auth/JWT stays exactly as-is; we just read the uid your AuthMiddleware put into context.
func WithProjectContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id == "" {
			next.ServeHTTP(w, r)
			return
		}

		uid, ok := UserIDFromCtx(r.Context())
		if !ok || uid == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		pc := clients.ProjectClient()
		out, err := pc.GetProject(r.Context(), &projectv1.GetProjectRequest{
			UserId:            uid,
			Id:                id,
			IncludeOwnerships: true,
		})
		if err != nil || out.GetProject() == nil {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}

		p := out.GetProject()
		ctx := context.WithValue(r.Context(), projectCtxKey, &ProjectContext{
			Project:    p,
			Ownerships: p.GetOwnerships(), // returns []*ownershipv1.Ownership
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectCtx returns the project context loaded by WithProjectContext.
func ProjectCtx(r *http.Request) (*ProjectContext, bool) {
	pc, ok := r.Context().Value(projectCtxKey).(*ProjectContext)
	return pc, ok
}
