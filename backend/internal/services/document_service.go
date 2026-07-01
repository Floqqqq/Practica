package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const MaxFileSize int64 = 20 * 1024 * 1024

var ErrInvalidFile = errors.New("invalid file")

type DocumentService struct {
	uploadDir string
}

type UploadResult struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Size     int64  `json:"size"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

func NewDocumentService(uploadDir string) *DocumentService {
	return &DocumentService{
		uploadDir: uploadDir,
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

	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	safeFileName := filepath.Base(fileHeader.Filename)
	savedFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeFileName)
	savedPath := filepath.Join(s.uploadDir, savedFileName)

	destinationFile, err := os.Create(savedPath)
	if err != nil {
		return nil, fmt.Errorf("create destination file: %w", err)
	}
	defer destinationFile.Close()

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek file: %w", err)
	}

	limitedReader := &io.LimitedReader{
		R: file,
		N: MaxFileSize + 1,
	}

	writtenBytes, err := io.Copy(destinationFile, limitedReader)
	if err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	if writtenBytes > MaxFileSize {
		_ = os.Remove(savedPath)
		return nil, fmt.Errorf("%w: file size exceeds 20 MB", ErrInvalidFile)
	}

	select {
	case <-ctx.Done():
		_ = os.Remove(savedPath)
		return nil, ctx.Err()
	default:
	}

	return &UploadResult{
		FileName: safeFileName,
		FilePath: savedPath,
		Size:     writtenBytes,
		Status:   "uploaded",
		Message:  "file uploaded successfully",
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
