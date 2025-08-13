package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/gusplusbus/trustflow/data_server/internal/domain"
	"github.com/gusplusbus/trustflow/data_server/internal/repo"
)

type OwnershipService struct {
	r repo.OwnershipRepo
}

func NewOwnershipService(r repo.OwnershipRepo) *OwnershipService {
	return &OwnershipService{r: r}
}

func (s *OwnershipService) Create(ctx context.Context, in *domain.Ownership) (*domain.Ownership, error) {
	if err := in.ValidateForCreate(); err != nil {
		return nil, err
	}
	in.Organization = strings.TrimSpace(in.Organization)
	in.Repository = strings.TrimSpace(in.Repository)
	if in.Organization == "" || in.Repository == "" {
		return nil, fmt.Errorf("organization/repository required")
	}
	return s.r.Create(ctx, in)
}

func (s *OwnershipService) Update(ctx context.Context, in *domain.Ownership) (*domain.Ownership, error) {
	if in.ID == "" || in.UserID == "" {
		return nil, fmt.Errorf("missing identifiers")
	}
	in.Organization = strings.TrimSpace(in.Organization)
	in.Repository = strings.TrimSpace(in.Repository)
	if in.Organization == "" || in.Repository == "" {
		return nil, fmt.Errorf("organization/repository required")
	}
	return s.r.Update(ctx, in)
}

func (s *OwnershipService) Delete(ctx context.Context, userID, id string) (bool, error) {
	if userID == "" || id == "" {
		return false, fmt.Errorf("missing identifiers")
	}
	return s.r.Delete(ctx, userID, id)
}

func (s *OwnershipService) ListByProject(ctx context.Context, userID, projectID string) ([]*domain.Ownership, error) {
	if userID == "" || projectID == "" {
		return nil, fmt.Errorf("missing identifiers")
	}
	return s.r.ListByProject(ctx, userID, projectID)
}
