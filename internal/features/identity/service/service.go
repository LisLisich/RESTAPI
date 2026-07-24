package identity_service

import (
	"context"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
)

const (
	emailVerificationTTL = 24 * time.Hour
	passwordResetTTL     = time.Hour
)

type IdentityService struct {
	registrationRepository RegistrationRepository
	passwordHasher         PasswordHasher
	tokenIssuer            TokenIssuer
	now                    func() time.Time
}

type RegistrationRepository interface {
	Register(ctx context.Context, registration Registration) (domain.Account, error)
	VerifyEmail(
		ctx context.Context,
		tokenHash []byte,
		verifiedAt time.Time,
	) (domain.Account, error)
	GetPasswordAccount(ctx context.Context, email string) (PasswordAccount, error)
	CreateSession(ctx context.Context, session StoredSession) error
	GetSession(
		ctx context.Context,
		tokenHash []byte,
		now time.Time,
	) (StoredAuthenticationSession, error)
	RequestPasswordReset(ctx context.Context, request PasswordResetRequest) error
	ResetPassword(
		ctx context.Context,
		tokenHash []byte,
		passwordHash string,
		resetAt time.Time,
	) error
	RevokeSession(ctx context.Context, tokenHash []byte, revokedAt time.Time) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password string, encodedHash string) (bool, error)
}

type TokenIssuer interface {
	Issue() (IssuedToken, error)
	Hash(rawToken string) []byte
}

type IssuedToken struct {
	Raw  string
	Hash []byte
}

type Registration struct {
	User                  domain.User
	Account               domain.Account
	PasswordHash          string
	VerificationToken     IssuedToken
	VerificationExpiresAt time.Time
	CreatedAt             time.Time
}

type PasswordAccount struct {
	Account      domain.Account
	PasswordHash string
}

type StoredSession struct {
	AccountUserID int
	TokenHash     []byte
	CSRFHash      []byte
	ExpiresAt     time.Time
	CreatedAt     time.Time
}

type BrowserSession struct {
	Account   domain.Account
	Token     string
	CSRFToken string
	ExpiresAt time.Time
}

type StoredAuthenticationSession struct {
	AccountUserID int
	CSRFHash      []byte
}

type Principal struct {
	UserID int
}

type PasswordResetRequest struct {
	AccountUserID int
	Email         string
	Token         IssuedToken
	ExpiresAt     time.Time
	CreatedAt     time.Time
}

func NewIdentityService(
	registrationRepository RegistrationRepository,
	passwordHasher PasswordHasher,
	tokenIssuer TokenIssuer,
	now func() time.Time,
) *IdentityService {
	return &IdentityService{
		registrationRepository: registrationRepository,
		passwordHasher:         passwordHasher,
		tokenIssuer:            tokenIssuer,
		now:                    now,
	}
}
