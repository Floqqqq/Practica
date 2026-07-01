package handlers

import (
	"errors"
	"net/http"

	"github.com/Floqqqq/Practica/backend/internal/services"
	"github.com/Floqqqq/Practica/backend/pkg/response"
)

type DocumentHandler struct {
	documentService *services.DocumentService
}

func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, services.MaxFileSize+1024*1024)

	if err := r.ParseMultipartForm(services.MaxFileSize + 1024*1024); err != nil {
		response.JSONError(w, http.StatusBadRequest, "invalid_request", "failed to parse multipart form")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		response.JSONError(w, http.StatusBadRequest, "file_required", "file field is required")
		return
	}
	defer file.Close()

	uploadResult, err := h.documentService.Upload(r.Context(), file, fileHeader)
	if err != nil {
		if errors.Is(err, services.ErrInvalidFile) {
			response.JSONError(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}

		response.JSONError(w, http.StatusInternalServerError, "internal_error", "failed to upload document")
		return
	}

	response.JSON(w, http.StatusOK, uploadResult)
}
