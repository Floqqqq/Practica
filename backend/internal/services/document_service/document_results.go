package document_service

import "github.com/Floqqqq/Practica/backend/internal/models"

func duplicateUploadResult(
	documentID string,
	safeFileName string,
	writtenBytes int64,
	contentHash string,
	duplicatedDocument *models.DocumentMetadata,
) *UploadResult {
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
	}
}

func successfulUploadResult(metadata models.DocumentMetadata, message string) *UploadResult {
	return &UploadResult{
		ID:             metadata.ID,
		FileName:       metadata.FileName,
		FilePath:       metadata.FilePath,
		TextPath:       metadata.TextPath,
		ChunksPath:     metadata.ChunksPath,
		Size:           metadata.Size,
		Status:         metadata.Status,
		Message:        message,
		ContentHash:    metadata.ContentHash,
		PagesCount:     metadata.PagesCount,
		ExtractedChars: metadata.ExtractedChars,
		ChunksCount:    metadata.ChunksCount,
		TextPreview:    metadata.TextPreview,
		Duplicate:      false,
	}
}
