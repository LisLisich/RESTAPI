package postgresintegration_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_pgx_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool/pgx"
	statistics_postgres_repository "github.com/LisLisich/fintask/internal/features/statistics/repository/postgres"
	task_postgres_repository "github.com/LisLisich/fintask/internal/features/tasks/repository/postgres"
	users_postgres_repository "github.com/LisLisich/fintask/internal/features/users/repository/postgres"
)

const integrationEnv = "FINTASK_POSTGRES_INTEGRATION"

func TestPostgresRepositoriesUserTaskLifecycle(t *testing.T) {
	pool := openIntegrationPool(t)
	requireMigratedSchema(t, pool)

	ctx := context.Background()
	usersRepository := users_postgres_repository.NewUserRepository(pool)
	tasksRepository := task_postgres_repository.NewTaskRepository(pool)
	statisticsRepository := statistics_postgres_repository.NewStatisticsRepository(pool)

	var createdUserID int
	createdTaskIDs := []int{}
	t.Cleanup(func() {
		for _, taskID := range createdTaskIDs {
			err := tasksRepository.DeleteTask(context.Background(), taskID)
			if err != nil && !errors.Is(err, core_errors.ErrNotFound) {
				t.Logf("cleanup task %d: %v", taskID, err)
			}
		}
		if createdUserID != 0 {
			err := usersRepository.DeleteUser(context.Background(), createdUserID)
			if err != nil && !errors.Is(err, core_errors.ErrNotFound) {
				t.Logf("cleanup user %d: %v", createdUserID, err)
			}
		}
	})

	unique := time.Now().UnixNano()
	phoneNumber := fmt.Sprintf("+79%09d", unique%1_000_000_000)
	user := domain.NewUserUninitialized(
		fmt.Sprintf("Integration User %d", unique),
		&phoneNumber,
	)

	createdUser, err := usersRepository.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	createdUserID = createdUser.ID

	gotUser, err := usersRepository.GetUser(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if gotUser.FullName != user.FullName || gotUser.PhoneNumber == nil || *gotUser.PhoneNumber != phoneNumber {
		t.Fatalf("unexpected user from postgres: %#v", gotUser)
	}

	userToPatch := createdUser
	userToPatch.FullName = "Integration User Updated"
	patchedUser, err := usersRepository.PatchUser(ctx, createdUser.ID, userToPatch)
	if err != nil {
		t.Fatalf("patch user: %v", err)
	}
	if patchedUser.Version != createdUser.Version+1 {
		t.Fatalf("expected user version %d, got %d", createdUser.Version+1, patchedUser.Version)
	}

	_, err = usersRepository.PatchUser(ctx, createdUser.ID, userToPatch)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected stale user patch to return ErrConflict, got %v", err)
	}

	taskCreatedAt := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	taskDescription := "integration task description"
	task := domain.NewTask(
		domain.UninitializedID,
		domain.UninitializedVersion,
		"integration task",
		&taskDescription,
		false,
		taskCreatedAt,
		nil,
		createdUser.ID,
	)

	createdTask, err := tasksRepository.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	createdTaskIDs = append(createdTaskIDs, createdTask.ID)

	gotTask, err := tasksRepository.GetTask(ctx, createdTask.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if gotTask.AuthorUserID != createdUser.ID || gotTask.Title != task.Title {
		t.Fatalf("unexpected task from postgres: %#v", gotTask)
	}

	limit := 10
	offset := 0
	userTasks, err := tasksRepository.GetTasks(ctx, &createdUser.ID, &limit, &offset)
	if err != nil {
		t.Fatalf("get tasks by user: %v", err)
	}
	requireOnlyTaskID(t, userTasks, createdTask.ID)

	from := taskCreatedAt.Add(-time.Minute)
	to := taskCreatedAt.Add(time.Minute)
	statisticTasks, err := statisticsRepository.GetTasks(ctx, &createdUser.ID, &from, &to)
	if err != nil {
		t.Fatalf("get statistic tasks: %v", err)
	}
	requireOnlyTaskID(t, statisticTasks, createdTask.ID)

	taskToPatch := createdTask
	taskToPatch.Title = "integration task updated"
	completedAt := taskCreatedAt.Add(time.Minute)
	taskToPatch.Completed = true
	taskToPatch.CompletedAt = &completedAt
	patchedTask, err := tasksRepository.PatchTask(ctx, createdTask.ID, taskToPatch)
	if err != nil {
		t.Fatalf("patch task: %v", err)
	}
	if patchedTask.Version != createdTask.Version+1 {
		t.Fatalf("expected task version %d, got %d", createdTask.Version+1, patchedTask.Version)
	}
	if !patchedTask.Completed || patchedTask.CompletedAt == nil {
		t.Fatalf("expected completed task, got %#v", patchedTask)
	}

	_, err = tasksRepository.PatchTask(ctx, createdTask.ID, taskToPatch)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected stale task patch to return ErrConflict, got %v", err)
	}

	orphanTask := domain.NewTaskUninitialized(
		"orphan integration task",
		&taskDescription,
		-1,
	)
	_, err = tasksRepository.CreateTask(ctx, orphanTask)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected task with missing author to return ErrNotFound, got %v", err)
	}

	if err := tasksRepository.DeleteTask(ctx, createdTask.ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}
	createdTaskIDs = nil
	if _, err := tasksRepository.GetTask(ctx, createdTask.ID); !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected deleted task to return ErrNotFound, got %v", err)
	}

	if err := usersRepository.DeleteUser(ctx, createdUser.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	createdUserID = 0
	if _, err := usersRepository.GetUser(ctx, createdUser.ID); !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected deleted user to return ErrNotFound, got %v", err)
	}
}

func openIntegrationPool(t *testing.T) *core_pgx_pool.Pool {
	t.Helper()

	if os.Getenv(integrationEnv) != "1" {
		t.Skipf(
			"set %s=1 and POSTGRES_USER/POSTGRES_PASSWORD/POSTGRES_DB to run PostgreSQL integration tests",
			integrationEnv,
		)
	}

	timeout := 5 * time.Second
	if timeoutEnv := os.Getenv("POSTGRES_TIMEOUT"); timeoutEnv != "" {
		parsedTimeout, err := time.ParseDuration(timeoutEnv)
		if err != nil {
			t.Fatalf("parse POSTGRES_TIMEOUT: %v", err)
		}
		timeout = parsedTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.Config{
		Host:     getenvDefault("POSTGRES_HOST", "127.0.0.1"),
		Port:     getenvDefault("POSTGRES_PORT", "5432"),
		User:     requiredEnv(t, "POSTGRES_USER"),
		Password: requiredEnv(t, "POSTGRES_PASSWORD"),
		Database: requiredEnv(t, "POSTGRES_DB"),
		Timeout:  timeout,
	})
	if err != nil {
		t.Fatalf("open postgres pool: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

func requireMigratedSchema(t *testing.T, pool *core_pgx_pool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), pool.OpTimeout())
	defer cancel()

	rows, err := pool.Query(ctx, "SELECT id FROM todoapp.users LIMIT 0")
	if err != nil {
		t.Fatalf("todoapp.users is unavailable, run migrations first: %v", err)
	}
	rows.Close()

	rows, err = pool.Query(ctx, "SELECT id FROM todoapp.tasks LIMIT 0")
	if err != nil {
		t.Fatalf("todoapp.tasks is unavailable, run migrations first: %v", err)
	}
	rows.Close()
}

func requireOnlyTaskID(t *testing.T, tasks []domain.Task, expectedID int) {
	t.Helper()

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d: %#v", len(tasks), tasks)
	}
	if tasks[0].ID != expectedID {
		t.Fatalf("expected task id %d, got %d", expectedID, tasks[0].ID)
	}
}

func requiredEnv(t *testing.T, name string) string {
	t.Helper()

	value := os.Getenv(name)
	if value == "" {
		t.Fatalf("%s must be set when %s=1", name, integrationEnv)
	}
	return value
}

func getenvDefault(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}
