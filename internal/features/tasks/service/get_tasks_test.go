package tasks_service

import (
	"context"
	"errors"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestGetTasksPassesQueryParamsToRepository(t *testing.T) {
	repository := &fakeTaskRepository{}
	service := NewTasksService(repository)
	userID := taskIntPtr(1)
	limit := taskIntPtr(20)
	offset := taskIntPtr(40)

	tasks, err := service.GetTasks(context.Background(), userID, limit, offset)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if !repository.getTasksCalled {
		t.Fatal("expected GetTasks to be called")
	}
	assertTaskIntPointerValue(t, repository.getTasksUserID, 1, "user id")
	assertTaskIntPointerValue(t, repository.getTasksLimit, 20, "limit")
	assertTaskIntPointerValue(t, repository.getTasksOffset, 40, "offset")
}

func TestGetTasksRejectsNegativeLimitBeforeRepository(t *testing.T) {
	repository := &fakeTaskRepository{}
	service := NewTasksService(repository)

	_, err := service.GetTasks(context.Background(), nil, taskIntPtr(-1), nil)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if repository.getTasksCalled {
		t.Fatal("expected GetTasks not to be called for negative limit")
	}
}

func TestGetTasksRejectsNegativeOffsetBeforeRepository(t *testing.T) {
	repository := &fakeTaskRepository{}
	service := NewTasksService(repository)

	_, err := service.GetTasks(context.Background(), nil, nil, taskIntPtr(-1))

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if repository.getTasksCalled {
		t.Fatal("expected GetTasks not to be called for negative offset")
	}
}

func TestGetTasksWrapsRepositoryError(t *testing.T) {
	repository := &fakeTaskRepository{
		getTasksErr: core_errors.ErrNotFound,
	}
	service := NewTasksService(repository)

	_, err := service.GetTasks(context.Background(), nil, nil, nil)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !repository.getTasksCalled {
		t.Fatal("expected GetTasks to be called")
	}
}

func assertTaskIntPointerValue(t *testing.T, actual *int, expected int, name string) {
	t.Helper()

	if actual == nil {
		t.Fatalf("expected %s to be %d, got nil", name, expected)
	}
	if *actual != expected {
		t.Fatalf("expected %s to be %d, got %d", name, expected, *actual)
	}
}
