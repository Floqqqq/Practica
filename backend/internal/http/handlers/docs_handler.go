package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Floqqqq/Practica/backend/pkg/response"
)

type DocsHandler struct {
	swaggerPaths []string
}

func NewDocsHandler(swaggerPaths ...string) *DocsHandler {
	return &DocsHandler{
		swaggerPaths: swaggerPaths,
	}
}

func (h *DocsHandler) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	if !isGetOrHead(r) {
		response.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}

	_, _ = fmt.Fprint(w, `<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <title>Practice API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: "/docs/swagger.yaml",
        dom_id: "#swagger-ui"
      });
    };
  </script>
</body>
</html>`)
}

func (h *DocsHandler) SwaggerYAML(w http.ResponseWriter, r *http.Request) {
	if !isGetOrHead(r) {
		response.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	swaggerFile, err := h.readSwaggerFile()
	if err != nil {
		response.JSONError(w, http.StatusInternalServerError, "internal_error", "failed to read swagger file")
		return
	}

	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(swaggerFile)
}

func (h *DocsHandler) readSwaggerFile() ([]byte, error) {
	for _, path := range h.swaggerPaths {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}

	return nil, os.ErrNotExist
}

func isGetOrHead(r *http.Request) bool {
	return r.Method == http.MethodGet || r.Method == http.MethodHead
}
