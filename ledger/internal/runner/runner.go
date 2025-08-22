package runner

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gusplusbus/trustflow/ledger/internal/dataserver"
	bucketv1 "github.com/gusplusbus/trustflow/data_server/gen/bucketv1"
)

type Config struct {
	DataServerGRPCAddr string
	Interval           time.Duration
	ListPageSize       int32
}

type Runner struct {
	cfg     Config
	buckets dataserver.Buckets
}

func New(cfg Config, buckets dataserver.Buckets) *Runner {
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	if cfg.ListPageSize <= 0 {
		cfg.ListPageSize = 50
	}
	return &Runner{cfg: cfg, buckets: buckets}
}

func (r *Runner) Start(ctx context.Context) {
	t := time.NewTicker(r.cfg.Interval)
	defer t.Stop()

	// Run immediately once
	r.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.tick(ctx)
		}
	}
}

func (r *Runner) tick(ctx context.Context) {
	// 0) Auto-close yesterday-and-older open buckets so they can be anchored
	if err := r.closeStaleOpenBuckets(ctx); err != nil {
		log.Printf("[runner] auto-close error: %v", err)
	}

	// 1) Anchor buckets that are ready
	statuses := []string{"closed", "needs_anchoring"}
	for _, st := range statuses {
		if err := r.handleStatus(ctx, st); err != nil {
			log.Printf("[runner] status=%s error: %v", st, err)
		}
	}
}

func (r *Runner) handleStatus(ctx context.Context, status string) error {
	page := ""
	total := 0
	for {
		resp, err := r.buckets.ListByStatus(ctx, status, r.cfg.ListPageSize, page)
		if err != nil {
			return err
		}
		for _, b := range resp.GetBuckets() {
			if err := r.anchorOne(ctx, b); err != nil {
				log.Printf("[runner] anchor ref=%s/%s/%s: %v",
					b.GetRef().GetScope().GetEntityKind(),
					b.GetRef().GetScope().GetEntityKey(),
					b.GetRef().GetBucketKey(), err)
				continue
			}
			total++
		}
		page = resp.GetNextPageToken()
		if page == "" {
			break
		}
	}
	if total > 0 {
		log.Printf("[runner] anchored %d buckets for status=%s", total, status)
	}
	return nil
}

func (r *Runner) anchorOne(ctx context.Context, b *bucketv1.BucketInfo) error {
	// Minimal manifest (anchor the verifiable root + metadata).
	manifest := map[string]any{
		"entity_kind": b.GetRef().GetScope().GetEntityKind(),
		"entity_key":  b.GetRef().GetScope().GetEntityKey(),
		"bucket_key":  b.GetRef().GetBucketKey(),
		"root_hash":   b.GetRootHash(),
		"leaf_count":  b.GetLeafCount(),
		"closed_at":   b.GetClosedAt(), // RFC3339 string in your proto
		"status":      b.GetStatus(),
	}
	raw, _ := json.Marshal(manifest)

	// TODO(real): PUT to IPFS → cid, send chain tx → txid.
	// Dev mode: deterministic "cid" + "tx" so we can exercise DS.SetBucketAnchored.
	cid := dataserver.DevCID(raw)
	tx := dataserver.DevTX("anchored")

	_, err := r.buckets.SetAnchored(ctx, b.GetRef(), cid, tx)
	return err
}

func (r *Runner) closeStaleOpenBuckets(ctx context.Context) error {
	today := time.Now().UTC().Format("2006-01-02")

	page := ""
	total := 0
	for {
		resp, err := r.buckets.ListByStatus(ctx, "open", r.cfg.ListPageSize, page)
		if err != nil {
			return err
		}
		for _, b := range resp.GetBuckets() {
			bkey := b.GetRef().GetBucketKey()
			// bucket keys are YYYY-MM-DD, so string compare is fine
			if bkey < today {
				if _, err := r.buckets.MarkClosed(ctx, b.GetRef()); err != nil {
					log.Printf("[runner] mark-closed %s/%s/%s: %v",
						b.GetRef().GetScope().GetEntityKind(),
						b.GetRef().GetScope().GetEntityKey(),
						bkey, err)
					continue
				}
				total++
			}
		}
		page = resp.GetNextPageToken()
		if page == "" {
			break
		}
	}
	if total > 0 {
		log.Printf("[runner] auto-closed %d stale open buckets (< %s)", total, today)
	}
	return nil
}
