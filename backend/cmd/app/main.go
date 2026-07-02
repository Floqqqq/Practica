package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Floqqqq/Practica/backend/internal/config"
	"github.com/Floqqqq/Practica/backend/internal/elastic"
	httpserver "github.com/Floqqqq/Practica/backend/internal/http"
)

func main() {
	cfg := config.Load()

	elasticClient, err := elastic.NewClient(cfg.ElasticsearchURL)
	if err != nil {
		log.Fatalf("failed to create elasticsearch client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := elasticClient.Ping(ctx); err != nil {
		log.Fatalf("failed to connect to elasticsearch: %v", err)
	}

	if err := elasticClient.EnsureDocumentsIndex(ctx); err != nil {
		log.Fatalf("failed to ensure documents index: %v", err)
	}

	router := httpserver.NewRouter(cfg.UploadDir, elasticClient)

	log.Printf("backend started on port %s", cfg.AppPort)
	log.Printf("elasticsearch connected: %s", cfg.ElasticsearchURL)
	log.Printf("documents index is ready: %s", elastic.DocumentsIndexName)

	if err := http.ListenAndServe(":"+cfg.AppPort, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
