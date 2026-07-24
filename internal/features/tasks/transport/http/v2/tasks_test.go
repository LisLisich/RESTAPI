package tasks_transport_http_v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	identity_service "github.com/LisLisich/fintask/internal/features/identity/service"
	identity_http_middleware "github.com/LisLisich/fintask/internal/features/identity/transport/http/middleware"
	"go.uber.org/zap"
)

type fakeTasksService struct {
	createdTask domain.Task
	createCall  bool

	getCall   bool
	gotUserID *int
	gotLimit  *int
	gotOffset *int

	getOwnedCall   bool
	gotOwnedTaskID int
	gotOwnedUserID int
	getOwnedErr    error

	patchOwnedCall bool
	patchedTaskID  int
	patchedUserID  int
	patchedPatch   domain.TaskPatch
	patchOwnedErr  error

	deleteOwnedCall bool
	deletedTaskID   int
	deletedUserID   int
	deleteOwnedErr  error
}

func (s *fakeTasksService) CreateTask(
	_ context.Context,
	task domain.Task,
) (domain.Task, error) {
	s.createCall = true
	s.createdTask = task
	task.ID = 10
	task.Version = 1
	return task, nil
}

func (s *fakeTasksService) GetTasks(
	_ context.Context,
	userID *int,
	limit *int,
	offset *int,
) ([]domain.Task, error) {
	s.getCall = true
	s.gotUserID = userID
	s.gotLimit = limit
	s.gotOffset = offset
	return []domain.Task{}, nil
}

func (s *fakeTasksService) GetOwnedTask(
	_ context.Context,
	id int,
	userID int,
) (domain.Task, error) {
	s.getOwnedCall = true
	s.gotOwnedTaskID = id
	s.gotOwnedUserID = userID
	if s.getOwnedErr != nil {
		return domain.Task{}, s.getOwnedErr
	}
	return domain.NewTask(
		id,
		1,
		"owned task",
		nil,
		false,
		time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC),
		nil,
		userID,
	), nil
}

func (s *fakeTasksService) PatchOwnedTask(
	_ context.Context,
	id int,
	userID int,
	patch domain.TaskPatch,
) (domain.Task, error) {
	s.patchOwnedCall = true
	s.patchedTaskID = id
	s.patchedUserID = userID
	s.patchedPatch = patch
	if s.patchOwnedErr != nil {
		return domain.Task{}, s.patchOwnedErr
	}
	return domain.NewTask(
		id,
		2,
		"updated title",
		nil,
		false,
		time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC),
		nil,
		userID,
	), nil
}

func (s *fakeTasksService) DeleteOwnedTask(
	_ context.Context,
	id int,
	userID int,
) error {
	s.deleteOwnedCall = true
	s.deletedTaskID = id
	s.deletedUserID = userID
	return s.deleteOwnedErr
}

func TestCreateTaskUsesAuthenticatedOwner(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service, nil)
	request := newV2TaskRequest(
		http.MethodPost,
		"/tasks",
		`{"title":"Prepare interview","author_user_id":999}`,
	)
	request = withPrincipal(request, 42)
	response := httptest.NewRecorder()

	handler.CreateTask(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body)
	}
	if !service.createCall {
		t.Fatal("expected CreateTask call")
	}
	if service.createdTask.AuthorUserID != 42 {
		t.Fatalf("expected authenticated owner 42, got %d", service.createdTask.AuthorUserID)
	}
	if strings.Contains(response.Body.String(), "999") {
		t.Fatal("client supplied author_user_id must be ignored")
	}
}

func TestGetTasksAlwaysFiltersByAuthenticatedOwner(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service, nil)
	request := newV2TaskRequest(
		http.MethodGet,
		"/tasks?user_id=999&limit=20&offset=10",
		"",
	)
	request = withPrincipal(request, 42)
	response := httptest.NewRecorder()

	handler.GetTasks(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.getCall || service.gotUserID == nil || *service.gotUserID != 42 {
		t.Fatal("expected tasks to be filtered by authenticated owner")
	}
	if service.gotLimit == nil || *service.gotLimit != 20 {
		t.Fatalf("expected limit 20, got %v", service.gotLimit)
	}
	if service.gotOffset == nil || *service.gotOffset != 10 {
		t.Fatalf("expected offset 10, got %v", service.gotOffset)
	}
}

func TestTaskRoutesUseSessionMiddleware(t *testing.T) {
	middlewareCalled := false
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(rw, r)
		})
	}
	handler := NewTasksHTTPHandler(&fakeTasksService{}, authMiddleware)
	routes := handler.Routes()
	request := newV2TaskRequest(http.MethodGet, "/tasks", "")
	request = withPrincipal(request, 42)
	response := httptest.NewRecorder()

	routes[1].WithMiddleware().ServeHTTP(response, request)

	if !middlewareCalled {
		t.Fatal("expected session middleware on v2 task route")
	}
}

func newV2TaskRequest(method string, target string, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	logger := &core_logger.Logger{Logger: zap.NewNop()}
	return request.WithContext(core_logger.ToContext(request.Context(), logger))
}

func withPrincipal(request *http.Request, userID int) *http.Request {
	return request.WithContext(
		identity_http_middleware.ContextWithPrincipal(
			request.Context(),
			identity_service.Principal{UserID: userID},
		),
	)
}
