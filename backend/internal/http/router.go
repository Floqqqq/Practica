package httpserver

import (
	"net/http"

	"github.com/Floqqqq/Practica/backend/internal/elastic"
	"github.com/Floqqqq/Practica/backend/internal/http/handlers"
	"github.com/Floqqqq/Practica/backend/internal/services"
	"github.com/Floqqqq/Practica/backend/pkg/response"
)

func NewRouter(
	uploadDir string,
	elasticClient *elastic.Client,
	cacheService *services.CacheService,
) http.Handler {
	mux := http.NewServeMux()

	documentService := services.NewDocumentService(uploadDir, elasticClient)
	documentHandler := handlers.NewDocumentHandler(documentService)

	searchService := services.NewSearchService(elasticClient, cacheService)
	searchHandler := handlers.NewSearchHandler(searchService)

	docsHandler := handlers.NewDocsHandler(
		"docs/swagger.yaml",
		"backend/docs/swagger.yaml",
	)

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
	mux.HandleFunc("/api/v1/search", searchHandler.Search)

	mux.HandleFunc("/docs", docsHandler.SwaggerUI)
	mux.HandleFunc("/docs/", docsHandler.SwaggerUI)
	mux.HandleFunc("/docs/swagger.yaml", docsHandler.SwaggerYAML)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response.JSONError(w, http.StatusNotFound, "not_found", "resource not found")
	})

	return mux
}
