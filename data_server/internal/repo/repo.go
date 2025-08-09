package repo

import (
	"context"

	"github.com/gusplusbus/trustflow/data_server/internal/domain"
)

type ProjectRepo interface {
	Create(ctx context.Context, in *domain.Project) (*domain.Project, error)
	Get(ctx context.Context, userID, id string) (*domain.Project, error)
	Update(ctx context.Context, in *domain.Project) (*domain.Project, error)
	Delete(ctx context.Context, userID, id string) (bool, error)
}
