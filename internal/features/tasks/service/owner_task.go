package tasks_service

import (
	"context"
	"fmt"

	"github.com/LisLisich/fintask/internal/core/domain"
)

func (s *TasksService) GetOwnedTask(
	ctx context.Context,
	id int,
	userID int,
) (domain.Task, error) {
	task, err := s.tasksRepository.GetTaskForUser(ctx, id, userID)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get owned task: %w", err)
	}
	return task, nil
}

func (s *TasksService) PatchOwnedTask(
	ctx context.Context,
	id int,
	userID int,
	patch domain.TaskPatch,
) (domain.Task, error) {
	task, err := s.tasksRepository.GetTaskForUser(ctx, id, userID)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get owned task for patch: %w", err)
	}
	if err := task.ApplyPatch(patch); err != nil {
		return domain.Task{}, fmt.Errorf("apply owned task patch: %w", err)
	}
	task, err = s.tasksRepository.PatchTaskForUser(ctx, id, userID, task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("patch owned task: %w", err)
	}
	return task, nil
}

func (s *TasksService) DeleteOwnedTask(
	ctx context.Context,
	id int,
	userID int,
) error {
	if err := s.tasksRepository.DeleteTaskForUser(ctx, id, userID); err != nil {
		return fmt.Errorf("delete owned task: %w", err)
	}
	return nil
}
