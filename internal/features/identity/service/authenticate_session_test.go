package identity_service

import (
	"context"
	"errors"
	"testing"
	"time"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestAuthenticateSessionReturnsPrincipalForActiveSession(t *testing.T) {
	now := time.Date(2026, time.July, 24, 16, 0, 0, 0, time.UTC)
	repository := &fakeRegistrationRepository{
		authSession: StoredAuthenticationSession{
			AccountUserID: 42,
			CSRFHash:      []byte("csrf-hash"),
		},
	}
	issuer := &fakeTokenIssuer{hash: []byte("session-hash")}
	service := NewIdentityService(repository, &fakePasswordHasher{}, issuer, func() time.Time {
		return now
	})

	principal, err := service.AuthenticateSession(
		context.Background(),
		"raw-session-token",
		"",
		false,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if principal.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", principal.UserID)
	}
	if string(repository.sessionHash) != "session-hash" {
		t.Fatalf("expected session hash, got %q", repository.sessionHash)
	}
	if !repository.sessionNow.Equal(now) {
		t.Fatalf("expected lookup time %v, got %v", now, repository.sessionNow)
	}
}

func TestAuthenticateSessionRequiresMatchingCSRFForMutation(t *testing.T) {
	repository := &fakeRegistrationRepository{
		authSession: StoredAuthenticationSession{
			AccountUserID: 42,
			CSRFHash:      []byte("csrf-hash"),
		},
	}
	issuer := &fakeTokenIssuer{}
	service := NewIdentityService(repository, &fakePasswordHasher{}, issuer, time.Now)

	issuer.hash = []byte("session-hash")
	_, err := service.AuthenticateSession(
		context.Background(),
		"raw-session-token",
		"",
		true,
	)
	if !errors.Is(err, core_errors.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for missing CSRF token, got %v", err)
	}

	issuer.hash = []byte("wrong-csrf-hash")
	_, err = service.AuthenticateSession(
		context.Background(),
		"raw-session-token",
		"wrong-csrf-token",
		true,
	)
	if !errors.Is(err, core_errors.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for mismatched CSRF token, got %v", err)
	}

	issuer.hash = []byte("csrf-hash")
	principal, err := service.AuthenticateSession(
		context.Background(),
		"raw-session-token",
		"valid-csrf-token",
		true,
	)
	if err != nil {
		t.Fatalf("expected matching CSRF token, got %v", err)
	}
	if principal.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", principal.UserID)
	}
}

func TestAuthenticateSessionRejectsMissingOrUnknownSession(t *testing.T) {
	tests := []struct {
		name       string
		rawSession string
		repository *fakeRegistrationRepository
	}{
		{
			name:       "missing cookie",
			rawSession: "",
			repository: &fakeRegistrationRepository{},
		},
		{
			name:       "unknown session",
			rawSession: "unknown-token",
			repository: &fakeRegistrationRepository{
				sessionErr: core_errors.ErrNotFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewIdentityService(
				tt.repository,
				&fakePasswordHasher{},
				&fakeTokenIssuer{hash: []byte("session-hash")},
				time.Now,
			)

			_, err := service.AuthenticateSession(
				context.Background(),
				tt.rawSession,
				"",
				false,
			)

			if !errors.Is(err, core_errors.ErrUnauthorized) {
				t.Fatalf("expected ErrUnauthorized, got %v", err)
			}
		})
	}
}
