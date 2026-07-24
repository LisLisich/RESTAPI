package identity_service

import (
	"context"
	"errors"
	"testing"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestVerifyEmailConsumesHashedTokenAtCurrentTime(t *testing.T) {
	now := time.Date(2026, time.July, 24, 14, 0, 0, 0, time.UTC)
	repository := &fakeRegistrationRepository{}
	hasher := &fakePasswordHasher{}
	issuer := &fakeTokenIssuer{hash: []byte("verification-hash")}
	service := NewIdentityService(repository, hasher, issuer, func() time.Time {
		return now
	})

	account, err := service.VerifyEmail(context.Background(), "raw-token")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if account.Status != "active" {
		t.Fatalf("expected active account, got %q", account.Status)
	}
	if !issuer.hashCalled || issuer.rawToken != "raw-token" {
		t.Fatal("expected raw token to be hashed")
	}
	if !repository.verifyCalled {
		t.Fatal("expected repository verification")
	}
	if string(repository.verifyHash) != "verification-hash" {
		t.Fatalf("expected verification hash, got %q", repository.verifyHash)
	}
	if !repository.verifiedAt.Equal(now) {
		t.Fatalf("expected verification time %v, got %v", now, repository.verifiedAt)
	}
}

func TestVerifyEmailRejectsBlankToken(t *testing.T) {
	repository := &fakeRegistrationRepository{}
	issuer := &fakeTokenIssuer{}
	service := NewIdentityService(repository, &fakePasswordHasher{}, issuer, time.Now)

	_, err := service.VerifyEmail(context.Background(), "   ")

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if issuer.hashCalled {
		t.Fatal("blank token must not be hashed")
	}
	if repository.verifyCalled {
		t.Fatal("repository must not be called for blank token")
	}
}

func TestVerifyEmailPreservesInvalidExpiredTokenError(t *testing.T) {
	repository := &fakeRegistrationRepository{verifyErr: core_errors.ErrInvalidArgument}
	issuer := &fakeTokenIssuer{hash: []byte("verification-hash")}
	service := NewIdentityService(repository, &fakePasswordHasher{}, issuer, time.Now)

	_, err := service.VerifyEmail(context.Background(), "expired-token")

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
