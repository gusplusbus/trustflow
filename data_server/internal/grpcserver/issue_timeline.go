package grpcserver

import (
	"context"

	pb "github.com/gusplusbus/trustflow/data_server/gen/issuetimelinev1"
	"github.com/gusplusbus/trustflow/data_server/internal/service"
)

type IssuesTimelineGRPC struct {
	pb.UnimplementedIssuesTimelineServiceServer
	svc *service.IssuesTimelineService
}

func NewIssuesTimelineGRPC(svc *service.IssuesTimelineService) *IssuesTimelineGRPC {
	return &IssuesTimelineGRPC{svc: svc}
}

func (g *IssuesTimelineGRPC) GetCheckpoint(ctx context.Context, req *pb.GetCheckpointRequest) (*pb.GetCheckpointResponse, error) {
	return g.svc.GetCheckpoint(ctx, req)
}

func (g *IssuesTimelineGRPC) AppendBatch(ctx context.Context, req *pb.AppendBatchRequest) (*pb.AppendBatchResponse, error) {
	return g.svc.AppendBatch(ctx, req)
}
