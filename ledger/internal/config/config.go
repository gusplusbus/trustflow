package config

import (
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr             string // e.g. :9091
	GitHubWebhookSecret  []byte // used to verify X-Hub-Signature-256
	APIURL               string // e.g. http://api:8080/internal/ledger/notify
	DataServerGRPCAddr   string // optional: e.g. data_server:9090 (stubbed)
	APITimeout           time.Duration
}

func mustEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf("missing env %s", key)
	}
	return v
}

func Load() Config {
	httpAddr := os.Getenv("LEDGER_HTTP_ADDR")
	if strings.TrimSpace(httpAddr) == "" {
		httpAddr = ":9091"
	}
	apiURL := os.Getenv("API_SYNC_URL")
	if strings.TrimSpace(apiURL) == "" {
		apiURL = "http://api:8080/internal/ledger/notify"
	}
	gh := []byte(mustEnv("GITHUB_WEBHOOK_SECRET"))

	return Config{
		HTTPAddr:            httpAddr,
		GitHubWebhookSecret: gh,
		APIURL:              apiURL,
		DataServerGRPCAddr:  os.Getenv("DATASERVER_GRPC_ADDR"),
		APITimeout:          6 * time.Second,
	}
}
