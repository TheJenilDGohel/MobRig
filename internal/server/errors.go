package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ProblemDetail implements RFC 7807 problem details for HTTP APIs.
type ProblemDetail struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

// writeJSON encodes data as application/json with the given HTTP status.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to encode response: %v"}`, err), http.StatusInternalServerError)
	}
}

// writeProblem writes an RFC 7807 problem detail response.
func writeProblem(w http.ResponseWriter, status int, title, detail, errType string) {
	if errType == "" {
		errType = "about:blank"
	}
	prob := ProblemDetail{
		Type:   errType,
		Title:  title,
		Status: status,
		Detail: detail,
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(prob)
}

// writeBadRequest writes an RFC 7807 400 Bad Request error.
func writeBadRequest(w http.ResponseWriter, detail string) {
	writeProblem(w, http.StatusBadRequest, "Bad Request", detail, "https://mobrig.dev/errors/bad-request")
}

// writeNotFound writes an RFC 7807 404 Not Found error.
func writeNotFound(w http.ResponseWriter, detail string) {
	writeProblem(w, http.StatusNotFound, "Not Found", detail, "https://mobrig.dev/errors/not-found")
}

// writeInternalError writes an RFC 7807 500 Internal Server Error.
func writeInternalError(w http.ResponseWriter, detail string) {
	writeProblem(w, http.StatusInternalServerError, "Internal Server Error", detail, "https://mobrig.dev/errors/internal-error")
}
