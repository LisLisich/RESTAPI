package identity_service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func (s *IdentityService) RequestPasswordReset(
	ctx context.Context,
	email string,
) error {
	normalizedAccount, err := domain.NewAccountUninitialized(0, email)
	if err != nil {
		return fmt.Errorf("validate password reset email: %w", err)
	}

	passwordAccount, err := s.registrationRepository.GetPasswordAccount(
		ctx,
		normalizedAccount.Email,
	)
	if errors.Is(err, core_errors.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get password reset account: %w", err)
	}
	if passwordAccount.Account.Status != domain.AccountStatusActive {
		return nil
	}

	token, err := s.tokenIssuer.Issue()
	if err != nil {
		return fmt.Errorf("issue password reset token: %w", err)
	}
	if err := validateIssuedToken(token); err != nil {
		return fmt.Errorf("validate password reset token: %w", err)
	}

	now := s.now().UTC()
	if err := s.registrationRepository.RequestPasswordReset(
		ctx,
		PasswordResetRequest{
			AccountUserID: passwordAccount.Account.UserID,
			Email:         passwordAccount.Account.Email,
			Token:         token,
			ExpiresAt:     now.Add(passwordResetTTL),
			CreatedAt:     now,
		},
	); err != nil {
		return fmt.Errorf("request password reset: %w", err)
	}
	return nil
}

func (s *IdentityService) ResetPassword(
	ctx context.Context,
	rawToken string,
	newPassword string,
) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return fmt.Errorf(
			"password reset token is required: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	if err := domain.ValidatePassword(newPassword); err != nil {
		return fmt.Errorf("validate new password: %w", err)
	}

	passwordHash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}
	if err := s.registrationRepository.ResetPassword(
		ctx,
		s.tokenIssuer.Hash(rawToken),
		passwordHash,
		s.now().UTC(),
	); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	return nil
}
