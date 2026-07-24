package tasks_transport_http_v2

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestGetTaskUsesAuthenticatedOwner(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service, nil)
	request := newOwnedTaskRequest(http.MethodGet, "/tasks/15", "", 15)
	request = withPrincipal(request, 42)
	response := httptest.NewRecorder()

	handler.GetTask(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.getOwnedCall ||
		service.gotOwnedTaskID != 15 ||
		service.gotOwnedUserID != 42 {
		t.Fatal("expected owner-scoped get")
	}
}

func TestGetTaskHidesForeignTaskAsNotFound(t *testing.T) {
	service := &fakeTasksService{
		getOwnedErr: fmt.Errorf("owned task: %w", core_errors.ErrNotFound),
	}
	handler := NewTasksHTTPHandler(service, nil)
	request := newOwnedTaskRequest(http.MethodGet, "/tasks/15", "", 15)
	request = withPrincipal(request, 42)
	response := httptest.NewRecorder()

	handler.GetTask(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestPatchTaskUsesAuthenticatedOwnerAndNullablePatch(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service, nil)
	request := newOwnedTaskRequest(
		http.MethodPatch,
		"/tasks/15",
		`{"title":"updated title","description":null}`,
		15,
	)
	request = withPrincipal(request, 42)
	response := httptest.NewRecorder()

	handler.PatchTask(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body)
	}
	if !service.patchOwnedCall ||
		service.patchedTaskID != 15 ||
		service.patchedUserID != 42 {
		t.Fatal("expected owner-scoped patch")
	}
	if !service.patchedPatch.Title.Set ||
		service.patchedPatch.Title.Value == nil ||
		*service.patchedPatch.Title.Value != "updated title" {
		t.Fatalf("unexpected title patch %+v", service.patchedPatch.Title)
	}
	if !service.patchedPatch.Description.Set ||
		service.patchedPatch.Description.Value != nil {
		t.Fatalf("expected explicit null description, got %+v", service.patchedPatch.Description)
	}
}

func TestPatchTaskRejectsEmptyPatch(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service, nil)
	request := newOwnedTaskRequest(http.MethodPatch, "/tasks/15", `{}`, 15)
	request = withPrincipal(request, 42)
	response := httptest.NewRecorder()

	handler.PatchTask(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.patchOwnedCall {
		t.Fatal("service must not be called for empty patch")
	}
}

func TestDeleteTaskUsesAuthenticatedOwner(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service, nil)
	request := newOwnedTaskRequest(http.MethodDelete, "/tasks/15", "", 15)
	request = withPrincipal(request, 42)
	response := httptest.NewRecorder()

	handler.DeleteTask(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if !service.deleteOwnedCall ||
		service.deletedTaskID != 15 ||
		service.deletedUserID != 42 {
		t.Fatal("expected owner-scoped delete")
	}
}

func TestTaskRoutesExposeFullProtectedCRUD(t *testing.T) {
	middlewareCalls := 0
	protectedMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			middlewareCalls++
			next.ServeHTTP(rw, r)
		})
	}
	handler := NewTasksHTTPHandler(&fakeTasksService{}, protectedMiddleware)
	routes := handler.Routes()

	if len(routes) != 5 {
		t.Fatalf("expected five routes, got %d", len(routes))
	}
	expected := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/tasks"},
		{method: http.MethodGet, path: "/tasks"},
		{method: http.MethodGet, path: "/tasks/{id}"},
		{method: http.MethodPatch, path: "/tasks/{id}"},
		{method: http.MethodDelete, path: "/tasks/{id}"},
	}
	for index, route := range routes {
		if route.Method != expected[index].method || route.Path != expected[index].path {
			t.Fatalf("unexpected route %d: %s %s", index, route.Method, route.Path)
		}
		request := newV2TaskRequest(route.Method, "/tasks/15", `{}`)
		request = withPrincipal(request, 42)
		request.SetPathValue("id", "15")
		response := httptest.NewRecorder()
		route.WithMiddleware().ServeHTTP(response, request)
	}
	if middlewareCalls != len(routes) {
		t.Fatalf("expected middleware on every route, got %d calls", middlewareCalls)
	}
}

func newOwnedTaskRequest(
	method string,
	target string,
	body string,
	taskID int,
) *http.Request {
	request := newV2TaskRequest(method, target, body)
	request.SetPathValue("id", fmt.Sprintf("%d", taskID))
	return request
}
