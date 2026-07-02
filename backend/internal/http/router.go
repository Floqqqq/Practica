package httpserver

import (
	"net/http"

	"github.com/Floqqqq/Practica/backend/internal/elastic"
	"github.com/Floqqqq/Practica/backend/internal/http/handlers"
	"github.com/Floqqqq/Practica/backend/internal/services"
	document_service "github.com/Floqqqq/Practica/backend/internal/services/document_service"
)

func NewRouter(
	uploadDir string,
	elasticClient *elastic.Client,
	cacheService *services.CacheService,
) http.Handler {
	mux := http.NewServeMux()

	documentHandler := newDocumentHandler(uploadDir, elasticClient)
	searchHandler := newSearchHandler(elasticClient, cacheService)
	docsHandler := newDocsHandler()

	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/api/v1/documents/upload", documentHandler.UploadDocument)
	mux.HandleFunc("/api/v1/search", searchHandler.Search)

	mux.HandleFunc("/docs", docsHandler.SwaggerUI)
	mux.HandleFunc("/docs/", docsHandler.SwaggerUI)
	mux.HandleFunc("/docs/swagger.yaml", docsHandler.SwaggerYAML)

	mux.HandleFunc("/", handlers.NotFound)

	return mux
}

func newDocumentHandler(
	uploadDir string,
	elasticClient *elastic.Client,
) *handlers.DocumentHandler {
	documentService := document_service.NewDocumentService(uploadDir, elasticClient)

	return handlers.NewDocumentHandler(documentService)
}

func newSearchHandler(
	elasticClient *elastic.Client,
	cacheService *services.CacheService,
) *handlers.SearchHandler {
	searchService := services.NewSearchService(elasticClient, cacheService)

	return handlers.NewSearchHandler(searchService)
}

func newDocsHandler() *handlers.DocsHandler {
	return handlers.NewDocsHandler(
		"docs/swagger.yaml",
		"backend/docs/swagger.yaml",
	)
}
