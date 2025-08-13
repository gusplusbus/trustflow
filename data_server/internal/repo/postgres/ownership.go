package postgres

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gusplusbus/trustflow/data_server/internal/domain"
	"github.com/gusplusbus/trustflow/data_server/internal/repo"
)

//go:embed queries/*.sql
var ownershipFS embed.FS

type OwnershipPG struct {
	db       *pgxpool.Pool
	qCreate  string
	qUpdate  string
	qDelete  string
	qListByP string
}

func NewOwnershipPG(db *pgxpool.Pool) (*OwnershipPG, error) {
	read := func(name string) string {
		b, err := ownershipFS.ReadFile("queries/" + name)
		if err != nil { panic(err) }
		return string(b)
	}
	return &OwnershipPG{
		db:       db,
		qCreate:  read("create_ownership.sql"),
		qUpdate:  read("update_ownership.sql"),
		qDelete:  read("delete_ownership.sql"),
		qListByP: read("list_ownership_by_project.sql"),
	}, nil
}

var _ repo.OwnershipRepo = (*OwnershipPG)(nil)

func (pg *OwnershipPG) Create(ctx context.Context, in *domain.Ownership) (*domain.Ownership, error) {
	out := *in
	err := pg.db.QueryRow(ctx, pg.qCreate,
		in.UserID, in.ProjectID, in.Organization, in.Repository, in.Provider, in.WebURL,
	).Scan(
		&out.ID, &out.CreatedAt, &out.UpdatedAt,
		&out.ProjectID, &out.UserID,
		&out.Organization, &out.Repository,
		&out.Provider, &out.WebURL,
	)
	if err != nil {
		return nil, fmt.Errorf("ownership create: %w", err)
	}
	return &out, nil
}

func (pg *OwnershipPG) Update(ctx context.Context, in *domain.Ownership) (*domain.Ownership, error) {
	out := *in
	err := pg.db.QueryRow(ctx, pg.qUpdate,
		in.Organization, in.Repository, in.Provider, in.WebURL,
		in.ID, in.UserID,
	).Scan(
		&out.ID, &out.CreatedAt, &out.UpdatedAt,
		&out.ProjectID, &out.UserID,
		&out.Organization, &out.Repository,
		&out.Provider, &out.WebURL,
	)
	if err != nil {
		return nil, fmt.Errorf("ownership update: %w", err)
	}
	return &out, nil
}

func (pg *OwnershipPG) Delete(ctx context.Context, userID, id string) (bool, error) {
	tag, err := pg.db.Exec(ctx, pg.qDelete, id, userID)
	if err != nil {
		return false, fmt.Errorf("ownership delete: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (pg *OwnershipPG) ListByProject(ctx context.Context, userID, projectID string) ([]*domain.Ownership, error) {
	rows, err := pg.db.Query(ctx, pg.qListByP, userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("ownership list_by_project: %w", err)
	}
	defer rows.Close()

	var out []*domain.Ownership
	for rows.Next() {
		var o domain.Ownership
		if err := rows.Scan(
			&o.ID, &o.CreatedAt, &o.UpdatedAt,
			&o.ProjectID, &o.UserID,
			&o.Organization, &o.Repository,
			&o.Provider, &o.WebURL,
		); err != nil {
			return nil, fmt.Errorf("ownership scan: %w", err)
		}
		out = append(out, &o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ownership rows: %w", err)
	}
	return out, nil
}
