package tasks_service

import (
	"context"
	"errors"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestGetTaskPassesIDToRepository(t *testing.T) {
	repository := &fakeTaskRepository{}
	service := NewTasksService(repository)

	task, err := service.GetTask(context.Background(), 15)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.getTaskCalled {
		t.Fatal("expected GetTask to be called")
	}
	if repository.gotTaskID != 15 {
		t.Fatalf("expected task id %d, got %d", 15, repository.gotTaskID)
	}
	if task.ID != 15 {
		t.Fatalf("expected returned task id %d, got %d", 15, task.ID)
	}
}

func TestGetTaskWrapsRepositoryError(t *testing.T) {
	repository := &fakeTaskRepository{
		getTaskErr: core_errors.ErrNotFound,
	}
	service := NewTasksService(repository)

	_, err := service.GetTask(context.Background(), 15)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !repository.getTaskCalled {
		t.Fatal("expected GetTask to be called")
	}
}
