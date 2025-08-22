package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	pb "github.com/gusplusbus/trustflow/data_server/gen/issuetimelinev1"
	"github.com/gusplusbus/trustflow/data_server/internal/repo/postgres"
	"github.com/gusplusbus/trustflow/data_server/internal/service/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service holds the legacy timeline repo plus optional bucket repo.
type IssuesTimelineService struct {
	repo       *postgres.IssuesTimelinePG
	bucketRepo *postgres.BucketRepo // nil => bucket writes disabled
	pool       *pgxpool.Pool        // tx handle when bucket writes enabled
}

func NewIssuesTimelineService(repo *postgres.IssuesTimelinePG) *IssuesTimelineService {
	return &IssuesTimelineService{repo: repo}
}

func NewIssuesTimelineServiceWithBuckets(repo *postgres.IssuesTimelinePG, bucket *postgres.BucketRepo, pool *pgxpool.Pool) *IssuesTimelineService {
	return &IssuesTimelineService{repo: repo, bucketRepo: bucket, pool: pool}
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

	// Build raw items for the legacy table (kept for compatibility)
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

	// 1) Legacy write
	inserted, err := s.repo.InsertMany(ctx, items)
	if err != nil {
		return nil, err
	}

	// 2) Bucketized write (if enabled)
	if s.bucketRepo != nil && len(items) > 0 {
		if err := s.appendToBuckets(ctx, req.GetGhIssueId(), items); err != nil {
			return nil, err
		}
	}

	// 3) Checkpoint
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

// appendToBuckets:
//  - canonicalize + hash each item (DAG-CBOR -> SHA-256)
//  - group by daily bucket_key (UTC "YYYY-MM-DD")
//  - Insert timeline_items (idempotent by provider_event_id)
//  - Insert new leaves with proper leaf_index
//  - Recompute Merkle root and upsert bucket row
//  - Auto-close buckets from days before today
func (s *IssuesTimelineService) appendToBuckets(ctx context.Context, ghIssueID int64, items []postgres.RawItem) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	entityKind := "issue"
	entityKey := fmt.Sprintf("gh#%d", ghIssueID)

	// Accumulate new leaf hashes by bucket_key
	type bucketAcc struct{ leaves [][]byte }
	acc := map[string]*bucketAcc{}

	for _, it := range items {
		// Canonicalize payload map for hashing
		var pm map[string]any
		if len(it.PayloadJSON) > 0 {
			_ = json.Unmarshal(it.PayloadJSON, &pm)
		}
		canon := crypto.CanonItem{
			Provider:        it.Provider,
			ProviderEventID: it.ProviderEventID,
			Type:            it.Type,
			Actor:           it.Actor,
			CreatedAt:       it.CreatedAt.UTC(),
			Payload:         pm,
		}
		_, itemHash, err := crypto.HashDAGCBOR(canon)
		if err != nil {
			return err
		}

		bKey := canon.CreatedAt.Format("2006-01-02")

		// Insert canonical item row (idempotent on provider_event_id)
		ok, err := s.bucketRepo.InsertItem(ctx, tx,
			entityKind, entityKey, it.Provider, it.ProviderEventID, it.Type, it.Actor,
			canon.CreatedAt, it.PayloadJSON, itemHash, bKey)
		if err != nil {
			return err
		}
		if !ok {
			// duplicate event; do not add another leaf
			continue
		}
		if acc[bKey] == nil {
			acc[bKey] = &bucketAcc{}
		}
		acc[bKey].leaves = append(acc[bKey].leaves, itemHash)
	}

	// Per bucket: rebuild root and write leaves + upsert bucket row
	for bKey, a := range acc {
		// Load existing leaves to compute base index and new root
		prevLeaves, err := s.bucketRepo.SelectLeaves(ctx, entityKind, entityKey, bKey)
		if err != nil && err.Error() != "no rows in result set" {
			return err
		}
		all := make([][]byte, 0, len(prevLeaves)+len(a.leaves))
		for _, v := range prevLeaves {
			all = append(all, v.LeafHash)
		}
		all = append(all, a.leaves...)
		root := crypto.BuildMerkleRoot(all)

		// Upsert bucket (root & leaf_count increment)
		if err := s.bucketRepo.UpsertBatch(ctx, tx, entityKind, entityKey, bKey, root, int32(len(a.leaves))); err != nil {
			return err
		}

		// Insert new leaves with proper leaf_index
		base := int32(len(prevLeaves))
		for i, leaf := range a.leaves {
			if err := s.bucketRepo.InsertLeaf(ctx, tx, entityKind, entityKey, bKey, base+int32(i), leaf); err != nil {
				return err
			}
		}

		// Auto-close any bucket before today so runners can anchor
		if isBeforeTodayUTC(bKey) {
			if _, err := s.bucketRepo.MarkClosed(ctx, entityKind, entityKey, bKey); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func isBeforeTodayUTC(bucketKey string) bool {
	today := time.Now().UTC().Format("2006-01-02")
	return bucketKey < today
}
