package grpcserver

import (
	"context"
	"time"

	ownershipv1 "github.com/gusplusbus/trustflow/data_server/gen/ownershipv1"
	"github.com/gusplusbus/trustflow/data_server/internal/domain"
	"github.com/gusplusbus/trustflow/data_server/internal/service"
)

type OwnershipServer struct {
	ownershipv1.UnimplementedOwnershipServiceServer
	svc *service.OwnershipService
}

func NewOwnershipServer(svc *service.OwnershipService) *OwnershipServer {
	return &OwnershipServer{svc: svc}
}

func toOwnershipProto(o *domain.Ownership) *ownershipv1.Ownership {
	return &ownershipv1.Ownership{
		Id:          o.ID,
		CreatedAt:   o.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   o.UpdatedAt.UTC().Format(time.RFC3339),
		ProjectId:   o.ProjectID,
		UserId:      o.UserID,
		Organization:o.Organization,
		Repository:  o.Repository,
		Provider:    o.Provider,
		WebUrl:      o.WebURL,
	}
}

func (s *OwnershipServer) Health(ctx context.Context, _ *ownershipv1.HealthRequest) (*ownershipv1.HealthResponse, error) {
	return &ownershipv1.HealthResponse{Status: "ok"}, nil
}

func (s *OwnershipServer) CreateOwnership(ctx context.Context, req *ownershipv1.CreateOwnershipRequest) (*ownershipv1.CreateOwnershipResponse, error) {
	in := &domain.Ownership{
		UserID:       req.GetUserId(),
		ProjectID:    req.GetProjectId(),
		Organization: req.GetOrganization(),
		Repository:   req.GetRepository(),
		Provider:     req.GetProvider(),
		WebURL:       req.GetWebUrl(),
	}
	o, err := s.svc.Create(ctx, in)
	if err != nil { return nil, err }
	return &ownershipv1.CreateOwnershipResponse{Ownership: toOwnershipProto(o)}, nil
}

func (s *OwnershipServer) UpdateOwnership(ctx context.Context, req *ownershipv1.UpdateOwnershipRequest) (*ownershipv1.UpdateOwnershipResponse, error) {
	in := &domain.Ownership{
		ID:           req.GetId(),
		UserID:       req.GetUserId(),
		Organization: req.GetOrganization(),
		Repository:   req.GetRepository(),
		Provider:     req.GetProvider(),
		WebURL:       req.GetWebUrl(),
	}
	o, err := s.svc.Update(ctx, in)
	if err != nil { return nil, err }
	return &ownershipv1.UpdateOwnershipResponse{Ownership: toOwnershipProto(o)}, nil
}

func (s *OwnershipServer) DeleteOwnership(ctx context.Context, req *ownershipv1.DeleteOwnershipRequest) (*ownershipv1.DeleteOwnershipResponse, error) {
	ok, err := s.svc.Delete(ctx, req.GetUserId(), req.GetId())
	if err != nil { return nil, err }
	return &ownershipv1.DeleteOwnershipResponse{Deleted: ok}, nil
}
