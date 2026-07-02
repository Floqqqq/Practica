package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Floqqqq/Practica/backend/internal/elastic"
	"github.com/Floqqqq/Practica/backend/internal/models"
	"github.com/google/uuid"
)

const MaxFileSize int64 = 20 * 1024 * 1024

var ErrInvalidFile = errors.New("invalid file")

type DocumentService struct {
	uploadDir     string
	parser        *ParserService
	chunker       *ChunkService
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
		parser:        NewParserService(),
		chunker:       NewChunkService(),
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

	if err := s.prepareUploadDirs(); err != nil {
		return nil, fmt.Errorf("prepare upload dirs: %w", err)
	}

	originalsDir := filepath.Join(s.uploadDir, "originals")
	extractedDir := filepath.Join(s.uploadDir, "extracted")
	chunksDir := filepath.Join(s.uploadDir, "chunks")

	savedFileName := fmt.Sprintf("%s_%s", documentID, safeFileName)
	savedPath := filepath.Join(originalsDir, savedFileName)

	contentHash, writtenBytes, err := saveFileWithHash(file, savedPath)
	if err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	if writtenBytes > MaxFileSize {
		_ = os.Remove(savedPath)
		return nil, fmt.Errorf("%w: file size exceeds 20 MB", ErrInvalidFile)
	}

	if duplicatedDocument, err := s.findMetadataByHash(contentHash); err == nil {
		_ = os.Remove(savedPath)

		return &UploadResult{
			ID:                 documentID,
			OriginalDocumentID: duplicatedDocument.ID,
			FileName:           safeFileName,
			Size:               writtenBytes,
			Status:             "duplicate",
			Message:            "file already uploaded; extraction, chunking and indexing skipped",
			ContentHash:        contentHash,
			PagesCount:         duplicatedDocument.PagesCount,
			ExtractedChars:     duplicatedDocument.ExtractedChars,
			ChunksCount:        duplicatedDocument.ChunksCount,
			TextPreview:        duplicatedDocument.TextPreview,
			Duplicate:          true,
		}, nil
	}

	select {
	case <-ctx.Done():
		_ = os.Remove(savedPath)
		return nil, ctx.Err()
	default:
	}

	parsedDocument, err := s.parser.ExtractText(savedPath)
	if err != nil {
		_ = os.Remove(savedPath)
		return nil, fmt.Errorf("%w: failed to extract text: %v", ErrInvalidFile, err)
	}

	textPath := filepath.Join(extractedDir, documentID+".txt")

	if err := os.WriteFile(textPath, []byte(parsedDocument.Text), 0644); err != nil {
		_ = os.Remove(savedPath)
		return nil, fmt.Errorf("save extracted text: %w", err)
	}

	chunks, err := s.chunker.SplitDocumentIntoChunks(documentID, safeFileName, parsedDocument)
	if err != nil {
		_ = os.Remove(savedPath)
		_ = os.Remove(textPath)
		return nil, fmt.Errorf("%w: failed to split text into chunks: %v", ErrInvalidFile, err)
	}

	chunksPath := filepath.Join(chunksDir, documentID+".json")

	if err := saveChunks(chunksPath, chunks); err != nil {
		_ = os.Remove(savedPath)
		_ = os.Remove(textPath)
		return nil, fmt.Errorf("save chunks: %w", err)
	}

	documentStatus := "chunked"
	documentMessage := "file uploaded, text extracted and split into chunks successfully"

	if s.elasticClient != nil {
		if err := s.elasticClient.IndexChunks(ctx, chunks); err != nil {
			_ = os.Remove(savedPath)
			_ = os.Remove(textPath)
			_ = os.Remove(chunksPath)
			return nil, fmt.Errorf("index chunks in elasticsearch: %w", err)
		}

		documentStatus = "indexed"
		documentMessage = "file uploaded, text extracted, split into chunks and indexed successfully"
	}

	textPreview := makeTextPreview(parsedDocument.Text, 300)

	metadata := models.DocumentMetadata{
		ID:             documentID,
		FileName:       safeFileName,
		FilePath:       savedPath,
		TextPath:       textPath,
		ChunksPath:     chunksPath,
		Size:           writtenBytes,
		Status:         documentStatus,
		ContentHash:    contentHash,
		PagesCount:     parsedDocument.PagesCount,
		ExtractedChars: parsedDocument.CharsCount,
		ChunksCount:    len(chunks),
		TextPreview:    textPreview,
		UploadedAt:     time.Now(),
	}

	if err := s.saveMetadata(metadata); err != nil {
		_ = os.Remove(savedPath)
		_ = os.Remove(textPath)
		_ = os.Remove(chunksPath)
		return nil, fmt.Errorf("save document metadata: %w", err)
	}

	return &UploadResult{
		ID:             documentID,
		FileName:       safeFileName,
		FilePath:       savedPath,
		TextPath:       textPath,
		ChunksPath:     chunksPath,
		Size:           writtenBytes,
		Status:         documentStatus,
		Message:        documentMessage,
		ContentHash:    contentHash,
		PagesCount:     parsedDocument.PagesCount,
		ExtractedChars: parsedDocument.CharsCount,
		ChunksCount:    len(chunks),
		TextPreview:    textPreview,
		Duplicate:      false,
	}, nil
}

func validateFile(file multipart.File, fileHeader *multipart.FileHeader) error {
	if fileHeader == nil {
		return fmt.Errorf("%w: file is required", ErrInvalidFile)
	}

	if strings.TrimSpace(fileHeader.Filename) == "" {
		return fmt.Errorf("%w: file name is empty", ErrInvalidFile)
	}

	if fileHeader.Size <= 0 {
		return fmt.Errorf("%w: file is empty", ErrInvalidFile)
	}

	if fileHeader.Size > MaxFileSize {
		return fmt.Errorf("%w: file size exceeds 20 MB", ErrInvalidFile)
	}

	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if extension != ".pdf" && extension != ".docx" {
		return fmt.Errorf("%w: only PDF and DOCX files are allowed", ErrInvalidFile)
	}

	buffer := make([]byte, 512)

	readBytes, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: failed to read file", ErrInvalidFile)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("%w: failed to reset file cursor", ErrInvalidFile)
	}

	contentType := http.DetectContentType(buffer[:readBytes])

	switch extension {
	case ".pdf":
		if !isPDF(buffer[:readBytes], contentType) {
			return fmt.Errorf("%w: invalid PDF file", ErrInvalidFile)
		}

	case ".docx":
		if !isDOCX(buffer[:readBytes], contentType) {
			return fmt.Errorf("%w: invalid DOCX file", ErrInvalidFile)
		}
	}

	return nil
}

func isPDF(fileHeader []byte, contentType string) bool {
	return strings.HasPrefix(string(fileHeader), "%PDF") ||
		contentType == "application/pdf"
}

func isDOCX(fileHeader []byte, contentType string) bool {
	hasZipSignature := len(fileHeader) >= 4 &&
		fileHeader[0] == 'P' &&
		fileHeader[1] == 'K' &&
		(fileHeader[2] == 3 || fileHeader[2] == 5 || fileHeader[2] == 7) &&
		(fileHeader[3] == 4 || fileHeader[3] == 6 || fileHeader[3] == 8)

	return hasZipSignature ||
		contentType == "application/zip" ||
		contentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
}

func saveFileWithHash(file multipart.File, destinationPath string) (string, int64, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", 0, fmt.Errorf("reset file cursor: %w", err)
	}

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return "", 0, fmt.Errorf("create destination file: %w", err)
	}
	defer destinationFile.Close()

	hasher := sha256.New()

	writer := io.MultiWriter(destinationFile, hasher)

	limitedReader := &io.LimitedReader{
		R: file,
		N: MaxFileSize + 1,
	}

	writtenBytes, err := io.Copy(writer, limitedReader)
	if err != nil {
		return "", 0, fmt.Errorf("copy file: %w", err)
	}

	contentHash := hex.EncodeToString(hasher.Sum(nil))

	return contentHash, writtenBytes, nil
}

func saveChunks(chunksPath string, chunks []models.Chunk) error {
	chunksBytes, err := json.MarshalIndent(chunks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(chunksPath, chunksBytes, 0644)
}

func (s *DocumentService) prepareUploadDirs() error {
	dirs := []string{
		s.uploadDir,
		filepath.Join(s.uploadDir, "originals"),
		filepath.Join(s.uploadDir, "extracted"),
		filepath.Join(s.uploadDir, "chunks"),
		filepath.Join(s.uploadDir, "metadata"),
		filepath.Join(s.uploadDir, "metadata", "hash_index"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

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
