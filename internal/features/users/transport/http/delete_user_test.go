package users_transport_http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestDeleteUserReturnsNoContentOnValidRequest(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newDeleteUserRequest()
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.DeleteUser(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if !service.deleteUserCalled {
		t.Fatal("expected service DeleteUser to be called")
	}
	if service.deletedUserID != 15 {
		t.Fatalf("expected user id %d, got %d", 15, service.deletedUserID)
	}
}

func TestDeleteUserReturnsBadRequestOnInvalidID(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newDeleteUserRequest()
	request.SetPathValue("id", "not-integer")
	response := httptest.NewRecorder()

	handler.DeleteUser(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.deleteUserCalled {
		t.Fatal("expected service DeleteUser not to be called")
	}
}

func TestDeleteUserReturnsNotFoundWhenServiceReturnsNotFound(t *testing.T) {
	service := &fakeUsersService{
		deleteUserErr: fmt.Errorf("delete user: %w", core_errors.ErrNotFound),
	}
	handler := NewUsersHTTPHandler(service)

	request := newDeleteUserRequest()
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.DeleteUser(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
	if !service.deleteUserCalled {
		t.Fatal("expected service DeleteUser to be called")
	}
}

func newDeleteUserRequest() *http.Request {
	return newUserRequestWithLogger(
		http.MethodDelete,
		"/users/15",
		strings.NewReader(""),
	)
}
