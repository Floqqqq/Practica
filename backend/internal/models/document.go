package models

import "time"

type ParsedPage struct {
	PageNumber int    `json:"page_number"`
	Text       string `json:"text"`
}

type ParsedDocument struct {
	Pages      []ParsedPage `json:"pages"`
	Text       string       `json:"text"`
	PagesCount int          `json:"pages_count"`
	CharsCount int          `json:"chars_count"`
}

type DocumentMetadata struct {
	ID             string    `json:"id"`
	FileName       string    `json:"file_name"`
	FilePath       string    `json:"file_path"`
	TextPath       string    `json:"text_path"`
	ChunksPath     string    `json:"chunks_path"`
	Size           int64     `json:"size"`
	Status         string    `json:"status"`
	ContentHash    string    `json:"content_hash"`
	PagesCount     int       `json:"pages_count"`
	ExtractedChars int       `json:"extracted_chars"`
	ChunksCount    int       `json:"chunks_count"`
	TextPreview    string    `json:"text_preview"`
	UploadedAt     time.Time `json:"uploaded_at"`
}
