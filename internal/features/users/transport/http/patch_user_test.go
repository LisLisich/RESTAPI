package users_transport_http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestPatchUserReturnsOKOnValidRequest(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newPatchUserRequest(`{
		"full_name": "Ivan Updated",
		"phone_number": null
	}`)
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.PatchUser(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.patchUserCalled {
		t.Fatal("expected service PatchUser to be called")
	}
	if service.patchedUserID != 15 {
		t.Fatalf("expected user id %d, got %d", 15, service.patchedUserID)
	}
	if !service.patchedPatch.FullName.Set || service.patchedPatch.FullName.Value == nil {
		t.Fatal("expected full name patch to be set")
	}
	if *service.patchedPatch.FullName.Value != "Ivan Updated" {
		t.Fatalf("expected full name %q, got %q", "Ivan Updated", *service.patchedPatch.FullName.Value)
	}
	if !service.patchedPatch.PhoneNumber.Set {
		t.Fatal("expected phone number patch to be set")
	}
	if service.patchedPatch.PhoneNumber.Value != nil {
		t.Fatal("expected phone number patch value to be nil")
	}
}

func TestPatchUserWritesStatusHeaderOnceOnValidRequest(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newPatchUserRequest(`{
		"full_name": "Ivan Updated"
	}`)
	request.SetPathValue("id", "15")
	response := newUserWriteHeaderCountingResponseWriter()

	handler.PatchUser(response, request)

	if response.writeHeaderCalls != 1 {
		t.Fatalf("expected WriteHeader to be called once, got %d", response.writeHeaderCalls)
	}
}

func TestPatchUserReturnsBadRequestOnInvalidID(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newPatchUserRequest(`{
		"full_name": "Ivan Updated"
	}`)
	request.SetPathValue("id", "not-integer")
	response := httptest.NewRecorder()

	handler.PatchUser(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.patchUserCalled {
		t.Fatal("expected service PatchUser not to be called")
	}
}

func TestPatchUserReturnsBadRequestOnInvalidRequest(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newPatchUserRequest(`{
		"full_name": null
	}`)
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.PatchUser(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.patchUserCalled {
		t.Fatal("expected service PatchUser not to be called")
	}
}

func TestPatchUserReturnsNotFoundWhenServiceReturnsNotFound(t *testing.T) {
	service := &fakeUsersService{
		patchUserErr: fmt.Errorf("patch user: %w", core_errors.ErrNotFound),
	}
	handler := NewUsersHTTPHandler(service)

	request := newPatchUserRequest(`{
		"full_name": "Ivan Updated"
	}`)
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.PatchUser(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
	if !service.patchUserCalled {
		t.Fatal("expected service PatchUser to be called")
	}
}

func TestPatchUserReturnsConflictWhenServiceReturnsConflict(t *testing.T) {
	service := &fakeUsersService{
		patchUserErr: fmt.Errorf("patch user: %w", core_errors.ErrConflict),
	}
	handler := NewUsersHTTPHandler(service)

	request := newPatchUserRequest(`{
		"full_name": "Ivan Updated"
	}`)
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.PatchUser(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, response.Code)
	}
	if !service.patchUserCalled {
		t.Fatal("expected service PatchUser to be called")
	}
}

func newPatchUserRequest(body string) *http.Request {
	return newUserRequestWithLogger(
		http.MethodPatch,
		"/users/15",
		strings.NewReader(body),
	)
}

type userWriteHeaderCountingResponseWriter struct {
	*httptest.ResponseRecorder
	writeHeaderCalls int
}

func newUserWriteHeaderCountingResponseWriter() *userWriteHeaderCountingResponseWriter {
	return &userWriteHeaderCountingResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

func (rw *userWriteHeaderCountingResponseWriter) WriteHeader(statusCode int) {
	rw.writeHeaderCalls++
	rw.ResponseRecorder.WriteHeader(statusCode)
}
