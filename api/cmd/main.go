package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gusplusbus/trustflow/api/internal/routes"
	"github.com/gusplusbus/trustflow/api/internal/queue"
	"github.com/gusplusbus/trustflow/api/internal/timeline"
)

func main() {
	port := getPort()
	r := routes.NewRouter()

	// Workers
	ctx, cancel := context.WithCancel(context.Background())
	queue.Start(ctx, timeline.Consumer)
	queue.WaitStarted()

	// HTTP server
	srv := &http.Server{Addr: ":"+port, Handler: r}

	// graceful shutdown
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Printf("shutting down ...")
		cancel()
		_ = srv.Close()
	}()

	log.Printf("🚀 Trustflow API is running on port %s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}
