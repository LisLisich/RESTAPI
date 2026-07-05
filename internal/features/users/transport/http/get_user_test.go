package users_transport_http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestGetUserReturnsOKOnValidRequest(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newGetUserRequest()
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.GetUser(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.getUserCalled {
		t.Fatal("expected service GetUser to be called")
	}
	if service.gotUserID != 15 {
		t.Fatalf("expected user id %d, got %d", 15, service.gotUserID)
	}

	var body UserDTOResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.ID != 15 {
		t.Fatalf("expected response user id %d, got %d", 15, body.ID)
	}
}

func TestGetUserReturnsBadRequestOnInvalidID(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newGetUserRequest()
	request.SetPathValue("id", "not-integer")
	response := httptest.NewRecorder()

	handler.GetUser(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.getUserCalled {
		t.Fatal("expected service GetUser not to be called")
	}
}

func TestGetUserReturnsNotFoundWhenServiceReturnsNotFound(t *testing.T) {
	service := &fakeUsersService{
		getUserErr: fmt.Errorf("get user: %w", core_errors.ErrNotFound),
	}
	handler := NewUsersHTTPHandler(service)

	request := newGetUserRequest()
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.GetUser(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
	if !service.getUserCalled {
		t.Fatal("expected service GetUser to be called")
	}
}

func newGetUserRequest() *http.Request {
	return newUserRequestWithLogger(
		http.MethodGet,
		"/users/15",
		strings.NewReader(""),
	)
}
