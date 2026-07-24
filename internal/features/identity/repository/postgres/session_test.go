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

type fakePasswordAccountRow struct {
	account      domain.Account
	passwordHash string
	err          error
}

type fakeAuthenticationSessionRow struct {
	accountUserID int
	csrfHash      []byte
	err           error
}

func (r fakeAuthenticationSessionRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*int) = r.accountUserID
	*dest[1].(*[]byte) = r.csrfHash
	return nil
}

func (r fakePasswordAccountRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*int) = r.account.UserID
	*dest[1].(*string) = r.account.Email
	*dest[2].(*domain.AccountStatus) = r.account.Status
	*dest[3].(**time.Time) = r.account.EmailVerifiedAt
	*dest[4].(*string) = r.passwordHash
	return nil
}

func TestGetPasswordAccountLoadsCredentialByNormalizedEmail(t *testing.T) {
	verifiedAt := time.Date(2026, time.July, 24, 14, 0, 0, 0, time.UTC)
	pool := &fakePool{
		queryRow: fakePasswordAccountRow{
			account: domain.Account{
				UserID:          42,
				Email:           "user@example.com",
				Status:          domain.AccountStatusActive,
				EmailVerifiedAt: &verifiedAt,
			},
			passwordHash: "argon2id-hash",
		},
	}
	repository := NewIdentityRepository(pool)

	passwordAccount, err := repository.GetPasswordAccount(
		context.Background(),
		"user@example.com",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if passwordAccount.Account.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", passwordAccount.Account.UserID)
	}
	if passwordAccount.PasswordHash != "argon2id-hash" {
		t.Fatalf("expected password hash, got %q", passwordAccount.PasswordHash)
	}
	if !strings.Contains(pool.queryRowSQL, "todoapp.password_credentials") {
		t.Fatalf("expected credentials join, got %q", pool.queryRowSQL)
	}
}

func TestGetPasswordAccountMapsNoRowsToNotFound(t *testing.T) {
	pool := &fakePool{
		queryRow: fakePasswordAccountRow{err: core_postgres_pool.ErrNoRows},
	}
	repository := NewIdentityRepository(pool)

	_, err := repository.GetPasswordAccount(context.Background(), "unknown@example.com")

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateSessionPersistsOnlyTokenHashes(t *testing.T) {
	pool := &fakePool{}
	repository := NewIdentityRepository(pool)
	session := identity_service.StoredSession{
		AccountUserID: 42,
		TokenHash:     []byte("session-hash"),
		CSRFHash:      []byte("csrf-hash"),
		CreatedAt:     time.Date(2026, time.July, 24, 15, 0, 0, 0, time.UTC),
		ExpiresAt:     time.Date(2026, time.July, 25, 3, 0, 0, 0, time.UTC),
	}

	err := repository.CreateSession(context.Background(), session)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(pool.execSQL, "INSERT INTO todoapp.sessions") {
		t.Fatalf("expected session insert, got %q", pool.execSQL)
	}
	if string(pool.execArgs[1].([]byte)) != "session-hash" {
		t.Fatalf("expected session hash, got %v", pool.execArgs[1])
	}
	if string(pool.execArgs[2].([]byte)) != "csrf-hash" {
		t.Fatalf("expected csrf hash, got %v", pool.execArgs[2])
	}
}

func TestGetSessionLoadsOnlyActiveUnexpiredAccountSession(t *testing.T) {
	pool := &fakePool{
		queryRow: fakeAuthenticationSessionRow{
			accountUserID: 42,
			csrfHash:      []byte("csrf-hash"),
		},
	}
	repository := NewIdentityRepository(pool)
	now := time.Date(2026, time.July, 24, 16, 0, 0, 0, time.UTC)

	session, err := repository.GetSession(
		context.Background(),
		[]byte("session-hash"),
		now,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if session.AccountUserID != 42 {
		t.Fatalf("expected user id 42, got %d", session.AccountUserID)
	}
	if !strings.Contains(pool.queryRowSQL, "session.revoked_at IS NULL") {
		t.Fatalf("expected revoked session guard, got %q", pool.queryRowSQL)
	}
	if !strings.Contains(pool.queryRowSQL, "session.expires_at > $2") {
		t.Fatalf("expected expiry guard, got %q", pool.queryRowSQL)
	}
	if !strings.Contains(pool.queryRowSQL, "account.status = 'active'") {
		t.Fatalf("expected active account guard, got %q", pool.queryRowSQL)
	}
}

func TestGetSessionMapsMissingSessionToNotFound(t *testing.T) {
	pool := &fakePool{
		queryRow: fakeAuthenticationSessionRow{err: core_postgres_pool.ErrNoRows},
	}
	repository := NewIdentityRepository(pool)

	_, err := repository.GetSession(
		context.Background(),
		[]byte("unknown-hash"),
		time.Now(),
	)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
