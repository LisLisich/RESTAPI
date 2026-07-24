package tasks_service

import (
	"context"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
)

type TasksService struct {
	tasksRepository TasksRepository
}

type TasksRepository interface {
	CreateTask(
		ctx context.Context,
		task domain.Task,
	) (domain.Task, error)
	GetTasks(
		ctx context.Context,
		userID *int,
		limit *int,
		offset *int,
	) ([]domain.Task, error)
	GetTask(
		ctx context.Context,
		id int,
	) (domain.Task, error)
	DeleteTask(
		ctx context.Context,
		id int,
	) error
	PatchTask(
		ctx context.Context,
		id int,
		patch domain.Task,
	) (domain.Task, error)
	GetTaskForUser(ctx context.Context, id int, userID int) (domain.Task, error)
	PatchTaskForUser(
		ctx context.Context,
		id int,
		userID int,
		task domain.Task,
	) (domain.Task, error)
	DeleteTaskForUser(ctx context.Context, id int, userID int) error
}

func NewTasksService(
	tasksRepository TasksRepository,
) *TasksService {
	return &TasksService{
		tasksRepository: tasksRepository,
	}
}
