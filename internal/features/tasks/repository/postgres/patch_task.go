package task_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"
)

func (r *TaskRepository) PatchTask(
	ctx context.Context,
	id int,
	patch domain.Task,
) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	query := `
	UPDATE todoapp.tasks
	SET 
		title=$1,
		description=$2,
		completed=$3,
		completed_at=$4,
		version=version+1
	WHERE id=$5 AND version=$6
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
	row := r.pool.QueryRow(
		ctx,
		query,
		patch.Title,
		patch.Description,
		patch.Completed,
		patch.CompletedAt,
		id,
		patch.Version,
	)
	var taskModel TaskModel
	err := row.Scan(
		&taskModel.ID,
		&taskModel.Version,
		&taskModel.Title,
		&taskModel.Description,
		&taskModel.Completed,
		&taskModel.CreatedAt,
		&taskModel.CompletedAt,
		&taskModel.AuthorUserID,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Task{}, fmt.Errorf(
				"task with id='%d': concurrently accessed: %w",
				id,
				core_errors.ErrConflict,
			)
		}
		return domain.Task{}, fmt.Errorf("scan error: %w", err)
	}
	taskDomain := taskDomainsFromModel(taskModel)
	return taskDomain, nil
}
