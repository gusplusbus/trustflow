package service

import (
	"bytes"
	"context"
	"errors"
	"time"

	bucketv1 "github.com/gusplusbus/trustflow/data_server/gen/bucketv1"
	"github.com/gusplusbus/trustflow/data_server/internal/repo/postgres"
	"github.com/gusplusbus/trustflow/data_server/internal/service/crypto"
)

// ---- DTO ----

type BucketDTO struct {
	EntityKind string
	EntityKey  string
	BucketKey  string
	RootHash   []byte
	LeafCount  int32
	Status     string
	CID        *string
	ClosedAt   *time.Time
	AnchoredTx *string
	AnchoredAt *time.Time
}

func (b BucketDTO) ToProto() *bucketv1.BucketInfo {
	var cid, tx, closed, anchored string
	if b.CID != nil {
		cid = *b.CID
	}
	if b.AnchoredTx != nil {
		tx = *b.AnchoredTx
	}
	if b.ClosedAt != nil {
		closed = b.ClosedAt.UTC().Format(time.RFC3339)
	}
	if b.AnchoredAt != nil {
		anchored = b.AnchoredAt.UTC().Format(time.RFC3339)
	}

	return &bucketv1.BucketInfo{
		Ref: &bucketv1.BucketRef{
			Scope:     &bucketv1.Scope{EntityKind: b.EntityKind, EntityKey: b.EntityKey},
			BucketKey: b.BucketKey,
		},
		RootHash:   b.RootHash,
		LeafCount:  uint32(b.LeafCount),
		Status:     b.Status,
		Cid:        cid,
		AnchoredTx: tx,
		AnchoredAt: anchored,
		ClosedAt:   closed,
	}
}

// ---- Service ----

type BucketService struct {
	repo *postgres.BucketRepo
}

func NewBucketService(r *postgres.BucketRepo) *BucketService { return &BucketService{repo: r} }

func (s *BucketService) ListBuckets(ctx context.Context, scope bucketv1.Scope, limit, offset int32) ([]BucketDTO, error) {
	rows, err := s.repo.ListByScope(ctx, scope.GetEntityKind(), scope.GetEntityKey(), limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]BucketDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, rowToDTO(r))
	}
	return out, nil
}

func (s *BucketService) GetBucket(ctx context.Context, ref bucketv1.BucketRef) (BucketDTO, error) {
	r, err := s.repo.GetBucket(ctx, ref.GetScope().GetEntityKind(), ref.GetScope().GetEntityKey(), ref.GetBucketKey())
	if err != nil {
		return BucketDTO{}, err
	}
	return rowToDTO(r), nil
}

func (s *BucketService) MarkBucketClosed(ctx context.Context, ref bucketv1.BucketRef) (BucketDTO, error) {
	r, err := s.repo.MarkClosed(ctx, ref.GetScope().GetEntityKind(), ref.GetScope().GetEntityKey(), ref.GetBucketKey())
	if err != nil {
		return BucketDTO{}, err
	}
	return rowToDTO(r), nil
}

func (s *BucketService) SetBucketAnchored(ctx context.Context, ref bucketv1.BucketRef, cid, anchoredTx string) (BucketDTO, error) {
	r, err := s.repo.SetAnchored(ctx, ref.GetScope().GetEntityKind(), ref.GetScope().GetEntityKey(), ref.GetBucketKey(), cid, anchoredTx)
	if err != nil {
		return BucketDTO{}, err
	}
	return rowToDTO(r), nil
}

func (s *BucketService) InclusionProof(ctx context.Context, ref bucketv1.BucketRef, providerEventID string) (*bucketv1.InclusionProofResponse, error) {
	// Locate the item (entity/bucket and its hash)
	loc, err := s.repo.GetItemForProof(ctx, providerEventID)
	if err != nil {
		return nil, err
	}
	// Ensure the requested ref matches the item’s actual bucket
	if loc.EntityKind != ref.GetScope().GetEntityKind() ||
		loc.EntityKey != ref.GetScope().GetEntityKey() ||
		loc.BucketKey != ref.GetBucketKey() {
		return nil, errors.New("item not in requested bucket")
	}

	// Fetch leaves in order, find index by matching hash (robust vs 0/1-based seq)
	lrows, err := s.repo.SelectLeaves(ctx, loc.EntityKind, loc.EntityKey, loc.BucketKey)
	if err != nil {
		return nil, err
	}
	if len(lrows) == 0 {
		return nil, errors.New("no leaves in bucket")
	}
	leaves := make([][]byte, len(lrows))
	idx := -1
	for i, v := range lrows {
		leaves[i] = v.LeafHash
		if bytes.Equal(v.LeafHash, loc.ItemHash) {
			idx = i
		}
	}
	if idx < 0 {
		return nil, errors.New("leaf for item not found in bucket")
	}

	leaf, path, root := crypto.BuildProof(leaves, idx)
	if leaf == nil {
		return nil, errors.New("failed to build proof")
	}

	steps := make([]*bucketv1.InclusionProofResponse_Step, 0, len(path))
	for _, st := range path {
		steps = append(steps, &bucketv1.InclusionProofResponse_Step{
			Sibling:       st.Sibling,
			SiblingIsLeft: st.SiblingIsLeft,
		})
	}
	return &bucketv1.InclusionProofResponse{
		LeafHash: leaf,
		Path:     steps,
		RootHash: root,
	}, nil
}

func (s *BucketService) ListByStatus(ctx context.Context, status string, limit int32, page string) (*bucketv1.ListBucketsByStatusResponse, error) {
    rows, next, err := s.repo.ListBucketsByStatus(ctx, status, int(limit), page)
    if err != nil {
        return nil, err
    }
    out := &bucketv1.ListBucketsByStatusResponse{NextPageToken: next}
    for _, r := range rows {
        bi := &bucketv1.BucketInfo{
            Ref: &bucketv1.BucketRef{
                Scope: &bucketv1.Scope{EntityKind: r.EntityKind, EntityKey: r.EntityKey},
                BucketKey: r.BucketKey,
            },
            RootHash:  r.RootHash,
            LeafCount: uint32(r.LeafCount),
            Status:    r.Status,
            Cid:       r.CID,
            AnchoredTx: r.AnchoredTX,
        }
        if !r.ClosedAt.IsZero() {
            bi.ClosedAt = r.ClosedAt.UTC().Format(time.RFC3339)
        }
        if !r.AnchoredAt.IsZero() {
            bi.AnchoredAt = r.AnchoredAt.UTC().Format(time.RFC3339)
        }
        out.Buckets = append(out.Buckets, bi)
    }
    return out, nil
}

// ---- helpers ----

func rowToDTO(r postgres.BucketRow) BucketDTO {
	return BucketDTO{
		EntityKind: r.EntityKind,
		EntityKey:  r.EntityKey,
		BucketKey:  r.BucketKey,
		RootHash:   r.RootHash,
		LeafCount:  r.LeafCount,
		Status:     r.Status,
		CID:        r.CID,
		ClosedAt:   r.ClosedAt,
		AnchoredTx: r.AnchoredTx,
		AnchoredAt: r.AnchoredAt,
	}
}
