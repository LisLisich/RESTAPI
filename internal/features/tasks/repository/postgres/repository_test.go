package task_postgres_repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool"
)

type fakePool struct {
	querySQL  string
	queryArgs []any
	queryRows core_postgres_pool.Rows
	queryErr  error

	queryRowSQL  string
	queryRowArgs []any
	queryRow     core_postgres_pool.Row

	execSQL  string
	execArgs []any
	execTag  core_postgres_pool.CommandTag
	execErr  error
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
	p.queryRowSQL = sql
	p.queryRowArgs = args
	if p.queryRow == nil {
		return &fakeRow{}
	}
	return p.queryRow
}

func (p *fakePool) Exec(ctx context.Context, sql string, arguments ...any) (core_postgres_pool.CommandTag, error) {
	p.execSQL = sql
	p.execArgs = arguments
	if p.execErr != nil {
		return nil, p.execErr
	}
	return p.execTag, nil
}

func (p *fakePool) Close() {}

func (p *fakePool) OpTimeout() time.Duration {
	return time.Second
}

type fakeRow struct {
	values []any
	err    error
}

var _ core_postgres_pool.Row = (*fakeRow)(nil)

func (r *fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	for i := range dest {
		switch d := dest[i].(type) {
		case *int:
			*d = r.values[i].(int)
		case *string:
			*d = r.values[i].(string)
		case **string:
			value, _ := r.values[i].(*string)
			*d = value
		case *bool:
			*d = r.values[i].(bool)
		case *time.Time:
			*d = r.values[i].(time.Time)
		case **time.Time:
			value, _ := r.values[i].(*time.Time)
			*d = value
		default:
			panic("unsupported scan destination")
		}
	}
	return nil
}

type fakeRows struct {
	rows   [][]any
	index  int
	err    error
	closed bool
}

var _ core_postgres_pool.Rows = (*fakeRows)(nil)

func (r *fakeRows) Close() {
	r.closed = true
}

func (r *fakeRows) Err() error {
	return r.err
}

func (r *fakeRows) Next() bool {
	return r.index < len(r.rows)
}

func (r *fakeRows) Scan(dest ...any) error {
	row := (&fakeRow{values: r.rows[r.index]}).Scan(dest...)
	r.index++
	return row
}

type fakeCommandTag struct {
	rowsAffected int64
}

var _ core_postgres_pool.CommandTag = (*fakeCommandTag)(nil)

func (t fakeCommandTag) RowsAffected() int64 {
	return t.rowsAffected
}

func TestCreateTaskMapsForeignKeyViolationToNotFound(t *testing.T) {
	pool := &fakePool{
		queryRow: &fakeRow{
			err: core_postgres_pool.ErrViolatesForeignKey,
		},
	}
	repository := NewTaskRepository(pool)

	_, err := repository.CreateTask(context.Background(), newRepositoryTask(0, 7, "valid title"))

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "user with id='7'") {
		t.Fatalf("expected author id context, got %q", err.Error())
	}
}

func TestGetTaskMapsNoRowsToNotFound(t *testing.T) {
	pool := &fakePool{
		queryRow: &fakeRow{
			err: core_postgres_pool.ErrNoRows,
		},
	}
	repository := NewTaskRepository(pool)

	_, err := repository.GetTask(context.Background(), 15)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "task with id='15'") {
		t.Fatalf("expected task id context, got %q", err.Error())
	}
}

func TestDeleteTaskMapsZeroRowsAffectedToNotFound(t *testing.T) {
	pool := &fakePool{
		execTag: fakeCommandTag{rowsAffected: 0},
	}
	repository := NewTaskRepository(pool)

	err := repository.DeleteTask(context.Background(), 15)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "task with id='15'") {
		t.Fatalf("expected task id context, got %q", err.Error())
	}
}

func TestPatchTaskMapsNoRowsToConflict(t *testing.T) {
	pool := &fakePool{
		queryRow: &fakeRow{
			err: core_postgres_pool.ErrNoRows,
		},
	}
	repository := NewTaskRepository(pool)

	_, err := repository.PatchTask(context.Background(), 15, newRepositoryTask(15, 1, "valid title"))

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if !strings.Contains(err.Error(), "concurrently accessed") {
		t.Fatalf("expected optimistic lock context, got %q", err.Error())
	}
}

func TestGetTasksAddsAuthorFilterWhenUserIDIsPresent(t *testing.T) {
	pool := &fakePool{
		queryRows: &fakeRows{},
	}
	repository := NewTaskRepository(pool)
	userID := 7
	limit := 20
	offset := 40

	_, err := repository.GetTasks(context.Background(), &userID, &limit, &offset)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(pool.querySQL, "WHERE author_user_id=$3") {
		t.Fatalf("expected author filter in query, got %q", pool.querySQL)
	}
	if len(pool.queryArgs) != 3 {
		t.Fatalf("expected 3 query args, got %d", len(pool.queryArgs))
	}
	if pool.queryArgs[2] != &userID {
		t.Fatalf("expected third query arg to be userID pointer")
	}
}

func newRepositoryTask(id int, authorUserID int, title string) domain.Task {
	createdAt := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	return domain.NewTask(
		id,
		1,
		title,
		stringPtr("description"),
		false,
		createdAt,
		nil,
		authorUserID,
	)
}

func stringPtr(value string) *string {
	return &value
}
