package handlers

import (
    "github.com/gusplusbus/trustflow/auth_server/internal/database"
    "sync/atomic"
)

// APIConfig holds shared dependencies for handlers.
type APIConfig struct {
    FileserverHits atomic.Int32
    DBQueries      *database.Queries
    Platform       string
}
