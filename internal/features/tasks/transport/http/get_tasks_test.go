package tasks_transport_http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetTasksReturnsOKOnValidQueryParams(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newGetTasksRequest("/tasks?user_id=1&limit=20&offset=40")
	response := httptest.NewRecorder()

	handler.GetTasks(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.getTasksCalled {
		t.Fatal("expected service GetTasks to be called")
	}
	assertIntPointerValue(t, service.getTasksUserID, 1, "user id")
	assertIntPointerValue(t, service.getTasksLimit, 20, "limit")
	assertIntPointerValue(t, service.getTasksOffset, 40, "offset")
}

func TestGetTasksPassesNilPointersWhenQueryParamsAreAbsent(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newGetTasksRequest("/tasks")
	response := httptest.NewRecorder()

	handler.GetTasks(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.getTasksCalled {
		t.Fatal("expected service GetTasks to be called")
	}
	if service.getTasksUserID != nil {
		t.Fatalf("expected user id to be nil, got %d", *service.getTasksUserID)
	}
	if service.getTasksLimit != nil {
		t.Fatalf("expected limit to be nil, got %d", *service.getTasksLimit)
	}
	if service.getTasksOffset != nil {
		t.Fatalf("expected offset to be nil, got %d", *service.getTasksOffset)
	}
}

func TestGetTasksReturnsBadRequestOnInvalidQueryParam(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newGetTasksRequest("/tasks?limit=not-integer")
	response := httptest.NewRecorder()

	handler.GetTasks(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.getTasksCalled {
		t.Fatal("expected service GetTasks not to be called")
	}
}

func TestGetTasksReturnsInternalServerErrorWhenServiceFails(t *testing.T) {
	service := &fakeTasksService{
		getTasksErr: errors.New("storage unavailable"),
	}
	handler := NewTasksHTTPHandler(service)

	request := newGetTasksRequest("/tasks")
	response := httptest.NewRecorder()

	handler.GetTasks(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
	if !service.getTasksCalled {
		t.Fatal("expected service GetTasks to be called")
	}
}

func newGetTasksRequest(target string) *http.Request {
	return newRequestWithLogger(
		http.MethodGet,
		target,
		strings.NewReader(""),
	)
}

func assertIntPointerValue(t *testing.T, actual *int, expected int, name string) {
	t.Helper()

	if actual == nil {
		t.Fatalf("expected %s to be %d, got nil", name, expected)
	}
	if *actual != expected {
		t.Fatalf("expected %s to be %d, got %d", name, expected, *actual)
	}
}
