package clients

import (
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	projectv1 "github.com/gusplusbus/trustflow/data_server/gen/projectv1"
)

var (
	conn   *grpc.ClientConn
	client projectv1.ProjectServiceClient
	once   sync.Once
)

func ProjectClient() projectv1.ProjectServiceClient {
	once.Do(func() {
		var err error
		addr := "data_server:9090" // 👈 container DNS name from docker-compose
		conn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("dial grpc: %v", err)
		}
		client = projectv1.NewProjectServiceClient(conn)
	})
	return client
}
