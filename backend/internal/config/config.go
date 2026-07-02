package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppPort          string
	UploadDir        string
	ElasticsearchURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	SearchCacheTTL time.Duration
}

func Load() Config {
	return Config{
		AppPort:          getEnv("APP_PORT", "8080"),
		UploadDir:        getEnv("UPLOAD_DIR", "uploads"),
		ElasticsearchURL: getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		SearchCacheTTL: time.Duration(getEnvInt("SEARCH_CACHE_TTL_SECONDS", 300)) * time.Second,
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsedValue
}
