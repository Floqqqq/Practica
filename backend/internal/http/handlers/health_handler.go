package handlers

import (
	"net/http"

	"github.com/Floqqqq/Practica/backend/pkg/response"
)

func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
