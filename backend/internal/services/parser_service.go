package services

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Floqqqq/Practica/backend/internal/models"
	"github.com/ledongthuc/pdf"
)

type ParserService struct{}

func NewParserService() *ParserService {
	return &ParserService{}
}

func (s *ParserService) ExtractText(filePath string) (*models.ParsedDocument, error) {
	extension := strings.ToLower(filepath.Ext(filePath))

	switch extension {
	case ".pdf":
		return s.extractPDF(filePath)
	case ".docx":
		return s.extractDOCX(filePath)
	default:
		return nil, fmt.Errorf("unsupported document format: %s", extension)
	}
}

func (s *ParserService) extractPDF(filePath string) (*models.ParsedDocument, error) {
	file, reader, err := pdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open PDF: %w", err)
	}
	defer file.Close()

	var pages []models.ParsedPage
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

func (s *ParserService) extractDOCX(filePath string) (*models.ParsedDocument, error) {
	docxFile, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("open DOCX archive: %w", err)
	}
	defer docxFile.Close()

	var fullText strings.Builder

	for _, file := range docxFile.File {
		if !isDOCXTextFile(file.Name) {
			continue
		}

		textPart, err := extractTextFromDOCXXML(file)
		if err != nil {
			return nil, fmt.Errorf("extract text from DOCX file %s: %w", file.Name, err)
		}

		textPart = normalizeExtractedText(textPart)
		if strings.TrimSpace(textPart) == "" {
			continue
		}

		fullText.WriteString(textPart)
		fullText.WriteString("\n")
	}

	text := normalizeExtractedText(fullText.String())
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("DOCX text is empty")
	}

	pages := []models.ParsedPage{
		{
			PageNumber: 1,
			Text:       text,
		},
	}

	return buildParsedDocument(pages, text)
}

func isDOCXTextFile(fileName string) bool {
	if fileName == "word/document.xml" {
		return true
	}

	if strings.HasPrefix(fileName, "word/header") && strings.HasSuffix(fileName, ".xml") {
		return true
	}

	if strings.HasPrefix(fileName, "word/footer") && strings.HasSuffix(fileName, ".xml") {
		return true
	}

	return false
}

func extractTextFromDOCXXML(file *zip.File) (string, error) {
	reader, err := file.Open()
	if err != nil {
		return "", err
	}
	defer reader.Close()

	decoder := xml.NewDecoder(reader)

	var text strings.Builder

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return "", err
		}

		switch element := token.(type) {
		case xml.StartElement:
			switch element.Name.Local {
			case "t":
				var value string
				if err := decoder.DecodeElement(&value, &element); err != nil {
					return "", err
				}
				text.WriteString(value)

			case "tab":
				text.WriteString("\t")
			}

		case xml.EndElement:
			if element.Name.Local == "p" {
				text.WriteString("\n")
			}
		}
	}

	return text.String(), nil
}

func buildParsedDocument(pages []models.ParsedPage, fullText string) (*models.ParsedDocument, error) {
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
