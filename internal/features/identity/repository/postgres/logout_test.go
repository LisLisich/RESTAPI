package identity_postgres_repository

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRevokeSessionIsIdempotent(t *testing.T) {
	pool := &fakePool{}
	repository := NewIdentityRepository(pool)
	revokedAt := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)

	err := repository.RevokeSession(
		context.Background(),
		[]byte("session-hash"),
		revokedAt,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(pool.execSQL, "UPDATE todoapp.sessions") {
		t.Fatalf("expected session update, got %q", pool.execSQL)
	}
	if !strings.Contains(pool.execSQL, "revoked_at IS NULL") {
		t.Fatalf("expected idempotent revocation guard, got %q", pool.execSQL)
	}
	if string(pool.execArgs[0].([]byte)) != "session-hash" {
		t.Fatalf("expected session hash, got %v", pool.execArgs[0])
	}
}
