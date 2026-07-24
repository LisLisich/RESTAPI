package identity_service

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func (s *IdentityService) AuthenticateSession(
	ctx context.Context,
	rawSessionToken string,
	rawCSRFToken string,
	requireCSRF bool,
) (Principal, error) {
	rawSessionToken = strings.TrimSpace(rawSessionToken)
	if rawSessionToken == "" {
		return Principal{}, invalidSessionError()
	}

	session, err := s.registrationRepository.GetSession(
		ctx,
		s.tokenIssuer.Hash(rawSessionToken),
		s.now().UTC(),
	)
	if errors.Is(err, core_errors.ErrNotFound) {
		return Principal{}, invalidSessionError()
	}
	if err != nil {
		return Principal{}, fmt.Errorf("get browser session: %w", err)
	}

	if requireCSRF {
		rawCSRFToken = strings.TrimSpace(rawCSRFToken)
		if rawCSRFToken == "" {
			return Principal{}, invalidCSRFError()
		}
		csrfHash := s.tokenIssuer.Hash(rawCSRFToken)
		if subtle.ConstantTimeCompare(csrfHash, session.CSRFHash) != 1 {
			return Principal{}, invalidCSRFError()
		}
	}

	return Principal{UserID: session.AccountUserID}, nil
}

func invalidSessionError() error {
	return fmt.Errorf("invalid or expired session: %w", core_errors.ErrUnauthorized)
}

func invalidCSRFError() error {
	return fmt.Errorf("invalid csrf token: %w", core_errors.ErrForbidden)
}
