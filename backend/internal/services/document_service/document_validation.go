package document_service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

const fileSignatureBufferSize = 512

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
	if !isSupportedDocumentExtension(extension) {
		return fmt.Errorf("%w: only PDF and DOCX files are allowed", ErrInvalidFile)
	}

	fileHeaderBytes, contentType, err := detectFileSignature(file)
	if err != nil {
		return err
	}

	if !isValidDocumentSignature(extension, fileHeaderBytes, contentType) {
		return fmt.Errorf("%w: invalid %s file", ErrInvalidFile, strings.TrimPrefix(strings.ToUpper(extension), "."))
	}

	return nil
}

func isSupportedDocumentExtension(extension string) bool {
	return extension == ".pdf" || extension == ".docx"
}

func detectFileSignature(file multipart.File) ([]byte, string, error) {
	buffer := make([]byte, fileSignatureBufferSize)

	readBytes, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, "", fmt.Errorf("%w: failed to read file", ErrInvalidFile)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, "", fmt.Errorf("%w: failed to reset file cursor", ErrInvalidFile)
	}

	fileHeaderBytes := buffer[:readBytes]
	contentType := http.DetectContentType(fileHeaderBytes)

	return fileHeaderBytes, contentType, nil
}

func isValidDocumentSignature(extension string, fileHeader []byte, contentType string) bool {
	switch extension {
	case ".pdf":
		return isPDF(fileHeader, contentType)
	case ".docx":
		return isDOCX(fileHeader, contentType)
	default:
		return false
	}
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
