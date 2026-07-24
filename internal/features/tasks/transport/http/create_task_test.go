package tasks_transport_http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	"go.uber.org/zap"
)

type fakeTasksService struct {
	createTaskCalled bool
	createdTask      domain.Task

	getTasksCalled bool
	getTasksUserID *int
	getTasksLimit  *int
	getTasksOffset *int
	getTasksErr    error

	getTaskCalled bool
	gotTaskID     int
	getTaskErr    error

	deleteTaskCalled bool
	deletedTaskID    int
	deleteTaskErr    error

	patchTaskCalled bool
	patchedTaskID   int
	patchedPatch    domain.TaskPatch
	patchTaskErr    error
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
	s.getTasksCalled = true
	s.getTasksUserID = userID
	s.getTasksLimit = limit
	s.getTasksOffset = offset

	if s.getTasksErr != nil {
		return nil, s.getTasksErr
	}

	return []domain.Task{
		newTask(10, 1, "first task"),
		newTask(11, 1, "second task"),
	}, nil
}

func (s *fakeTasksService) GetTask(
	ctx context.Context,
	id int,
) (domain.Task, error) {
	s.getTaskCalled = true
	s.gotTaskID = id

	if s.getTaskErr != nil {
		return domain.Task{}, s.getTaskErr
	}

	return newTask(id, 1, "found task"), nil
}

func (s *fakeTasksService) DeleteTask(
	ctx context.Context,
	id int,
) error {
	s.deleteTaskCalled = true
	s.deletedTaskID = id

	if s.deleteTaskErr != nil {
		return s.deleteTaskErr
	}

	return nil
}

func (s *fakeTasksService) PatchTask(
	ctx context.Context,
	id int,
	patch domain.TaskPatch,
) (domain.Task, error) {
	s.patchTaskCalled = true
	s.patchedTaskID = id
	s.patchedPatch = patch

	if s.patchTaskErr != nil {
		return domain.Task{}, s.patchTaskErr
	}

	return newPatchedTask(id), nil
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
	return newRequestWithLogger(
		http.MethodPost,
		"/tasks",
		strings.NewReader(body),
	)
}

func newRequestWithLogger(method string, target string, body *strings.Reader) *http.Request {
	request := httptest.NewRequest(method, target, body)
	logger := &core_logger.Logger{
		Logger: zap.NewNop(),
	}
	return request.WithContext(core_logger.ToContext(request.Context(), logger))
}

func ptr[T any](value T) *T {
	return &value
}

func newTask(id int, authorUserID int, title string) domain.Task {
	return domain.NewTask(
		id,
		1,
		title,
		ptr("description"),
		false,
		time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC),
		nil,
		authorUserID,
	)
}

func newPatchedTask(id int) domain.Task {
	return domain.NewTask(
		id,
		2,
		"new title",
		nil,
		true,
		time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC),
		ptr(time.Date(2026, 7, 3, 12, 1, 0, 0, time.UTC)),
		1,
	)
}
