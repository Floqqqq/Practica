package document_service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Floqqqq/Practica/backend/internal/models"
)

func (s *DocumentService) saveMetadata(metadata models.DocumentMetadata) error {
	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	metadataDir := filepath.Join(s.uploadDir, "metadata")
	hashIndexDir := filepath.Join(metadataDir, "hash_index")

	metadataByIDPath := filepath.Join(metadataDir, metadata.ID+".json")
	metadataByHashPath := filepath.Join(hashIndexDir, metadata.ContentHash+".json")

	if err := os.WriteFile(metadataByIDPath, metadataBytes, 0644); err != nil {
		return err
	}

	if err := os.WriteFile(metadataByHashPath, metadataBytes, 0644); err != nil {
		return err
	}

	return nil
}

func (s *DocumentService) findMetadataByHash(contentHash string) (*models.DocumentMetadata, error) {
	metadataPath := filepath.Join(s.uploadDir, "metadata", "hash_index", contentHash+".json")

	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata models.DocumentMetadata

	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func makeTextPreview(text string, limit int) string {
	runes := []rune(strings.TrimSpace(text))

	if len(runes) <= limit {
		return string(runes)
	}

	return string(runes[:limit]) + "..."
}
