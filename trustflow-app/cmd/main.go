package main

import (
	"log"
	"net/http"

	"github.com/gusplusbus/trustflow/trustflow-app/internal/routes"
)

func main() {
	r := routes.NewRouter()

	log.Println("🚀 Trustflow Go App running on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
