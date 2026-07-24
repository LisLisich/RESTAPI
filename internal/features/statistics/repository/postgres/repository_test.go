package statistics_postgres_repository

import (
	"context"
	"strings"
	"testing"
	"time"

	core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"
)

type fakePool struct {
	querySQL  string
	queryArgs []any
	queryRows core_postgres_pool.Rows
	queryErr  error
}

var _ core_postgres_pool.Pool = (*fakePool)(nil)

func (p *fakePool) Query(ctx context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
	p.querySQL = sql
	p.queryArgs = args
	if p.queryErr != nil {
		return nil, p.queryErr
	}
	return p.queryRows, nil
}

func (p *fakePool) QueryRow(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
	return nil
}

func (p *fakePool) Exec(ctx context.Context, sql string, arguments ...any) (core_postgres_pool.CommandTag, error) {
	return nil, nil
}

func (p *fakePool) Close() {}

func (p *fakePool) OpTimeout() time.Duration {
	return time.Second
}

type fakeRows struct {
	closed bool
}

var _ core_postgres_pool.Rows = (*fakeRows)(nil)

func (r *fakeRows) Close() {
	r.closed = true
}

func (r *fakeRows) Err() error {
	return nil
}

func (r *fakeRows) Next() bool {
	return false
}

func (r *fakeRows) Scan(dest ...any) error {
	return nil
}

func TestGetTasksBuildsFiltersInArgumentOrder(t *testing.T) {
	pool := &fakePool{
		queryRows: &fakeRows{},
	}
	repository := NewStatisticsRepository(pool)
	userID := 7
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)

	_, err := repository.GetTasks(context.Background(), &userID, &from, &to)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(pool.querySQL, "author_user_id=$1") {
		t.Fatalf("expected user filter placeholder, got %q", pool.querySQL)
	}
	if !strings.Contains(pool.querySQL, "created_at>=$2") {
		t.Fatalf("expected from filter placeholder, got %q", pool.querySQL)
	}
	if !strings.Contains(pool.querySQL, "created_at<$3") {
		t.Fatalf("expected to filter placeholder, got %q", pool.querySQL)
	}
	if len(pool.queryArgs) != 3 {
		t.Fatalf("expected 3 query args, got %d", len(pool.queryArgs))
	}
	if pool.queryArgs[0] != &userID || pool.queryArgs[1] != &from || pool.queryArgs[2] != &to {
		t.Fatalf("expected userID, from, to pointers in query args")
	}
}

func TestGetTasksOmitsWhereWhenFiltersAreAbsent(t *testing.T) {
	pool := &fakePool{
		queryRows: &fakeRows{},
	}
	repository := NewStatisticsRepository(pool)

	_, err := repository.GetTasks(context.Background(), nil, nil, nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if strings.Contains(pool.querySQL, " WHERE ") {
		t.Fatalf("expected query without WHERE, got %q", pool.querySQL)
	}
	if len(pool.queryArgs) != 0 {
		t.Fatalf("expected no query args, got %d", len(pool.queryArgs))
	}
}
