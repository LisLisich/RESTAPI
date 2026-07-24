package task_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"
)

func (r *TaskRepository) GetTaskForUser(
	ctx context.Context,
	id int,
	userID int,
) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		SELECT
			id,
			version,
			title,
			description,
			completed,
			created_at,
			completed_at,
			author_user_id
		FROM todoapp.tasks
		WHERE id=$1 AND author_user_id=$2;
	`
	var taskModel TaskModel
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&taskModel.ID,
		&taskModel.Version,
		&taskModel.Title,
		&taskModel.Description,
		&taskModel.Completed,
		&taskModel.CreatedAt,
		&taskModel.CompletedAt,
		&taskModel.AuthorUserID,
	)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return domain.Task{}, fmt.Errorf(
			"task with id='%d' for user id='%d': %w",
			id,
			userID,
			core_errors.ErrNotFound,
		)
	}
	if err != nil {
		return domain.Task{}, fmt.Errorf("scan owned task: %w", err)
	}
	return taskDomainsFromModel(taskModel), nil
}

func (r *TaskRepository) PatchTaskForUser(
	ctx context.Context,
	id int,
	userID int,
	task domain.Task,
) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		UPDATE todoapp.tasks
		SET
			title=$1,
			description=$2,
			completed=$3,
			completed_at=$4,
			version=version+1
		WHERE id=$5 AND author_user_id=$6 AND version=$7
		RETURNING
			id,
			version,
			title,
			description,
			completed,
			created_at,
			completed_at,
			author_user_id;
	`
	var taskModel TaskModel
	err := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Completed,
		task.CompletedAt,
		id,
		userID,
		task.Version,
	).Scan(
		&taskModel.ID,
		&taskModel.Version,
		&taskModel.Title,
		&taskModel.Description,
		&taskModel.Completed,
		&taskModel.CreatedAt,
		&taskModel.CompletedAt,
		&taskModel.AuthorUserID,
	)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return domain.Task{}, fmt.Errorf(
			"task with id='%d' for user id='%d' was concurrently accessed: %w",
			id,
			userID,
			core_errors.ErrConflict,
		)
	}
	if err != nil {
		return domain.Task{}, fmt.Errorf("scan patched owned task: %w", err)
	}
	return taskDomainsFromModel(taskModel), nil
}

func (r *TaskRepository) DeleteTaskForUser(
	ctx context.Context,
	id int,
	userID int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		DELETE FROM todoapp.tasks
		WHERE id=$1 AND author_user_id=$2;
	`
	tag, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete owned task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"task with id='%d' for user id='%d': %w",
			id,
			userID,
			core_errors.ErrNotFound,
		)
	}
	return nil
}
