package tasks_transport_http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestPatchTaskReturnsOKOnValidRequest(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newPatchTaskRequest(`{
		"title": "new title",
		"description": null,
		"completed": true
	}`)
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.PatchTask(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.patchTaskCalled {
		t.Fatal("expected service PatchTask to be called")
	}
	if service.patchedTaskID != 15 {
		t.Fatalf("expected task id %d, got %d", 15, service.patchedTaskID)
	}
	if !service.patchedPatch.Title.Set || service.patchedPatch.Title.Value == nil {
		t.Fatal("expected title patch to be set")
	}
	if *service.patchedPatch.Title.Value != "new title" {
		t.Fatalf("expected title %q, got %q", "new title", *service.patchedPatch.Title.Value)
	}
	if !service.patchedPatch.Description.Set {
		t.Fatal("expected description patch to be set")
	}
	if service.patchedPatch.Description.Value != nil {
		t.Fatal("expected description patch value to be nil")
	}
	if !service.patchedPatch.Completed.Set || service.patchedPatch.Completed.Value == nil {
		t.Fatal("expected completed patch to be set")
	}
	if !*service.patchedPatch.Completed.Value {
		t.Fatal("expected completed patch value to be true")
	}
}

func TestPatchTaskWritesStatusHeaderOnceOnValidRequest(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newPatchTaskRequest(`{
		"title": "new title"
	}`)
	request.SetPathValue("id", "15")
	response := newWriteHeaderCountingResponseWriter()

	handler.PatchTask(response, request)

	if response.writeHeaderCalls != 1 {
		t.Fatalf("expected WriteHeader to be called once, got %d", response.writeHeaderCalls)
	}
}

func TestPatchTaskReturnsBadRequestOnInvalidID(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newPatchTaskRequest(`{
		"title": "new title"
	}`)
	request.SetPathValue("id", "not-integer")
	response := httptest.NewRecorder()

	handler.PatchTask(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.patchTaskCalled {
		t.Fatal("expected service PatchTask not to be called")
	}
}

func TestPatchTaskReturnsBadRequestOnInvalidRequest(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newPatchTaskRequest(`{
		"title": null
	}`)
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.PatchTask(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.patchTaskCalled {
		t.Fatal("expected service PatchTask not to be called")
	}
}

func TestPatchTaskReturnsNotFoundWhenServiceReturnsNotFound(t *testing.T) {
	service := &fakeTasksService{
		patchTaskErr: fmt.Errorf("patch task: %w", core_errors.ErrNotFound),
	}
	handler := NewTasksHTTPHandler(service)

	request := newPatchTaskRequest(`{
		"title": "new title"
	}`)
	request.SetPathValue("id", "15")
	response := httptest.NewRecorder()

	handler.PatchTask(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
	if !service.patchTaskCalled {
		t.Fatal("expected service PatchTask to be called")
	}
}

func newPatchTaskRequest(body string) *http.Request {
	return newRequestWithLogger(
		http.MethodPatch,
		"/tasks/15",
		strings.NewReader(body),
	)
}

type writeHeaderCountingResponseWriter struct {
	*httptest.ResponseRecorder
	writeHeaderCalls int
}

func newWriteHeaderCountingResponseWriter() *writeHeaderCountingResponseWriter {
	return &writeHeaderCountingResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

func (rw *writeHeaderCountingResponseWriter) WriteHeader(statusCode int) {
	rw.writeHeaderCalls++
	rw.ResponseRecorder.WriteHeader(statusCode)
}
