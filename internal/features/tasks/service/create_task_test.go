package tasks_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

type fakeTaskRepository struct {
	createTaskCalled bool
	createdTask      domain.Task
	createTaskErr    error

	getTasksCalled bool
	getTasksUserID *int
	getTasksLimit  *int
	getTasksOffset *int
	getTasksErr    error

	getTaskCalled bool
	gotTaskID     int
	getTaskErr    error
	storedTask    domain.Task

	deleteTaskCalled bool
	deletedTaskID    int
	deleteTaskErr    error

	patchTaskCalled bool
	patchedTaskID   int
	patchedTask     domain.Task
	patchTaskErr    error

	getOwnedTaskCalled bool
	ownedTaskID        int
	ownedTaskUserID    int
	getOwnedTaskErr    error

	patchOwnedTaskCalled bool
	patchedOwnedTaskID   int
	patchedOwnedUserID   int
	patchedOwnedTask     domain.Task
	patchOwnedTaskErr    error

	deleteOwnedTaskCalled bool
	deletedOwnedTaskID    int
	deletedOwnedUserID    int
	deleteOwnedTaskErr    error
}

var _ TasksRepository = (*fakeTaskRepository)(nil)

func (f *fakeTaskRepository) CreateTask(
	ctx context.Context,
	task domain.Task,
) (domain.Task, error) {
	f.createTaskCalled = true
	f.createdTask = task

	if f.createTaskErr != nil {
		return domain.Task{}, f.createTaskErr
	}

	return task, nil
}

func (f *fakeTaskRepository) GetTask(
	ctx context.Context,
	id int,
) (domain.Task, error) {
	f.getTaskCalled = true
	f.gotTaskID = id

	if f.getTaskErr != nil {
		return domain.Task{}, f.getTaskErr
	}
	if f.storedTask.ID != 0 {
		return f.storedTask, nil
	}

	return newServiceTask(id, 1, "found task"), nil
}

func (f *fakeTaskRepository) GetTasks(
	ctx context.Context,
	userID *int,
	limit *int,
	offset *int,
) ([]domain.Task, error) {
	f.getTasksCalled = true
	f.getTasksUserID = userID
	f.getTasksLimit = limit
	f.getTasksOffset = offset

	if f.getTasksErr != nil {
		return nil, f.getTasksErr
	}

	return []domain.Task{
		newServiceTask(10, 1, "first task"),
	}, nil
}

func (f *fakeTaskRepository) DeleteTask(
	ctx context.Context,
	id int,
) error {
	f.deleteTaskCalled = true
	f.deletedTaskID = id

	if f.deleteTaskErr != nil {
		return f.deleteTaskErr
	}

	return nil
}

func (f *fakeTaskRepository) PatchTask(
	ctx context.Context,
	id int,
	task domain.Task,
) (domain.Task, error) {
	f.patchTaskCalled = true
	f.patchedTaskID = id
	f.patchedTask = task

	if f.patchTaskErr != nil {
		return domain.Task{}, f.patchTaskErr
	}

	return task, nil
}

func (f *fakeTaskRepository) GetTaskForUser(
	_ context.Context,
	id int,
	userID int,
) (domain.Task, error) {
	f.getOwnedTaskCalled = true
	f.ownedTaskID = id
	f.ownedTaskUserID = userID
	if f.getOwnedTaskErr != nil {
		return domain.Task{}, f.getOwnedTaskErr
	}
	if f.storedTask.ID != 0 {
		return f.storedTask, nil
	}
	return newServiceTask(id, userID, "owned task"), nil
}

func (f *fakeTaskRepository) PatchTaskForUser(
	_ context.Context,
	id int,
	userID int,
	task domain.Task,
) (domain.Task, error) {
	f.patchOwnedTaskCalled = true
	f.patchedOwnedTaskID = id
	f.patchedOwnedUserID = userID
	f.patchedOwnedTask = task
	if f.patchOwnedTaskErr != nil {
		return domain.Task{}, f.patchOwnedTaskErr
	}
	return task, nil
}

func (f *fakeTaskRepository) DeleteTaskForUser(
	_ context.Context,
	id int,
	userID int,
) error {
	f.deleteOwnedTaskCalled = true
	f.deletedOwnedTaskID = id
	f.deletedOwnedUserID = userID
	return f.deleteOwnedTaskErr
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

func TestCreateTaskPassesValidDomainToRepository(t *testing.T) {
	repository := &fakeTaskRepository{}
	service := NewTasksService(repository)
	task := domain.NewTaskUninitialized("valid title", taskStringPtr("description"), 1)

	createdTask, err := service.CreateTask(context.Background(), task)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.createTaskCalled {
		t.Fatal("expected CreateTask to be called")
	}
	if repository.createdTask.Title != "valid title" {
		t.Fatalf("expected title %q, got %q", "valid title", repository.createdTask.Title)
	}
	if createdTask.Title != "valid title" {
		t.Fatalf("expected created task title %q, got %q", "valid title", createdTask.Title)
	}
}

func TestCreateTaskWrapsRepositoryError(t *testing.T) {
	repository := &fakeTaskRepository{
		createTaskErr: core_errors.ErrNotFound,
	}
	service := NewTasksService(repository)
	task := domain.NewTaskUninitialized("valid title", taskStringPtr("description"), 1)

	_, err := service.CreateTask(context.Background(), task)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !repository.createTaskCalled {
		t.Fatal("expected CreateTask to be called")
	}
}

func taskStringPtr(value string) *string {
	return &value
}

func taskIntPtr(value int) *int {
	return &value
}

func newServiceTask(id int, authorUserID int, title string) domain.Task {
	return domain.NewTask(
		id,
		1,
		title,
		taskStringPtr("description"),
		false,
		time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC),
		nil,
		authorUserID,
	)
}
