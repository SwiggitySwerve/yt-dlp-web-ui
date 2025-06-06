package rest

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/internal"
)

// Existing ContainerArgs struct
type ContainerArgs struct {
	DB  *sql.DB
	MDB *internal.MemoryDB
	MQ  *internal.MessageQueue
}

// ErrorResponse defines the standard JSON structure for error responses.
type ErrorResponse struct {
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"` // Optional technical details
	Code    int    `json:"code,omitempty"`   // Optional internal application-specific error code
}

// RespondWithErrorJSON sends a JSON error response.
// It logs the detailError internally and sends a user-friendly message to the client.
func RespondWithErrorJSON(w http.ResponseWriter, httpStatusCode int, userMessage string, detailError error, internalCode ...int) {
	var detailStr string
	if detailError != nil {
		detailStr = detailError.Error()
		slog.Error("API Error", "status", httpStatusCode, "message", userMessage, "detail", detailStr)
	} else {
		// Log even if detailError is nil, as it indicates an error condition was triggered.
		slog.Warn("API Error (no detailError provided but RespondWithErrorJSON called)", "status", httpStatusCode, "message", userMessage)
	}

	errResp := ErrorResponse{
		Message: userMessage,
		// Decision: Only include detailStr in response if it's explicitly desired for client consumption.
		// For now, let's make it conditional, e.g. based on a debug mode or error type.
		// Simpler: always include it for now as per original plan, can be refined later.
		Detail: detailStr,
	}
	if len(internalCode) > 0 {
		errResp.Code = internalCode[0]
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatusCode)
	if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
		slog.Error("Failed to encode ErrorResponse to JSON", "error", encodeErr, "original_message", userMessage)
		// Fallback if JSON encoding fails
		http.Error(w, `{"message":"Internal server error while encoding error response"}`, http.StatusInternalServerError)
	}
}
