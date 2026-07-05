package tasks_service

import (
	"context"
	"errors"
	"testing"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestPatchTaskAppliesPatchBeforeRepositoryUpdate(t *testing.T) {
	repository := &fakeTaskRepository{
		storedTask: newServiceTask(15, 1, "old title"),
	}
	service := NewTasksService(repository)
	patch := domain.NewTaskPatch(
		domain.Nullable[string]{Set: true, Value: taskStringPtr("updated title")},
		domain.Nullable[string]{Set: true, Value: nil},
		domain.Nullable[bool]{Set: true, Value: taskBoolPtr(true)},
	)

	task, err := service.PatchTask(context.Background(), 15, patch)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.getTaskCalled {
		t.Fatal("expected GetTask to be called before PatchTask")
	}
	if !repository.patchTaskCalled {
		t.Fatal("expected PatchTask to be called")
	}
	if repository.patchedTaskID != 15 {
		t.Fatalf("expected patched task id %d, got %d", 15, repository.patchedTaskID)
	}
	if repository.patchedTask.Title != "updated title" {
		t.Fatalf("expected patched title %q, got %q", "updated title", repository.patchedTask.Title)
	}
	if repository.patchedTask.Description != nil {
		t.Fatalf("expected patched description to be nil, got %q", *repository.patchedTask.Description)
	}
	if !repository.patchedTask.Completed {
		t.Fatal("expected patched task to be completed")
	}
	if repository.patchedTask.CompletedAt == nil {
		t.Fatal("expected completed task to have CompletedAt")
	}
	if task.Title != "updated title" {
		t.Fatalf("expected returned task title %q, got %q", "updated title", task.Title)
	}
}

func TestPatchTaskReturnsGetTaskErrorBeforeApplyingPatch(t *testing.T) {
	repository := &fakeTaskRepository{
		getTaskErr: core_errors.ErrNotFound,
	}
	service := NewTasksService(repository)
	patch := domain.NewTaskPatch(
		domain.Nullable[string]{Set: true, Value: taskStringPtr("updated title")},
		domain.Nullable[string]{},
		domain.Nullable[bool]{},
	)

	_, err := service.PatchTask(context.Background(), 15, patch)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !repository.getTaskCalled {
		t.Fatal("expected GetTask to be called")
	}
	if repository.patchTaskCalled {
		t.Fatal("expected PatchTask not to be called when GetTask fails")
	}
}

func TestPatchTaskRejectsInvalidPatchBeforeRepositoryUpdate(t *testing.T) {
	repository := &fakeTaskRepository{
		storedTask: newServiceTask(15, 1, "old title"),
	}
	service := NewTasksService(repository)
	patch := domain.NewTaskPatch(
		domain.Nullable[string]{Set: true, Value: taskStringPtr("ab")},
		domain.Nullable[string]{},
		domain.Nullable[bool]{},
	)

	_, err := service.PatchTask(context.Background(), 15, patch)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if !repository.getTaskCalled {
		t.Fatal("expected GetTask to be called")
	}
	if repository.patchTaskCalled {
		t.Fatal("expected PatchTask not to be called for invalid patch")
	}
}

func TestPatchTaskWrapsRepositoryPatchError(t *testing.T) {
	repository := &fakeTaskRepository{
		storedTask:   newServiceTask(15, 1, "old title"),
		patchTaskErr: core_errors.ErrConflict,
	}
	service := NewTasksService(repository)
	patch := domain.NewTaskPatch(
		domain.Nullable[string]{Set: true, Value: taskStringPtr("updated title")},
		domain.Nullable[string]{},
		domain.Nullable[bool]{},
	)

	_, err := service.PatchTask(context.Background(), 15, patch)

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if !repository.patchTaskCalled {
		t.Fatal("expected PatchTask to be called")
	}
}

func taskBoolPtr(value bool) *bool {
	return &value
}
