package parser_service

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Floqqqq/Practica/backend/internal/models"
)

type ParserService struct{}

func NewParserService() *ParserService {
	return &ParserService{}
}

func (s *ParserService) ExtractText(filePath string) (*models.ParsedDocument, error) {
	extension := strings.ToLower(filepath.Ext(filePath))

	switch extension {
	case ".pdf":
		return extractPDF(filePath)
	case ".docx":
		return extractDOCX(filePath)
	default:
		return nil, fmt.Errorf("unsupported document format: %s", extension)
	}
}
