package identity_service

import (
	"context"
	"fmt"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

type RegisterAccountInput struct {
	FullName string
	Email    string
	Password string
}

func (s *IdentityService) RegisterAccount(
	ctx context.Context,
	input RegisterAccountInput,
) (domain.Account, error) {
	user := domain.NewUserUninitialized(input.FullName, nil)
	if err := user.Validate(); err != nil {
		return domain.Account{}, fmt.Errorf("validate registration user: %w", err)
	}

	account, err := domain.NewAccountUninitialized(user.ID, input.Email)
	if err != nil {
		return domain.Account{}, fmt.Errorf("validate registration account: %w", err)
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return domain.Account{}, fmt.Errorf("validate registration password: %w", err)
	}

	passwordHash, err := s.passwordHasher.Hash(input.Password)
	if err != nil {
		return domain.Account{}, fmt.Errorf("hash registration password: %w", err)
	}
	verificationToken, err := s.tokenIssuer.Issue()
	if err != nil {
		return domain.Account{}, fmt.Errorf("issue email verification token: %w", err)
	}
	if verificationToken.Raw == "" || len(verificationToken.Hash) == 0 {
		return domain.Account{}, fmt.Errorf(
			"issue email verification token: empty token: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	now := s.now().UTC()
	account, err = s.registrationRepository.Register(
		ctx,
		Registration{
			User:                  user,
			Account:               account,
			PasswordHash:          passwordHash,
			VerificationToken:     verificationToken,
			VerificationExpiresAt: now.Add(emailVerificationTTL),
			CreatedAt:             now,
		},
	)
	if err != nil {
		return domain.Account{}, fmt.Errorf("register account: %w", err)
	}

	return account, nil
}
