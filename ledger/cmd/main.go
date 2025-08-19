package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gusplusbus/trustflow/ledger/internal/config"
	"github.com/gusplusbus/trustflow/ledger/internal/webhook"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/webhook/github", webhook.NewGitHubHandler(cfg))

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("[ledger] listening on %s", cfg.HTTPAddr)
	log.Fatal(srv.ListenAndServe())
}
