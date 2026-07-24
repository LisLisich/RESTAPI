package tasks_transport_http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestDeleteTaskReturnsNoContentOnValidRequest(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newDeleteTaskRequest()
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.DeleteTask(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if !service.deleteTaskCalled {
		t.Fatal("expected service DeleteTask to be called")
	}
	if service.deletedTaskID != 15 {
		t.Fatalf("expected task id %d, got %d", 15, service.deletedTaskID)
	}
}

func TestDeleteTaskReturnsBadRequestOnInvalidID(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newDeleteTaskRequest()
	request.SetPathValue("id", "not-integer")
	response := httptest.NewRecorder()

	handler.DeleteTask(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.deleteTaskCalled {
		t.Fatal("expected service DeleteTask not to be called")
	}
}

func TestDeleteTaskReturnsNotFoundWhenServiceReturnsNotFound(t *testing.T) {
	service := &fakeTasksService{
		deleteTaskErr: fmt.Errorf("delete task: %w", core_errors.ErrNotFound),
	}
	handler := NewTasksHTTPHandler(service)

	request := newDeleteTaskRequest()
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.DeleteTask(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
	if !service.deleteTaskCalled {
		t.Fatal("expected service DeleteTask to be called")
	}
}

func newDeleteTaskRequest() *http.Request {
	return newRequestWithLogger(
		http.MethodDelete,
		"/tasks/15",
		strings.NewReader(""),
	)
}
