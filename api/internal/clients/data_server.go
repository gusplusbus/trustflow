package clients

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"

	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
	ownershipv1 "github.com/gusplusbus/trustflow/data_server/gen/ownershipv1"
)

var (
	onceConn    sync.Once
	conn        *grpc.ClientConn
	projectCli  projectv1.ProjectServiceClient
	ownershipCli ownershipv1.OwnershipServiceClient
)

// dialDataServer dials the data_server once and initializes both clients.
func dialDataServer() {
	// Allow override via env; default to docker-compose service name.
	addr := os.Getenv("DATA_SERVER_ADDR")
	if addr == "" {
		// Use gRPC's DNS resolver explicitly; works well in container nets
		addr = "dns:///data_server:9090"
	}

	c, err := grpc.DialContext(
		context.Background(),
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
			Backoff: backoff.Config{
				BaseDelay:  200 * time.Millisecond,
				Multiplier: 1.6,
				MaxDelay:   4 * time.Second,
				Jitter:     0.2,
			},
		}),
		// Retry UNAVAILABLE on both services
		grpc.WithDefaultServiceConfig(`{
		 "loadBalancingPolicy":"pick_first",
		 "methodConfig":[{
		   "name":[
		     {"service":"trustflow.project.v1.ProjectService"},
		     {"service":"trustflow.ownership.v1.OwnershipService"}
		   ],
		   "retryPolicy":{
		     "MaxAttempts":4,
		     "InitialBackoff":"0.4s",
		     "MaxBackoff":"3s",
		     "BackoffMultiplier":1.6,
		     "RetryableStatusCodes":["UNAVAILABLE"]
		   }
		 }]
		}`),
		grpc.WithBlock(), // wait for connection (with backoff)
	)
	if err != nil {
		log.Fatalf("grpc dial to data_server failed: %v", err)
	}
	conn = c
	projectCli = projectv1.NewProjectServiceClient(conn)
	ownershipCli = ownershipv1.NewOwnershipServiceClient(conn)
}

// ProjectClient returns a singleton ProjectServiceClient.
func ProjectClient() projectv1.ProjectServiceClient {
	onceConn.Do(dialDataServer)
	return projectCli
}

// OwnershipClient returns a singleton OwnershipServiceClient.
func OwnershipClient() ownershipv1.OwnershipServiceClient {
	onceConn.Do(dialDataServer)
	return ownershipCli
}
