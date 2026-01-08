package responder

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/diagnosis/go-toolkit/errors"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()

	testdata := map[string]string{"message": "success"}
	JSON(w, 200, testdata, "test-correlation-123")

	if w.Code != 200 {
		t.Errorf("expected code 200 got: %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected header application/json got: %s", w.Header().Get("Content-Type"))
	}

	if w.Header().Get("X-Correlation-ID") != "test-correlation-123" {
		t.Errorf("expected header test-correlation-123 got: %s", w.Header().Get("X-Correlation-ID"))
	}

	var res SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	if err != nil {
		t.Fatalf("failed to unmarshal response body")
	}
	data, ok := res.Data.(map[string]any)
	if !ok {
		t.Errorf("failed to get map of data")
	}
	if data["message"] != "success" {

		t.Errorf("expected response message success got: %v", data["message"])
	}
	if res.CorrelationID != "test-correlation-123" {
		t.Errorf("expected correlation id is test-correlation-123 got: %s", res.CorrelationID)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	err := errors.NotFound("not found", "mahmut is not found in DB")

	Error(w, err, "correlation-id-123")

	code := w.Code
	headerContentType := w.Header().Get("Content-Type")
	headerCorrelationID := w.Header().Get("X-Correlation-ID")

	if code != 404 {
		t.Errorf("expected httpStatus 404 got: %d", code)
	}
	if headerCorrelationID != "correlation-id-123" {
		t.Errorf("expected correlation header correlation-id123, got : %s", headerCorrelationID)
	}
	if headerContentType != "application/json" {
		t.Errorf("expected content-type is application/json, got: %s", headerContentType)
	}
	var errRes ErrorResponse
	e := json.Unmarshal(w.Body.Bytes(), &errRes)
	if e != nil {
		t.Fatalf("failed to unmarshal errorReponse")
	}
	message := errRes.Error.Message
	status := errRes.Error.Status
	correlationIDInResoponse := errRes.Error.CorrelationID
	if status != errors.CodeNotFound {
		t.Errorf("expected %v got %v", errors.CodeNotFound, status)
	}
	if message != "not found" {
		t.Errorf("expected %s, got %s", "not found", message)
	}
	if correlationIDInResoponse != "correlation-id-123" {
		t.Errorf("expected %s, got %s", "correlation-id-123", correlationIDInResoponse)
	}
}

func TestError_RegularError(t *testing.T) {
	w := httptest.NewRecorder()

	regularErr := fmt.Errorf("unidentified error")
	Error(w, regularErr, "test-123")

	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var errRes ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errRes)
	if err != nil {
		t.Fatal("failed to unmarshal error response")
	}
	if errRes.Error.Status != errors.CodeDefaultError {
		t.Errorf("expected CodeDefaultError for regular error")
	}
}
