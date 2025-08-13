package repo

import (
	"context"

	"github.com/gusplusbus/trustflow/data_server/internal/domain"
)

/* Projects */
type ProjectRepo interface {
	Create(ctx context.Context, in *domain.Project) (*domain.Project, error)
	Get(ctx context.Context, userID, id string) (*domain.Project, error)
	Update(ctx context.Context, in *domain.Project) (*domain.Project, error)
	Delete(ctx context.Context, userID, id string) (bool, error)
	List(ctx context.Context, p RepoListParams) (*ListResult, error)
}

type RepoListParams struct {
	UserID     string
	Page       int
	PageSize   int
	SortColumn string // created_at | updated_at | title | team_size | duration_estimate
	SortDir    string // "ASC" | "DESC"
	Q          string // free-text
}

type ListResult struct {
	Projects []*domain.Project
	Total    int64
}

/* Ownerships — matches ownership.proto (Create/Update/Delete) + ListByProject for hydration */
type OwnershipRepo interface {
	Create(ctx context.Context, in *domain.Ownership) (*domain.Ownership, error)
	Update(ctx context.Context, in *domain.Ownership) (*domain.Ownership, error)
	Delete(ctx context.Context, userID, id string) (bool, error)
	ListByProject(ctx context.Context, userID, projectID string) ([]*domain.Ownership, error)
}

