package identity_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestLoginAPIStoresRefreshHashAndReturnsTokenPair(t *testing.T) {
	now := time.Date(2026, time.July, 25, 14, 0, 0, 0, time.UTC)
	repository := &fakeRegistrationRepository{
		passwordAccount: PasswordAccount{
			Account: domain.Account{
				UserID: 42,
				Email:  "user@example.com",
				Status: domain.AccountStatusActive,
			},
			PasswordHash: "argon2id-hash",
		},
	}
	passwordHasher := &fakePasswordHasher{verifyResult: true}
	tokenIssuer := &fakeTokenIssuer{
		tokens: []IssuedToken{
			{Raw: "refresh-token", Hash: []byte("refresh-hash")},
			{Raw: "access-jti", Hash: []byte("unused-jti-hash")},
			{Raw: "family-id", Hash: []byte("unused-family-hash")},
		},
	}
	accessIssuer := &fakeAccessTokenIssuer{token: "signed-access-token"}
	service := NewIdentityService(
		repository,
		passwordHasher,
		tokenIssuer,
		func() time.Time { return now },
		WithAccessTokenIssuer(accessIssuer),
	)

	pair, err := service.LoginAPI(
		context.Background(),
		LoginInput{Email: "user@example.com", Password: "strong-password"},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pair.AccessToken != "signed-access-token" || pair.RefreshToken != "refresh-token" {
		t.Fatalf("unexpected token pair %+v", pair)
	}
	if !repository.createRefreshCalled {
		t.Fatal("expected refresh token persistence")
	}
	if string(repository.createdRefresh.TokenHash) != "refresh-hash" {
		t.Fatalf("expected refresh hash, got %q", repository.createdRefresh.TokenHash)
	}
	if repository.createdRefresh.AccountUserID != 42 {
		t.Fatalf("expected user id 42, got %d", repository.createdRefresh.AccountUserID)
	}
	if repository.createdRefresh.FamilyID != "family-id" {
		t.Fatalf("expected family id, got %q", repository.createdRefresh.FamilyID)
	}
	if !repository.createdRefresh.ExpiresAt.Equal(now.Add(30 * 24 * time.Hour)) {
		t.Fatalf("unexpected refresh expiry %v", repository.createdRefresh.ExpiresAt)
	}
	if !accessIssuer.called || accessIssuer.userID != 42 || accessIssuer.jwtID != "access-jti" {
		t.Fatal("expected Ed25519 access token issue")
	}
}

func TestRefreshAPIRotatesRefreshToken(t *testing.T) {
	now := time.Date(2026, time.July, 25, 15, 0, 0, 0, time.UTC)
	repository := &fakeRegistrationRepository{rotateRefreshUserID: 42}
	tokenIssuer := &fakeTokenIssuer{
		hash: []byte("old-refresh-hash"),
		tokens: []IssuedToken{
			{Raw: "new-refresh-token", Hash: []byte("new-refresh-hash")},
			{Raw: "new-access-jti", Hash: []byte("unused")},
		},
	}
	accessIssuer := &fakeAccessTokenIssuer{token: "new-access-token"}
	service := NewIdentityService(
		repository,
		&fakePasswordHasher{},
		tokenIssuer,
		func() time.Time { return now },
		WithAccessTokenIssuer(accessIssuer),
	)

	pair, err := service.RefreshAPI(context.Background(), "old-refresh-token")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.rotateRefreshCalled {
		t.Fatal("expected refresh rotation")
	}
	if string(repository.oldRefreshHash) != "old-refresh-hash" {
		t.Fatalf("unexpected old hash %q", repository.oldRefreshHash)
	}
	if string(repository.rotatedRefresh.TokenHash) != "new-refresh-hash" {
		t.Fatalf("unexpected new hash %q", repository.rotatedRefresh.TokenHash)
	}
	if pair.AccessToken != "new-access-token" || pair.RefreshToken != "new-refresh-token" {
		t.Fatalf("unexpected token pair %+v", pair)
	}
}

func TestRefreshAPIRejectsReusedToken(t *testing.T) {
	repository := &fakeRegistrationRepository{
		rotateRefreshErr: core_errors.ErrUnauthorized,
	}
	service := NewIdentityService(
		repository,
		&fakePasswordHasher{},
		&fakeTokenIssuer{
			hash:  []byte("old-hash"),
			token: IssuedToken{Raw: "new-token", Hash: []byte("new-hash")},
		},
		time.Now,
		WithAccessTokenIssuer(&fakeAccessTokenIssuer{}),
	)

	_, err := service.RefreshAPI(context.Background(), "reused-token")

	if !errors.Is(err, core_errors.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
