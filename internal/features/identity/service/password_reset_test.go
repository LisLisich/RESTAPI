package identity_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestRequestPasswordResetCreatesExpiringOutboxRequest(t *testing.T) {
	now := time.Date(2026, time.July, 25, 10, 0, 0, 0, time.UTC)
	repository := &fakeRegistrationRepository{
		passwordAccount: PasswordAccount{
			Account: domain.Account{
				UserID: 42,
				Email:  "user@example.com",
				Status: domain.AccountStatusActive,
			},
		},
	}
	issuer := &fakeTokenIssuer{
		token: IssuedToken{
			Raw:  "raw-reset-token",
			Hash: []byte("reset-token-hash"),
		},
	}
	service := NewIdentityService(repository, &fakePasswordHasher{}, issuer, func() time.Time {
		return now
	})

	err := service.RequestPasswordReset(context.Background(), " USER@Example.COM ")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repository.gotEmail != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", repository.gotEmail)
	}
	if !repository.passwordResetCalled {
		t.Fatal("expected password reset request persistence")
	}
	request := repository.passwordResetRequest
	if request.AccountUserID != 42 || request.Email != "user@example.com" {
		t.Fatalf("unexpected reset account: %+v", request)
	}
	if request.Token.Raw != "raw-reset-token" || string(request.Token.Hash) != "reset-token-hash" {
		t.Fatalf("unexpected reset token: %+v", request.Token)
	}
	if !request.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("expected one-hour expiry, got %v", request.ExpiresAt)
	}
}

func TestRequestPasswordResetDoesNotRevealUnknownOrInactiveAccount(t *testing.T) {
	tests := []struct {
		name       string
		repository *fakeRegistrationRepository
	}{
		{
			name: "unknown account",
			repository: &fakeRegistrationRepository{
				getPasswordErr: core_errors.ErrNotFound,
			},
		},
		{
			name: "inactive account",
			repository: &fakeRegistrationRepository{
				passwordAccount: PasswordAccount{
					Account: domain.Account{
						UserID: 42,
						Email:  "user@example.com",
						Status: domain.AccountStatusPendingVerification,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issuer := &fakeTokenIssuer{}
			service := NewIdentityService(tt.repository, &fakePasswordHasher{}, issuer, time.Now)

			err := service.RequestPasswordReset(context.Background(), "user@example.com")

			if err != nil {
				t.Fatalf("expected indistinguishable success, got %v", err)
			}
			if issuer.issueCalled || tt.repository.passwordResetCalled {
				t.Fatal("must not create reset token for unknown or inactive account")
			}
		})
	}
}

func TestResetPasswordHashesPasswordAndToken(t *testing.T) {
	now := time.Date(2026, time.July, 25, 11, 0, 0, 0, time.UTC)
	repository := &fakeRegistrationRepository{}
	hasher := &fakePasswordHasher{hash: "new-argon2id-hash"}
	issuer := &fakeTokenIssuer{hash: []byte("reset-token-hash")}
	service := NewIdentityService(repository, hasher, issuer, func() time.Time {
		return now
	})

	err := service.ResetPassword(
		context.Background(),
		"raw-reset-token",
		"new-strong-password",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !hasher.hashCalled || hasher.password != "new-strong-password" {
		t.Fatal("expected new password to be hashed")
	}
	if !repository.resetPasswordCalled {
		t.Fatal("expected repository password reset")
	}
	if string(repository.resetTokenHash) != "reset-token-hash" {
		t.Fatalf("expected token hash, got %q", repository.resetTokenHash)
	}
	if repository.resetPasswordHash != "new-argon2id-hash" {
		t.Fatalf("expected new password hash, got %q", repository.resetPasswordHash)
	}
	if !repository.resetAt.Equal(now) {
		t.Fatalf("expected reset time %v, got %v", now, repository.resetAt)
	}
}

func TestResetPasswordRejectsInvalidInputBeforeHashing(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		password string
	}{
		{name: "blank token", token: " ", password: "new-strong-password"},
		{name: "short password", token: "raw-token", password: "short"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakeRegistrationRepository{}
			hasher := &fakePasswordHasher{}
			service := NewIdentityService(repository, hasher, &fakeTokenIssuer{}, time.Now)

			err := service.ResetPassword(context.Background(), tt.token, tt.password)

			if !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if hasher.hashCalled || repository.resetPasswordCalled {
				t.Fatal("invalid reset input must be rejected before dependencies")
			}
		})
	}
}
