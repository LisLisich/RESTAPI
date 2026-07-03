package tasks_transport_http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	"go.uber.org/zap"
)

type fakeTasksService struct {
	createTaskCalled bool
	createdTask      domain.Task
}

var _ TasksService = (*fakeTasksService)(nil)

func (s *fakeTasksService) CreateTask(
	ctx context.Context,
	task domain.Task,
) (domain.Task, error) {
	s.createTaskCalled = true
	s.createdTask = task

	return domain.NewTask(
		10,
		1,
		task.Title,
		task.Description,
		false,
		task.CreatedAt,
		nil,
		task.AuthorUserID,
	), nil
}

func (s *fakeTasksService) GetTasks(
	ctx context.Context,
	userID *int,
	limit *int,
	offset *int,
) ([]domain.Task, error) {
	return nil, nil
}

func (s *fakeTasksService) GetTask(
	ctx context.Context,
	id int,
) (domain.Task, error) {
	return domain.Task{}, nil
}

func (s *fakeTasksService) DeleteTask(
	ctx context.Context,
	id int,
) error {
	return nil
}

func (s *fakeTasksService) PatchTask(
	ctx context.Context,
	id int,
	patch domain.TaskPatch,
) (domain.Task, error) {
	return domain.Task{}, nil
}

func TestCreateTaskReturnsCreatedOnValidRequest(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newCreateTaskRequest(`{
	"title": "valid title",
	"description": "valid description",
	"author_user_id": 1
	}`)
	response := httptest.NewRecorder()

	handler.CreateTask(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}
	if !service.createTaskCalled {
		t.Fatal("expected service CreateTask to be called")
	}
	if service.createdTask.Title != "valid title" {
		t.Fatalf("expected task title %q, got %q", "valid title", service.createdTask.Title)
	}
	if service.createdTask.AuthorUserID != 1 {
		t.Fatalf("expected author user id %d, got %d", 1, service.createdTask.AuthorUserID)
	}
}

func TestCreateTaskReturnsBadRequestOnInvalidRequest(t *testing.T) {
	service := &fakeTasksService{}
	handler := NewTasksHTTPHandler(service)

	request := newCreateTaskRequest(`{
	"title": "",
	"description": "valid description",
	"author_user_id": 1
	}`)
	response := httptest.NewRecorder()

	handler.CreateTask(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.createTaskCalled {
		t.Fatal("expected service CreateTask not to be called")
	}
}

func newCreateTaskRequest(body string) *http.Request {
	request := httptest.NewRequest(
		http.MethodPost,
		"/tasks",
		strings.NewReader(body),
	)
	logger := &core_logger.Logger{
		Logger: zap.NewNop(),
	}
	return request.WithContext(core_logger.ToContext(request.Context(), logger))
}
