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
)

type fakeAccountRow struct {
	userID          int
	email           string
	status          domain.AccountStatus
	emailVerifiedAt *time.Time
	err             error
}

func (r fakeAccountRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*int) = r.userID
	*dest[1].(*string) = r.email
	*dest[2].(*domain.AccountStatus) = r.status
	*dest[3].(**time.Time) = r.emailVerifiedAt
	return nil
}

func TestVerifyEmailConsumesTokenAndActivatesAccountAtomically(t *testing.T) {
	verifiedAt := time.Date(2026, time.July, 24, 14, 0, 0, 0, time.UTC)
	pool := &fakePool{
		queryRow: fakeAccountRow{
			userID:          42,
			email:           "user@example.com",
			status:          domain.AccountStatusActive,
			emailVerifiedAt: &verifiedAt,
		},
	}
	repository := NewIdentityRepository(pool)

	account, err := repository.VerifyEmail(
		context.Background(),
		[]byte("verification-hash"),
		verifiedAt,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if account.UserID != 42 || account.Status != domain.AccountStatusActive {
		t.Fatalf("expected active account 42, got %+v", account)
	}
	if !strings.Contains(pool.queryRowSQL, "UPDATE todoapp.account_tokens") {
		t.Fatalf("expected token update, got %q", pool.queryRowSQL)
	}
	if !strings.Contains(pool.queryRowSQL, "UPDATE todoapp.accounts") {
		t.Fatalf("expected account update, got %q", pool.queryRowSQL)
	}
	if !strings.Contains(pool.queryRowSQL, "consumed_at IS NULL") {
		t.Fatalf("expected one-time token guard, got %q", pool.queryRowSQL)
	}
	if !strings.Contains(pool.queryRowSQL, "expires_at > $2") {
		t.Fatalf("expected expiry guard, got %q", pool.queryRowSQL)
	}
}

func TestVerifyEmailRejectsUnknownExpiredOrConsumedToken(t *testing.T) {
	pool := &fakePool{
		queryRow: fakeAccountRow{err: core_postgres_pool.ErrNoRows},
	}
	repository := NewIdentityRepository(pool)

	_, err := repository.VerifyEmail(
		context.Background(),
		[]byte("invalid-hash"),
		time.Now(),
	)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
