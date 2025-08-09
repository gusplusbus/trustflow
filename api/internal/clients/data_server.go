package clients

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
)

var (
	once   sync.Once
	conn   *grpc.ClientConn
	client projectv1.ProjectServiceClient
)

func ProjectClient() projectv1.ProjectServiceClient {
	once.Do(func() {
		target := "dns:///data_server:9090" // force gRPC’s DNS resolver

		var err error
		conn, err = grpc.DialContext(
			context.Background(),
			target,
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
			// transient UNAVAILABLE retries
			grpc.WithDefaultServiceConfig(`{
				"loadBalancingPolicy":"pick_first",
				"methodConfig":[{
					"name":[{"service":"trustflow.project.v1.ProjectService"}],
					"retryPolicy":{
						"MaxAttempts":4,
						"InitialBackoff":"0.4s",
						"MaxBackoff":"3s",
						"BackoffMultiplier":1.6,
						"RetryableStatusCodes":["UNAVAILABLE"]
					}
				}]
			}`),
			grpc.WithBlock(), // wait until a connection is made (with backoff)
		)
		if err != nil {
			log.Fatalf("grpc dial failed: %v", err)
		}
		client = projectv1.NewProjectServiceClient(conn)
	})
	return client
}
