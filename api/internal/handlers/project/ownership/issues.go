package ownership

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gusplusbus/trustflow/api/internal/middleware"
	ghprov "github.com/gusplusbus/trustflow/api/internal/providers/github"
)

// ---------- helpers ----------

func parseInt(s string, def int) int {
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
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func pick(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
func mapLabels(ls []struct{ Name string `json:"name"` }) []string {
	out := make([]string, 0, len(ls))
	for _, l := range ls {
		name := strings.TrimSpace(l.Name)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}


func HandleIssues(w http.ResponseWriter, r *http.Request) {
	// Always derive owner/repo from project context (first ownership)
	pc, ok := middleware.ProjectCtx(r)
	if !ok || pc == nil || len(pc.Ownerships) == 0 {
		http.Error(w, "no ownership configured for this project", http.StatusBadRequest)
		return
	}
	owner := strings.TrimSpace(pc.Ownerships[0].GetOrganization())
	repo := strings.TrimSpace(pc.Ownerships[0].GetRepository())
	if owner == "" || repo == "" {
		http.Error(w, "ownership is missing organization or repository", http.StatusBadRequest)
		return
	}

	// Filters (owner/repo intentionally NOT read from query)
	q := r.URL.Query()
	state := pick(q.Get("state"), "open") // open|closed|all
	labels := strings.TrimSpace(q.Get("labels"))
	assignee := q.Get("assignee") // "", "*", or login
	since := strings.TrimSpace(q.Get("since"))
	perPage := clamp(parseInt(q.Get("per_page"), 50), 1, 100)
	page := max(parseInt(q.Get("page"), 1), 1)
	search := strings.TrimSpace(q.Get("search"))
	useSearch := search != ""

	// GitHub installation token for the repo
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

	// Build upstream URL
	var upstream string
	if useSearch {
		terms := []string{fmt.Sprintf("repo:%s/%s", owner, repo), "is:issue"}
		if state != "all" {
			terms = append(terms, "state:"+state)
		}
		if labels != "" {
			for _, lbl := range strings.Split(labels, ",") {
				lbl = strings.TrimSpace(lbl)
				if lbl != "" {
					terms = append(terms, `label:"`+strings.ReplaceAll(lbl, `"`, `\"`)+`"`)
				}
			}
		}
		switch assignee {
		case "*":
			terms = append(terms, "assignee:*")
		default:
			if assignee != "" {
				terms = append(terms, "assignee:"+assignee)
			}
		}
		if since != "" {
			if t, err := time.Parse(time.RFC3339, since); err == nil {
				terms = append(terms, "updated:>="+t.Format("2006-01-02"))
			}
		}
		if search != "" {
			terms = append(terms, search)
		}
		params := url.Values{}
		params.Set("q", strings.Join(terms, " "))
		params.Set("per_page", strconv.Itoa(perPage))
		params.Set("page", strconv.Itoa(page))
		upstream = "https://api.github.com/search/issues?" + params.Encode()
	} else {
		params := url.Values{}
		params.Set("state", state)
		if labels != "" {
			params.Set("labels", labels)
		}
		if assignee != "" {
			params.Set("assignee", assignee)
		}
		if since != "" {
			params.Set("since", since)
		}
		params.Set("per_page", strconv.Itoa(perPage))
		params.Set("page", strconv.Itoa(page))
		upstream = fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?%s",
			url.PathEscape(owner), url.PathEscape(repo), params.Encode())
	}

	// Execute request
	req, _ := http.NewRequestWithContext(r.Context(), "GET", upstream, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "trustflow/ownership-issues")
	client := &http.Client{Timeout: 12 * time.Second}

	res, err := client.Do(req)
	if err != nil {
		http.Error(w, "github request: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	ri := &rateInfo{}
	if lim, _ := strconv.Atoi(res.Header.Get("X-RateLimit-Limit")); lim > 0 {
		ri.Limit = lim
	}
	if rem, _ := strconv.Atoi(res.Header.Get("X-RateLimit-Remaining")); rem >= 0 {
		ri.Remaining = rem
	}
	if rs, _ := strconv.Atoi(res.Header.Get("X-RateLimit-Reset")); rs > 0 {
		ri.Reset = rs
	}

	if res.StatusCode != 200 {
		var body struct{ Message string `json:"message"` }
		_ = json.NewDecoder(res.Body).Decode(&body)
		http.Error(w, fmt.Sprintf("github (%d): %s", res.StatusCode, strings.TrimSpace(body.Message)), http.StatusBadGateway)
		return
	}

	out := listResp{Rate: ri}

	// Decode response
	if useSearch {
		var sr struct {
			Total int `json:"total_count"`
			Items []struct {
				ID          int64  `json:"id"`
				Number      int    `json:"number"`
				Title       string `json:"title"`
				State       string `json:"state"`
				HTMLURL     string `json:"html_url"`
				CreatedAt   string `json:"created_at"`
				UpdatedAt   string `json:"updated_at"`
				User        *struct{ Login string `json:"login"` } `json:"user"`
				Labels      []struct{ Name string `json:"name"` } `json:"labels"`
				PullRequest *struct{}                               `json:"pull_request"`
			} `json:"items"`
		}
		if err := json.NewDecoder(res.Body).Decode(&sr); err != nil {
			http.Error(w, "decode search response: "+err.Error(), http.StatusBadGateway)
			return
		}
		out.Total = sr.Total
		for _, it := range sr.Items {
			if it.PullRequest != nil {
				continue
			}
			out.Items = append(out.Items, issueItem{
				ID:        it.ID,
				Number:    it.Number,
				Title:     it.Title,
				State:     it.State,
				HTMLURL:   it.HTMLURL,
				CreatedAt: it.CreatedAt,
				UpdatedAt: it.UpdatedAt,
				UserLogin: func() string { if it.User != nil { return it.User.Login } ; return "" }(),
				Labels:    mapLabels(it.Labels),
			})
		}
	} else {
		var arr []struct {
			ID          int64  `json:"id"`
			Number      int    `json:"number"`
			Title       string `json:"title"`
			State       string `json:"state"`
			HTMLURL     string `json:"html_url"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
			User        *struct{ Login string `json:"login"` } `json:"user"`
			Labels      []struct{ Name string `json:"name"` }   `json:"labels"`
			PullRequest *struct{}                                `json:"pull_request"`
		}
		if err := json.NewDecoder(res.Body).Decode(&arr); err != nil {
			http.Error(w, "decode issues response: "+err.Error(), http.StatusBadGateway)
			return
		}
		out.Total = -1
		for _, it := range arr {
			if it.PullRequest != nil {
				continue
			}
			out.Items = append(out.Items, issueItem{
				ID:        it.ID,
				Number:    it.Number,
				Title:     it.Title,
				State:     it.State,
				HTMLURL:   it.HTMLURL,
				CreatedAt: it.CreatedAt,
				UpdatedAt: it.UpdatedAt,
				UserLogin: func() string { if it.User != nil { return it.User.Login } ; return "" }(),
				Labels:    mapLabels(it.Labels),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(out)
}
