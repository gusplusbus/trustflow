package service

import (
	"context"
	"fmt"

	"github.com/gusplusbus/trustflow/data_server/internal/domain"
	"github.com/gusplusbus/trustflow/data_server/internal/repo"
)

type ProjectService struct {
	r repo.ProjectRepo
}

func NewProjectService(r repo.ProjectRepo) *ProjectService { return &ProjectService{r: r} }

// You can add instrumentation here (timers, spans) without touching repo/grpc.

func (s *ProjectService) Create(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	if err := p.ValidateForCreate(); err != nil { return nil, err }
	return s.r.Create(ctx, p)
}

func (s *ProjectService) Get(ctx context.Context, userID, id string) (*domain.Project, error) {
	if userID == "" || id == "" { return nil, fmt.Errorf("missing identifiers") }
	return s.r.Get(ctx, userID, id)
}

func (s *ProjectService) Update(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	if p.ID == "" || p.UserID == "" { return nil, fmt.Errorf("missing identifiers") }
	if err := p.ValidateForUpdate(); err != nil { return nil, err }
	return s.r.Update(ctx, p)
}

func (s *ProjectService) Delete(ctx context.Context, userID, id string) (bool, error) {
	if userID == "" || id == "" { return false, fmt.Errorf("missing identifiers") }
	return s.r.Delete(ctx, userID, id)
}
