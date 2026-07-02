package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Floqqqq/Practica/backend/internal/models"
	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCacheService(client *redis.Client, ttl time.Duration) *CacheService {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	return &CacheService{
		client: client,
		ttl:    ttl,
	}
}

func (s *CacheService) Enabled() bool {
	return s != nil && s.client != nil
}

func (s *CacheService) BuildSearchCacheKey(query string, page int, limit int) string {
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	queryHash := sha256.Sum256([]byte(normalizedQuery))

	return fmt.Sprintf("search:%x:page:%d:limit:%d", queryHash, page, limit)
}

func (s *CacheService) GetSearchResponse(
	ctx context.Context,
	key string,
) (*models.SearchResponse, bool, error) {
	if !s.Enabled() {
		return nil, false, nil
	}

	data, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	var cachedResponse models.SearchResponse

	if err := json.Unmarshal(data, &cachedResponse); err != nil {
		_ = s.client.Del(ctx, key).Err()
		return nil, false, err
	}

	cachedResponse.Cached = true

	return &cachedResponse, true, nil
}

func (s *CacheService) SetSearchResponse(
	ctx context.Context,
	key string,
	searchResponse *models.SearchResponse,
) error {
	if !s.Enabled() || searchResponse == nil {
		return nil
	}

	responseForCache := *searchResponse
	responseForCache.Cached = false

	data, err := json.Marshal(responseForCache)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, key, data, s.ttl).Err()
}
