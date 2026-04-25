// Package handlers provides HTTP handler functions for the REST API.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// apiInfo represents the metadata returned by the home endpoint.
type apiInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Docs    string `json:"docs"`
	Time    string `json:"timestamp"`
}

// HomeHandler returns a JSON response with general information about the API.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	info := apiInfo{
		Name:    "Go GORM REST API",
		Version: "v1.0.0",
		Status:  "ok",
		Message: "Welcome! The API is up and running. Please refer to the documentation for available endpoints.",
		Docs:    "/api/v1",
		Time:    time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(info)
}
