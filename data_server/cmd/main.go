package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	bucketv1 "github.com/gusplusbus/trustflow/data_server/gen/bucketv1"
	issuev1 "github.com/gusplusbus/trustflow/data_server/gen/issuev1"
	issuetimelinev1 "github.com/gusplusbus/trustflow/data_server/gen/issuetimelinev1"
	ownershipv1 "github.com/gusplusbus/trustflow/data_server/gen/ownershipv1"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"

	"github.com/gusplusbus/trustflow/data_server/internal/grpcserver"
	"github.com/gusplusbus/trustflow/data_server/internal/repo/postgres"
	"github.com/gusplusbus/trustflow/data_server/internal/service"
	"github.com/gusplusbus/trustflow/data_server/internal/service/dbwrap"
)

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s not set", key)
	}
	return v
}

func connectWithRetry(dsn string, maxWait time.Duration) *pgxpool.Pool {
	deadline := time.Now().Add(maxWait)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		pool, err := pgxpool.New(ctx, dsn)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				cancel()
				return pool
			}
			pool.Close()
		}
		cancel()
		if time.Now().After(deadline) {
			log.Fatalf("database still not ready after %s", maxWait)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":9090"
	}
	dsn := mustEnv("DATABASE_URL")

	// DB
	pool := connectWithRetry(dsn, 60*time.Second)
	defer pool.Close()

	// Repos
	projectRepo, err := postgres.NewProjectPG(pool)
	if err != nil {
		log.Fatalf("project repo init: %v", err)
	}
	ownershipRepo, err := postgres.NewOwnershipPG(pool)
	if err != nil {
		log.Fatalf("ownership repo init: %v", err)
	}
	issueRepo, err := postgres.NewIssuePG(pool)
	if err != nil {
		log.Fatalf("issue repo init: %v", err)
	}
	issuesTimelineRepo, err := postgres.NewIssuesTimelinePG(pool)
	if err != nil {
		log.Fatalf("issues timeline repo init: %v", err)
	}

	// NEW: bucket repo
	bucketRepo := postgres.NewBucketRepo(pool, postgres.LoadEmbeddedQueries())

	// Services
	projectSvc := service.NewProjectService(projectRepo, ownershipRepo)
	ownershipSvc := service.NewOwnershipService(ownershipRepo)
	issueSvc := service.NewIssueService(projectRepo, ownershipRepo, issueRepo, dbwrap.PoolExec{Pool: pool})

	// IMPORTANT: use the bucket-aware constructor
	issuesTimelineSvc := service.NewIssuesTimelineServiceWithBuckets(issuesTimelineRepo, bucketRepo, pool)
	bucketSvc := service.NewBucketService(bucketRepo)

	// gRPC
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	s := grpc.NewServer()

	// Servers
	projectSrv := grpcserver.NewProjectServer(projectSvc, ownershipSvc)
	ownershipSrv := grpcserver.NewOwnershipServer(ownershipSvc)
	issueSrv := grpcserver.NewIssueServer(issueSvc)
	issuesTimelineSrv := grpcserver.NewIssuesTimelineGRPC(issuesTimelineSvc)
	bucketSrv := grpcserver.NewBucketServer(bucketSvc)

	// Register
	projectv1.RegisterProjectServiceServer(s, projectSrv)
	ownershipv1.RegisterOwnershipServiceServer(s, ownershipSrv)
	issuev1.RegisterIssueServiceServer(s, issueSrv)
	issuetimelinev1.RegisterIssuesTimelineServiceServer(s, issuesTimelineSrv)
	bucketv1.RegisterBucketServiceServer(s, bucketSrv)

	log.Printf("gRPC services listening on %s (Project, Ownership, Issue, IssuesTimeline, Bucket)", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
