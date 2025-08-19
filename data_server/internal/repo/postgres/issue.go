package postgres

import (
	"context"
	"database/sql"
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
  qInsertMany   string
  qExistsByGhID string
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
    qInsertMany:  read("insert_many_issues.sql"),
    qExistsByGhID: read("exists_by_gh_id.sql"),
	}, nil
}

type IssueRepo interface {
	InsertMany(ctx context.Context, in []*domain.Issue) ([]*domain.Issue, int, error)
	ListByProject(ctx context.Context, userID, projectID string) ([]*domain.Issue, error)
	ExistsByGhID(ctx context.Context, ghIssueID int64) (bool, error)
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

func (pg *IssuePG) InsertMany(ctx context.Context, in []*domain.Issue) ([]*domain.Issue, int, error) {
	n := len(in)
	if n == 0 {
		return nil, 0, nil
	}

	userIDs := make([]string, 0, n)
	projectIDs := make([]string, 0, n)
	orgs := make([]string, 0, n)
	repos := make([]string, 0, n)
	ghiIDs := make([]int64, 0, n)
	ghNums := make([]int32, 0, n)
	titles := make([]string, 0, n)
	states := make([]string, 0, n)
	htmls := make([]string, 0, n)
	labels := make([][]string, 0, n)
	userLogins := make([]string, 0, n)
	ghCreated := make([]*time.Time, 0, n) // nullable
	ghUpdated := make([]*time.Time, 0, n) // nullable
	created := make([]time.Time, 0, n)
	updated := make([]time.Time, 0, n)

	for _, it := range in {
		userIDs = append(userIDs, it.UserID)
		projectIDs = append(projectIDs, it.ProjectID)
		orgs = append(orgs, it.Organization)
		repos = append(repos, it.Repository)
		ghiIDs = append(ghiIDs, it.GHIssueID)
		ghNums = append(ghNums, int32(it.GHNumber))
		titles = append(titles, it.Title)
		states = append(states, it.State)
		htmls = append(htmls, it.HTMLURL)
		userLogins = append(userLogins, it.GHUserLogin)
		labels = append(labels, it.Labels)

		if it.GHCreatedAt.IsZero() { ghCreated = append(ghCreated, nil) } else { t := it.GHCreatedAt; ghCreated = append(ghCreated, &t) }
		if it.GHUpdatedAt.IsZero() { ghUpdated = append(ghUpdated, nil) } else { t := it.GHUpdatedAt; ghUpdated = append(ghUpdated, &t) }

		created = append(created, it.CreatedAt)
		updated = append(updated, it.UpdatedAt)
	}

	rows, err := pg.db.Query(ctx, pg.qInsertMany,
		userIDs, projectIDs, orgs, repos,
		ghiIDs, ghNums,
		titles, states, htmls,
		labels, userLogins,
		ghCreated, ghUpdated,
		created, updated,
	)
	if err != nil { return nil, 0, fmt.Errorf("issue insert_many: %w", err) }
	defer rows.Close()

	out := make([]*domain.Issue, 0, n)

  for rows.Next() {
    var it domain.Issue

    // use nullable temps for the two GH timestamps
    var ghCreated sql.NullTime
    var ghUpdated sql.NullTime

    if err := rows.Scan(
      &it.ID, &it.CreatedAt, &it.UpdatedAt,
      &it.ProjectID, &it.UserID,
      &it.Organization, &it.Repository,
      &it.GHIssueID, &it.GHNumber,
      &it.Title, &it.State, &it.HTMLURL,
      &it.Labels, &it.GHUserLogin,
      &ghCreated, &ghUpdated, // <-- changed
    ); err != nil {
      return nil, 0, fmt.Errorf("issue insert_many scan: %w", err)
    }

    if ghCreated.Valid { it.GHCreatedAt = ghCreated.Time }
    if ghUpdated.Valid { it.GHUpdatedAt = ghUpdated.Time }

    out = append(out, &it)
  }

  if err := rows.Err(); err != nil { return nil, 0, fmt.Errorf("issue insert_many rows: %w", err) }

  dups := n - len(out)
  return out, dups, nil
}

func (pg *IssuePG) ExistsByGhID(ctx context.Context, ghIssueID int64) (bool, error) {
	var exists bool
	err := pg.db.QueryRow(ctx, pg.qExistsByGhID, ghIssueID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// no row means false
			return false, nil
		}
		return false, fmt.Errorf("issue exists_by_gh_id: %w", err)
	}
	return exists, nil
}

