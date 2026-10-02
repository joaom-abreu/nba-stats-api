package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type requestIDKey struct{}

type paginationMeta struct {
	Limit    int `json:"limit"`
	Offset   int `json:"offset"`
	Returned int `json:"returned"`
}

type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func requestID(r *http.Request) string {
	id, _ := r.Context().Value(requestIDKey{}).(string)
	return id
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}

func writeError(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	code string,
	message string,
) {
	response := struct {
		Error apiError `json:"error"`
	}{
		Error: apiError{
			Code:      code,
			Message:   message,
			RequestID: requestID(r),
		},
	}

	writeJSON(w, statusCode, response)
}
