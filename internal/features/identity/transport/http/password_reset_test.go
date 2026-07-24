package identity_transport_http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestRequestPasswordResetReturnsAccepted(t *testing.T) {
	service := &fakeIdentityService{}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequestForPath(
		"/auth/password-reset/request",
		`{"email":"user@example.com"}`,
	)
	response := httptest.NewRecorder()

	handler.RequestPasswordReset(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, response.Code)
	}
	if !service.passwordResetRequested || service.passwordResetEmail != "user@example.com" {
		t.Fatal("expected password reset request")
	}
}

func TestConfirmPasswordResetReturnsNoContent(t *testing.T) {
	service := &fakeIdentityService{}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequestForPath(
		"/auth/password-reset/confirm",
		`{"token":"raw-reset-token","new_password":"new-strong-password"}`,
	)
	response := httptest.NewRecorder()

	handler.ConfirmPasswordReset(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if !service.resetPasswordCalled {
		t.Fatal("expected ResetPassword call")
	}
	if service.resetToken != "raw-reset-token" ||
		service.resetPassword != "new-strong-password" {
		t.Fatal("unexpected reset password input")
	}
}

func TestConfirmPasswordResetRejectsConsumedToken(t *testing.T) {
	service := &fakeIdentityService{
		resetPasswordErr: fmt.Errorf("reset: %w", core_errors.ErrInvalidArgument),
	}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequestForPath(
		"/auth/password-reset/confirm",
		`{"token":"consumed-token","new_password":"new-strong-password"}`,
	)
	response := httptest.NewRecorder()

	handler.ConfirmPasswordReset(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
