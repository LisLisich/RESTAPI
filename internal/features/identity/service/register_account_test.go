package identity_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

type fakeRegistrationRepository struct {
	registerCalled bool
	registration   Registration
	err            error

	verifyCalled bool
	verifyHash   []byte
	verifiedAt   time.Time
	verifyErr    error

	passwordAccount PasswordAccount
	getPasswordErr  error
	getPasswordCall bool
	gotEmail        string

	createSessionCalled bool
	createdSession      StoredSession
	createSessionErr    error

	authSession StoredAuthenticationSession
	getSession  bool
	sessionHash []byte
	sessionNow  time.Time
	sessionErr  error

	passwordResetCalled  bool
	passwordResetRequest PasswordResetRequest
	passwordResetErr     error

	resetPasswordCalled bool
	resetTokenHash      []byte
	resetPasswordHash   string
	resetAt             time.Time
	resetPasswordErr    error

	revokeSessionCalled bool
	revokeSessionHash   []byte
	revokeSessionAt     time.Time
	revokeSessionErr    error

	createRefreshCalled bool
	createdRefresh      StoredRefreshToken
	createRefreshErr    error

	rotateRefreshCalled bool
	oldRefreshHash      []byte
	rotatedRefresh      StoredRefreshToken
	rotateRefreshUserID int
	rotateRefreshErr    error
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

func (r *fakeRegistrationRepository) GetPasswordAccount(
	_ context.Context,
	email string,
) (PasswordAccount, error) {
	r.getPasswordCall = true
	r.gotEmail = email
	if r.getPasswordErr != nil {
		return PasswordAccount{}, r.getPasswordErr
	}
	return r.passwordAccount, nil
}

func (r *fakeRegistrationRepository) CreateSession(
	_ context.Context,
	session StoredSession,
) error {
	r.createSessionCalled = true
	r.createdSession = session
	return r.createSessionErr
}

func (r *fakeRegistrationRepository) GetSession(
	_ context.Context,
	tokenHash []byte,
	now time.Time,
) (StoredAuthenticationSession, error) {
	r.getSession = true
	r.sessionHash = tokenHash
	r.sessionNow = now
	if r.sessionErr != nil {
		return StoredAuthenticationSession{}, r.sessionErr
	}
	return r.authSession, nil
}

func (r *fakeRegistrationRepository) RequestPasswordReset(
	_ context.Context,
	request PasswordResetRequest,
) error {
	r.passwordResetCalled = true
	r.passwordResetRequest = request
	return r.passwordResetErr
}

func (r *fakeRegistrationRepository) ResetPassword(
	_ context.Context,
	tokenHash []byte,
	passwordHash string,
	resetAt time.Time,
) error {
	r.resetPasswordCalled = true
	r.resetTokenHash = tokenHash
	r.resetPasswordHash = passwordHash
	r.resetAt = resetAt
	return r.resetPasswordErr
}

func (r *fakeRegistrationRepository) RevokeSession(
	_ context.Context,
	tokenHash []byte,
	revokedAt time.Time,
) error {
	r.revokeSessionCalled = true
	r.revokeSessionHash = tokenHash
	r.revokeSessionAt = revokedAt
	return r.revokeSessionErr
}

func (r *fakeRegistrationRepository) CreateRefreshToken(
	_ context.Context,
	token StoredRefreshToken,
) error {
	r.createRefreshCalled = true
	r.createdRefresh = token
	return r.createRefreshErr
}

func (r *fakeRegistrationRepository) RotateRefreshToken(
	_ context.Context,
	oldTokenHash []byte,
	newToken StoredRefreshToken,
) (int, error) {
	r.rotateRefreshCalled = true
	r.oldRefreshHash = oldTokenHash
	r.rotatedRefresh = newToken
	if r.rotateRefreshErr != nil {
		return 0, r.rotateRefreshErr
	}
	return r.rotateRefreshUserID, nil
}

func (r *fakeRegistrationRepository) VerifyEmail(
	_ context.Context,
	tokenHash []byte,
	verifiedAt time.Time,
) (domain.Account, error) {
	r.verifyCalled = true
	r.verifyHash = tokenHash
	r.verifiedAt = verifiedAt
	if r.verifyErr != nil {
		return domain.Account{}, r.verifyErr
	}

	account, err := domain.NewAccountUninitialized(42, "user@example.com")
	if err != nil {
		return domain.Account{}, err
	}
	if err := account.VerifyEmail(verifiedAt); err != nil {
		return domain.Account{}, err
	}
	return account, nil
}

type fakePasswordHasher struct {
	hashCalled bool
	password   string
	hash       string
	err        error

	verifyCalled bool
	verifyResult bool
	verifyErr    error
	encodedHash  string
}

func (h *fakePasswordHasher) Hash(password string) (string, error) {
	h.hashCalled = true
	h.password = password
	return h.hash, h.err
}

func (h *fakePasswordHasher) Verify(password string, encodedHash string) (bool, error) {
	h.verifyCalled = true
	h.password = password
	h.encodedHash = encodedHash
	return h.verifyResult, h.verifyErr
}

type fakeTokenIssuer struct {
	issueCalled bool
	token       IssuedToken
	tokens      []IssuedToken
	err         error

	hashCalled bool
	rawToken   string
	hash       []byte
}

func (i *fakeTokenIssuer) Issue() (IssuedToken, error) {
	i.issueCalled = true
	if len(i.tokens) > 0 {
		token := i.tokens[0]
		i.tokens = i.tokens[1:]
		return token, i.err
	}
	return i.token, i.err
}

func (i *fakeTokenIssuer) Hash(rawToken string) []byte {
	i.hashCalled = true
	i.rawToken = rawToken
	return i.hash
}

type fakeAccessTokenIssuer struct {
	called bool
	userID int
	jwtID  string
	token  string
	err    error
}

func (issuer *fakeAccessTokenIssuer) Issue(userID int, jwtID string) (string, error) {
	issuer.called = true
	issuer.userID = userID
	issuer.jwtID = jwtID
	return issuer.token, issuer.err
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
