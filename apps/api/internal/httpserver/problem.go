package httpserver

import (
	"encoding/json"
	"net/http"
)

// problem is an RFC 9457 "problem details" document. The type member is a URN
// of the form urn:brynn:problem:<code>. Codes are part of the API contract and
// must not change once released.
type problem struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// writeProblem sends an RFC 9457 error response. detail must never contain
// internal state, stack traces, SQL, or user data.
func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem{
		Type:      "urn:brynn:problem:" + code,
		Title:     title,
		Status:    status,
		Detail:    detail,
		RequestID: RequestIDFrom(r.Context()),
	})
}

// writeJSON sends v as a JSON response with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
