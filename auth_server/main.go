package main

import (
	"database/sql"
  "github.com/gusplusbus/trustflow/auth_server/internal/routes"
  "github.com/gusplusbus/trustflow/auth_server/internal/database"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set in environment variables")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
	  platform = "dev"
  }

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to database:", err)
	}

	queries := database.New(db)

	// Register routes
	mux := routes.Register(queries, platform)

	// Start server
	server := &http.Server{
		Addr:    ":4000",
		Handler: mux,
	}

	fmt.Println("Server starting on http://localhost:4000")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Server error:", err)
	}
}
