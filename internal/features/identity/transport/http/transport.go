package identity_transport_http

import (
	"context"
	"net/http"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_http_middleware "github.com/LisLisich/fintask/internal/core/transport/http/middleware"
	core_http_server "github.com/LisLisich/fintask/internal/core/transport/http/server"
	identity_service "github.com/LisLisich/fintask/internal/features/identity/service"
)

type IdentityService interface {
	RegisterAccount(
		ctx context.Context,
		input identity_service.RegisterAccountInput,
	) (domain.Account, error)
	VerifyEmail(ctx context.Context, rawToken string) (domain.Account, error)
	Login(
		ctx context.Context,
		input identity_service.LoginInput,
	) (identity_service.BrowserSession, error)
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, rawToken string, newPassword string) error
	Logout(ctx context.Context, rawSessionToken string) error
	LoginAPI(
		ctx context.Context,
		input identity_service.LoginInput,
	) (identity_service.TokenPair, error)
	RefreshAPI(ctx context.Context, rawRefreshToken string) (identity_service.TokenPair, error)
}

type IdentityHTTPHandler struct {
	identityService     IdentityService
	googleLoginService  *identity_service.GoogleLoginService
	protectedMiddleware core_http_middleware.Middleware
}

func NewIdentityHTTPHandler(
	identityService IdentityService,
	protectedMiddleware ...core_http_middleware.Middleware,
) *IdentityHTTPHandler {
	handler := &IdentityHTTPHandler{identityService: identityService}
	if len(protectedMiddleware) > 0 {
		handler.protectedMiddleware = protectedMiddleware[0]
	}
	return handler
}

func (h *IdentityHTTPHandler) SetGoogleLoginService(
	service *identity_service.GoogleLoginService,
) {
	h.googleLoginService = service
}

func (h *IdentityHTTPHandler) Routes() []core_http_server.Route {
	routes := []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/auth/register",
			Handler: h.RegisterAccount,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/verify-email",
			Handler: h.VerifyEmail,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/login",
			Handler: h.Login,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/password-reset/request",
			Handler: h.RequestPasswordReset,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/password-reset/confirm",
			Handler: h.ConfirmPasswordReset,
		},
		{
			Method:     http.MethodPost,
			Path:       "/auth/logout",
			Handler:    h.Logout,
			Middleware: h.protectedRouteMiddleware(),
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/token",
			Handler: h.LoginAPI,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/token/refresh",
			Handler: h.RefreshAPI,
		},
	}
	if h.googleLoginService != nil {
		routes = append(
			routes,
			core_http_server.Route{
				Method:  http.MethodGet,
				Path:    "/auth/google/start",
				Handler: h.StartGoogleLogin,
			},
			core_http_server.Route{
				Method:  http.MethodGet,
				Path:    "/auth/google/callback",
				Handler: h.CompleteGoogleLogin,
			},
		)
	}
	return routes
}

func (h *IdentityHTTPHandler) protectedRouteMiddleware() []core_http_middleware.Middleware {
	if h.protectedMiddleware == nil {
		return nil
	}
	return []core_http_middleware.Middleware{h.protectedMiddleware}
}
