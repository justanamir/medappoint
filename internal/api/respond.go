package api

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

// JSON writes a JSON response with status code.
func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ErrorJSON writes a standardized JSON error.
func ErrorJSON(w http.ResponseWriter, status int, msg string, details interface{}) {
	JSON(w, status, ErrorResponse{Error: msg, Details: details})
}
