package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gusplusbus/trustflow/data_server/internal/domain"
	"github.com/gusplusbus/trustflow/data_server/internal/repo"
)

//go:embed queries/*.sql
var issueFS embed.FS

type IssuePG struct {
	db            *pgxpool.Pool
	qCreate       string
	qGetByUnique  string
	qListByProj   string
}

func NewIssuePG(db *pgxpool.Pool) (*IssuePG, error) {
	read := func(name string) string {
		b, err := issueFS.ReadFile("queries/" + name)
		if err != nil { panic(err) }
		return string(b)
	}
	return &IssuePG{
		db:           db,
		qCreate:      read("create_issue.sql"),
		qGetByUnique: read("get_issue_by_unique.sql"),
		qListByProj:  read("list_issues_by_project.sql"),
	}, nil
}

var _ repo.IssueRepo = (*IssuePG)(nil)

// Create inserts a new issue. If the unique key already exists, it does NOT
// overwrite — it returns the existing row (matching your “don’t override old ones” rule).
func (pg *IssuePG) Create(ctx context.Context, in *domain.Issue) (*domain.Issue, error) {
	out := domain.Issue{}
	// First try insert … ON CONFLICT DO NOTHING RETURNING …
	err := pg.db.QueryRow(ctx, pg.qCreate,
		in.ProjectID, in.UserID,
		in.Organization, in.Repository,
		in.GHIssueID, in.GHNumber,
		in.Title, in.State, in.HTMLURL,
		in.Labels, in.GHUserLogin,
		in.GHCreatedAt, in.GHUpdatedAt,
	).Scan(
		&out.ID, &out.CreatedAt, &out.UpdatedAt,
		&out.ProjectID, &out.UserID,
		&out.Organization, &out.Repository,
		&out.GHIssueID, &out.GHNumber,
		&out.Title, &out.State, &out.HTMLURL,
		&out.Labels, &out.GHUserLogin,
		&out.GHCreatedAt, &out.GHUpdatedAt,
	)
	if err == nil {
		return &out, nil
	}
	// If no row returned because of conflict, fetch the existing row.
	if errors.Is(err, pgx.ErrNoRows) {
		return pg.getByUnique(ctx, in.ProjectID, in.Organization, in.Repository, in.GHNumber)
	}
	return nil, fmt.Errorf("issue create: %w", err)
}

func (pg *IssuePG) getByUnique(ctx context.Context, projectID, org, repoName string, number int32) (*domain.Issue, error) {
	var out domain.Issue
	err := pg.db.QueryRow(ctx, pg.qGetByUnique, projectID, org, repoName, number).Scan(
		&out.ID, &out.CreatedAt, &out.UpdatedAt,
		&out.ProjectID, &out.UserID,
		&out.Organization, &out.Repository,
		&out.GHIssueID, &out.GHNumber,
		&out.Title, &out.State, &out.HTMLURL,
		&out.Labels, &out.GHUserLogin,
		&out.GHCreatedAt, &out.GHUpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("issue get_by_unique: %w", err)
	}
	return &out, nil
}

func (pg *IssuePG) ListByProject(ctx context.Context, userID, projectID string) ([]*domain.Issue, error) {
	rows, err := pg.db.Query(ctx, pg.qListByProj, userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("issue list_by_project: %w", err)
	}
	defer rows.Close()

	var out []*domain.Issue
	for rows.Next() {
		var it domain.Issue
		if err := rows.Scan(
			&it.ID, &it.CreatedAt, &it.UpdatedAt,
			&it.ProjectID, &it.UserID,
			&it.Organization, &it.Repository,
			&it.GHIssueID, &it.GHNumber,
			&it.Title, &it.State, &it.HTMLURL,
			&it.Labels, &it.GHUserLogin,
			&it.GHCreatedAt, &it.GHUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("issue scan: %w", err)
		}
		out = append(out, &it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("issue rows: %w", err)
	}
	return out, nil
}

// (Optional) tiny helper you might want later for “freshness windows”
func (pg *IssuePG) Now(ctx context.Context) (time.Time, error) {
	var t time.Time
	if err := pg.db.QueryRow(ctx, "select now()").Scan(&t); err != nil {
		return time.Time{}, err
	}
	return t, nil
}
