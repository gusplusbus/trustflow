package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gusplusbus/trustflow/data_server/internal/domain"
	"github.com/gusplusbus/trustflow/data_server/internal/repo"
)

type IssueService struct {
	projects ProjectRepoLike // minimal interface: GetByID(ctx, userID, projectID) (*domain.Project, error)
	own      OwnershipRepoLike // optional if you want to ensure ownership exists
	issues   repo.IssueRepo
	db       DBLike // for outbox insert
}

func NewIssueService(p ProjectRepoLike, o OwnershipRepoLike, i repo.IssueRepo, db DBLike) *IssueService {
	return &IssueService{projects: p, own: o, issues: i, db: db}
}

// Minimal interfaces to avoid import cycles; implement with your existing repos.
type ProjectRepoLike interface {
	Get(ctx context.Context, userID, id string) (*domain.Project, error)
}
type OwnershipRepoLike interface {
	ListByProject(ctx context.Context, userID, projectID string) ([]*domain.Ownership, error)
}
type DBLike interface {
	Exec(ctx context.Context, sql string, args ...any) (any, error)
}

func (s *IssueService) Import(ctx context.Context, userID, projectID string, sel []domain.Issue) ([]*domain.Issue, int, error) {
	// 1) scope check: make sure project belongs to user
	if _, err := s.projects.Get(ctx, userID, projectID); err != nil {
		return nil, 0, err
	}

	// 2) minimally require at least one ownership to copy org/repo onto rows
	owns, err := s.own.ListByProject(ctx, userID, projectID)
	if err != nil { return nil, 0, err }
	if len(owns) == 0 { return nil, 0, errors.New("no ownership for project") }
	org := owns[0].Organization
	repoName := owns[0].Repository

	// 3) stamp rows and insert
	now := time.Now().UTC()
	rows := make([]*domain.Issue, 0, len(sel))
	for i := range sel {
		it := sel[i] // copy
		it.UserID, it.ProjectID = userID, projectID
		it.Organization, it.Repository = org, repoName
		it.CreatedAt, it.UpdatedAt = now, now
		rows = append(rows, &it)
	}
	inserted, dups, err := s.issues.InsertMany(ctx, rows)
	if err != nil { return nil, 0, err }

	// 4) outbox for each inserted
	for _, it := range inserted {
		payload, _ := json.Marshal(map[string]any{
			"type": "project_issue.imported.v1",
			"issue": map[string]any{
				"id": it.ID, "project_id": it.ProjectID, "user_id": it.UserID,
				"organization": it.Organization, "repository": it.Repository,
				"gh_issue_id": it.GHIssueID, "gh_number": it.GHNumber,
				"title": it.Title, "state": it.State, "html_url": it.HTMLURL,
        "user_login": it.GHUserLogin, "labels": it.Labels,
				"gh_created_at": it.GHCreatedAt, "gh_updated_at": it.GHUpdatedAt,
			},
		})
		_, _ = s.db.Exec(ctx,
			`INSERT INTO issue_outbox (event_type, payload) VALUES ($1, $2)`,
			"project_issue.imported.v1", payload,
		)
	}

	return inserted, dups, nil
}

func (s *IssueService) List(ctx context.Context, userID, projectID string) ([]*domain.Issue, error) {
	// scope check matches your existing style
	if _, err := s.projects.Get(ctx, userID, projectID); err != nil {
		return nil, err
	}
	return s.issues.ListByProject(ctx, userID, projectID)
}

func (s *IssueService) ExistsByGhID(ctx context.Context, ghIssueID int64) (bool, error) {
	return s.issues.ExistsByGhID(ctx, ghIssueID)
}
