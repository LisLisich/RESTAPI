package identity_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestLoginCreatesHashedBrowserSession(t *testing.T) {
	now := time.Date(2026, time.July, 24, 15, 0, 0, 0, time.UTC)
	account, err := domain.NewAccountUninitialized(42, "user@example.com")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	if err := account.VerifyEmail(now.Add(-time.Hour)); err != nil {
		t.Fatalf("activate account: %v", err)
	}
	repository := &fakeRegistrationRepository{
		passwordAccount: PasswordAccount{
			Account:      account,
			PasswordHash: "argon2id-hash",
		},
	}
	hasher := &fakePasswordHasher{verifyResult: true}
	issuer := &fakeTokenIssuer{
		tokens: []IssuedToken{
			{Raw: "session-token", Hash: []byte("session-hash")},
			{Raw: "csrf-token", Hash: []byte("csrf-hash")},
		},
	}
	service := NewIdentityService(repository, hasher, issuer, func() time.Time {
		return now
	})

	session, err := service.Login(
		context.Background(),
		LoginInput{
			Email:    " USER@Example.COM ",
			Password: "strong-password",
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repository.gotEmail != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", repository.gotEmail)
	}
	if !hasher.verifyCalled || hasher.encodedHash != "argon2id-hash" {
		t.Fatal("expected stored Argon2id hash to be verified")
	}
	if session.Token != "session-token" || session.CSRFToken != "csrf-token" {
		t.Fatalf("expected raw tokens in browser session, got %+v", session)
	}
	if !repository.createSessionCalled {
		t.Fatal("expected session persistence")
	}
	if string(repository.createdSession.TokenHash) != "session-hash" {
		t.Fatalf("expected session hash, got %q", repository.createdSession.TokenHash)
	}
	if string(repository.createdSession.CSRFHash) != "csrf-hash" {
		t.Fatalf("expected csrf hash, got %q", repository.createdSession.CSRFHash)
	}
	if repository.createdSession.AccountUserID != 42 {
		t.Fatalf("expected account user id 42, got %d", repository.createdSession.AccountUserID)
	}
	if !session.ExpiresAt.Equal(now.Add(12 * time.Hour)) {
		t.Fatalf("unexpected session expiry %v", session.ExpiresAt)
	}
}

func TestLoginDoesNotRevealWhetherEmailOrPasswordIsWrong(t *testing.T) {
	tests := []struct {
		name       string
		repository *fakeRegistrationRepository
		hasher     *fakePasswordHasher
	}{
		{
			name: "unknown email",
			repository: &fakeRegistrationRepository{
				getPasswordErr: core_errors.ErrNotFound,
			},
			hasher: &fakePasswordHasher{},
		},
		{
			name: "wrong password",
			repository: &fakeRegistrationRepository{
				passwordAccount: PasswordAccount{
					Account: domain.Account{
						UserID: 42,
						Email:  "user@example.com",
						Status: domain.AccountStatusActive,
					},
					PasswordHash: "argon2id-hash",
				},
			},
			hasher: &fakePasswordHasher{verifyResult: false},
		},
		{
			name: "unverified account",
			repository: &fakeRegistrationRepository{
				passwordAccount: PasswordAccount{
					Account: domain.Account{
						UserID: 42,
						Email:  "user@example.com",
						Status: domain.AccountStatusPendingVerification,
					},
					PasswordHash: "argon2id-hash",
				},
			},
			hasher: &fakePasswordHasher{verifyResult: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewIdentityService(
				tt.repository,
				tt.hasher,
				&fakeTokenIssuer{},
				time.Now,
			)

			_, err := service.Login(
				context.Background(),
				LoginInput{Email: "user@example.com", Password: "strong-password"},
			)

			if !errors.Is(err, core_errors.ErrUnauthorized) {
				t.Fatalf("expected ErrUnauthorized, got %v", err)
			}
			if tt.repository.createSessionCalled {
				t.Fatal("must not create session for rejected login")
			}
		})
	}
}
