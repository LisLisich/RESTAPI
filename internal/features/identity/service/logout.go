package identity_service

import (
	"context"
	"fmt"
	"strings"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func (s *IdentityService) Logout(
	ctx context.Context,
	rawSessionToken string,
) error {
	rawSessionToken = strings.TrimSpace(rawSessionToken)
	if rawSessionToken == "" {
		return fmt.Errorf("session token is required: %w", core_errors.ErrUnauthorized)
	}
	if err := s.registrationRepository.RevokeSession(
		ctx,
		s.tokenIssuer.Hash(rawSessionToken),
		s.now().UTC(),
	); err != nil {
		return fmt.Errorf("revoke browser session: %w", err)
	}
	return nil
}
