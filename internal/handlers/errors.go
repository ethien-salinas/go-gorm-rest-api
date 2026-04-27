package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorResponse represents the JSON body returned on any HTTP error.
type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
		slog.Default().Error("handler: failed to encode error response", "error", err)
	}
}

// writeValidationErrors writes a 400 response with {"errors": [...]} for multi-rule violations.
func writeValidationErrors(w http.ResponseWriter, violations []string) {
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(map[string]any{"errors": violations}); err != nil {
		slog.Default().Error("handler: failed to encode validation error response", "error", err)
	}
}
