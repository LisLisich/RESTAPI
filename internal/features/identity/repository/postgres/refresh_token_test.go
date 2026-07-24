package identity_postgres_repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"
	identity_service "github.com/LisLisich/fintask/internal/features/identity/service"
)

func TestCreateRefreshTokenPersistsHash(t *testing.T) {
	pool := &fakePool{}
	repository := NewIdentityRepository(pool)
	token := identity_service.StoredRefreshToken{
		AccountUserID: 42,
		FamilyID:      "family-id",
		TokenHash:     []byte("refresh-hash"),
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:     time.Now(),
	}

	err := repository.CreateRefreshToken(context.Background(), token)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(pool.execSQL, "INSERT INTO todoapp.refresh_tokens") {
		t.Fatalf("unexpected query %q", pool.execSQL)
	}
	if string(pool.execArgs[2].([]byte)) != "refresh-hash" {
		t.Fatalf("expected hash, got %v", pool.execArgs[2])
	}
}

func TestRotateRefreshTokenConsumesAndReplacesAtomically(t *testing.T) {
	pool := &fakePool{queryRow: fakeResetPasswordRow{userID: 42}}
	repository := NewIdentityRepository(pool)

	userID, err := repository.RotateRefreshToken(
		context.Background(),
		[]byte("old-hash"),
		identity_service.StoredRefreshToken{
			TokenHash: []byte("new-hash"),
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
			CreatedAt: time.Now(),
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID != 42 {
		t.Fatalf("expected user id 42, got %d", userID)
	}
	if !strings.Contains(pool.queryRowSQL, "used_at IS NULL") ||
		!strings.Contains(pool.queryRowSQL, "INSERT INTO todoapp.refresh_tokens") {
		t.Fatalf("expected atomic rotation, got %q", pool.queryRowSQL)
	}
}

func TestRotateRefreshTokenRejectsReuse(t *testing.T) {
	pool := &fakePool{
		queryRow: fakeResetPasswordRow{err: core_postgres_pool.ErrNoRows},
	}
	repository := NewIdentityRepository(pool)

	_, err := repository.RotateRefreshToken(
		context.Background(),
		[]byte("used-hash"),
		identity_service.StoredRefreshToken{
			TokenHash: []byte("new-hash"),
			ExpiresAt: time.Now().Add(time.Hour),
			CreatedAt: time.Now(),
		},
	)

	if !errors.Is(err, core_errors.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
