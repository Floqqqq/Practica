package parser_service

import (
	"fmt"
	"strings"

	"github.com/Floqqqq/Practica/backend/internal/models"
	"github.com/ledongthuc/pdf"
)

func extractPDF(filePath string) (*models.ParsedDocument, error) {
	file, reader, err := pdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open PDF: %w", err)
	}
	defer file.Close()

	pages := make([]models.ParsedPage, 0)
	var fullText strings.Builder

	for pageNumber := 1; pageNumber <= reader.NumPage(); pageNumber++ {
		page := reader.Page(pageNumber)
		if page.V.IsNull() {
			continue
		}

		pageText, err := page.GetPlainText(nil)
		if err != nil {
			return nil, fmt.Errorf("extract text from PDF page %d: %w", pageNumber, err)
		}

		pageText = normalizeExtractedText(pageText)
		if strings.TrimSpace(pageText) == "" {
			continue
		}

		pages = append(pages, models.ParsedPage{
			PageNumber: pageNumber,
			Text:       pageText,
		})

		fullText.WriteString(fmt.Sprintf("\n\n--- PAGE %d ---\n", pageNumber))
		fullText.WriteString(pageText)
	}

	return buildParsedDocument(pages, fullText.String())
}
