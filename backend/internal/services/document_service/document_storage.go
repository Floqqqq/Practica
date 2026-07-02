package document_service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/Floqqqq/Practica/backend/internal/models"
)

type documentPaths struct {
	OriginalFile  string
	ExtractedText string
	Chunks        string
}

func newDocumentPaths(uploadDir string, documentID string, safeFileName string) documentPaths {
	return documentPaths{
		OriginalFile:  filepath.Join(uploadDir, "originals", fmt.Sprintf("%s_%s", documentID, safeFileName)),
		ExtractedText: filepath.Join(uploadDir, "extracted", documentID+".txt"),
		Chunks:        filepath.Join(uploadDir, "chunks", documentID+".json"),
	}
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

func writeTextFile(path string, text string) error {
	return os.WriteFile(path, []byte(text), 0644)
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

func removeFiles(paths ...string) {
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}

		_ = os.Remove(path)
	}
}
