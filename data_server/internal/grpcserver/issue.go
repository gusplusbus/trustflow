package grpcserver

import (
	"context"
	"time"
  "strings"
	issuev1 "github.com/gusplusbus/trustflow/data_server/gen/issuev1"
	"github.com/gusplusbus/trustflow/data_server/internal/domain"
	"github.com/gusplusbus/trustflow/data_server/internal/service"
)

// just below imports or near the method (unexported helper)
func parseRFC3339OrZero(s string) time.Time {
    s = strings.TrimSpace(s)
    if s == "" {
        return time.Time{}
    }
    // GitHub timestamps are RFC3339 (e.g., 2024-06-20T12:34:56Z)
    if t, err := time.Parse(time.RFC3339, s); err == nil {
        return t.UTC()
    }
    // fall back: try time.RFC3339Nano if you see nano precision in your data
    if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
        return t.UTC()
    }
    return time.Time{}
}

type IssueServer struct {
	issuev1.UnimplementedIssueServiceServer
	svc *service.IssueService
}

func NewIssueServer(svc *service.IssueService) *IssueServer { return &IssueServer{svc: svc} }

func (s *IssueServer) Health(ctx context.Context, _ *issuev1.HealthRequest) (*issuev1.HealthResponse, error) {
	return &issuev1.HealthResponse{Status: "ok"}, nil
}

func (s *IssueServer) ImportIssues(ctx context.Context, req *issuev1.ImportIssuesRequest) (*issuev1.ImportIssuesResponse, error) {
    // map request to domain rows (API already fetched GH details, we should persist all of them)
    rows := make([]domain.Issue, 0, len(req.GetIssues()))
    for _, sel := range req.GetIssues() {
        rows = append(rows, domain.Issue{
            GHIssueID:    sel.GetId(),
            GHNumber:     sel.GetNumber(),
            Title:        sel.GetTitle(),
            State:        sel.GetState(),
            HTMLURL:      sel.GetHtmlUrl(),
            GHUserLogin:  sel.GetUserLogin(),
            Labels:       sel.GetLabels(),
            GHCreatedAt:  parseRFC3339OrZero(sel.GetGhCreatedAt()),
            GHUpdatedAt:  parseRFC3339OrZero(sel.GetGhUpdatedAt()),
        })
    }

    inserted, dups, err := s.svc.Import(ctx, req.GetUserId(), req.GetProjectId(), rows)
    if err != nil { return nil, err }

    out := &issuev1.ImportIssuesResponse{Duplicates: int32(dups)}
    for _, it := range inserted {
        out.Imported = append(out.Imported, toIssueProto(it))
    }
    return out, nil
}

func (s *IssueServer) ListIssues(ctx context.Context, req *issuev1.ListIssuesRequest) (*issuev1.ListIssuesResponse, error) {
	rows, err := s.svc.List(ctx, req.GetUserId(), req.GetProjectId())
	if err != nil { return nil, err }
	out := &issuev1.ListIssuesResponse{}
	for _, it := range rows {
		out.Issues = append(out.Issues, toIssueProto(it))
	}
	return out, nil
}

func toIssueProto(it *domain.Issue) *issuev1.Issue {
	return &issuev1.Issue{
		Id:          it.ID,
		CreatedAt:   it.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   it.UpdatedAt.UTC().Format(time.RFC3339),
		ProjectId:   it.ProjectID,
		UserId:      it.UserID,
		Organization: it.Organization,
		Repository:  it.Repository,
		GhIssueId:   it.GHIssueID,
		GhNumber:    it.GHNumber,
		Title:       it.Title,
		State:       it.State,
		HtmlUrl:     it.HTMLURL,
		UserLogin:   it.GHUserLogin,
		Labels:      it.Labels,
		GhCreatedAt: it.GHCreatedAt.UTC().Format(time.RFC3339),
		GhUpdatedAt: it.GHUpdatedAt.UTC().Format(time.RFC3339),
	}
}
