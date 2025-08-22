package dataserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	bucketv1 "github.com/gusplusbus/trustflow/data_server/gen/bucketv1"
)

type Buckets interface {
	ListByStatus(ctx context.Context, status string, limit int32, pageToken string) (*bucketv1.ListBucketsByStatusResponse, error)
	SetAnchored(ctx context.Context, ref *bucketv1.BucketRef, cid, anchoredTx string) (*bucketv1.SetBucketAnchoredResponse, error)
}

type bucketClient struct {
	cc  *grpc.ClientConn
	api bucketv1.BucketServiceClient
}

func NewBucketClient(addr string, dialTimeout time.Duration) (Buckets, func() error, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, nil, err
	}
	return &bucketClient{cc: cc, api: bucketv1.NewBucketServiceClient(cc)}, cc.Close, nil
}

func (c *bucketClient) ListByStatus(ctx context.Context, status string, limit int32, pageToken string) (*bucketv1.ListBucketsByStatusResponse, error) {
	req := &bucketv1.ListBucketsByStatusRequest{Status: status, Limit: limit, PageToken: pageToken}
	return c.api.ListBucketsByStatus(ctx, req)
}

func (c *bucketClient) SetAnchored(ctx context.Context, ref *bucketv1.BucketRef, cid, anchoredTx string) (*bucketv1.SetBucketAnchoredResponse, error) {
	req := &bucketv1.SetBucketAnchoredRequest{Ref: ref, Cid: cid, AnchoredTx: anchoredTx}
	return c.api.SetBucketAnchored(ctx, req)
}

// --- Helpers you can reuse if you want a quick "CID-ish" dev stub ---
// DevCID returns a deterministic string for bytes (NOT a real IPFS CID).
func DevCID(b []byte) string {
	sum := sha256.Sum256(b)
	return "devcid-" + hex.EncodeToString(sum[:8])
}

// DevTX returns a fake tx id for dev.
func DevTX(prefix string) string {
	now := time.Now().UTC().UnixNano()
	return fmt.Sprintf("%s-%d", prefix, now)
}
