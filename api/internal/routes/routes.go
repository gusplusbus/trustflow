package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gusplusbus/trustflow/api/internal/handlers"
	"github.com/gusplusbus/trustflow/api/internal/ledger"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/healthz", handlers.HealthCheck).Methods(http.MethodGet)

	RegisterProjectRoutes(api)
	// internal routes (not exposed to end-users / UI)
	internal := r.PathPrefix("/internal").Subrouter()
	internal.HandleFunc("/ledger/notify", ledger.HandleNotify).Methods(http.MethodPost)

	return r
}
