package services

import (
	"context"
	"errors"
	"strings"

	"github.com/Floqqqq/Practica/backend/internal/elastic"
	"github.com/Floqqqq/Practica/backend/internal/models"
)

var ErrInvalidSearchQuery = errors.New("invalid search query")

type SearchService struct {
	elasticClient *elastic.Client
	cacheService  *CacheService
}

func NewSearchService(
	elasticClient *elastic.Client,
	cacheService *CacheService,
) *SearchService {
	return &SearchService{
		elasticClient: elasticClient,
		cacheService:  cacheService,
	}
}

func (s *SearchService) Search(
	ctx context.Context,
	query string,
	page int,
	limit int,
) (*models.SearchResponse, error) {
	query = strings.TrimSpace(query)

	if query == "" {
		return nil, ErrInvalidSearchQuery
	}

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 50 {
		limit = 50
	}

	if s.cacheService != nil && s.cacheService.Enabled() {
		cacheKey := s.cacheService.BuildSearchCacheKey(query, page, limit)

		cachedResponse, found, err := s.cacheService.GetSearchResponse(ctx, cacheKey)
		if err == nil && found {
			return cachedResponse, nil
		}

		searchResponse, err := s.elasticClient.SearchChunks(ctx, query, page, limit)
		if err != nil {
			return nil, err
		}

		searchResponse.Cached = false

		_ = s.cacheService.SetSearchResponse(ctx, cacheKey, searchResponse)

		return searchResponse, nil
	}

	searchResponse, err := s.elasticClient.SearchChunks(ctx, query, page, limit)
	if err != nil {
		return nil, err
	}

	searchResponse.Cached = false

	return searchResponse, nil
}
