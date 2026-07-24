package tasks_service

import (
	"context"
	"errors"
	"testing"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestGetOwnedTaskPassesTaskAndOwnerIDs(t *testing.T) {
	repository := &fakeTaskRepository{}
	service := NewTasksService(repository)

	task, err := service.GetOwnedTask(context.Background(), 15, 42)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.getOwnedTaskCalled ||
		repository.ownedTaskID != 15 ||
		repository.ownedTaskUserID != 42 {
		t.Fatal("expected owner-scoped repository lookup")
	}
	if task.ID != 15 || task.AuthorUserID != 42 {
		t.Fatalf("unexpected task %+v", task)
	}
}

func TestPatchOwnedTaskAppliesDomainPatchBeforeOwnedUpdate(t *testing.T) {
	repository := &fakeTaskRepository{
		storedTask: newServiceTask(15, 42, "old title"),
	}
	service := NewTasksService(repository)
	patch := domain.NewTaskPatch(
		domain.Nullable[string]{Set: true, Value: taskStringPtr("updated title")},
		domain.Nullable[string]{Set: true, Value: nil},
		domain.Nullable[bool]{Set: true, Value: taskBoolPtr(true)},
	)

	task, err := service.PatchOwnedTask(context.Background(), 15, 42, patch)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.getOwnedTaskCalled || !repository.patchOwnedTaskCalled {
		t.Fatal("expected owned get followed by owned patch")
	}
	if repository.patchedOwnedTaskID != 15 || repository.patchedOwnedUserID != 42 {
		t.Fatal("expected task and owner ids in update")
	}
	if repository.patchedOwnedTask.Title != "updated title" ||
		repository.patchedOwnedTask.Description != nil ||
		!repository.patchedOwnedTask.Completed ||
		repository.patchedOwnedTask.CompletedAt == nil {
		t.Fatalf("domain patch was not applied: %+v", repository.patchedOwnedTask)
	}
	if task.Title != "updated title" {
		t.Fatalf("unexpected returned task %+v", task)
	}
}

func TestPatchOwnedTaskDoesNotUpdateForeignOrInvalidTask(t *testing.T) {
	tests := []struct {
		name       string
		repository *fakeTaskRepository
		patch      domain.TaskPatch
		wantErr    error
	}{
		{
			name: "foreign task",
			repository: &fakeTaskRepository{
				getOwnedTaskErr: core_errors.ErrNotFound,
			},
			patch: domain.NewTaskPatch(
				domain.Nullable[string]{Set: true, Value: taskStringPtr("updated title")},
				domain.Nullable[string]{},
				domain.Nullable[bool]{},
			),
			wantErr: core_errors.ErrNotFound,
		},
		{
			name: "invalid patch",
			repository: &fakeTaskRepository{
				storedTask: newServiceTask(15, 42, "old title"),
			},
			patch: domain.NewTaskPatch(
				domain.Nullable[string]{Set: true, Value: taskStringPtr("ab")},
				domain.Nullable[string]{},
				domain.Nullable[bool]{},
			),
			wantErr: core_errors.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewTasksService(tt.repository)

			_, err := service.PatchOwnedTask(context.Background(), 15, 42, tt.patch)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.repository.patchOwnedTaskCalled {
				t.Fatal("repository update must not run")
			}
		})
	}
}

func TestDeleteOwnedTaskPassesTaskAndOwnerIDs(t *testing.T) {
	repository := &fakeTaskRepository{}
	service := NewTasksService(repository)

	err := service.DeleteOwnedTask(context.Background(), 15, 42)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.deleteOwnedTaskCalled ||
		repository.deletedOwnedTaskID != 15 ||
		repository.deletedOwnedUserID != 42 {
		t.Fatal("expected owner-scoped delete")
	}
}
