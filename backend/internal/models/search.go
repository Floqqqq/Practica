package models

type SearchResult struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id,omitempty"`
	FileName   string  `json:"file_name"`
	Page       int     `json:"page"`
	ChunkIndex int     `json:"chunk_index,omitempty"`
	Text       string  `json:"text"`
	Highlight  string  `json:"highlight,omitempty"`
	Score      float64 `json:"score"`
}

type SearchResponse struct {
	Query   string         `json:"query"`
	Page    int            `json:"page"`
	Limit   int            `json:"limit"`
	Total   int64          `json:"total"`
	Cached  bool           `json:"cached"`
	Results []SearchResult `json:"results"`
}
