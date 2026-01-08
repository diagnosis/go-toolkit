package responder

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/diagnosis/go-toolkit/errors"
)

type SuccessResponse struct {
	Data          any       `json:"data"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	Message       string    `json:"message,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

type ErrorResponse struct {
	Error struct {
		Status        errors.Status `json:"status"`
		Message       string        `json:"message"`
		CorrelationID string        `json:"correlation_id,omitempty"`
		Timestamp     time.Time     `json:"timestamp"`
	} `json:"error"`
}

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
		log.Printf("Failed to encode JSON response: %v", err)
	}
}
func Error(w http.ResponseWriter, err error, correlationID string) {
	statusErr, ok := errors.AsStatusErr(err)
	if !ok {
		statusErr = errors.DefaultError("Internal server error", "unknown error occurred", err)
	}
	response := ErrorResponse{}
	response.Error.Status = statusErr.Status
	response.Error.Message = statusErr.Message
	response.Error.Timestamp = time.Now().UTC()
	response.Error.CorrelationID = correlationID

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)
	w.WriteHeader(statusErr.HTTPStatus)

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		log.Printf("Failed to encode error response: %v", encodeErr)
	}

}
