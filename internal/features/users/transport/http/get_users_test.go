package users_transport_http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetUsersReturnsOKOnValidQueryParams(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newGetUsersRequest("/users?limit=20&offset=40")
	response := httptest.NewRecorder()

	handler.GetUsers(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.getUsersCalled {
		t.Fatal("expected service GetUsers to be called")
	}
	assertUserIntPointerValue(t, service.getUsersLimit, 20, "limit")
	assertUserIntPointerValue(t, service.getUsersOffset, 40, "offset")
}

func TestGetUsersPassesNilPointersWhenQueryParamsAreAbsent(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newGetUsersRequest("/users")
	response := httptest.NewRecorder()

	handler.GetUsers(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.getUsersCalled {
		t.Fatal("expected service GetUsers to be called")
	}
	if service.getUsersLimit != nil {
		t.Fatalf("expected limit to be nil, got %d", *service.getUsersLimit)
	}
	if service.getUsersOffset != nil {
		t.Fatalf("expected offset to be nil, got %d", *service.getUsersOffset)
	}
}

func TestGetUsersReturnsBadRequestOnInvalidQueryParam(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newGetUsersRequest("/users?offset=not-integer")
	response := httptest.NewRecorder()

	handler.GetUsers(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.getUsersCalled {
		t.Fatal("expected service GetUsers not to be called")
	}
}

func TestGetUsersReturnsInternalServerErrorWhenServiceFails(t *testing.T) {
	service := &fakeUsersService{
		getUsersErr: errors.New("storage unavailable"),
	}
	handler := NewUsersHTTPHandler(service)

	request := newGetUsersRequest("/users")
	response := httptest.NewRecorder()

	handler.GetUsers(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
	if !service.getUsersCalled {
		t.Fatal("expected service GetUsers to be called")
	}
}

func newGetUsersRequest(target string) *http.Request {
	return newUserRequestWithLogger(
		http.MethodGet,
		target,
		strings.NewReader(""),
	)
}

func assertUserIntPointerValue(t *testing.T, actual *int, expected int, name string) {
	t.Helper()

	if actual == nil {
		t.Fatalf("expected %s to be %d, got nil", name, expected)
	}
	if *actual != expected {
		t.Fatalf("expected %s to be %d, got %d", name, expected, *actual)
	}
}
