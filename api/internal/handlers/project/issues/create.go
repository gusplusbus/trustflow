package issues

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"strconv"

	"github.com/gusplusbus/trustflow/api/internal/middleware"
	ghprov "github.com/gusplusbus/trustflow/api/internal/providers/github"
)

// ---------- request/response types (local to this package) ----------

type importIssue struct {
	ID     int64 `json:"id"`
	Number int   `json:"number"`
}

type importReq struct {
	Issues []importIssue `json:"issues"`
}

type issueItem struct {
	ID        int64    `json:"id"`
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	State     string   `json:"state"`
	HTMLURL   string   `json:"html_url"`
	UserLogin string   `json:"user_login,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type rateInfo struct {
	Limit     int `json:"limit"`
	Remaining int `json:"remaining"`
	Reset     int `json:"reset"`
}

type listResp struct {
	Items []issueItem `json:"items"`
	Total int         `json:"total"`
	Rate  *rateInfo   `json:"rate,omitempty"`
}

// ---------- helpers ----------

func mapLabels(ls []struct{ Name string `json:"name"` }) []string {
	out := make([]string, 0, len(ls))
	for _, l := range ls {
		if n := strings.TrimSpace(l.Name); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// HandleCreate fetches the selected GitHub issues by number for the project's repo.
// It enforces: issue must be OPEN and have NO assignees. PRs are ignored.
// It does not persist anything yet; it just returns the selected issues' details.
func HandleCreate(w http.ResponseWriter, r *http.Request) {
	// Project + ownership context
	pc, ok := middleware.ProjectCtx(r)
	if !ok || pc == nil || pc.Project == nil || len(pc.Ownerships) == 0 {
		http.Error(w, "no ownership configured for this project", http.StatusBadRequest)
		return
	}
	owner := strings.TrimSpace(pc.Ownerships[0].GetOrganization())
	repo := strings.TrimSpace(pc.Ownerships[0].GetRepository())
	if owner == "" || repo == "" {
		http.Error(w, "ownership is missing organization or repository", http.StatusBadRequest)
		return
	}

	// Parse selection
	var req importReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(req.Issues) == 0 {
		http.Error(w, "issues array required", http.StatusBadRequest)
		return
	}

	// GitHub installation token
	ver, err := ghprov.NewVerifierFromEnv()
	if err != nil {
		http.Error(w, "github verifier: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tok, err := ver.InstallationTokenForRepo(r.Context(), owner, repo)
	if err != nil {
		http.Error(w, "installation token: "+err.Error(), http.StatusBadGateway)
		return
	}

	client := &http.Client{Timeout: 12 * time.Second}
	items := make([]issueItem, 0, len(req.Issues))
	var ri *rateInfo
	skipped := 0

	for _, sel := range req.Issues {
		if sel.Number <= 0 {
			skipped++
			continue
		}
		u := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d",
			url.PathEscape(owner), url.PathEscape(repo), sel.Number)

		reqGit, _ := http.NewRequestWithContext(r.Context(), "GET", u, nil)
		reqGit.Header.Set("Authorization", "Bearer "+tok)
		reqGit.Header.Set("Accept", "application/vnd.github+json")
		reqGit.Header.Set("User-Agent", "trustflow/issues-create")

		res, err := client.Do(reqGit)
		if err != nil {
			http.Error(w, "github request: "+err.Error(), http.StatusBadGateway)
			return
		}
		// capture latest rate snapshot
		curRI := &rateInfo{}
		if lim, _ := strconv.Atoi(res.Header.Get("X-RateLimit-Limit")); lim > 0 { curRI.Limit = lim }
		if rem, _ := strconv.Atoi(res.Header.Get("X-RateLimit-Remaining")); rem >= 0 { curRI.Remaining = rem }
		if rs, _ := strconv.Atoi(res.Header.Get("X-RateLimit-Reset")); rs > 0 { curRI.Reset = rs }
		ri = curRI

		if res.StatusCode != http.StatusOK {
			var body struct{ Message string `json:"message"` }
			_ = json.NewDecoder(res.Body).Decode(&body)
			res.Body.Close()
			http.Error(w, fmt.Sprintf("github (%d): %s", res.StatusCode, strings.TrimSpace(body.Message)), http.StatusBadGateway)
			return
		}

		var gh struct {
			ID          int64  `json:"id"`
			Number      int    `json:"number"`
			Title       string `json:"title"`
			State       string `json:"state"`
			HTMLURL     string `json:"html_url"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
			User        *struct{ Login string `json:"login"` } `json:"user"`
			Labels      []struct{ Name string `json:"name"` }   `json:"labels"`
			Assignees   []struct{ Login string `json:"login"` }  `json:"assignees"`
			PullRequest *struct{}                                `json:"pull_request"`
		}
		if err := json.NewDecoder(res.Body).Decode(&gh); err != nil {
			res.Body.Close()
			http.Error(w, "decode issue response: "+err.Error(), http.StatusBadGateway)
			return
		}
		res.Body.Close()

		// Skip PRs
		if gh.PullRequest != nil {
			skipped++
			continue
		}
		// Enforce: open & unassigned
		if strings.ToLower(strings.TrimSpace(gh.State)) != "open" || len(gh.Assignees) > 0 {
			skipped++
			continue
		}

		items = append(items, issueItem{
			ID:        gh.ID,
			Number:    gh.Number,
			Title:     gh.Title,
			State:     gh.State,
			HTMLURL:   gh.HTMLURL,
			CreatedAt: gh.CreatedAt,
			UpdatedAt: gh.UpdatedAt,
			UserLogin: func() string { if gh.User != nil { return gh.User.Login }; return "" }(),
			Labels:    mapLabels(gh.Labels),
		})
	}

	// If nothing qualified, explain clearly
	if len(items) == 0 {
		http.Error(w, "no issues matched the import criteria (must be open and unassigned)", http.StatusUnprocessableEntity)
		return
	}

	// Normal success path
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if skipped > 0 {
		w.Header().Set("X-Skipped-Count", strconv.Itoa(skipped))
	}
	_ = json.NewEncoder(w).Encode(listResp{
		Items: items,
		Total: len(items),
		Rate:  ri,
	})
}
