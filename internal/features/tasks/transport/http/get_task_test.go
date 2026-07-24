package tasks_transport_http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
)

func TestGetTaskReturnsOKOnValidRequest(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newGetTaskRequest()
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.GetTask(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.getTaskCalled {
		t.Fatal("expected service GetTask to be called")
	}
	if service.gotTaskID != 15 {
		t.Fatalf("expected task id %d, got %d", 15, service.gotTaskID)
	}

	var body TaskDTOResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.ID != 15 {
		t.Fatalf("expected response task id %d, got %d", 15, body.ID)
	}
}

func TestGetTaskReturnsBadRequestOnInvalidID(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newGetTaskRequest()
	request.SetPathValue("id", "not-integer")
	response := httptest.NewRecorder()

	handler.GetTask(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.getTaskCalled {
		t.Fatal("expected service GetTask not to be called")
	}
}

func TestGetTaskReturnsNotFoundWhenServiceReturnsNotFound(t *testing.T) {
	service := &fakeTasksService{
		getTaskErr: fmt.Errorf("get task: %w", core_errors.ErrNotFound),
	}
	handler := NewTasksHTTPHandler(service)

	request := newGetTaskRequest()
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.GetTask(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
	if !service.getTaskCalled {
		t.Fatal("expected service GetTask to be called")
	}

	var body core_http_response.ErrorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response body: %v", err)
	}
	if body.Message != "failed to get task" {
		t.Fatalf("expected message %q, got %q", "failed to get task", body.Message)
	}
}

func newGetTaskRequest() *http.Request {
	return newRequestWithLogger(
		http.MethodGet,
		"/tasks/15",
		strings.NewReader(""),
	)
}
