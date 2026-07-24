package tasks_service

import (
	"context"
	"errors"
	"strings"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestDeleteTaskPassesIDToRepository(t *testing.T) {
	repository := &fakeTaskRepository{}
	service := NewTasksService(repository)

	err := service.DeleteTask(context.Background(), 15)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.deleteTaskCalled {
		t.Fatal("expected DeleteTask to be called")
	}
	if repository.deletedTaskID != 15 {
		t.Fatalf("expected task id %d, got %d", 15, repository.deletedTaskID)
	}
}

func TestDeleteTaskWrapsRepositoryErrorWithTaskContext(t *testing.T) {
	repository := &fakeTaskRepository{
		deleteTaskErr: core_errors.ErrNotFound,
	}
	service := NewTasksService(repository)

	err := service.DeleteTask(context.Background(), 15)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "delete task") {
		t.Fatalf("expected error to contain %q, got %q", "delete task", err.Error())
	}
	if !repository.deleteTaskCalled {
		t.Fatal("expected DeleteTask to be called")
	}
}
