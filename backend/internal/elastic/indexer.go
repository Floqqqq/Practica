package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/Floqqqq/Practica/backend/internal/models"
)

func (c *Client) IndexChunks(ctx context.Context, chunks []models.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	var body bytes.Buffer
	encoder := json.NewEncoder(&body)

	indexedAt := time.Now().UTC().Format(time.RFC3339)

	for _, chunk := range chunks {
		meta := map[string]any{
			"index": map[string]any{
				"_index": DocumentsIndexName,
				"_id":    chunk.ChunkID,
			},
		}

		if err := encoder.Encode(meta); err != nil {
			return fmt.Errorf("encode bulk meta: %w", err)
		}

		document := map[string]any{
			"chunk_id":     chunk.ChunkID,
			"document_id":  chunk.DocumentID,
			"file_name":    chunk.FileName,
			"page_number":  chunk.PageNumber,
			"chunk_index":  chunk.ChunkIndex,
			"text":         chunk.Text,
			"start_offset": chunk.StartOffset,
			"end_offset":   chunk.EndOffset,
			"chars_count":  chunk.CharsCount,
			"indexed_at":   indexedAt,
		}

		if err := encoder.Encode(document); err != nil {
			return fmt.Errorf("encode bulk document: %w", err)
		}
	}

	response, err := c.es.Bulk(
		bytes.NewReader(body.Bytes()),
		c.es.Bulk.WithContext(ctx),
		c.es.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("bulk index chunks: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read bulk response: %w", err)
	}

	if response.IsError() {
		return fmt.Errorf("bulk index failed: %s: %s", response.Status(), string(responseBody))
	}

	var bulkResponse struct {
		Errors bool `json:"errors"`
	}

	if err := json.Unmarshal(responseBody, &bulkResponse); err != nil {
		return fmt.Errorf("decode bulk response: %w", err)
	}

	if bulkResponse.Errors {
		return fmt.Errorf("bulk index finished with item errors: %s", string(responseBody))
	}

	return nil
}
