package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
	"github.com/gusplusbus/trustflow/data_server/internal/grpcserver"
	"github.com/gusplusbus/trustflow/data_server/internal/repo/postgres"
	"github.com/gusplusbus/trustflow/data_server/internal/service"
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

	// Wait for Postgres to be ready
	pool := connectWithRetry(dsn, 60*time.Second)
	defer pool.Close()

	// Wire: repo -> service -> gRPC server
	pgRepo, err := postgres.NewProjectPG(pool)
	if err != nil {
		log.Fatalf("repo init: %v", err)
	}
	svc := service.NewProjectService(pgRepo)
	grpcSrv := grpcserver.NewProjectServer(svc)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	projectv1.RegisterProjectServiceServer(s, grpcSrv)

	log.Printf("gRPC Project service listening on %s", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
