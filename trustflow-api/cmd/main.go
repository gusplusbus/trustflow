package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gusplusbus/trustflow/trustflow-api/internal/routes"
)

func main() {
	port := getPort()
	r := routes.NewRouter()

	log.Printf("🚀 Trustflow API is running on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}
