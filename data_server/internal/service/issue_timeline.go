package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	pb "github.com/gusplusbus/trustflow/data_server/gen/issuetimelinev1"
	"github.com/gusplusbus/trustflow/data_server/internal/repo/postgres"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NOTE: repo is the concrete *IssuesTimelinePG you wired in main.go
type IssuesTimelineService struct {
	repo *postgres.IssuesTimelinePG
}

func NewIssuesTimelineService(repo *postgres.IssuesTimelinePG) *IssuesTimelineService {
	return &IssuesTimelineService{repo: repo}
}

func (s *IssuesTimelineService) GetCheckpoint(ctx context.Context, req *pb.GetCheckpointRequest) (*pb.GetCheckpointResponse, error) {
	if req.GetGhIssueId() == 0 {
		return nil, errors.New("gh_issue_id required")
	}
	pid, err := s.repo.GetProjectIssueIDByGhID(ctx, req.GetGhIssueId())
	if err != nil {
		// Not managed yet -> treat as empty checkpoint
		return &pb.GetCheckpointResponse{Cursor: ""}, nil
	}

	ck, err := s.repo.GetCheckpoint(ctx, pid)
	if err != nil {
		return nil, err
	}

	out := &pb.GetCheckpointResponse{Cursor: ck.Cursor}
	if !ck.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(ck.UpdatedAt)
	}
	return out, nil
}

func (s *IssuesTimelineService) AppendBatch(ctx context.Context, req *pb.AppendBatchRequest) (*pb.AppendBatchResponse, error) {
	if req.GetGhIssueId() == 0 {
		return nil, errors.New("gh_issue_id required")
	}
	pid, err := s.repo.GetProjectIssueIDByGhID(ctx, req.GetGhIssueId())
	if err != nil {
		return nil, err
	}

	items := make([]postgres.RawItem, 0, len(req.GetItems()))
	for _, it := range req.GetItems() {
		payload := it.GetPayloadJson()
		if len(payload) == 0 {
			payload, _ = json.Marshal(map[string]any{})
		}
		var actor *string
		if it.GetActor() != "" {
			a := it.GetActor()
			actor = &a
		}
		items = append(items, postgres.RawItem{
			ProjectIssueID:  pid,
			Provider:        it.GetProvider(),
			ProviderEventID: it.GetProviderEventId(),
			Type:            it.GetType(),
			Actor:           actor,
			CreatedAt:       it.GetCreatedAt().AsTime().UTC(),
			PayloadJSON:     payload,
		})
	}

	inserted, err := s.repo.InsertMany(ctx, items)
	if err != nil {
		return nil, err
	}

	// compute lastAt for checkpoint (even if inserts dedup to 0)
	var lastAt *time.Time
	if len(req.GetItems()) > 0 {
		max := req.GetItems()[0].GetCreatedAt().AsTime()
		for _, it := range req.GetItems() {
			t := it.GetCreatedAt().AsTime()
			if t.After(max) {
				max = t
			}
		}
		lastAt = &max
	}

	if err := s.repo.UpsertCheckpoint(ctx, pid, req.GetEndCursor(), lastAt); err != nil {
		return nil, err
	}

	return &pb.AppendBatchResponse{
		Inserted:     uint32(inserted),
		LatestCursor: req.GetEndCursor(),
	}, nil
}
