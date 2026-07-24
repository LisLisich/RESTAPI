package identity_postgres_repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
)

type fakePool struct {
	tx       *fakeTx
	beginErr error
}

func (p *fakePool) Begin(context.Context) (core_postgres_pool.Tx, error) {
	if p.beginErr != nil {
		return nil, p.beginErr
	}
	return p.tx, nil
}

func (*fakePool) Query(context.Context, string, ...any) (core_postgres_pool.Rows, error) {
	panic("unexpected pool Query call")
}

func (*fakePool) QueryRow(context.Context, string, ...any) core_postgres_pool.Row {
	panic("unexpected pool QueryRow call")
}

func (*fakePool) Exec(context.Context, string, ...any) (core_postgres_pool.CommandTag, error) {
	panic("unexpected pool Exec call")
}

func (*fakePool) Close() {}

func (*fakePool) OpTimeout() time.Duration {
	return time.Second
}

type fakeTx struct {
	userID     int
	querySQL   string
	queryArgs  []any
	execSQL    []string
	execArgs   [][]any
	execErrAt  int
	execErr    error
	committed  bool
	rolledBack bool
}

func (*fakeTx) Query(context.Context, string, ...any) (core_postgres_pool.Rows, error) {
	panic("unexpected transaction Query call")
}

func (tx *fakeTx) QueryRow(_ context.Context, sql string, args ...any) core_postgres_pool.Row {
	tx.querySQL = sql
	tx.queryArgs = args
	return fakeRow{userID: tx.userID}
}

func (tx *fakeTx) Exec(
	_ context.Context,
	sql string,
	args ...any,
) (core_postgres_pool.CommandTag, error) {
	tx.execSQL = append(tx.execSQL, sql)
	tx.execArgs = append(tx.execArgs, args)
	if tx.execErrAt == len(tx.execSQL) {
		return nil, tx.execErr
	}
	return fakeCommandTag{}, nil
}

func (tx *fakeTx) Commit(context.Context) error {
	tx.committed = true
	return nil
}

func (tx *fakeTx) Rollback(context.Context) error {
	tx.rolledBack = true
	return nil
}

type fakeRow struct {
	userID int
}

func (r fakeRow) Scan(dest ...any) error {
	*dest[0].(*int) = r.userID
	return nil
}

type fakeCommandTag struct{}

func (fakeCommandTag) RowsAffected() int64 {
	return 1
}

func TestRegisterPersistsIdentityAtomically(t *testing.T) {
	tx := &fakeTx{userID: 42}
	repository := NewIdentityRepository(&fakePool{tx: tx})
	registration := newRegistration()

	account, err := repository.Register(context.Background(), registration)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if account.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", account.UserID)
	}
	if !tx.committed {
		t.Fatal("expected transaction commit")
	}
	if tx.rolledBack {
		t.Fatal("did not expect rollback after successful commit")
	}
	if !strings.Contains(tx.querySQL, "INSERT INTO todoapp.users") {
		t.Fatalf("expected user insert, got %q", tx.querySQL)
	}
	if len(tx.execSQL) != 4 {
		t.Fatalf("expected 4 dependent inserts, got %d", len(tx.execSQL))
	}
	expectedTables := []string{
		"todoapp.accounts",
		"todoapp.password_credentials",
		"todoapp.account_tokens",
		"todoapp.outbox_events",
	}
	for index, table := range expectedTables {
		if !strings.Contains(tx.execSQL[index], table) {
			t.Fatalf("expected insert into %s, got %q", table, tx.execSQL[index])
		}
	}
	if got := tx.execArgs[1][1]; got != registration.PasswordHash {
		t.Fatalf("expected password hash argument, got %v", got)
	}
	if got := tx.execArgs[2][2]; string(got.([]byte)) != "verification-hash" {
		t.Fatalf("expected token hash argument, got %v", got)
	}
	outboxPayload, ok := tx.execArgs[3][3].([]byte)
	if !ok {
		t.Fatalf("expected JSON payload bytes, got %T", tx.execArgs[3][3])
	}
	if !strings.Contains(string(outboxPayload), registration.VerificationToken.Raw) {
		t.Fatal("expected raw verification token only in outbox payload")
	}
}

func TestRegisterRollsBackWhenDependentInsertFails(t *testing.T) {
	tx := &fakeTx{
		userID:    42,
		execErrAt: 3,
		execErr:   errors.New("insert token"),
	}
	repository := NewIdentityRepository(&fakePool{tx: tx})

	_, err := repository.Register(context.Background(), newRegistration())

	if err == nil {
		t.Fatal("expected registration error")
	}
	if tx.committed {
		t.Fatal("must not commit incomplete registration")
	}
	if !tx.rolledBack {
		t.Fatal("expected rollback")
	}
}

func TestRegisterMapsUniqueEmailToConflict(t *testing.T) {
	tx := &fakeTx{
		userID:    42,
		execErrAt: 1,
		execErr:   core_postgres_pool.ErrViolatesUnique,
	}
	repository := NewIdentityRepository(&fakePool{tx: tx})

	_, err := repository.Register(context.Background(), newRegistration())

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if !tx.rolledBack {
		t.Fatal("expected rollback")
	}
}

func newRegistration() identity_service.Registration {
	account, err := domain.NewAccountUninitialized(0, "user@example.com")
	if err != nil {
		panic(err)
	}
	return identity_service.Registration{
		User:         domain.NewUserUninitialized("Ivan Ivanov", nil),
		Account:      account,
		PasswordHash: "argon2id-hash",
		VerificationToken: identity_service.IssuedToken{
			Raw:  "raw-verification-token",
			Hash: []byte("verification-hash"),
		},
		VerificationExpiresAt: time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC),
		CreatedAt:             time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC),
	}
}
