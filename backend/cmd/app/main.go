package main

import (
	"log"
	"net/http"
	"os"

	httpserver "github.com/Floqqqq/Practica/backend/internal/http"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "uploads"
	}

	router := httpserver.NewRouter(uploadDir)

	log.Printf("backend started on port %s", port)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
