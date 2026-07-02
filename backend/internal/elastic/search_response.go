package elastic

import (
	"encoding/json"
	"fmt"

	"github.com/Floqqqq/Practica/backend/internal/models"
)

type elasticSearchResponse struct {
	Hits elasticSearchHits `json:"hits"`
}

type elasticSearchHits struct {
	Total elasticSearchTotal `json:"total"`
	Hits  []elasticSearchHit `json:"hits"`
}

type elasticSearchTotal struct {
	Value int64 `json:"value"`
}

type elasticSearchHit struct {
	Score     float64             `json:"_score"`
	Source    elasticChunkSource  `json:"_source"`
	Highlight map[string][]string `json:"highlight"`
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

func decodeSearchResponse(responseBody []byte) (*elasticSearchResponse, error) {
	var elasticResponse elasticSearchResponse

	if err := json.Unmarshal(responseBody, &elasticResponse); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	return &elasticResponse, nil
}

func mapElasticHitsToSearchResults(hits []elasticSearchHit) []models.SearchResult {
	results := make([]models.SearchResult, 0, len(hits))

	for _, hit := range hits {
		results = append(results, models.SearchResult{
			ChunkID:    hit.Source.ChunkID,
			DocumentID: hit.Source.DocumentID,
			FileName:   hit.Source.FileName,
			Page:       hit.Source.PageNumber,
			ChunkIndex: hit.Source.ChunkIndex,
			Text:       hit.Source.Text,
			Highlight:  extractHighlight(hit.Highlight),
			Score:      hit.Score,
		})
	}

	return results
}

func extractHighlight(highlight map[string][]string) string {
	fragments, ok := highlight["text"]
	if !ok || len(fragments) == 0 {
		return ""
	}

	return fragments[0]
}
