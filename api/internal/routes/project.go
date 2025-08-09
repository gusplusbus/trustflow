package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gusplusbus/trustflow/api/internal/handlers"
	project "github.com/gusplusbus/trustflow/api/internal/handlers/project"
)

func RegisterProjectRoutes(api *mux.Router) {
	api.
		Handle("/projects", handlers.AuthMiddleware(http.HandlerFunc(project.HandleCreate))).
		Methods(http.MethodPost)

	api.
		Handle("/projects", handlers.AuthMiddleware(http.HandlerFunc(project.HandleList))).
		Methods(http.MethodGet)

	api.
		Handle("/projects/{id}", handlers.AuthMiddleware(http.HandlerFunc(project.HandleGet))).
		Methods(http.MethodGet)

	api.
		Handle("/projects/{id}", handlers.AuthMiddleware(http.HandlerFunc(project.HandleUpdate))).
		Methods(http.MethodPut)

	api.
		Handle("/projects/{id}", handlers.AuthMiddleware(http.HandlerFunc(project.HandleDelete))).
		Methods(http.MethodDelete)
}
