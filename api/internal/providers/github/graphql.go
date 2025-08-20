package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Minimal GraphQL client using installation token
type GraphQLClient struct {
	http *http.Client
}

func NewGraphQLClient(timeout time.Duration) *GraphQLClient {
	if timeout <= 0 {
		timeout = 12 * time.Second
	}
	return &GraphQLClient{http: &http.Client{Timeout: timeout}}
}

const issueTimelineQuery = `
query IssueTimeline($owner: String!, $repo: String!, $number: Int!, $pageSize: Int!, $after: String) {
  repository(owner: $owner, name: $repo) {
    issue(number: $number) {
      id
      timelineItems(first: $pageSize, after: $after) {
        pageInfo { hasNextPage endCursor }
        nodes {
          __typename
          ... on LabeledEvent { id createdAt actor { login } label { name } }
          ... on UnlabeledEvent { id createdAt actor { login } label { name } }
          ... on AssignedEvent { id createdAt actor { login } assignee { __typename ... on User { login } } }
          ... on UnassignedEvent { id createdAt actor { login } assignee { __typename ... on User { login } } }
          ... on ClosedEvent { id createdAt actor { login } }
          ... on ReopenedEvent { id createdAt actor { login } }
          ... on MilestonedEvent { id createdAt actor { login } milestoneTitle }
          ... on DemilestonedEvent { id createdAt actor { login } milestoneTitle }
          ... on RenamedTitleEvent { id createdAt actor { login } currentTitle previousTitle }
          ... on IssueComment { id createdAt author { login } body }
        }
      }
    }
  }
}`

type page struct {
	EndCursor   string
	HasNextPage bool
	IssueID     string
	Items       []map[string]any // raw nodes
}

func (c *GraphQLClient) FetchIssueTimelinePage(ctx context.Context, token, owner, repo string, number int, pageSize int, after string) (*page, error) {
	reqBody := map[string]any{
		"query": issueTimelineQuery,
		"variables": map[string]any{
			"owner":    owner,
			"repo":     repo,
			"number":   number,
			"pageSize": pageSize,
			"after":    func() any { if after == "" { return nil }; return after }(),
		},
	}
	buf, _ := json.Marshal(reqBody)

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/graphql", bytes.NewReader(buf))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "trustflow/graphql-issue-timeline")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graphql do: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var e struct{ Message string `json:"message"` }
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return nil, fmt.Errorf("graphql (%d): %s", resp.StatusCode, e.Message)
	}

	var out struct {
		Data struct {
			Repository struct {
				Issue struct {
					Id            string `json:"id"`
					TimelineItems struct {
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
						Nodes []map[string]any `json:"nodes"`
					} `json:"timelineItems"`
				} `json:"issue"`
			} `json:"repository"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode graphql: %w", err)
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("graphql: %s", out.Errors[0].Message)
	}

	p := &page{
		EndCursor:   out.Data.Repository.Issue.TimelineItems.PageInfo.EndCursor,
		HasNextPage: out.Data.Repository.Issue.TimelineItems.PageInfo.HasNextPage,
		IssueID:     out.Data.Repository.Issue.Id,
		Items:       out.Data.Repository.Issue.TimelineItems.Nodes,
	}
	return p, nil
}
