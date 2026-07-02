package parser_service

import (
	"errors"
	"strings"
	"unicode"

	"github.com/Floqqqq/Practica/backend/internal/models"
)

func buildParsedDocument(
	pages []models.ParsedPage,
	fullText string,
) (*models.ParsedDocument, error) {
	text := normalizeExtractedText(fullText)

	if strings.TrimSpace(text) == "" {
		return nil, errors.New("extracted text is empty")
	}

	return &models.ParsedDocument{
		Pages:      pages,
		Text:       text,
		PagesCount: len(pages),
		CharsCount: len([]rune(text)),
	}, nil
}

func normalizeExtractedText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	text = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}

		if unicode.IsControl(r) {
			return -1
		}

		return r
	}, text)

	lines := strings.Split(text, "\n")
	normalizedLines := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			continue
		}

		normalizedLines = append(normalizedLines, line)
	}

	return strings.Join(normalizedLines, "\n")
}
