package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const DocumentsIndexName = "documents"

func (c *Client) EnsureDocumentsIndex(ctx context.Context) error {
	exists, err := c.indexExists(ctx, DocumentsIndexName)
	if err != nil {
		return fmt.Errorf("check documents index: %w", err)
	}

	if exists {
		return nil
	}

	if err := c.createDocumentsIndex(ctx); err != nil {
		return fmt.Errorf("create documents index: %w", err)
	}

	return nil
}

func (c *Client) indexExists(ctx context.Context, indexName string) (bool, error) {
	response, err := c.es.Indices.Exists(
		[]string{indexName},
		c.es.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected status while checking index %s: %s", indexName, response.Status())
	}
}

func (c *Client) createDocumentsIndex(ctx context.Context) error {
	indexBody := documentsIndexBody()

	bodyBytes, err := json.Marshal(indexBody)
	if err != nil {
		return fmt.Errorf("marshal documents index body: %w", err)
	}

	response, err := c.es.Indices.Create(
		DocumentsIndexName,
		c.es.Indices.Create.WithContext(ctx),
		c.es.Indices.Create.WithBody(bytes.NewReader(bodyBytes)),
	)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.IsError() {
		return fmt.Errorf("elasticsearch create index failed: %s", response.Status())
	}

	return nil
}

func documentsIndexBody() map[string]any {
	return map[string]any{
		"settings": map[string]any{
			"analysis": map[string]any{
				"filter": map[string]any{
					"russian_stop": map[string]any{
						"type":      "stop",
						"stopwords": "_russian_",
					},
					"russian_stemmer": map[string]any{
						"type":     "stemmer",
						"language": "russian",
					},
				},
				"analyzer": map[string]any{
					"analysis_ru": map[string]any{
						"type":      "custom",
						"tokenizer": "standard",
						"filter": []string{
							"lowercase",
							"russian_stop",
							"russian_stemmer",
						},
					},
				},
			},
		},
		"mappings": map[string]any{
			"properties": map[string]any{
				"chunk_id": map[string]any{
					"type": "keyword",
				},
				"document_id": map[string]any{
					"type": "keyword",
				},
				"file_name": map[string]any{
					"type": "keyword",
				},
				"page_number": map[string]any{
					"type": "integer",
				},
				"chunk_index": map[string]any{
					"type": "integer",
				},
				"text": map[string]any{
					"type":            "text",
					"analyzer":        "analysis_ru",
					"search_analyzer": "analysis_ru",
				},
				"start_offset": map[string]any{
					"type": "integer",
				},
				"end_offset": map[string]any{
					"type": "integer",
				},
				"chars_count": map[string]any{
					"type": "integer",
				},
				"indexed_at": map[string]any{
					"type": "date",
				},
			},
		},
	}
}
