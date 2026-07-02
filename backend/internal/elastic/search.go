package elastic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Floqqqq/Practica/backend/internal/models"
)

func (c *Client) SearchChunks(
	ctx context.Context,
	query string,
	page int,
	limit int,
) (*models.SearchResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	page, limit = normalizeSearchPagination(page, limit)

	bodyBytes, err := buildSearchRequestBody(query, page, limit)
	if err != nil {
		return nil, err
	}

	response, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(DocumentsIndexName),
		c.es.Search.WithBody(bytes.NewReader(bodyBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("search chunks: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read search response: %w", err)
	}

	if response.IsError() {
		return nil, fmt.Errorf("search failed: %s: %s", response.Status(), string(responseBody))
	}

	elasticResponse, err := decodeSearchResponse(responseBody)
	if err != nil {
		return nil, err
	}

	return &models.SearchResponse{
		Query:   query,
		Page:    page,
		Limit:   limit,
		Total:   elasticResponse.Hits.Total.Value,
		Cached:  false,
		Results: mapElasticHitsToSearchResults(elasticResponse.Hits.Hits),
	}, nil
}

func normalizeSearchPagination(page int, limit int) (int, int) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	return page, limit
}
