package routes

import (
  "github.com/gusplusbus/trustflow/auth_server/internal/handlers"
  "github.com/gusplusbus/trustflow/auth_server/internal/database"
    "net/http"
    "sync/atomic"
)

func Register(dbQueries *database.Queries, platform string) *http.ServeMux {
    apiCfg := &handlers.APIConfig{
        DBQueries:      dbQueries,
        Platform:       platform,
        FileserverHits: atomic.Int32{},
    }

    mux := http.NewServeMux()

    // Health
    mux.HandleFunc("GET /api/healthz", handlers.HandleHealth)

    // Users
    mux.HandleFunc("POST /api/users", apiCfg.HandleCreateUser)
    mux.Handle("PUT /api/users", apiCfg.MiddlewareAuth(http.HandlerFunc(apiCfg.HandleUpdateUser))) 
    mux.HandleFunc("POST /api/login", apiCfg.HandleLogin)
    mux.HandleFunc("POST /api/refresh", apiCfg.HandleRefreshToken)
    mux.HandleFunc("POST /api/revoke", apiCfg.HandleRevokeRefreshToken)

    // Metrics/admin
    mux.HandleFunc("GET /admin/metrics", apiCfg.HandlerMetrics)
    mux.HandleFunc("POST /admin/reset", apiCfg.HandlerReset)

    // Static file serving with middleware (auth middleware will come later)
    fileServer := http.FileServer(http.Dir("."))
    mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileServer)))

    return mux
}
