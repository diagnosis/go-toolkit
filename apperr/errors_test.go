package apperr

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

func TestStatus_Code_Exhaustive(t *testing.T){
	for s := CodeBadRequest; s <= CodeEmailNotVerified; s++ {
		if s.Code() == "unknown"{
			t.Errorf("Status %d has no code name in codeNames", s)
		}
	}
}

func TestHelpers(t *testing.T) {
	tests := []struct {
		name       string
		err        *StatusErr
		wantStatus Status
		wantCode string
		wantHTTP   int
	}{
		{"NotFound", NotFound("test", "test"), CodeNotFound,CodeNotFound.Code() ,404},
		{"BadRequest", BadRequest("test", "test"), CodeBadRequest, CodeBadRequest.Code(), 400},
		{"Internal", Internal("test", "test"), CodeInternalError, CodeInternalError.Code(), 500},
		{"Unauthorized", Unauthorized("test", "test"), CodeUnauthorized, CodeUnauthorized.Code(), 401},
		{"Forbidden", Forbidden("test", "test"), CodeForbidden, CodeForbidden.Code(), 403},
		{"Conflict", Conflict("test", "test"), CodeConflict, CodeConflict.Code(),409},
		{"Validation", Validation("test", "test"), CodeValidationError, CodeValidationError.Code(), 400},
		{"Database", Database("test", "test"), CodeDatabaseError, CodeDatabaseError.Code(),500},
		{"TokenError", TokenError("test", "test"), CodeTokenErr, CodeTokenErr.Code(),401},
		{"InvalidCredentials", InvalidCredentials("test", "test"), CodeInvalidCredentials, CodeInvalidCredentials.Code(),401},
		{"AccountInactive", AccountInactive("test", "test"), CodeAccountInactive, CodeAccountInactive.Code(), 403},
		{"EmailExists", EmailExists("test", "test"), CodeEmailExists, CodeEmailExists.Code(),409},
		{"DefaultError", DefaultError("test", "test"), CodeDefaultError, CodeDefaultError.Code(),500},
		{ "EmailNotVerified", EmailNotVerified("test", "test"),  CodeEmailNotVerified, CodeEmailNotVerified.Code(),403},
		{"UnprocessableContent", UnprocessableContent("test", "test"), CodeUnprocessableContent, CodeUnprocessableContent.Code(), 422},
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
	converted, ok := AsStatusErr(statusErr)
	if !ok {
		t.Error("AsStatusErr should succeed for StatusErr")
	}
	if converted.Status != CodeBadRequest {
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
