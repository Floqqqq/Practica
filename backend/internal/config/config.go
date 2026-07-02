package config

import "os"

type Config struct {
	AppPort          string
	UploadDir        string
	ElasticsearchURL string
}

func Load() Config {
	return Config{
		AppPort:          getEnv("APP_PORT", "8080"),
		UploadDir:        getEnv("UPLOAD_DIR", "uploads"),
		ElasticsearchURL: getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}
