package identity_service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour
)

func (s *IdentityService) LoginAPI(
	ctx context.Context,
	input LoginInput,
) (TokenPair, error) {
	passwordAccount, err := s.authenticatePasswordAccount(ctx, input)
	if err != nil {
		return TokenPair{}, err
	}
	if s.accessTokenIssuer == nil {
		return TokenPair{}, fmt.Errorf("access token issuer is not configured")
	}

	refreshToken, err := s.tokenIssuer.Issue()
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue refresh token: %w", err)
	}
	accessJWTID, err := s.tokenIssuer.Issue()
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue access jwt id: %w", err)
	}
	familyID, err := s.tokenIssuer.Issue()
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue refresh family id: %w", err)
	}
	if err := validateIssuedToken(refreshToken); err != nil {
		return TokenPair{}, err
	}

	accessToken, err := s.accessTokenIssuer.Issue(
		passwordAccount.Account.UserID,
		accessJWTID.Raw,
	)
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue access token: %w", err)
	}
	now := s.now().UTC()
	if err := s.registrationRepository.CreateRefreshToken(
		ctx,
		StoredRefreshToken{
			AccountUserID: passwordAccount.Account.UserID,
			FamilyID:      familyID.Raw,
			TokenHash:     refreshToken.Hash,
			ExpiresAt:     now.Add(refreshTokenTTL),
			CreatedAt:     now,
		},
	); err != nil {
		return TokenPair{}, fmt.Errorf("store refresh token: %w", err)
	}
	return newTokenPair(accessToken, refreshToken.Raw), nil
}

func (s *IdentityService) RefreshAPI(
	ctx context.Context,
	rawRefreshToken string,
) (TokenPair, error) {
	rawRefreshToken = strings.TrimSpace(rawRefreshToken)
	if rawRefreshToken == "" {
		return TokenPair{}, fmt.Errorf("refresh token is required: %w", core_errors.ErrUnauthorized)
	}
	if s.accessTokenIssuer == nil {
		return TokenPair{}, fmt.Errorf("access token issuer is not configured")
	}

	newRefreshToken, err := s.tokenIssuer.Issue()
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue rotated refresh token: %w", err)
	}
	accessJWTID, err := s.tokenIssuer.Issue()
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue access jwt id: %w", err)
	}
	now := s.now().UTC()
	userID, err := s.registrationRepository.RotateRefreshToken(
		ctx,
		s.tokenIssuer.Hash(rawRefreshToken),
		StoredRefreshToken{
			TokenHash: newRefreshToken.Hash,
			ExpiresAt: now.Add(refreshTokenTTL),
			CreatedAt: now,
		},
	)
	if err != nil {
		return TokenPair{}, fmt.Errorf("rotate refresh token: %w", err)
	}
	accessToken, err := s.accessTokenIssuer.Issue(userID, accessJWTID.Raw)
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue rotated access token: %w", err)
	}
	return newTokenPair(accessToken, newRefreshToken.Raw), nil
}

func (s *IdentityService) authenticatePasswordAccount(
	ctx context.Context,
	input LoginInput,
) (PasswordAccount, error) {
	normalizedAccount, err := domain.NewAccountUninitialized(0, input.Email)
	if err != nil {
		return PasswordAccount{}, fmt.Errorf("validate login email: %w", err)
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return PasswordAccount{}, fmt.Errorf("validate login password: %w", err)
	}
	passwordAccount, err := s.registrationRepository.GetPasswordAccount(
		ctx,
		normalizedAccount.Email,
	)
	if errors.Is(err, core_errors.ErrNotFound) {
		return PasswordAccount{}, invalidCredentialsError()
	}
	if err != nil {
		return PasswordAccount{}, fmt.Errorf("get password account: %w", err)
	}
	if passwordAccount.Account.Status != domain.AccountStatusActive {
		return PasswordAccount{}, invalidCredentialsError()
	}
	matches, err := s.passwordHasher.Verify(input.Password, passwordAccount.PasswordHash)
	if err != nil {
		return PasswordAccount{}, fmt.Errorf("verify password hash: %w", err)
	}
	if !matches {
		return PasswordAccount{}, invalidCredentialsError()
	}
	return passwordAccount, nil
}

func newTokenPair(accessToken string, refreshToken string) TokenPair {
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(accessTokenTTL.Seconds()),
	}
}
