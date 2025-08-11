package grpcserver

import (
	"context"
	"time"

	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
	"github.com/gusplusbus/trustflow/data_server/internal/domain"
	"github.com/gusplusbus/trustflow/data_server/internal/service"
)

type ProjectServer struct {
	projectv1.UnimplementedProjectServiceServer
	svc *service.ProjectService
}

func NewProjectServer(svc *service.ProjectService) *ProjectServer { return &ProjectServer{svc: svc} }

func toProto(p *domain.Project) *projectv1.Project {
	return &projectv1.Project{
		Id:                   p.ID,
		CreatedAt:            p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:            p.UpdatedAt.UTC().Format(time.RFC3339),
		Title:                p.Title,
		Description:          p.Description,
		DurationEstimate:     p.DurationEstimate,
		TeamSize:             p.TeamSize,
		ApplicationCloseTime: p.ApplicationCloseTime,
	}
}

/* Health */
func (s *ProjectServer) Health(ctx context.Context, _ *projectv1.HealthRequest) (*projectv1.HealthResponse, error) {
	return &projectv1.HealthResponse{Status: "ok"}, nil
}

/* Create */
func (s *ProjectServer) CreateProject(ctx context.Context, req *projectv1.CreateProjectRequest) (*projectv1.CreateProjectResponse, error) {
	in := &domain.Project{
		UserID:               req.GetUserId(),
		Title:                req.GetTitle(),
		Description:          req.GetDescription(),
		DurationEstimate:     req.GetDurationEstimate(),
		TeamSize:             req.GetTeamSize(),
		ApplicationCloseTime: req.GetApplicationCloseTime(),
	}
	p, err := s.svc.Create(ctx, in)
	if err != nil { return nil, err }
	return &projectv1.CreateProjectResponse{Project: toProto(p)}, nil
}

/* Get */
func (s *ProjectServer) GetProject(ctx context.Context, req *projectv1.GetProjectRequest) (*projectv1.GetProjectResponse, error) {
	p, err := s.svc.Get(ctx, req.GetUserId(), req.GetId())
	if err != nil { return nil, err }
	return &projectv1.GetProjectResponse{Project: toProto(p)}, nil
}

/* Update */
func (s *ProjectServer) UpdateProject(ctx context.Context, req *projectv1.UpdateProjectRequest) (*projectv1.UpdateProjectResponse, error) {
	in := &domain.Project{
		ID:                   req.GetId(),
		UserID:               req.GetUserId(),
		Title:                req.GetTitle(),
		Description:          req.GetDescription(),
		DurationEstimate:     req.GetDurationEstimate(),
		TeamSize:             req.GetTeamSize(),
		ApplicationCloseTime: req.GetApplicationCloseTime(),
	}
	p, err := s.svc.Update(ctx, in)
	if err != nil { return nil, err }
	return &projectv1.UpdateProjectResponse{Project: toProto(p)}, nil
}

/* Delete */
func (s *ProjectServer) DeleteProject(ctx context.Context, req *projectv1.DeleteProjectRequest) (*projectv1.DeleteProjectResponse, error) {
	ok, err := s.svc.Delete(ctx, req.GetUserId(), req.GetId())
	if err != nil { return nil, err }
	return &projectv1.DeleteProjectResponse{Deleted: ok}, nil
}

/* List */
func (s *ProjectServer) ListProjects(ctx context.Context, req *projectv1.ListProjectsRequest) (*projectv1.ListProjectsResponse, error) {
  res, err := s.svc.ListProjects(ctx, service.ListParams{
    UserID:   req.GetUserId(),
    Page:     int(req.GetPage()),
    PageSize: int(req.GetPageSize()),
    SortBy:   req.GetSortBy(),
    SortDir:  req.GetSortDir(),
    Q:        req.GetQ(),
  })
  if err != nil { return nil, err }
  out := make([]*projectv1.Project, 0, len(res.Projects))
  for _, p := range res.Projects { out = append(out, toProto(p)) }
  return &projectv1.ListProjectsResponse{
    Projects: out,
    Total:    res.Total,
  }, nil
}
