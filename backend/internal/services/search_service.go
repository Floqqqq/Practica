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
}

func NewSearchService(elasticClient *elastic.Client) *SearchService {
	return &SearchService{
		elasticClient: elasticClient,
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

	return s.elasticClient.SearchChunks(ctx, query, page, limit)
}
