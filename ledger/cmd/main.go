package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gusplusbus/trustflow/ledger/internal/config"
	"github.com/gusplusbus/trustflow/ledger/internal/dataserver"
	"github.com/gusplusbus/trustflow/ledger/internal/runner"
	"github.com/gusplusbus/trustflow/ledger/internal/webhook"
)

func main() {
	cfg := config.Load()

	// --- Runner (background) ---
	bcli, closeBuckets, err := dataserver.NewBucketClient(cfg.DataServerGRPCAddr, 2*time.Second)
	if err != nil {
		log.Fatalf("[runner] buckets dial: %v", err)
	}
	defer func() { _ = closeBuckets() }()

	r := runner.New(runner.Config{
		DataServerGRPCAddr: cfg.DataServerGRPCAddr,
		Interval:           30 * time.Second,
		ListPageSize:       50,
	}, bcli)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Start(ctx)

	// --- HTTP (webhook) ---
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/webhook/github", webhook.NewGitHubHandler(cfg)) // unchanged

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("[ledger] listening on %s", cfg.HTTPAddr)
	log.Fatal(srv.ListenAndServe())
}
