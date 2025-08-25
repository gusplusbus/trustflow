package grpcserver

import (
	"context"

	bucketv1 "github.com/gusplusbus/trustflow/data_server/gen/bucketv1"
	"github.com/gusplusbus/trustflow/data_server/internal/service"
)

type BucketServer struct {
	bucketv1.UnimplementedBucketServiceServer
	svc *service.BucketService
}

func NewBucketServer(svc *service.BucketService) *BucketServer {
	return &BucketServer{svc: svc}
}

func (s *BucketServer) ListBuckets(ctx context.Context, req *bucketv1.ListBucketsRequest) (*bucketv1.ListBucketsResponse, error) {
	if req.GetLimit() <= 0 {
		req.Limit = 50
	}
	rows, err := s.svc.ListBuckets(ctx, *req.GetScope(), req.GetLimit(), 0)
	if err != nil {
		return nil, err
	}
	out := make([]*bucketv1.BucketInfo, 0, len(rows))
	for _, b := range rows {
		out = append(out, b.ToProto())
	}
	return &bucketv1.ListBucketsResponse{Buckets: out}, nil
}

func (s *BucketServer) GetBucket(ctx context.Context, req *bucketv1.GetBucketRequest) (*bucketv1.GetBucketResponse, error) {
	b, err := s.svc.GetBucket(ctx, *req.GetRef())
	if err != nil {
		return nil, err
	}
	return &bucketv1.GetBucketResponse{Bucket: b.ToProto()}, nil
}

func (s *BucketServer) InclusionProof(ctx context.Context, req *bucketv1.InclusionProofRequest) (*bucketv1.InclusionProofResponse, error) {
	return s.svc.InclusionProof(ctx, *req.GetRef(), req.GetProviderEventId())
}

func (s *BucketServer) MarkBucketClosed(ctx context.Context, req *bucketv1.MarkBucketClosedRequest) (*bucketv1.MarkBucketClosedResponse, error) {
	b, err := s.svc.MarkBucketClosed(ctx, *req.GetRef())
	if err != nil {
		return nil, err
	}
	return &bucketv1.MarkBucketClosedResponse{Bucket: b.ToProto()}, nil
}

func (s *BucketServer) SetBucketAnchored(ctx context.Context, req *bucketv1.SetBucketAnchoredRequest) (*bucketv1.SetBucketAnchoredResponse, error) {
	b, err := s.svc.SetBucketAnchored(ctx, *req.GetRef(), req.GetCid(), req.GetAnchoredTx())
	if err != nil {
		return nil, err
	}
	return &bucketv1.SetBucketAnchoredResponse{Bucket: b.ToProto()}, nil
}

func (s *BucketServer) ListBucketsByStatus(ctx context.Context, req *bucketv1.ListBucketsByStatusRequest) (*bucketv1.ListBucketsByStatusResponse, error) {
    return s.svc.ListByStatus(ctx, req.GetStatus(), req.GetLimit(), req.GetPageToken())
}
