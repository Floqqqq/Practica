package elastic

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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
	response, err := c.es.Indices.Create(
		DocumentsIndexName,
		c.es.Indices.Create.WithContext(ctx),
		c.es.Indices.Create.WithBody(strings.NewReader(documentsIndexMapping)),
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
