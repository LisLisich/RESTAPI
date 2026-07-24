package identity_service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

const browserSessionTTL = 12 * time.Hour

type LoginInput struct {
	Email    string
	Password string
}

func (s *IdentityService) Login(
	ctx context.Context,
	input LoginInput,
) (BrowserSession, error) {
	normalizedAccount, err := domain.NewAccountUninitialized(0, input.Email)
	if err != nil {
		return BrowserSession{}, fmt.Errorf("validate login email: %w", err)
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return BrowserSession{}, fmt.Errorf("validate login password: %w", err)
	}

	passwordAccount, err := s.registrationRepository.GetPasswordAccount(
		ctx,
		normalizedAccount.Email,
	)
	if errors.Is(err, core_errors.ErrNotFound) {
		return BrowserSession{}, invalidCredentialsError()
	}
	if err != nil {
		return BrowserSession{}, fmt.Errorf("get password account: %w", err)
	}
	if passwordAccount.Account.Status != domain.AccountStatusActive {
		return BrowserSession{}, invalidCredentialsError()
	}

	passwordMatches, err := s.passwordHasher.Verify(
		input.Password,
		passwordAccount.PasswordHash,
	)
	if err != nil {
		return BrowserSession{}, fmt.Errorf("verify password hash: %w", err)
	}
	if !passwordMatches {
		return BrowserSession{}, invalidCredentialsError()
	}

	sessionToken, err := s.tokenIssuer.Issue()
	if err != nil {
		return BrowserSession{}, fmt.Errorf("issue session token: %w", err)
	}
	csrfToken, err := s.tokenIssuer.Issue()
	if err != nil {
		return BrowserSession{}, fmt.Errorf("issue csrf token: %w", err)
	}
	if err := validateIssuedToken(sessionToken); err != nil {
		return BrowserSession{}, fmt.Errorf("validate session token: %w", err)
	}
	if err := validateIssuedToken(csrfToken); err != nil {
		return BrowserSession{}, fmt.Errorf("validate csrf token: %w", err)
	}

	now := s.now().UTC()
	expiresAt := now.Add(browserSessionTTL)
	if err := s.registrationRepository.CreateSession(
		ctx,
		StoredSession{
			AccountUserID: passwordAccount.Account.UserID,
			TokenHash:     sessionToken.Hash,
			CSRFHash:      csrfToken.Hash,
			ExpiresAt:     expiresAt,
			CreatedAt:     now,
		},
	); err != nil {
		return BrowserSession{}, fmt.Errorf("create browser session: %w", err)
	}

	return BrowserSession{
		Account:   passwordAccount.Account,
		Token:     sessionToken.Raw,
		CSRFToken: csrfToken.Raw,
		ExpiresAt: expiresAt,
	}, nil
}

func validateIssuedToken(token IssuedToken) error {
	if token.Raw == "" || len(token.Hash) == 0 {
		return fmt.Errorf("issued token is empty")
	}
	return nil
}

func invalidCredentialsError() error {
	return fmt.Errorf("invalid credentials: %w", core_errors.ErrUnauthorized)
}
