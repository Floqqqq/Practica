package document_service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/Floqqqq/Practica/backend/internal/elastic"
	"github.com/Floqqqq/Practica/backend/internal/models"
	baseServices "github.com/Floqqqq/Practica/backend/internal/services"
	parser_service "github.com/Floqqqq/Practica/backend/internal/services/parser_service"
	"github.com/google/uuid"
)

const MaxFileSize int64 = 20 * 1024 * 1024

var ErrInvalidFile = errors.New("invalid file")

type DocumentService struct {
	uploadDir     string
	parser        *parser_service.ParserService
	chunker       *baseServices.ChunkService
	elasticClient *elastic.Client
}

type UploadResult struct {
	ID                 string `json:"id"`
	OriginalDocumentID string `json:"original_document_id,omitempty"`
	FileName           string `json:"file_name"`
	FilePath           string `json:"file_path,omitempty"`
	TextPath           string `json:"text_path,omitempty"`
	ChunksPath         string `json:"chunks_path,omitempty"`
	Size               int64  `json:"size"`
	Status             string `json:"status"`
	Message            string `json:"message"`
	ContentHash        string `json:"content_hash"`
	PagesCount         int    `json:"pages_count"`
	ExtractedChars     int    `json:"extracted_chars"`
	ChunksCount        int    `json:"chunks_count"`
	TextPreview        string `json:"text_preview,omitempty"`
	Duplicate          bool   `json:"duplicate"`
}

func NewDocumentService(uploadDir string, elasticClients ...*elastic.Client) *DocumentService {
	var elasticClient *elastic.Client

	if len(elasticClients) > 0 {
		elasticClient = elasticClients[0]
	}

	return &DocumentService{
		uploadDir:     uploadDir,
		parser:        parser_service.NewParserService(),
		chunker:       baseServices.NewChunkService(),
		elasticClient: elasticClient,
	}
}

func (s *DocumentService) Upload(
	ctx context.Context,
	file multipart.File,
	fileHeader *multipart.FileHeader,
) (*UploadResult, error) {
	if err := validateFile(file, fileHeader); err != nil {
		return nil, err
	}

	documentID := uuid.NewString()
	safeFileName := filepath.Base(fileHeader.Filename)
	paths := newDocumentPaths(s.uploadDir, documentID, safeFileName)

	if err := s.prepareUploadDirs(); err != nil {
		return nil, fmt.Errorf("prepare upload dirs: %w", err)
	}

	contentHash, writtenBytes, err := saveFileWithHash(file, paths.OriginalFile)
	if err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	if writtenBytes > MaxFileSize {
		removeFiles(paths.OriginalFile)
		return nil, fmt.Errorf("%w: file size exceeds 20 MB", ErrInvalidFile)
	}

	if duplicatedDocument, err := s.findMetadataByHash(contentHash); err == nil {
		removeFiles(paths.OriginalFile)

		return duplicateUploadResult(
			documentID,
			safeFileName,
			writtenBytes,
			contentHash,
			duplicatedDocument,
		), nil
	}

	if err := ensureContextIsActive(ctx); err != nil {
		removeFiles(paths.OriginalFile)
		return nil, err
	}

	parsedDocument, err := s.extractAndSaveText(paths)
	if err != nil {
		removeFiles(paths.OriginalFile)
		return nil, err
	}

	chunks, err := s.createAndSaveChunks(documentID, safeFileName, parsedDocument, paths)
	if err != nil {
		removeFiles(paths.OriginalFile, paths.ExtractedText)
		return nil, err
	}

	documentStatus, documentMessage, err := s.indexChunksIfConfigured(ctx, chunks)
	if err != nil {
		removeFiles(paths.OriginalFile, paths.ExtractedText, paths.Chunks)
		return nil, err
	}

	textPreview := makeTextPreview(parsedDocument.Text, 300)

	metadata := buildDocumentMetadata(
		documentID,
		safeFileName,
		paths,
		writtenBytes,
		documentStatus,
		contentHash,
		parsedDocument,
		chunks,
		textPreview,
	)

	if err := s.saveMetadata(metadata); err != nil {
		removeFiles(paths.OriginalFile, paths.ExtractedText, paths.Chunks)
		return nil, fmt.Errorf("save document metadata: %w", err)
	}

	return successfulUploadResult(metadata, documentMessage), nil
}

func (s *DocumentService) extractAndSaveText(
	paths documentPaths,
) (*models.ParsedDocument, error) {
	parsedDocument, err := s.parser.ExtractText(paths.OriginalFile)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to extract text: %v", ErrInvalidFile, err)
	}

	if err := writeTextFile(paths.ExtractedText, parsedDocument.Text); err != nil {
		return nil, fmt.Errorf("save extracted text: %w", err)
	}

	return parsedDocument, nil
}

func (s *DocumentService) createAndSaveChunks(
	documentID string,
	safeFileName string,
	parsedDocument *models.ParsedDocument,
	paths documentPaths,
) ([]models.Chunk, error) {
	chunks, err := s.chunker.SplitDocumentIntoChunks(documentID, safeFileName, parsedDocument)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to split text into chunks: %v", ErrInvalidFile, err)
	}

	if err := saveChunks(paths.Chunks, chunks); err != nil {
		return nil, fmt.Errorf("save chunks: %w", err)
	}

	return chunks, nil
}

func (s *DocumentService) indexChunksIfConfigured(
	ctx context.Context,
	chunks []models.Chunk,
) (string, string, error) {
	if s.elasticClient == nil {
		return "chunked", "file uploaded, text extracted and split into chunks successfully", nil
	}

	if err := s.elasticClient.IndexChunks(ctx, chunks); err != nil {
		return "", "", fmt.Errorf("index chunks in elasticsearch: %w", err)
	}

	return "indexed", "file uploaded, text extracted, split into chunks and indexed successfully", nil
}

func ensureContextIsActive(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func buildDocumentMetadata(
	documentID string,
	safeFileName string,
	paths documentPaths,
	writtenBytes int64,
	documentStatus string,
	contentHash string,
	parsedDocument *models.ParsedDocument,
	chunks []models.Chunk,
	textPreview string,
) models.DocumentMetadata {
	return models.DocumentMetadata{
		ID:             documentID,
		FileName:       safeFileName,
		FilePath:       paths.OriginalFile,
		TextPath:       paths.ExtractedText,
		ChunksPath:     paths.Chunks,
		Size:           writtenBytes,
		Status:         documentStatus,
		ContentHash:    contentHash,
		PagesCount:     parsedDocument.PagesCount,
		ExtractedChars: parsedDocument.CharsCount,
		ChunksCount:    len(chunks),
		TextPreview:    textPreview,
		UploadedAt:     time.Now(),
	}
}
