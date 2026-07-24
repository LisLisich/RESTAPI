package identity_service

import (
	"context"
	"fmt"
	"strings"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func (s *IdentityService) VerifyEmail(
	ctx context.Context,
	rawToken string,
) (domain.Account, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return domain.Account{}, fmt.Errorf(
			"email verification token is required: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	account, err := s.registrationRepository.VerifyEmail(
		ctx,
		s.tokenIssuer.Hash(rawToken),
		s.now().UTC(),
	)
	if err != nil {
		return domain.Account{}, fmt.Errorf("verify email: %w", err)
	}
	return account, nil
}
