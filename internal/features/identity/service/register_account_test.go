package identity_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

type fakeRegistrationRepository struct {
	registerCalled bool
	registration   Registration
	err            error
}

func (r *fakeRegistrationRepository) Register(
	_ context.Context,
	registration Registration,
) (domain.Account, error) {
	r.registerCalled = true
	r.registration = registration
	if r.err != nil {
		return domain.Account{}, r.err
	}

	account := registration.Account
	account.UserID = 42
	return account, nil
}

type fakePasswordHasher struct {
	hashCalled bool
	password   string
	hash       string
	err        error
}

func (h *fakePasswordHasher) Hash(password string) (string, error) {
	h.hashCalled = true
	h.password = password
	return h.hash, h.err
}

type fakeTokenIssuer struct {
	issueCalled bool
	token       IssuedToken
	err         error
}

func (i *fakeTokenIssuer) Issue() (IssuedToken, error) {
	i.issueCalled = true
	return i.token, i.err
}

func TestRegisterAccountBuildsAtomicRegistration(t *testing.T) {
	now := time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC)
	repository := &fakeRegistrationRepository{}
	hasher := &fakePasswordHasher{hash: "argon2id-hash"}
	issuer := &fakeTokenIssuer{
		token: IssuedToken{
			Raw:  "email-verification-token",
			Hash: []byte("token-hash"),
		},
	}
	service := NewIdentityService(repository, hasher, issuer, func() time.Time {
		return now
	})

	account, err := service.RegisterAccount(
		context.Background(),
		RegisterAccountInput{
			FullName: "Ivan Ivanov",
			Email:    "  USER@Example.COM ",
			Password: "strong-password",
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if account.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", account.UserID)
	}
	if account.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", account.Email)
	}
	if !hasher.hashCalled || hasher.password != "strong-password" {
		t.Fatal("expected password to be hashed")
	}
	if !issuer.issueCalled {
		t.Fatal("expected verification token to be issued")
	}
	if !repository.registerCalled {
		t.Fatal("expected repository registration")
	}
	if repository.registration.PasswordHash != "argon2id-hash" {
		t.Fatalf("expected password hash, got %q", repository.registration.PasswordHash)
	}
	if repository.registration.PasswordHash == "strong-password" {
		t.Fatal("plaintext password must not be passed to repository")
	}
	if repository.registration.VerificationToken.Raw != "email-verification-token" {
		t.Fatalf(
			"expected raw verification token, got %q",
			repository.registration.VerificationToken.Raw,
		)
	}
	wantExpiry := now.Add(24 * time.Hour)
	if !repository.registration.VerificationExpiresAt.Equal(wantExpiry) {
		t.Fatalf(
			"expected verification expiry %v, got %v",
			wantExpiry,
			repository.registration.VerificationExpiresAt,
		)
	}
}

func TestRegisterAccountRejectsInvalidInputBeforeDependencies(t *testing.T) {
	tests := []struct {
		name  string
		input RegisterAccountInput
	}{
		{
			name: "invalid full name",
			input: RegisterAccountInput{
				FullName: "Iv",
				Email:    "user@example.com",
				Password: "strong-password",
			},
		},
		{
			name: "invalid email",
			input: RegisterAccountInput{
				FullName: "Ivan Ivanov",
				Email:    "not-an-email",
				Password: "strong-password",
			},
		},
		{
			name: "short password",
			input: RegisterAccountInput{
				FullName: "Ivan Ivanov",
				Email:    "user@example.com",
				Password: "short",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakeRegistrationRepository{}
			hasher := &fakePasswordHasher{hash: "argon2id-hash"}
			issuer := &fakeTokenIssuer{}
			service := NewIdentityService(repository, hasher, issuer, time.Now)

			_, err := service.RegisterAccount(context.Background(), tt.input)

			if !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if hasher.hashCalled {
				t.Fatal("password hasher must not be called for invalid input")
			}
			if issuer.issueCalled {
				t.Fatal("token issuer must not be called for invalid input")
			}
			if repository.registerCalled {
				t.Fatal("repository must not be called for invalid input")
			}
		})
	}
}

func TestRegisterAccountPreservesConflict(t *testing.T) {
	repository := &fakeRegistrationRepository{err: core_errors.ErrConflict}
	hasher := &fakePasswordHasher{hash: "argon2id-hash"}
	issuer := &fakeTokenIssuer{
		token: IssuedToken{Raw: "raw-token", Hash: []byte("token-hash")},
	}
	service := NewIdentityService(repository, hasher, issuer, time.Now)

	_, err := service.RegisterAccount(
		context.Background(),
		RegisterAccountInput{
			FullName: "Ivan Ivanov",
			Email:    "user@example.com",
			Password: "strong-password",
		},
	)

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}
