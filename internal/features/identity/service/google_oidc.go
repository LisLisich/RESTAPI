package identity_service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

const googleLoginAttemptTTL = 10 * time.Minute

type GoogleIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

type GoogleLoginAttempt struct {
	StateHash    []byte
	Nonce        string
	CodeVerifier string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type GoogleLoginStart struct {
	AuthorizationURL string
	State            string
	ExpiresAt        time.Time
}

type GoogleLoginRepository interface {
	StoreGoogleLoginAttempt(ctx context.Context, attempt GoogleLoginAttempt) error
	ConsumeGoogleLoginAttempt(
		ctx context.Context,
		stateHash []byte,
		consumedAt time.Time,
	) (GoogleLoginAttempt, error)
	ResolveGoogleIdentity(
		ctx context.Context,
		identity GoogleIdentity,
		createdAt time.Time,
	) (int, error)
	CreateSession(ctx context.Context, session StoredSession) error
}

type GoogleOIDCClient interface {
	AuthorizationURL(state string, nonce string, codeChallenge string) string
	Exchange(
		ctx context.Context,
		code string,
		codeVerifier string,
		expectedNonce string,
	) (GoogleIdentity, error)
}

type GoogleLoginService struct {
	repository  GoogleLoginRepository
	client      GoogleOIDCClient
	tokenIssuer TokenIssuer
	now         func() time.Time
}

func NewGoogleLoginService(
	repository GoogleLoginRepository,
	client GoogleOIDCClient,
	tokenIssuer TokenIssuer,
	now func() time.Time,
) *GoogleLoginService {
	return &GoogleLoginService{
		repository: repository, client: client, tokenIssuer: tokenIssuer, now: now,
	}
}

func (service *GoogleLoginService) Begin(ctx context.Context) (GoogleLoginStart, error) {
	state, err := service.tokenIssuer.Issue()
	if err != nil {
		return GoogleLoginStart{}, fmt.Errorf("issue OIDC state: %w", err)
	}
	nonce, err := service.tokenIssuer.Issue()
	if err != nil {
		return GoogleLoginStart{}, fmt.Errorf("issue OIDC nonce: %w", err)
	}
	verifier, err := service.tokenIssuer.Issue()
	if err != nil {
		return GoogleLoginStart{}, fmt.Errorf("issue PKCE verifier: %w", err)
	}
	if err := validateIssuedToken(state); err != nil {
		return GoogleLoginStart{}, err
	}
	if err := validateIssuedToken(nonce); err != nil {
		return GoogleLoginStart{}, err
	}
	if err := validateIssuedToken(verifier); err != nil {
		return GoogleLoginStart{}, err
	}
	now := service.now().UTC()
	attempt := GoogleLoginAttempt{
		StateHash:    state.Hash,
		Nonce:        nonce.Raw,
		CodeVerifier: verifier.Raw,
		ExpiresAt:    now.Add(googleLoginAttemptTTL),
		CreatedAt:    now,
	}
	if err := service.repository.StoreGoogleLoginAttempt(ctx, attempt); err != nil {
		return GoogleLoginStart{}, fmt.Errorf("store Google login attempt: %w", err)
	}
	challengeHash := sha256.Sum256([]byte(verifier.Raw))
	return GoogleLoginStart{
		AuthorizationURL: service.client.AuthorizationURL(
			state.Raw,
			nonce.Raw,
			base64.RawURLEncoding.EncodeToString(challengeHash[:]),
		),
		State:     state.Raw,
		ExpiresAt: attempt.ExpiresAt,
	}, nil
}

func (service *GoogleLoginService) Complete(
	ctx context.Context,
	code string,
	queryState string,
	cookieState string,
) (BrowserSession, error) {
	if code == "" {
		return BrowserSession{}, fmt.Errorf(
			"authorization code is required: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	if queryState == "" ||
		cookieState == "" ||
		subtle.ConstantTimeCompare([]byte(queryState), []byte(cookieState)) != 1 {
		return BrowserSession{}, fmt.Errorf("OIDC state mismatch: %w", core_errors.ErrForbidden)
	}
	now := service.now().UTC()
	attempt, err := service.repository.ConsumeGoogleLoginAttempt(
		ctx,
		service.tokenIssuer.Hash(queryState),
		now,
	)
	if err != nil {
		return BrowserSession{}, fmt.Errorf("consume Google login attempt: %w", err)
	}
	identity, err := service.client.Exchange(
		ctx,
		code,
		attempt.CodeVerifier,
		attempt.Nonce,
	)
	if err != nil {
		return BrowserSession{}, fmt.Errorf("exchange Google authorization code: %w", err)
	}
	if identity.Subject == "" || identity.Email == "" || !identity.EmailVerified {
		return BrowserSession{}, fmt.Errorf("Google identity is incomplete: %w", core_errors.ErrUnauthorized)
	}
	userID, err := service.repository.ResolveGoogleIdentity(ctx, identity, now)
	if err != nil {
		return BrowserSession{}, fmt.Errorf("resolve Google identity: %w", err)
	}
	sessionToken, err := service.tokenIssuer.Issue()
	if err != nil {
		return BrowserSession{}, fmt.Errorf("issue Google session token: %w", err)
	}
	csrfToken, err := service.tokenIssuer.Issue()
	if err != nil {
		return BrowserSession{}, fmt.Errorf("issue Google CSRF token: %w", err)
	}
	if err := validateIssuedToken(sessionToken); err != nil {
		return BrowserSession{}, fmt.Errorf("validate Google session token: %w", err)
	}
	if err := validateIssuedToken(csrfToken); err != nil {
		return BrowserSession{}, fmt.Errorf("validate Google CSRF token: %w", err)
	}
	expiresAt := now.Add(browserSessionTTL)
	if err := service.repository.CreateSession(ctx, StoredSession{
		AccountUserID: userID,
		TokenHash:     sessionToken.Hash,
		CSRFHash:      csrfToken.Hash,
		ExpiresAt:     expiresAt,
		CreatedAt:     now,
	}); err != nil {
		return BrowserSession{}, fmt.Errorf("create Google browser session: %w", err)
	}
	return BrowserSession{
		Account: domain.Account{
			UserID: userID,
			Email:  identity.Email,
			Status: domain.AccountStatusActive,
		},
		Token: sessionToken.Raw, CSRFToken: csrfToken.Raw, ExpiresAt: expiresAt,
	}, nil
}
