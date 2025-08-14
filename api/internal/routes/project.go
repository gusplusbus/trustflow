package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/gusplusbus/trustflow/api/internal/middleware"
	project "github.com/gusplusbus/trustflow/api/internal/handlers/project"
	"github.com/gusplusbus/trustflow/api/internal/handlers/project/ownership"
)

func RegisterProjectRoutes(api *mux.Router) {
	// ----- Non project-scoped (still authenticated) -----
	api.
		Handle("/projects", middleware.AuthMiddleware(http.HandlerFunc(project.HandleCreate))).
		Methods(http.MethodPost)

	api.
		Handle("/projects", middleware.AuthMiddleware(http.HandlerFunc(project.HandleList))).
		Methods(http.MethodGet)

	// ----- Project-scoped subrouter: Auth -> ProjectCtx -----
	projectScoped := api.PathPrefix("/projects/{id}").Subrouter()
	projectScoped.Use(middleware.AuthMiddleware)     // leave your JWT as-is
	projectScoped.Use(middleware.WithProjectContext) // loads Project + Ownerships into ctx

	projectScoped.Handle("", http.HandlerFunc(project.HandleGet)).Methods(http.MethodGet)
	projectScoped.Handle("", http.HandlerFunc(project.HandleUpdate)).Methods(http.MethodPut)
	projectScoped.Handle("", http.HandlerFunc(project.HandleDelete)).Methods(http.MethodDelete)

	// Ownership endpoints (no owner/repo in query; use context, pick first ownership)
	projectScoped.Handle("/ownership", http.HandlerFunc(ownership.HandleCreate)).Methods(http.MethodPost)
	projectScoped.Handle("/ownership/issues", http.HandlerFunc(ownership.HandleIssues)).Methods(http.MethodGet)

	// (optional) If you add a delete ownership handler later, it belongs here too:
	// projectScoped.Handle("/ownership", http.HandlerFunc(ownership.HandleDelete)).Methods(http.MethodDelete)
}
