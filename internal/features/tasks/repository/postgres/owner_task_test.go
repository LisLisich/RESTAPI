package task_postgres_repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool"
)

func TestGetTaskForUserFiltersByTaskAndOwner(t *testing.T) {
	createdAt := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)
	pool := &fakePool{
		queryRow: &fakeRow{
			values: []any{15, 1, "owned task", (*string)(nil), false, createdAt, (*time.Time)(nil), 42},
		},
	}
	repository := NewTaskRepository(pool)

	task, err := repository.GetTaskForUser(context.Background(), 15, 42)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task.ID != 15 || task.AuthorUserID != 42 {
		t.Fatalf("unexpected task %+v", task)
	}
	if !strings.Contains(pool.queryRowSQL, "id=$1 AND author_user_id=$2") {
		t.Fatalf("expected owner filter, got %q", pool.queryRowSQL)
	}
	if len(pool.queryRowArgs) != 2 || pool.queryRowArgs[0] != 15 || pool.queryRowArgs[1] != 42 {
		t.Fatalf("unexpected query args %v", pool.queryRowArgs)
	}
}

func TestGetTaskForUserHidesForeignTaskAsNotFound(t *testing.T) {
	pool := &fakePool{
		queryRow: &fakeRow{err: core_postgres_pool.ErrNoRows},
	}
	repository := NewTaskRepository(pool)

	_, err := repository.GetTaskForUser(context.Background(), 15, 42)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPatchTaskForUserUsesOwnerAndVersionGuards(t *testing.T) {
	task := newRepositoryTask(15, 42, "updated title")
	pool := &fakePool{
		queryRow: &fakeRow{
			values: []any{
				task.ID,
				2,
				task.Title,
				task.Description,
				task.Completed,
				task.CreatedAt,
				task.CompletedAt,
				task.AuthorUserID,
			},
		},
	}
	repository := NewTaskRepository(pool)

	_, err := repository.PatchTaskForUser(context.Background(), 15, 42, task)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(
		pool.queryRowSQL,
		"WHERE id=$5 AND author_user_id=$6 AND version=$7",
	) {
		t.Fatalf("expected owner and version guards, got %q", pool.queryRowSQL)
	}
	if pool.queryRowArgs[5] != 42 || pool.queryRowArgs[6] != task.Version {
		t.Fatalf("unexpected owner/version args %v", pool.queryRowArgs)
	}
}

func TestPatchTaskForUserMapsMissingVersionToConflict(t *testing.T) {
	pool := &fakePool{
		queryRow: &fakeRow{err: core_postgres_pool.ErrNoRows},
	}
	repository := NewTaskRepository(pool)

	_, err := repository.PatchTaskForUser(
		context.Background(),
		15,
		42,
		newRepositoryTask(15, 42, "updated title"),
	)

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestDeleteTaskForUserUsesOwnerGuard(t *testing.T) {
	pool := &fakePool{
		execTag: fakeCommandTag{rowsAffected: 1},
	}
	repository := NewTaskRepository(pool)

	err := repository.DeleteTaskForUser(context.Background(), 15, 42)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(pool.execSQL, "WHERE id=$1 AND author_user_id=$2") {
		t.Fatalf("expected owner guard, got %q", pool.execSQL)
	}
	if len(pool.execArgs) != 2 || pool.execArgs[0] != 15 || pool.execArgs[1] != 42 {
		t.Fatalf("unexpected delete args %v", pool.execArgs)
	}
}

func TestDeleteTaskForUserHidesForeignTaskAsNotFound(t *testing.T) {
	pool := &fakePool{
		execTag: fakeCommandTag{rowsAffected: 0},
	}
	repository := NewTaskRepository(pool)

	err := repository.DeleteTaskForUser(context.Background(), 15, 42)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
