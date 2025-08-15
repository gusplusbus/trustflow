package issues

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	issuev1 "github.com/gusplusbus/trustflow/data_server/gen/issuev1"
	"github.com/gusplusbus/trustflow/api/internal/clients"
	"github.com/gusplusbus/trustflow/api/internal/middleware"
)

type issueDTO struct {
	ID            string   `json:"id"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	ProjectID     string   `json:"project_id"`
	UserID        string   `json:"user_id"`
	Organization  string   `json:"organization"`
	Repository    string   `json:"repository"`
	GHIssueID     int64    `json:"gh_issue_id"`
	GHNumber      int32    `json:"gh_number"`
	Title         string   `json:"title"`
	State         string   `json:"state"`
	HTMLURL       string   `json:"html_url"`
	UserLogin     string   `json:"user_login"`
	Labels        []string `json:"labels"`
	GHCreatedAt   string   `json:"gh_created_at"`
	GHUpdatedAt   string   `json:"gh_updated_at"`
}

func HandleList(w http.ResponseWriter, r *http.Request) {
	 	uid, ok := middleware.UserIDFromCtx(r.Context())
	if !ok || uid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
  // Project scope (already fetched by WithProjectContext)
	pc, ok := middleware.ProjectCtx(r)
	if !ok || pc.Project == nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	projectID := mux.Vars(r)["id"]
	if projectID == "" {
		http.Error(w, "missing project id", http.StatusBadRequest)
		return
	}

	cli := clients.IssueClient()
	out, err := cli.ListIssues(r.Context(), &issuev1.ListIssuesRequest{
		UserId:    uid,
		ProjectId: projectID,
	})
	if err != nil {
		// Project scope errors from data_server will surface here; keep simple for now.
		http.Error(w, "failed to list issues", http.StatusInternalServerError)
		return
	}

	// Map proto → JSON DTO
	items := make([]issueDTO, 0, len(out.GetIssues()))
	for _, it := range out.GetIssues() {
		items = append(items, issueDTO{
			ID:           it.GetId(),
			CreatedAt:    it.GetCreatedAt(),
			UpdatedAt:    it.GetUpdatedAt(),
			ProjectID:    it.GetProjectId(),
			UserID:       it.GetUserId(),
			Organization: it.GetOrganization(),
			Repository:   it.GetRepository(),
			GHIssueID:    it.GetGhIssueId(),
			GHNumber:     it.GetGhNumber(),
			Title:        it.GetTitle(),
			State:        it.GetState(),
			HTMLURL:      it.GetHtmlUrl(),
			UserLogin:    it.GetUserLogin(),
			Labels:       it.GetLabels(),
			GHCreatedAt:  it.GetGhCreatedAt(),
			GHUpdatedAt:  it.GetGhUpdatedAt(),
		})
	}

  w.Header().Set("Content-Type", "application/json")
  _ = json.NewEncoder(w).Encode(map[string]any{
    "items": items,          // ← was "issues"
    "total": len(items),     // optional but nice for the UI
  })
}
