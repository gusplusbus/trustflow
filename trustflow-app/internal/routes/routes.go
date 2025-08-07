package routes

import (
	"net/http"
	"github.com/gorilla/mux"
  "github.com/gusplusbus/trustflow/trustflow-app/internal/handlers"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	// API
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/healthz", handlers.HealthCheck).Methods("GET")

	// Static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	return r
}
