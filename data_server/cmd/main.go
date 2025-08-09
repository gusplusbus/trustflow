package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
)

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

type server struct {
	projectv1.UnimplementedProjectServiceServer
	db *pgxpool.Pool
}

func (s *server) Health(ctx context.Context, _ *projectv1.HealthRequest) (*projectv1.HealthResponse, error) {
	return &projectv1.HealthResponse{Status: "ok"}, nil
}

func (s *server) CreateProject(ctx context.Context, req *projectv1.CreateProjectRequest) (*projectv1.CreateProjectResponse, error) {
	if req.GetUserId() == "" || req.GetTitle() == "" || req.GetDescription() == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	var (
		id        string
		createdAt time.Time
		updatedAt time.Time
	)

	err := s.db.QueryRow(ctx, `
		INSERT INTO projects (
			id, created_at, updated_at, user_id, title, description, duration_estimate, team_size, application_close_time
		)
		VALUES (
			gen_random_uuid(), now() AT TIME ZONE 'utc', now() AT TIME ZONE 'utc',
			$1, $2, $3, $4, $5, $6
		)
		RETURNING id, created_at, updated_at
	`,
		req.UserId,
		req.Title,
		req.Description,
		req.DurationEstimate,
		req.TeamSize,
		req.ApplicationCloseTime,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert project: %w", err)
	}

	pr := &projectv1.Project{
		Id:                   id,
		CreatedAt:            createdAt.UTC().Format(time.RFC3339),
		UpdatedAt:            updatedAt.UTC().Format(time.RFC3339),
		Title:                req.Title,
		Description:          req.Description,
		DurationEstimate:     req.DurationEstimate,
		TeamSize:             req.TeamSize,
		ApplicationCloseTime: req.ApplicationCloseTime,
	}
	return &projectv1.CreateProjectResponse{Project: pr}, nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s not set", key)
	}
	return v
}

func ensureSchema(ctx context.Context, db *pgxpool.Pool) error {
	if _, err := db.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto;`); err != nil {
		return fmt.Errorf("enable pgcrypto: %w", err)
	}
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS projects (
		  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		  user_id UUID NOT NULL,
		  title TEXT NOT NULL,
		  description TEXT NOT NULL,
		  duration_estimate INTEGER NOT NULL,
		  team_size INTEGER NOT NULL,
		  application_close_time TEXT
		);
	`)
	if err != nil {
		return fmt.Errorf("create table projects: %w", err)
	}
	return nil
}

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":9090"
	}
	dsn := mustEnv("DATABASE_URL")

	// wait for postgres, then ensure schema
	pool := connectWithRetry(dsn, 60*time.Second)
	defer pool.Close()

	if err := ensureSchema(context.Background(), pool); err != nil {
		log.Fatalf("schema: %v", err)
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	projectv1.RegisterProjectServiceServer(s, &server{db: pool})

	log.Printf("gRPC Project service listening on %s", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
