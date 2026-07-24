package identity_http_middleware

import (
	"context"
	"fmt"
	"net/http"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_middleware "github.com/LisLisich/RESTAPI/internal/core/transport/http/middleware"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
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

type principalContextKey struct{}

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
