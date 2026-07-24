package identity_jwt_provider

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"strings"
	"testing"
	"time"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestEd25519ProviderIssuesAndVerifiesAccessToken(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	now := time.Date(2026, time.July, 25, 13, 0, 0, 0, time.UTC)
	provider := NewEd25519Provider(
		privateKey,
		publicKey,
		"fintask",
		"fintask-cli",
		15*time.Minute,
		func() time.Time { return now },
	)

	token, err := provider.Issue(42, "access-jti")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	claims, err := provider.Verify(token)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.Subject != "42" || claims.JWTID != "access-jti" {
		t.Fatalf("unexpected claims %+v", claims)
	}
	if claims.Issuer != "fintask" || claims.Audience != "fintask-cli" {
		t.Fatalf("unexpected issuer/audience %+v", claims)
	}
	if claims.ExpiresAt != now.Add(15*time.Minute).Unix() {
		t.Fatalf("unexpected expiry %d", claims.ExpiresAt)
	}
}

func TestEd25519ProviderRejectsTamperedToken(t *testing.T) {
	provider := newTestProvider(t, time.Now)
	token, err := provider.Issue(42, "jti")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	parts := strings.Split(token, ".")
	parts[1] = parts[1] + "a"

	_, err = provider.Verify(strings.Join(parts, "."))

	if !errors.Is(err, core_errors.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestEd25519ProviderRejectsExpiredToken(t *testing.T) {
	now := time.Date(2026, time.July, 25, 13, 0, 0, 0, time.UTC)
	current := now
	provider := newTestProvider(t, func() time.Time { return current })
	token, err := provider.Issue(42, "jti")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	current = now.Add(16 * time.Minute)

	_, err = provider.Verify(token)

	if !errors.Is(err, core_errors.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func newTestProvider(t *testing.T, now func() time.Time) *Ed25519Provider {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return NewEd25519Provider(
		privateKey,
		publicKey,
		"fintask",
		"fintask-cli",
		15*time.Minute,
		now,
	)
}
