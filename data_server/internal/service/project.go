package service

import (
	"context"
	"fmt"

	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
	"github.com/gusplusbus/trustflow/data_server/internal/domain"
	"github.com/gusplusbus/trustflow/data_server/internal/repo"
)

type ProjectService struct {
	r             repo.ProjectRepo
	ownershipRepo repo.OwnershipRepo
}

func NewProjectService(r repo.ProjectRepo, o repo.OwnershipRepo) *ProjectService {
	return &ProjectService{r: r, ownershipRepo: o}
}

func (s *ProjectService) Create(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	if err := p.ValidateForCreate(); err != nil {
		return nil, err
	}
	return s.r.Create(ctx, p)
}

func (s *ProjectService) Get(ctx context.Context, userID, id string, includeOwnerships bool) (*domain.Project, error) {
	if userID == "" || id == "" {
		return nil, fmt.Errorf("missing identifiers")
	}

	project, err := s.r.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if includeOwnerships {
		owns, err := s.ownershipRepo.ListByProject(ctx, userID, id)
		if err != nil {
			return nil, err
		}
		project.Ownerships = owns
	}

	return project, nil
}

func (s *ProjectService) Update(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	if p.ID == "" || p.UserID == "" {
		return nil, fmt.Errorf("missing identifiers")
	}
	if err := p.ValidateForUpdate(); err != nil {
		return nil, err
	}
	return s.r.Update(ctx, p)
}

func (s *ProjectService) Delete(ctx context.Context, userID, id string) (bool, error) {
	if userID == "" || id == "" {
		return false, fmt.Errorf("missing identifiers")
	}
	return s.r.Delete(ctx, userID, id)
}

type ListParams struct {
	UserID            string
	Page              int
	PageSize          int
	SortBy            projectv1.SortBy
	SortDir           projectv1.SortDir
	Q                 string
	IncludeOwnerships bool
}

type ListResult struct {
	Projects []*domain.Project
	Total    int64
}

func (s *ProjectService) ListProjects(ctx context.Context, p ListParams) (*ListResult, error) {
	if p.UserID == "" {
		return nil, fmt.Errorf("missing user_id")
	}
	if p.Page < 0 {
		p.Page = 0
	}
	if p.PageSize <= 0 {
		p.PageSize = 25
	}
	if p.PageSize > 200 {
		p.PageSize = 200
	}

	sortCol := "created_at"
	switch p.SortBy {
	case projectv1.SortBy_SORT_BY_UPDATED_AT:
		sortCol = "updated_at"
	case projectv1.SortBy_SORT_BY_TITLE:
		sortCol = "title"
	case projectv1.SortBy_SORT_BY_TEAM_SIZE:
		sortCol = "team_size"
	case projectv1.SortBy_SORT_BY_DURATION:
		sortCol = "duration_estimate"
	}

	dir := "DESC"
	if p.SortDir == projectv1.SortDir_SORT_DIR_ASC {
		dir = "ASC"
	}

	res, err := s.r.List(ctx, repo.RepoListParams{
		UserID:     p.UserID,
		Page:       p.Page,
		PageSize:   p.PageSize,
		SortColumn: sortCol,
		SortDir:    dir,
		Q:          p.Q,
	})
	if err != nil {
		return nil, err
	}

	if p.IncludeOwnerships {
		for _, proj := range res.Projects {
			owns, err := s.ownershipRepo.ListByProject(ctx, p.UserID, proj.ID)
			if err != nil {
				return nil, err
			}
			proj.Ownerships = owns
		}
	}

	return &ListResult{Projects: res.Projects, Total: res.Total}, nil
}
