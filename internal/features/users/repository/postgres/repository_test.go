package users_postgres_repository

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

func TestGetUserMapsNoRowsToNotFound(t *testing.T) {
	pool := &fakePool{
		queryRow: &fakeRow{err: core_postgres_pool.ErrNoRows},
	}
	repository := NewUserRepository(pool)

	_, err := repository.GetUser(context.Background(), 15)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "user with id='15'") {
		t.Fatalf("expected user id context, got %q", err.Error())
	}
}

func TestDeleteUserMapsZeroRowsAffectedToNotFound(t *testing.T) {
	pool := &fakePool{
		execTag: fakeCommandTag{rowsAffected: 0},
	}
	repository := NewUserRepository(pool)

	err := repository.DeleteUser(context.Background(), 15)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "user with id='15'") {
		t.Fatalf("expected user id context, got %q", err.Error())
	}
}

func TestPatchUserMapsNoRowsToConflict(t *testing.T) {
	pool := &fakePool{
		queryRow: &fakeRow{err: core_postgres_pool.ErrNoRows},
	}
	repository := NewUserRepository(pool)

	_, err := repository.PatchUser(context.Background(), 15, newRepositoryUser(15, "Ivan Ivanov"))

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if !strings.Contains(err.Error(), "concurrently accessed") {
		t.Fatalf("expected optimistic lock context, got %q", err.Error())
	}
}

func TestGetUsersUsesPaginationArguments(t *testing.T) {
	pool := &fakePool{
		queryRows: &fakeRows{},
	}
	repository := NewUserRepository(pool)
	limit := 20
	offset := 40

	_, err := repository.GetUsers(context.Background(), &limit, &offset)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(pool.querySQL, "LIMIT $1") {
		t.Fatalf("expected LIMIT placeholder, got %q", pool.querySQL)
	}
	if !strings.Contains(pool.querySQL, "OFFSET $2") {
		t.Fatalf("expected OFFSET placeholder, got %q", pool.querySQL)
	}
	if len(pool.queryArgs) != 2 {
		t.Fatalf("expected 2 query args, got %d", len(pool.queryArgs))
	}
	if pool.queryArgs[0] != &limit || pool.queryArgs[1] != &offset {
		t.Fatalf("expected limit and offset pointers as query args")
	}
}

func TestCreateUserScansReturnedUser(t *testing.T) {
	phoneNumber := "+79998887766"
	pool := &fakePool{
		queryRow: &fakeRow{
			values: []any{10, 1, "Ivan Ivanov", &phoneNumber},
		},
	}
	repository := NewUserRepository(pool)

	user, err := repository.CreateUser(context.Background(), newRepositoryUser(0, "Ivan Ivanov"))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != 10 {
		t.Fatalf("expected user id %d, got %d", 10, user.ID)
	}
	if user.PhoneNumber == nil || *user.PhoneNumber != phoneNumber {
		t.Fatalf("expected phone number %q, got %v", phoneNumber, user.PhoneNumber)
	}
}

func newRepositoryUser(id int, fullName string) domain.User {
	return domain.NewUser(
		id,
		1,
		fullName,
		userStringPtr("+79998887766"),
	)
}

func userStringPtr(value string) *string {
	return &value
}
