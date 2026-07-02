package handlers

import (
	"errors"
	"net/http"

	document_service "github.com/Floqqqq/Practica/backend/internal/services/document_service"
	"github.com/Floqqqq/Practica/backend/pkg/response"
)

type DocumentHandler struct {
	documentService *document_service.DocumentService
}

func NewDocumentHandler(documentService *document_service.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "validation_error", "file is required")
		return
	}
	defer file.Close()

	result, err := h.documentService.Upload(r.Context(), file, fileHeader)
	if err != nil {
		if errors.Is(err, document_service.ErrInvalidFile) {
			response.JSONError(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}

		response.JSONError(w, http.StatusInternalServerError, "internal_error", "failed to upload document")
		return
	}

	response.JSON(w, http.StatusOK, result)
}
