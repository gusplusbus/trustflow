package handlers

import (
	"fmt"
	"html/template"
	"net/http"
)

func (cfg *APIConfig) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("admin/metrics.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, cfg.FileserverHits.Load())
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

func (cfg *APIConfig) HandlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err := cfg.DBQueries.DeleteAllUsers(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting users: %v", err), http.StatusInternalServerError)
		return
	}

	cfg.FileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "All users deleted and hits reset")
}

func (cfg *APIConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

