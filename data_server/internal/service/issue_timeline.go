package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	pb "github.com/gusplusbus/trustflow/data_server/gen/issuetimelinev1"
	"github.com/gusplusbus/trustflow/data_server/internal/repo/postgres"
	"github.com/gusplusbus/trustflow/data_server/internal/service/crypto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type IssuesTimelineService struct {
	repo    *postgres.IssuesTimelinePG
	buckets *postgres.BucketRepo    // nil => bucket mode off
	pool    *pgxpool.Pool           // for tx when bucket mode on
}

func NewIssuesTimelineService(repo *postgres.IssuesTimelinePG) *IssuesTimelineService {
	return &IssuesTimelineService{repo: repo}
}

func NewIssuesTimelineServiceWithBuckets(repo *postgres.IssuesTimelinePG, buckets *postgres.BucketRepo, pool *pgxpool.Pool) *IssuesTimelineService {
	return &IssuesTimelineService{repo: repo, buckets: buckets, pool: pool}
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

	// Build raw items for the legacy table
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

	// 1) Legacy write (kept for compatibility)
	inserted, err := s.repo.InsertMany(ctx, items)
	if err != nil {
		return nil, err
	}

	// 2) Bucketized write (if enabled)
	if s.buckets != nil && len(items) > 0 {
		if err := s.appendToBuckets(ctx, pid, items); err != nil {
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

// appendToBuckets does:
//  - hash each item canonically (DAG-CBOR)
//  - derive bucket_key = created_at (UTC) YYYY-MM-DD
//  - upsert bucket roots (batched)
//  - insert timeline_items (dedup by provider_event_id) and timeline_bucket_leaves
// Everything is wrapped in a single SQL tx.
func (s *IssuesTimelineService) appendToBuckets(ctx context.Context, projectIssueID int64, items []postgres.RawItem) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// We’ll group per bucket_key (daily)
	type leaf struct {
		EntityKind string
		EntityKey  string
		BucketKey  string
		LeafHash   []byte
		// the repo likely uses an index internally, so we only pass hashes
	}

	// You already know the entity_kind/entity_key at ingestion time.
	// If your IssuesTimelinePG has helpers to fetch them by projectIssueID, call them here.
	entityKind, entityKey, err := s.repo.ResolveEntity(ctx, tx, projectIssueID)
	if err != nil {
		return err
	}

	leavesByBucket := map[string][]leaf{}
	for _, it := range items {
		// Canonicalize for hashing
		payloadMap := map[string]any{}
		_ = json.Unmarshal(it.PayloadJSON, &payloadMap) // ignore error; empty map on failure

		canon := crypto.CanonItem{
			Provider:        it.Provider,
			ProviderEventID: it.ProviderEventID,
			Type:            it.Type,
			Actor:           it.Actor,
			CreatedAt:       it.CreatedAt.UTC(),
			Payload:         payloadMap,
		}
		_, h, err := crypto.HashDAGCBOR(canon)
		if err != nil {
			return err
		}

		bkey := it.CreatedAt.UTC().Format("2006-01-02")

		// Insert timeline_items (dedup by provider_event_id)
		if err := s.buckets.InsertTimelineItem(ctx, tx, postgres.InsertTimelineItemParams{
			EntityKind:       entityKind,
			EntityKey:        entityKey,
			Provider:         it.Provider,
			ProviderEventID:  it.ProviderEventID,
			Type:             it.Type,
			Actor:            it.Actor,
			CreatedAt:        it.CreatedAt.UTC(),
			PayloadJSON:      it.PayloadJSON,
			ItemHash:         h,
			BucketKey:        bkey,
			ProjectIssueID:   projectIssueID, // if your schema uses it for seq; otherwise drop
		}); err != nil {
			return err
		}

		leavesByBucket[bkey] = append(leavesByBucket[bkey], leaf{
			EntityKind: entityKind,
			EntityKey:  entityKey,
			BucketKey:  bkey,
			LeafHash:   h,
		})
	}

	// Upsert bucket roots per bucket (batched)
	for bkey, ls := range leavesByBucket {
		hashes := make([][]byte, 0, len(ls))
		for _, l := range ls {
			hashes = append(hashes, l.LeafHash)
		}
		// Use your repo’s rolling-root helper (tl_upsert_bucket_batch.sql)
		if err := s.buckets.UpsertBucketBatch(ctx, tx, postgres.UpsertBucketBatchParams{
			EntityKind: entityKind,
			EntityKey:  entityKey,
			BucketKey:  bkey,
			LeafHashes: hashes,
		}); err != nil {
			return err
		}

		// Insert leaves with indexes for inclusion proofs later
		if err := s.buckets.InsertLeaves(ctx, tx, postgres.InsertLeavesParams{
			EntityKind: entityKind,
			EntityKey:  entityKey,
			BucketKey:  bkey,
			LeafHashes: hashes,
		}); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
