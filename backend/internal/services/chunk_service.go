package services

import (
	"errors"
	"strings"

	"github.com/Floqqqq/Practica/backend/internal/models"
	"github.com/google/uuid"
)

const (
	DefaultChunkSize    = 1000
	DefaultChunkOverlap = 100
)

type ChunkService struct {
	chunkSize int
	overlap   int
}

func NewChunkService() *ChunkService {
	return &ChunkService{
		chunkSize: DefaultChunkSize,
		overlap:   DefaultChunkOverlap,
	}
}

func (s *ChunkService) SplitDocumentIntoChunks(
	documentID string,
	fileName string,
	parsedDocument *models.ParsedDocument,
) ([]models.Chunk, error) {
	if strings.TrimSpace(documentID) == "" {
		return nil, errors.New("document id is required")
	}

	if strings.TrimSpace(fileName) == "" {
		return nil, errors.New("file name is required")
	}

	if parsedDocument == nil {
		return nil, errors.New("parsed document is required")
	}

	if len(parsedDocument.Pages) == 0 {
		return nil, errors.New("parsed document has no pages")
	}

	if s.chunkSize <= 0 {
		return nil, errors.New("chunk size must be positive")
	}

	if s.overlap < 0 {
		return nil, errors.New("chunk overlap cannot be negative")
	}

	if s.overlap >= s.chunkSize {
		return nil, errors.New("chunk overlap must be less than chunk size")
	}

	chunks := make([]models.Chunk, 0)
	chunkIndex := 0

	for _, page := range parsedDocument.Pages {
		pageText := strings.TrimSpace(page.Text)
		if pageText == "" {
			continue
		}

		pageChunks := s.splitPageIntoChunks(documentID, fileName, page.PageNumber, pageText, chunkIndex)

		chunks = append(chunks, pageChunks...)
		chunkIndex += len(pageChunks)
	}

	if len(chunks) == 0 {
		return nil, errors.New("no chunks were created")
	}

	return chunks, nil
}

func (s *ChunkService) splitPageIntoChunks(
	documentID string,
	fileName string,
	pageNumber int,
	pageText string,
	startChunkIndex int,
) []models.Chunk {
	runes := []rune(pageText)

	if len(runes) <= s.chunkSize {
		return []models.Chunk{
			{
				ChunkID:     uuid.NewString(),
				DocumentID:  documentID,
				FileName:    fileName,
				PageNumber:  pageNumber,
				ChunkIndex:  startChunkIndex,
				Text:        string(runes),
				StartOffset: 0,
				EndOffset:   len(runes),
				CharsCount:  len(runes),
			},
		}
	}

	step := s.chunkSize - s.overlap

	chunks := make([]models.Chunk, 0)
	chunkIndex := startChunkIndex

	for start := 0; start < len(runes); start += step {
		end := start + s.chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunkText := strings.TrimSpace(string(runes[start:end]))
		if chunkText != "" {
			chunks = append(chunks, models.Chunk{
				ChunkID:     uuid.NewString(),
				DocumentID:  documentID,
				FileName:    fileName,
				PageNumber:  pageNumber,
				ChunkIndex:  chunkIndex,
				Text:        chunkText,
				StartOffset: start,
				EndOffset:   end,
				CharsCount:  len([]rune(chunkText)),
			})

			chunkIndex++
		}

		if end == len(runes) {
			break
		}
	}

	return chunks
}
