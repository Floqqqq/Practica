package handlers

import (
	"net/http"

	"github.com/Floqqqq/Practica/backend/pkg/response"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	response.JSONError(w, http.StatusNotFound, "not_found", "resource not found")
}
