package domain

import (
	"errors"
	"testing"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestNewAccountUninitializedNormalizesEmail(t *testing.T) {
	account, err := NewAccountUninitialized(42, "  User.Name@Example.COM ")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if account.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", account.UserID)
	}
	if account.Email != "user.name@example.com" {
		t.Fatalf("expected normalized email, got %q", account.Email)
	}
	if account.Status != AccountStatusPendingVerification {
		t.Fatalf("expected pending status, got %q", account.Status)
	}
}

func TestNewAccountUninitializedRejectsInvalidEmail(t *testing.T) {
	_, err := NewAccountUninitialized(42, "not-an-email")

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestAccountVerifyEmailActivatesAccount(t *testing.T) {
	account, err := NewAccountUninitialized(42, "user@example.com")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	verifiedAt := time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC)

	if err := account.VerifyEmail(verifiedAt); err != nil {
		t.Fatalf("verify account: %v", err)
	}

	if account.Status != AccountStatusActive {
		t.Fatalf("expected active status, got %q", account.Status)
	}
	if account.EmailVerifiedAt == nil || !account.EmailVerifiedAt.Equal(verifiedAt) {
		t.Fatalf("expected verification time %v, got %v", verifiedAt, account.EmailVerifiedAt)
	}
}

func TestValidatePasswordUsesRuneLength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "minimum latin length", password: "long-password", wantErr: false},
		{name: "minimum unicode length", password: "пароль-длинный", wantErr: false},
		{name: "too short", password: "short", wantErr: true},
		{name: "too long", password: string(make([]rune, 129)), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
