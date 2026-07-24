package identity_transport_http

import (
	"context"
	"net/http"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_http_server "github.com/LisLisich/RESTAPI/internal/core/transport/http/server"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
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
}

type IdentityHTTPHandler struct {
	identityService IdentityService
}

func NewIdentityHTTPHandler(identityService IdentityService) *IdentityHTTPHandler {
	return &IdentityHTTPHandler{identityService: identityService}
}

func (h *IdentityHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
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
	}
}
