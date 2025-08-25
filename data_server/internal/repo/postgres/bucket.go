package postgres

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ... your existing types (BucketRow, LeafRow, ItemLoc, BucketRepo, etc.)

//go:embed queries/*.sql
var bucketFS embed.FS

// LoadEmbeddedQueries reads the bucket SQL files into a map for NewBucketRepo.
func LoadEmbeddedQueries() map[string]string {
	names := []string{
		"tl_insert_item.sql",
		"tl_insert_leaf.sql",
		"tl_select_leaves_for_bucket.sql",
		"tl_upsert_bucket_per_leaf.sql",
		"tl_upsert_bucket_batch.sql",
		"tl_get_bucket.sql",
		"tl_list_buckets_by_scope.sql",
		"tl_list_buckets_by_status.sql",
		"tl_mark_bucket_closed.sql",
		"tl_set_bucket_anchored.sql",
		"tl_get_item_by_provider_event.sql",
	}
	m := make(map[string]string, len(names))
	for _, n := range names {
		b, err := bucketFS.ReadFile("queries/" + n)
		if err != nil {
			panic(fmt.Errorf("read %s: %w", n, err))
		}
		m[n] = string(b)
	}
	return m
}

type BucketRow struct {
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

type LeafRow struct {
	LeafIndex int32
	LeafHash  []byte
}

type ItemLoc struct {
	EntityKind string
	EntityKey  string
	BucketKey  string
	Seq        int64
	ItemHash   []byte
}

type BucketRepo struct {
	db *pgxpool.Pool
	q  map[string]string // load embedded SQL like your other repos
}

func NewBucketRepo(db *pgxpool.Pool, queries map[string]string) *BucketRepo {
	return &BucketRepo{db: db, q: queries}
}

// Items (idempotent insert)
func (r *BucketRepo) InsertItem(ctx context.Context, tx pgx.Tx,
	entityKind, entityKey, provider, providerEventID, typ string, actor *string,
	createdAt time.Time, payloadJSON []byte, itemHash []byte, bucketKey string,
) (bool, error) {
	var a any
	if actor == nil { a = nil } else { a = *actor }
	ct, err := tx.Exec(ctx, r.q["tl_insert_item.sql"],
		entityKind, entityKey, provider, providerEventID, typ, a,
		createdAt, payloadJSON, itemHash, bucketKey)
	return ct.RowsAffected() == 1, err
}

// Leaves
func (r *BucketRepo) InsertLeaf(ctx context.Context, tx pgx.Tx,
	entityKind, entityKey, bucketKey string, leafIndex int32, leafHash []byte,
) error {
	_, err := tx.Exec(ctx, r.q["tl_insert_leaf.sql"], entityKind, entityKey, bucketKey, leafIndex, leafHash)
	return err
}

func (r *BucketRepo) SelectLeaves(ctx context.Context,
	entityKind, entityKey, bucketKey string,
) ([]LeafRow, error) {
	rows, err := r.db.Query(ctx, r.q["tl_select_leaves_for_bucket.sql"], entityKind, entityKey, bucketKey)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []LeafRow
	for rows.Next() {
		var lr LeafRow
		if err := rows.Scan(&lr.LeafIndex, &lr.LeafHash); err != nil { return nil, err }
		out = append(out, lr)
	}
	return out, rows.Err()
}

// Buckets
func (r *BucketRepo) UpsertPerLeaf(ctx context.Context, tx pgx.Tx,
	entityKind, entityKey, bucketKey string, newRoot []byte,
) error {
	_, err := tx.Exec(ctx, r.q["tl_upsert_bucket_per_leaf.sql"], entityKind, entityKey, bucketKey, newRoot)
	return err
}

func (r *BucketRepo) UpsertBatch(ctx context.Context, tx pgx.Tx,
	entityKind, entityKey, bucketKey string, newRoot []byte, appended int32,
) error {
	_, err := tx.Exec(ctx, r.q["tl_upsert_bucket_batch.sql"], entityKind, entityKey, bucketKey, newRoot, appended)
	return err
}

func (r *BucketRepo) GetBucket(ctx context.Context,
	entityKind, entityKey, bucketKey string,
) (BucketRow, error) {
	var b BucketRow
	err := r.db.QueryRow(ctx, r.q["tl_get_bucket.sql"], entityKind, entityKey, bucketKey).
		Scan(&b.EntityKind, &b.EntityKey, &b.BucketKey, &b.RootHash, &b.LeafCount, &b.Status, &b.CID, &b.ClosedAt, &b.AnchoredTx, &b.AnchoredAt)
	return b, err
}

func (r *BucketRepo) ListByScope(ctx context.Context,
	entityKind, entityKey string, limit, offset int32,
) ([]BucketRow, error) {
	rows, err := r.db.Query(ctx, r.q["tl_list_buckets_by_scope.sql"], entityKind, entityKey, limit, offset)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []BucketRow
	for rows.Next() {
		var b BucketRow
		if err := rows.Scan(&b.EntityKind, &b.EntityKey, &b.BucketKey, &b.RootHash, &b.LeafCount, &b.Status, &b.CID, &b.ClosedAt, &b.AnchoredTx, &b.AnchoredAt); err != nil { return nil, err }
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BucketRepo) ListByStatus(ctx context.Context,
	status string, limit, offset int32,
) ([]BucketRow, error) {
	rows, err := r.db.Query(ctx, r.q["tl_list_buckets_by_status.sql"], status, limit, offset)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []BucketRow
	for rows.Next() {
		var b BucketRow
		if err := rows.Scan(&b.EntityKind, &b.EntityKey, &b.BucketKey, &b.RootHash, &b.LeafCount, &b.Status, &b.CID, &b.ClosedAt, &b.AnchoredTx, &b.AnchoredAt); err != nil { return nil, err }
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BucketRepo) MarkClosed(ctx context.Context,
	entityKind, entityKey, bucketKey string,
) (BucketRow, error) {
	var b BucketRow
	err := r.db.QueryRow(ctx, r.q["tl_mark_bucket_closed.sql"], entityKind, entityKey, bucketKey).
		Scan(&b.EntityKind, &b.EntityKey, &b.BucketKey, &b.RootHash, &b.LeafCount, &b.Status, &b.CID, &b.ClosedAt, &b.AnchoredTx, &b.AnchoredAt)
	return b, err
}

func (r *BucketRepo) SetAnchored(ctx context.Context,
	entityKind, entityKey, bucketKey, cid, anchoredTx string,
) (BucketRow, error) {
	var b BucketRow
	err := r.db.QueryRow(ctx, r.q["tl_set_bucket_anchored.sql"], entityKind, entityKey, bucketKey, cid, anchoredTx).
		Scan(&b.EntityKind, &b.EntityKey, &b.BucketKey, &b.RootHash, &b.LeafCount, &b.Status, &b.CID, &b.ClosedAt, &b.AnchoredTx, &b.AnchoredAt)
	return b, err
}

func (r *BucketRepo) GetItemForProof(ctx context.Context, providerEventID string) (ItemLoc, error) {
	var v ItemLoc
	err := r.db.QueryRow(ctx, r.q["tl_get_item_by_provider_event.sql"], providerEventID).
		Scan(&v.EntityKind, &v.EntityKey, &v.BucketKey, &v.Seq, &v.ItemHash)
	return v, err
}

