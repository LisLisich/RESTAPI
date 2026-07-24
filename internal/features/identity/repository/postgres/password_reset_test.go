package identity_postgres_repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
)

func TestRequestPasswordResetReplacesTokenAndCreatesOutboxAtomically(t *testing.T) {
	tx := &fakeTx{}
	repository := NewIdentityRepository(&fakePool{tx: tx})
	request := identity_service.PasswordResetRequest{
		AccountUserID: 42,
		Email:         "user@example.com",
		Token: identity_service.IssuedToken{
			Raw:  "raw-reset-token",
			Hash: []byte("reset-token-hash"),
		},
		CreatedAt: time.Date(2026, time.July, 25, 10, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, time.July, 25, 11, 0, 0, 0, time.UTC),
	}

	err := repository.RequestPasswordReset(context.Background(), request)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !tx.committed || tx.rolledBack {
		t.Fatal("expected committed transaction")
	}
	if len(tx.execSQL) != 3 {
		t.Fatalf("expected three statements, got %d", len(tx.execSQL))
	}
	if !strings.Contains(tx.execSQL[0], "UPDATE todoapp.account_tokens") {
		t.Fatalf("expected previous token invalidation, got %q", tx.execSQL[0])
	}
	if !strings.Contains(tx.execSQL[1], "INSERT INTO todoapp.account_tokens") {
		t.Fatalf("expected reset token insert, got %q", tx.execSQL[1])
	}
	if !strings.Contains(tx.execSQL[2], "INSERT INTO todoapp.outbox_events") {
		t.Fatalf("expected outbox insert, got %q", tx.execSQL[2])
	}
	if string(tx.execArgs[1][2].([]byte)) != "reset-token-hash" {
		t.Fatalf("expected reset token hash, got %v", tx.execArgs[1][2])
	}
	payload := string(tx.execArgs[2][3].([]byte))
	if !strings.Contains(payload, "raw-reset-token") {
		t.Fatalf("expected raw token in outbox payload, got %q", payload)
	}
}

func TestRequestPasswordResetRollsBackOnOutboxFailure(t *testing.T) {
	tx := &fakeTx{
		execErrAt: 3,
		execErr:   errors.New("outbox unavailable"),
	}
	repository := NewIdentityRepository(&fakePool{tx: tx})

	err := repository.RequestPasswordReset(
		context.Background(),
		identity_service.PasswordResetRequest{
			AccountUserID: 42,
			Email:         "user@example.com",
			Token: identity_service.IssuedToken{
				Raw:  "raw-token",
				Hash: []byte("token-hash"),
			},
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
		},
	)

	if err == nil {
		t.Fatal("expected error")
	}
	if tx.committed || !tx.rolledBack {
		t.Fatal("expected rollback without commit")
	}
}

type fakeResetPasswordRow struct {
	userID int
	err    error
}

func (r fakeResetPasswordRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*int) = r.userID
	return nil
}

func TestResetPasswordConsumesTokenUpdatesCredentialAndRevokesSessions(t *testing.T) {
	pool := &fakePool{queryRow: fakeResetPasswordRow{userID: 42}}
	repository := NewIdentityRepository(pool)
	resetAt := time.Date(2026, time.July, 25, 11, 0, 0, 0, time.UTC)

	err := repository.ResetPassword(
		context.Background(),
		[]byte("reset-token-hash"),
		"new-argon2id-hash",
		resetAt,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	for _, table := range []string{
		"todoapp.account_tokens",
		"todoapp.password_credentials",
		"todoapp.sessions",
	} {
		if !strings.Contains(pool.queryRowSQL, table) {
			t.Fatalf("expected query to update %s, got %q", table, pool.queryRowSQL)
		}
	}
	if !strings.Contains(pool.queryRowSQL, "consumed_at IS NULL") ||
		!strings.Contains(pool.queryRowSQL, "expires_at > $3") {
		t.Fatalf("expected one-time unexpired token guard, got %q", pool.queryRowSQL)
	}
}

func TestResetPasswordRejectsInvalidExpiredOrConsumedToken(t *testing.T) {
	pool := &fakePool{
		queryRow: fakeResetPasswordRow{err: core_postgres_pool.ErrNoRows},
	}
	repository := NewIdentityRepository(pool)

	err := repository.ResetPassword(
		context.Background(),
		[]byte("invalid-token-hash"),
		"new-argon2id-hash",
		time.Now(),
	)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
