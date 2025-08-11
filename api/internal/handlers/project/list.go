package project

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gusplusbus/trustflow/api/internal/clients"
	"github.com/gusplusbus/trustflow/api/internal/handlers"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
)

func HandleList(w http.ResponseWriter, r *http.Request) {
	uid, ok := handlers.UserIDFromCtx(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query params
	qp := r.URL.Query()
	page := clamp(intFromQuery(qp.Get("page"), 0), 0, 1_000_000) // 0-based
	pageSize := clamp(intFromQuery(qp.Get("page_size"), 20), 1, 200)
	sortBy := parseSortBy(qp.Get("sort_by"))
	sortDir := parseSortDir(qp.Get("sort_dir"))
	q := qp.Get("q")

	// Call data_server with a sane timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cl := clients.ProjectClient()
	out, err := cl.ListProjects(ctx, &projectv1.ListProjectsRequest{
		UserId:   uid,
		Page:     int32(page),
		PageSize: int32(pageSize),
		SortBy:   sortBy,
		SortDir:  sortDir,
		Q:        q,
	})
	if err != nil {
		http.Error(w, "gRPC: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Projects  []*projectv1.Project `json:"projects"`
		Total     int64                `json:"total"`
		Page      int                  `json:"page"`
		PageSize  int                  `json:"page_size"`
		SortBy    string               `json:"sort_by"`
		SortDir   string               `json:"sort_dir"`
		Q         string               `json:"q"`
	}{
		Projects: out.GetProjects(),
		Total:    out.GetTotal(),
		Page:     page,
		PageSize: pageSize,
		SortBy:   sortByString(sortBy),
		SortDir:  sortDirString(sortDir),
		Q:        q,
	})
}

// ---- helpers ----

func intFromQuery(s string, def int) int {
	if s == "" {
		return def
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func parseSortBy(s string) projectv1.SortBy {
	if s == "" {
		return projectv1.SortBy_SORT_BY_CREATED_AT
	}
	switch strings.ToLower(s) {
	case "created_at", "createdat", "created":
		return projectv1.SortBy_SORT_BY_CREATED_AT
	case "updated_at", "updatedat", "updated":
		return projectv1.SortBy_SORT_BY_UPDATED_AT
	case "title":
		return projectv1.SortBy_SORT_BY_TITLE
	case "team_size", "teamsize":
		return projectv1.SortBy_SORT_BY_TEAM_SIZE
	case "duration", "duration_estimate", "durationestimate":
		return projectv1.SortBy_SORT_BY_DURATION
	default:
		// allow numeric enum values if someone passes them
		if n, err := strconv.Atoi(s); err == nil {
			return projectv1.SortBy(n)
		}
		return projectv1.SortBy_SORT_BY_CREATED_AT
	}
}

func parseSortDir(s string) projectv1.SortDir {
	if s == "" {
		return projectv1.SortDir_SORT_DIR_DESC
	}
	switch strings.ToLower(s) {
	case "asc", "ascending":
		return projectv1.SortDir_SORT_DIR_ASC
	case "desc", "descending":
		return projectv1.SortDir_SORT_DIR_DESC
	default:
		if n, err := strconv.Atoi(s); err == nil {
			return projectv1.SortDir(n)
		}
		return projectv1.SortDir_SORT_DIR_DESC
	}
}

func sortByString(v projectv1.SortBy) string {
	switch v {
	case projectv1.SortBy_SORT_BY_CREATED_AT:
		return "created_at"
	case projectv1.SortBy_SORT_BY_UPDATED_AT:
		return "updated_at"
	case projectv1.SortBy_SORT_BY_TITLE:
		return "title"
	case projectv1.SortBy_SORT_BY_TEAM_SIZE:
		return "team_size"
	case projectv1.SortBy_SORT_BY_DURATION:
		return "duration"
	default:
		return "created_at"
	}
}

func sortDirString(v projectv1.SortDir) string {
	switch v {
	case projectv1.SortDir_SORT_DIR_ASC:
		return "asc"
	case projectv1.SortDir_SORT_DIR_DESC:
		return "desc"
	default:
		return "desc"
	}
}
