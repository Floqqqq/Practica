package elastic

import (
	"bytes"
	"context"
	"encoding/json"
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

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	from := (page - 1) * limit

	searchBody := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"text"},
			},
		},
		"from": from,
		"size": limit,
		"highlight": map[string]any{
			"pre_tags":  []string{"<mark>"},
			"post_tags": []string{"</mark>"},
			"fields": map[string]any{
				"text": map[string]any{
					"fragment_size":       300,
					"number_of_fragments": 1,
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("marshal search body: %w", err)
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

	var elasticResponse struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Score     float64             `json:"_score"`
				Source    elasticChunkSource  `json:"_source"`
				Highlight map[string][]string `json:"highlight"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.Unmarshal(responseBody, &elasticResponse); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	results := make([]models.SearchResult, 0, len(elasticResponse.Hits.Hits))

	for _, hit := range elasticResponse.Hits.Hits {
		highlight := ""

		if fragments, ok := hit.Highlight["text"]; ok && len(fragments) > 0 {
			highlight = fragments[0]
		}

		results = append(results, models.SearchResult{
			ChunkID:    hit.Source.ChunkID,
			DocumentID: hit.Source.DocumentID,
			FileName:   hit.Source.FileName,
			Page:       hit.Source.PageNumber,
			ChunkIndex: hit.Source.ChunkIndex,
			Text:       hit.Source.Text,
			Highlight:  highlight,
			Score:      hit.Score,
		})
	}

	return &models.SearchResponse{
		Query:   query,
		Page:    page,
		Limit:   limit,
		Total:   elasticResponse.Hits.Total.Value,
		Cached:  false,
		Results: results,
	}, nil
}

type elasticChunkSource struct {
	ChunkID     string `json:"chunk_id"`
	DocumentID  string `json:"document_id"`
	FileName    string `json:"file_name"`
	PageNumber  int    `json:"page_number"`
	ChunkIndex  int    `json:"chunk_index"`
	Text        string `json:"text"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	CharsCount  int    `json:"chars_count"`
}
