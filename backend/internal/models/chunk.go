package models

type Chunk struct {
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
