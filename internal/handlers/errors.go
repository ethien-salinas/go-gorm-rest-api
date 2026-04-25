package handlers

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents the JSON body returned on any HTTP error.
type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
