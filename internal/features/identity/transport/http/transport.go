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
	}
}
