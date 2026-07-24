package statistics_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

type fakeStatisticsRepository struct {
	getTasksCalled bool
	getTasksUserID *int
	getTasksFrom   *time.Time
	getTasksTo     *time.Time
	getTasksErr    error
	tasks          []domain.Task
}

var _ StatisticsRepository = (*fakeStatisticsRepository)(nil)

func (r *fakeStatisticsRepository) GetTasks(
	ctx context.Context,
	userID *int,
	from *time.Time,
	to *time.Time,
) ([]domain.Task, error) {
	r.getTasksCalled = true
	r.getTasksUserID = userID
	r.getTasksFrom = from
	r.getTasksTo = to

	if r.getTasksErr != nil {
		return nil, r.getTasksErr
	}
	return r.tasks, nil
}

func TestGetStatisticsCalculatesCompletedRateAndAverageCompletionTime(t *testing.T) {
	createdAt := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	completedAtOne := createdAt.Add(2 * time.Hour)
	completedAtTwo := createdAt.Add(4 * time.Hour)
	repository := &fakeStatisticsRepository{
		tasks: []domain.Task{
			newStatisticsTask(1, true, createdAt, &completedAtOne),
			newStatisticsTask(2, true, createdAt, &completedAtTwo),
			newStatisticsTask(3, false, createdAt, nil),
		},
	}
	service := NewStatisticsService(repository)

	statistics, err := service.GetStatistics(context.Background(), nil, nil, nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if statistics.TasksCreated != 3 {
		t.Fatalf("expected tasks created %d, got %d", 3, statistics.TasksCreated)
	}
	if statistics.TasksCompleted != 2 {
		t.Fatalf("expected tasks completed %d, got %d", 2, statistics.TasksCompleted)
	}
	if statistics.TasksCompletedRate == nil || *statistics.TasksCompletedRate != float64(2)/float64(3)*100 {
		t.Fatalf("expected completed rate %v, got %v", float64(2)/float64(3)*100, statistics.TasksCompletedRate)
	}
	if statistics.TasksAverageCompletionTime == nil || *statistics.TasksAverageCompletionTime != 3*time.Hour {
		t.Fatalf("expected average completion time %v, got %v", 3*time.Hour, statistics.TasksAverageCompletionTime)
	}
}

func TestGetStatisticsRejectsInvalidDateRangeBeforeRepository(t *testing.T) {
	repository := &fakeStatisticsRepository{}
	service := NewStatisticsService(repository)
	from := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	to := from

	_, err := service.GetStatistics(context.Background(), nil, &from, &to)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if repository.getTasksCalled {
		t.Fatal("expected GetTasks not to be called for invalid date range")
	}
}

func TestGetStatisticsWrapsRepositoryError(t *testing.T) {
	repository := &fakeStatisticsRepository{
		getTasksErr: core_errors.ErrNotFound,
	}
	service := NewStatisticsService(repository)

	_, err := service.GetStatistics(context.Background(), nil, nil, nil)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !repository.getTasksCalled {
		t.Fatal("expected GetTasks to be called")
	}
}

func newStatisticsTask(id int, completed bool, createdAt time.Time, completedAt *time.Time) domain.Task {
	return domain.NewTask(
		id,
		1,
		"valid title",
		nil,
		completed,
		createdAt,
		completedAt,
		1,
	)
}
