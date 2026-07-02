package parser_service

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Floqqqq/Practica/backend/internal/models"
)

func extractDOCX(filePath string) (*models.ParsedDocument, error) {
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
