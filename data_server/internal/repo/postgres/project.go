package postgres

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gusplusbus/trustflow/data_server/internal/domain"
	"github.com/gusplusbus/trustflow/data_server/internal/repo"
)

//go:embed queries/*.sql
var projectFS embed.FS

type ProjectPG struct {
	db      *pgxpool.Pool
	qCreate string
	qGet    string
	qUpdate string
	qDelete string
	qList   string
	qCount  string
}

func NewProjectPG(db *pgxpool.Pool) (*ProjectPG, error) {
	read := func(name string) string {
		b, err := projectFS.ReadFile("queries/" + name)
		if err != nil { panic(err) }
		return string(b)
	}
	return &ProjectPG{
		db:      db,
		qCreate: read("create_project.sql"),
		qGet:    read("get_project.sql"),
		qUpdate: read("update_project.sql"),
		qDelete: read("delete_project.sql"),
		qList:   read("list_projects.sql"),
		qCount:  read("count_projects.sql"),
	}, nil
}

var _ repo.ProjectRepo = (*ProjectPG)(nil)

func (pg *ProjectPG) Create(ctx context.Context, in *domain.Project) (*domain.Project, error) {
	out := *in
	err := pg.db.QueryRow(ctx, pg.qCreate,
		in.UserID, in.Title, in.Description, in.DurationEstimate, in.TeamSize, in.ApplicationCloseTime,
	).Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt, &out.Title, &out.Description,
		&out.DurationEstimate, &out.TeamSize, &out.ApplicationCloseTime)
	if err != nil { return nil, fmt.Errorf("create: %w", err) }
	return &out, nil
}

func (pg *ProjectPG) Get(ctx context.Context, userID, id string) (*domain.Project, error) {
	var out domain.Project
	out.ID = id
	out.UserID = userID
	err := pg.db.QueryRow(ctx, pg.qGet, id, userID).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt, &out.Title, &out.Description,
			&out.DurationEstimate, &out.TeamSize, &out.ApplicationCloseTime)
	if err != nil { return nil, fmt.Errorf("get: %w", err) }
	return &out, nil
}

func (pg *ProjectPG) Update(ctx context.Context, in *domain.Project) (*domain.Project, error) {
	out := *in
	err := pg.db.QueryRow(ctx, pg.qUpdate,
		in.Title, in.Description, in.DurationEstimate, in.TeamSize, in.ApplicationCloseTime,
		in.ID, in.UserID,
	).Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt, &out.Title, &out.Description,
		&out.DurationEstimate, &out.TeamSize, &out.ApplicationCloseTime)
	if err != nil { return nil, fmt.Errorf("update: %w", err) }
	return &out, nil
}

func (pg *ProjectPG) Delete(ctx context.Context, userID, id string) (bool, error) {
	tag, err := pg.db.Exec(ctx, pg.qDelete, id, userID)
	if err != nil { return false, fmt.Errorf("delete: %w", err) }
	return tag.RowsAffected() > 0, nil
}

func (pg *ProjectPG) List(ctx context.Context, p repo.RepoListParams) (*repo.ListResult, error) {
	sortCol := map[string]bool{
		"created_at": true, "updated_at": true, "title": true,
		"team_size": true, "duration_estimate": true,
	}
	if !sortCol[p.SortColumn] {
		p.SortColumn = "created_at"
	}
	dir := "DESC"
	if p.SortDir == "ASC" {
		dir = "ASC"
	}

	offset := p.Page * p.PageSize
	if offset < 0 { offset = 0 }
	if p.PageSize <= 0 { p.PageSize = 25 }
	if p.PageSize > 200 { p.PageSize = 200 }

	var total int64
	if err := pg.db.QueryRow(ctx, pg.qCount, p.UserID, p.Q).Scan(&total); err != nil {
		return nil, fmt.Errorf("list count: %w", err)
	}

	sql := fmt.Sprintf(pg.qList, p.SortColumn, dir)
	rows, err := pg.db.Query(ctx, sql, p.UserID, p.Q, offset, p.PageSize)
	if err != nil {
		return nil, fmt.Errorf("list query: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.Project, 0, p.PageSize)
	for rows.Next() {
		var pr domain.Project
		pr.UserID = p.UserID
		if err := rows.Scan(
			&pr.ID,
			&pr.CreatedAt,
			&pr.UpdatedAt,
			&pr.Title,
			&pr.Description,
			&pr.DurationEstimate,
			&pr.TeamSize,
			&pr.ApplicationCloseTime,
		); err != nil {
			return nil, fmt.Errorf("list scan: %w", err)
		}
		out = append(out, &pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list rows: %w", err)
	}

	return &repo.ListResult{Projects: out, Total: total}, nil
}

// (optional) helper
func toRFC3339(t time.Time) string { return t.UTC().Format(time.RFC3339) }
