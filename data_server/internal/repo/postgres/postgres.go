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
var qFS embed.FS

type ProjectPG struct {
	db    *pgxpool.Pool
	qCreate string
	qGet    string
	qUpdate string
	qDelete string
}

func NewProjectPG(db *pgxpool.Pool) (*ProjectPG, error) {
	read := func(name string) string {
		b, err := qFS.ReadFile("queries/" + name)
		if err != nil { panic(err) }
		return string(b)
	}
	return &ProjectPG{
		db:     db,
		qCreate: read("create_project.sql"),
		qGet:    read("get_project.sql"),
		qUpdate: read("update_project.sql"),
		qDelete: read("delete_project.sql"),
	}, nil
}

var _ repo.ProjectRepo = (*ProjectPG)(nil)

func (pg *ProjectPG) Create(ctx context.Context, in *domain.Project) (*domain.Project, error) {
	out := *in // copy
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

// (optional) small helper if you ever need time conversions in one spot
func toRFC3339(t time.Time) string { return t.UTC().Format(time.RFC3339) }
