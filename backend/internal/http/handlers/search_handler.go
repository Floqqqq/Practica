package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Floqqqq/Practica/backend/internal/services"
	"github.com/Floqqqq/Practica/backend/pkg/response"
)

type SearchHandler struct {
	searchService *services.SearchService
}

func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	query := r.URL.Query().Get("q")
	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 10)

	searchResult, err := h.searchService.Search(r.Context(), query, page, limit)
	if err != nil {
		if errors.Is(err, services.ErrInvalidSearchQuery) {
			response.JSONError(w, http.StatusBadRequest, "validation_error", "query parameter q is required")
			return
		}

		response.JSONError(w, http.StatusInternalServerError, "internal_error", "failed to search documents")
		return
	}

	response.JSON(w, http.StatusOK, searchResult)
}

func parsePositiveInt(value string, defaultValue int) int {
	parsedValue, err := strconv.Atoi(value)
	if err != nil || parsedValue < 1 {
		return defaultValue
	}

	return parsedValue
}
