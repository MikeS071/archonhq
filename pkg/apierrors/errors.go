package apierrors

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse follows docs/ERROR_MODEL.md.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code          string         `json:"code"`
	Message       string         `json:"message"`
	Details       map[string]any `json:"details,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
}

func Write(w http.ResponseWriter, status int, code, message, correlationID string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorBody{
			Code:          code,
			Message:       message,
			Details:       details,
			CorrelationID: correlationID,
		},
	})
}
