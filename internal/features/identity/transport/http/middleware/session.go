package identity_http_middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	core_http_middleware "github.com/LisLisich/fintask/internal/core/transport/http/middleware"
	core_http_response "github.com/LisLisich/fintask/internal/core/transport/http/response"
	identity_jwt_provider "github.com/LisLisich/fintask/internal/features/identity/provider/jwt"
	identity_service "github.com/LisLisich/fintask/internal/features/identity/service"
)

const CSRFHeaderName = "X-CSRF-Token"

type SessionAuthenticator interface {
	AuthenticateSession(
		ctx context.Context,
		rawSessionToken string,
		rawCSRFToken string,
		requireCSRF bool,
	) (identity_service.Principal, error)
}

type AccessTokenVerifier interface {
	Verify(token string) (identity_jwt_provider.Claims, error)
}

type principalContextKey struct{}

func Authentication(
	sessionAuthenticator SessionAuthenticator,
	accessTokenVerifier AccessTokenVerifier,
	sessionCookieName string,
) core_http_middleware.Middleware {
	sessionMiddleware := Session(sessionAuthenticator, sessionCookieName)

	return func(next http.Handler) http.Handler {
		sessionHandler := sessionMiddleware(next)
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			authorization := strings.TrimSpace(r.Header.Get("Authorization"))
			if authorization == "" {
				sessionHandler.ServeHTTP(rw, r)
				return
			}

			parts := strings.Fields(authorization)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				authenticationError(rw, r, "invalid authorization header")
				return
			}
			claims, err := accessTokenVerifier.Verify(parts[1])
			if err != nil {
				authenticationError(rw, r, err.Error())
				return
			}
			userID, err := strconv.Atoi(claims.Subject)
			if err != nil || userID <= 0 {
				authenticationError(rw, r, "invalid access token subject")
				return
			}

			ctx := ContextWithPrincipal(
				r.Context(),
				identity_service.Principal{UserID: userID},
			)
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}

func Session(
	authenticator SessionAuthenticator,
	sessionCookieName string,
) core_http_middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			log := core_logger.FromContext(r.Context())
			responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

			sessionCookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				responseHandler.ErrorResponse(
					fmt.Errorf("session cookie is required: %w", core_errors.ErrUnauthorized),
					"authentication required",
				)
				return
			}

			requireCSRF := requestRequiresCSRF(r.Method)
			principal, err := authenticator.AuthenticateSession(
				r.Context(),
				sessionCookie.Value,
				r.Header.Get(CSRFHeaderName),
				requireCSRF,
			)
			if err != nil {
				responseHandler.ErrorResponse(err, "session authentication failed")
				return
			}

			ctx := ContextWithPrincipal(r.Context(), principal)
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}

func authenticationError(rw http.ResponseWriter, r *http.Request, reason string) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(r.Context()),
		rw,
	)
	responseHandler.ErrorResponse(
		fmt.Errorf("%s: %w", reason, core_errors.ErrUnauthorized),
		"authentication required",
	)
}

func PrincipalFromContext(ctx context.Context) (identity_service.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(identity_service.Principal)
	return principal, ok
}

func ContextWithPrincipal(
	ctx context.Context,
	principal identity_service.Principal,
) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func requestRequiresCSRF(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}
