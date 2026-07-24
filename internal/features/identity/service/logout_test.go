package identity_service

import (
	"context"
	"errors"
	"testing"
	"time"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestLogoutRevokesHashedSession(t *testing.T) {
	now := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)
	repository := &fakeRegistrationRepository{}
	issuer := &fakeTokenIssuer{hash: []byte("session-hash")}
	service := NewIdentityService(repository, &fakePasswordHasher{}, issuer, func() time.Time {
		return now
	})

	err := service.Logout(context.Background(), "raw-session-token")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.revokeSessionCalled {
		t.Fatal("expected session revocation")
	}
	if string(repository.revokeSessionHash) != "session-hash" {
		t.Fatalf("expected session hash, got %q", repository.revokeSessionHash)
	}
	if !repository.revokeSessionAt.Equal(now) {
		t.Fatalf("expected revocation time %v, got %v", now, repository.revokeSessionAt)
	}
}

func TestLogoutRejectsBlankSessionToken(t *testing.T) {
	repository := &fakeRegistrationRepository{}
	service := NewIdentityService(
		repository,
		&fakePasswordHasher{},
		&fakeTokenIssuer{},
		time.Now,
	)

	err := service.Logout(context.Background(), " ")

	if !errors.Is(err, core_errors.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
	if repository.revokeSessionCalled {
		t.Fatal("repository must not be called without session token")
	}
}
