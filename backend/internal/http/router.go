package httpserver

import (
	"net/http"

	"github.com/Floqqqq/Practica/backend/internal/http/handlers"
	"github.com/Floqqqq/Practica/backend/internal/services"
	"github.com/Floqqqq/Practica/backend/pkg/response"
)

func NewRouter(uploadDir string) http.Handler {
	mux := http.NewServeMux()

	documentService := services.NewDocumentService(uploadDir)
	documentHandler := handlers.NewDocumentHandler(documentService)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}

		response.JSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	mux.HandleFunc("/api/v1/documents/upload", documentHandler.UploadDocument)

	return mux
}
