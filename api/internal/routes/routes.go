package routes

import (
	"github.com/gorilla/mux"
  "github.com/gusplusbus/trustflow/api/internal/handlers"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	// API
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/healthz", handlers.HealthCheck).Methods("GET")


	return r
}
