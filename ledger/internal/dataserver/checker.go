package dataserver

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	issuev1 "github.com/gusplusbus/trustflow/data_server/gen/issuev1"
)

type Checker interface {
	// We only need ghIssueID now; owner/repo/number can be ignored here.
	IsManaged(ctx context.Context, owner, repo string, ghIssueID int64, number int) (bool, error)
}

type GRPCChecker struct {
	Addr    string                // e.g. "data_server:9090"
	Timeout time.Duration         // e.g. 900 * time.Millisecond
	cli     issuev1.IssueServiceClient
	conn    *grpc.ClientConn
}

func NewGRPCChecker(addr string, timeout time.Duration) (*GRPCChecker, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { return nil, err }
	return &GRPCChecker{
		Addr:    addr,
		Timeout: timeout,
		cli:     issuev1.NewIssueServiceClient(conn),
		conn:    conn,
	}, nil
}

func (c *GRPCChecker) Close() error {
	if c.conn != nil { return c.conn.Close() }
	return nil
}

func (c *GRPCChecker) IsManaged(ctx context.Context, _ string, _ string, ghIssueID int64, _ int) (bool, error) {
	if c.cli == nil {
		return true, nil // fail-open or return error; your call
	}
	var cancel context.CancelFunc
	if c.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}
	resp, err := c.cli.ExistsByGhID(ctx, &issuev1.ExistsByGhIDRequest{GhIssueId: ghIssueID})
	if err != nil {
		log.Printf("[ledger] ExistsByGhID RPC error: %v", err)
		return false, err
	}
	return resp.GetExists(), nil
}
