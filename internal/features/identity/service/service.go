package identity_service

import (
	"context"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
)

const emailVerificationTTL = 24 * time.Hour

type IdentityService struct {
	registrationRepository RegistrationRepository
	passwordHasher         PasswordHasher
	tokenIssuer            TokenIssuer
	now                    func() time.Time
}

type RegistrationRepository interface {
	Register(ctx context.Context, registration Registration) (domain.Account, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
}

type TokenIssuer interface {
	Issue() (IssuedToken, error)
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
