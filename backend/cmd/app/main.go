package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Floqqqq/Practica/backend/internal/config"
	"github.com/Floqqqq/Practica/backend/internal/elastic"
	httpserver "github.com/Floqqqq/Practica/backend/internal/http"
	"github.com/Floqqqq/Practica/backend/internal/services"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	elasticClient, err := elastic.NewClient(cfg.ElasticsearchURL)
	if err != nil {
		log.Fatalf("failed to create elasticsearch client: %v", err)
	}

	elasticCtx, elasticCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer elasticCancel()

	if err := elasticClient.Ping(elasticCtx); err != nil {
		log.Fatalf("failed to connect to elasticsearch: %v", err)
	}

	if err := elasticClient.EnsureDocumentsIndex(elasticCtx); err != nil {
		log.Fatalf("failed to ensure documents index: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	cacheService := services.NewCacheService(redisClient, cfg.SearchCacheTTL)

	redisCtx, redisCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer redisCancel()

	if err := redisClient.Ping(redisCtx).Err(); err != nil {
		log.Printf("redis unavailable, search cache disabled: %v", err)
		_ = redisClient.Close()
		cacheService = services.NewCacheService(nil, cfg.SearchCacheTTL)
	} else {
		log.Printf("redis connected: %s", cfg.RedisAddr)
	}

	router := httpserver.NewRouter(cfg.UploadDir, elasticClient, cacheService)

	log.Printf("backend started on port %s", cfg.AppPort)
	log.Printf("elasticsearch connected: %s", cfg.ElasticsearchURL)
	log.Printf("documents index is ready: %s", elastic.DocumentsIndexName)

	if err := http.ListenAndServe(":"+cfg.AppPort, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
