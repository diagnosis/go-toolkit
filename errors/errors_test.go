package errors

import (
	"errors"
	"net/http"
	"testing"
)

// Testing New constructor
func TestNew(t *testing.T) {
	err := New(CodeNotFound,
		"Not Found",
		"Failed to Fetch Data",
		http.StatusNotFound,
		nil,
	)

	if err.Status != CodeNotFound {
		t.Errorf("expected codeNotFound, got %v", err.Status)
	}
	if err.Message != "Not Found" {
		t.Errorf("expected Not Found, got %v", err.Message)
	}
	if err.InternalMessage != "Failed to Fetch Data" {
		t.Errorf("expected Failed to Fetch Data, got %v", err.InternalMessage)
	}
	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected statusNotfound, got %v", err.HTTPStatus)
	}
}

func TestErrorWrapping(t *testing.T) {
	newErr := errors.New("not trading rule broken")
	myErr := Internal("brain needs to obey", "stop messing around and focus go", newErr)
	if !errors.Is(myErr, newErr) {
		t.Error("failed to unwrapped root error")
	}

}

func TestHelpers(t *testing.T) {
	tests := []struct {
		name       string
		err        *StatusErr
		wantStatus Status
		wantHTTP   int
	}{
		{"NotFound", NotFound("test", "test"), CodeNotFound, 404},
		{"BadRequest", BadRequest("test", "test"), CodeBadRequest, 400},
		{"Internal", Internal("test", "test"), CodeInternalError, 500},
		{"Unauthorized", Unauthorized("test", "test"), CodeUnauthorized, 401},
		{"Forbidden", Forbidden("test", "test"), CodeForbidden, 403},
		{"Conflict", Conflict("test", "test"), CodeConflict, 409},
		{"Validation", Validation("test", "test"), CodeValidationError, 400},
		{"Database", Database("test", "test"), CodeDatabaseError, 500},
		{"TokenError", TokenError("test", "test"), CodeTokenErr, 401},
		{"InvalidCredentials", InvalidCredentials("test", "test"), CodeInvalidCredentials, 401},
		{"AccountInactive", AccountInactive("test", "test"), CodeAccountInactive, 403},
		{"EmailExists", EmailExists("test", "test"), CodeEmailExists, 409},
		{"DefaultError", DefaultError("test", "test"), CodeDefaultError, 500},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Status != tt.wantStatus {
				t.Errorf("%s: wrong status, got %v want %v", tt.name, tt.err.Status, tt.wantStatus)
			}
			if tt.err.HTTPStatus != tt.wantHTTP {
				t.Errorf("%s: wrong httpStatus, got %v want %v", tt.name, tt.err.HTTPStatus, tt.wantHTTP)
			}
		})
	}
}

func TestAsStatusErr(t *testing.T) {
	statusErr := BadRequest("bad", "bad")
	coverted, ok := AsStatusErr(statusErr)
	if !ok {
		t.Error("AsStatusErr should succeed for StatusErr")
	}
	if coverted.Status != CodeBadRequest {
		t.Error("Converted error should preserve status")
	}

	regularError := errors.New("i am regular")
	_, ok = AsStatusErr(regularError)

	if ok {
		t.Error("AsStatusErr should fail for regular error")
	}

	_, ok = AsStatusErr(nil)
	if ok {
		t.Error("AsStatusErr should fail for nil")
	}
}

func TestIsStatusErr(t *testing.T) {
	if !IsStatusErr(BadRequest("bad", "bad")) {
		t.Error("IsStatusErr should return true for StatusErr")
	}

	if IsStatusErr(errors.New("mahmut")) {
		t.Error("IsStatusErr should return false for regularErr")
	}

	if IsStatusErr(nil) {
		t.Error("IsStatusErr should return false for nil")
	}
}
