package identity_service

import (
	"context"
	"errors"
	"testing"
	"time"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

type fakeGoogleRepository struct {
	attempt    GoogleLoginAttempt
	identity   GoogleIdentity
	userID     int
	resolveErr error
	session    StoredSession
}

func (repository *fakeGoogleRepository) StoreGoogleLoginAttempt(
	_ context.Context,
	attempt GoogleLoginAttempt,
) error {
	repository.attempt = attempt
	return nil
}

func (repository *fakeGoogleRepository) ConsumeGoogleLoginAttempt(
	context.Context,
	[]byte,
	time.Time,
) (GoogleLoginAttempt, error) {
	return repository.attempt, nil
}

func (repository *fakeGoogleRepository) ResolveGoogleIdentity(
	_ context.Context,
	identity GoogleIdentity,
	_ time.Time,
) (int, error) {
	repository.identity = identity
	return repository.userID, repository.resolveErr
}

func (repository *fakeGoogleRepository) CreateSession(
	_ context.Context,
	session StoredSession,
) error {
	repository.session = session
	return nil
}

type fakeGoogleClient struct {
	exchangedIdentity GoogleIdentity
	nonce             string
	verifier          string
}

func (client *fakeGoogleClient) AuthorizationURL(
	state string,
	nonce string,
	codeChallenge string,
) string {
	return "https://accounts.example/auth?state=" + state
}

func (client *fakeGoogleClient) Exchange(
	_ context.Context,
	_ string,
	codeVerifier string,
	expectedNonce string,
) (GoogleIdentity, error) {
	client.verifier = codeVerifier
	client.nonce = expectedNonce
	return client.exchangedIdentity, nil
}

func TestGoogleLoginUsesPKCEAndCreatesSession(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	repository := &fakeGoogleRepository{userID: 42}
	client := &fakeGoogleClient{exchangedIdentity: GoogleIdentity{
		Subject:       "google-sub",
		Email:         "user@example.com",
		EmailVerified: true,
		Name:          "Test User",
	}}
	tokenIssuer := &fakeTokenIssuer{tokens: []IssuedToken{
		{Raw: "state", Hash: []byte("state-hash")},
		{Raw: "nonce", Hash: []byte("nonce-hash")},
		{Raw: "verifier", Hash: []byte("verifier-hash")},
		{Raw: "session", Hash: []byte("session-hash")},
		{Raw: "csrf", Hash: []byte("csrf-hash")},
	}}
	service := NewGoogleLoginService(repository, client, tokenIssuer, func() time.Time {
		return now
	})

	start, err := service.Begin(t.Context())
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	result, err := service.Complete(t.Context(), "code", start.State, start.State)

	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if result.Account.UserID != 42 || result.Token != "session" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if client.nonce != "nonce" || client.verifier != "verifier" {
		t.Fatal("expected stored nonce and PKCE verifier")
	}
}

func TestGoogleLoginRejectsStateCookieMismatch(t *testing.T) {
	service := NewGoogleLoginService(
		&fakeGoogleRepository{},
		&fakeGoogleClient{},
		&fakeTokenIssuer{},
		time.Now,
	)

	_, err := service.Complete(t.Context(), "code", "query-state", "cookie-state")

	if !errors.Is(err, core_errors.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
