package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gusplusbus/trustflow/api/internal/handlers"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/healthz", handlers.HealthCheck).Methods(http.MethodGet)

	api.Handle("/projects",
		handlers.AuthMiddleware(http.HandlerFunc(handlers.HandleCreateProject)),
	).Methods(http.MethodPost)

	return r
}
