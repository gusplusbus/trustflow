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

  walletv1 "github.com/gusplusbus/trustflow/data_server/gen/walletv1"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
	ownershipv1 "github.com/gusplusbus/trustflow/data_server/gen/ownershipv1"
	issuev1 "github.com/gusplusbus/trustflow/data_server/gen/issuev1"
  issuetimelinev1 "github.com/gusplusbus/trustflow/data_server/gen/issuetimelinev1"
)

var (
	onceConn     sync.Once
	grpcConn     *grpc.ClientConn
	projectCli   projectv1.ProjectServiceClient
	ownershipCli ownershipv1.OwnershipServiceClient
	issueCli     issuev1.IssueServiceClient
  timelineCli  issuetimelinev1.IssuesTimelineServiceClient
  walletCli    walletv1.WalletServiceClient
)

// dialDataServer dials the data_server once and initializes all clients.
func dialDataServer() {
	addr := os.Getenv("DATA_SERVER_ADDR")
	if addr == "" {
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
		grpc.WithDefaultServiceConfig(`{
		 "loadBalancingPolicy":"pick_first",
		 "methodConfig":[{
		   "name":[
		     {"service":"trustflow.project.v1.ProjectService"},
		     {"service":"trustflow.ownership.v1.OwnershipService"},
		     {"service":"trustflow.issue.v1.IssueService"},
         {"service":"trustflow.issues_timeline.v1.IssuesTimelineService"}
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
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("grpc dial to data_server failed: %v", err)
	}
	grpcConn = c
	projectCli = projectv1.NewProjectServiceClient(grpcConn)
	ownershipCli = ownershipv1.NewOwnershipServiceClient(grpcConn)
	issueCli = issuev1.NewIssueServiceClient(grpcConn)
  timelineCli  = issuetimelinev1.NewIssuesTimelineServiceClient(grpcConn)
  walletCli = walletv1.NewWalletServiceClient(grpcConn)
}

func ProjectClient() projectv1.ProjectServiceClient {
	onceConn.Do(dialDataServer)
	return projectCli
}

func OwnershipClient() ownershipv1.OwnershipServiceClient {
	onceConn.Do(dialDataServer)
	return ownershipCli
}

func IssueClient() issuev1.IssueServiceClient {
	onceConn.Do(dialDataServer)
	return issueCli
}

func TimelineClient() issuetimelinev1.IssuesTimelineServiceClient {
  onceConn.Do(dialDataServer)
  return timelineCli
}

func WalletClient() walletv1.WalletServiceClient {
  onceConn.Do(dialDataServer)
  return walletCli
}
