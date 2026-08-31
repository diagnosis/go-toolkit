// Package responder writes standard JSON success and error envelopes to
// http.ResponseWriter, tagging every response with a correlation ID.
package responder

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/diagnosis/go-toolkit/v3/apperr"
	"github.com/diagnosis/go-toolkit/v3/logger"
)

// SuccessResponse is the JSON envelope written by JSON for successful requests.
type SuccessResponse struct {
	Data          any       `json:"data"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	Message       string    `json:"message,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// ErrorResponse is the JSON envelope written by Error for failed requests.
type ErrorResponse struct {
	Error struct {
		Status        apperr.Status     `json:"status"`
		Message       string            `json:"message"`
		Details       map[string]string `json:"details,omitempty"`
		CorrelationID string            `json:"correlation_id,omitempty"`
		Timestamp     time.Time         `json:"timestamp"`
	} `json:"error"`
}

// JSON writes data wrapped in a SuccessResponse envelope with the given HTTP
// status code, setting the Content-Type and X-Correlation-ID headers.
func JSON(w http.ResponseWriter, status int, data any, correlationID string) {
	response := &SuccessResponse{
		Data:          data,
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		logger.Get().Error("failed to encode JSON response", "error", err, "correlation_id", correlationID)
	}
}

// Error writes err as an ErrorResponse envelope. If err is an *apperr.StatusErr
// (directly or wrapped), its HTTP status, application status code, message,
// and details are used; any other error is reported as a generic 500
// internal server error.
func Error(w http.ResponseWriter, err error, correlationID string) {
	statusErr, ok := apperr.AsStatusErr(err)
	if !ok {
		statusErr = apperr.DefaultError("Internal server error", "unknown error occurred", err)
	}
	response := ErrorResponse{}
	response.Error.Status = statusErr.Status
	response.Error.Message = statusErr.Message
	response.Error.Details = statusErr.Details
	response.Error.Timestamp = time.Now().UTC()
	response.Error.CorrelationID = correlationID

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)
	w.WriteHeader(statusErr.HTTPStatus)

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		logger.Get().Error("failed to encode error response", "error", encodeErr, "correlation_id", correlationID)
	}
}
