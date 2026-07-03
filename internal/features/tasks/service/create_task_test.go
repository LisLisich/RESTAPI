package tasks_service

import (
	"context"
	"errors"
	"testing"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

type fakeTaskRepository struct {
	createTaskCalled bool
}

var _ TasksRepository = (*fakeTaskRepository)(nil)

func (f *fakeTaskRepository) CreateTask(
	ctx context.Context,
	task domain.Task,
) (domain.Task, error) {
	f.createTaskCalled = true
	return task, nil
}

func (f *fakeTaskRepository) GetTask(
	ctx context.Context,
	id int,
) (domain.Task, error) {
	return domain.Task{}, nil
}

func (f *fakeTaskRepository) GetTasks(
	ctx context.Context,
	userID *int,
	limit *int,
	offset *int,
) ([]domain.Task, error) {
	return []domain.Task{}, nil
}

func (f *fakeTaskRepository) DeleteTask(
	ctx context.Context,
	id int,
) error {
	return nil
}

func (f *fakeTaskRepository) PatchTask(
	ctx context.Context,
	id int,
	task domain.Task,
) (domain.Task, error) {
	return domain.Task{}, nil
}

func TestCreateTaskRejectsInvalidDomainBeforeRepository(t *testing.T) {
	repository := &fakeTaskRepository{}
	service := NewTasksService(repository)
	task := domain.NewTaskUninitialized("ab", nil, 1)
	_, err := service.CreateTask(context.Background(), task)
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if repository.createTaskCalled {
		t.Fatalf("expected CreateTask not to be called for invalid task")
	}
}
