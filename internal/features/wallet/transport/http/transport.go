package wallet_transport_http

import (
	"context"
	"net/http"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_http_middleware "github.com/LisLisich/RESTAPI/internal/core/transport/http/middleware"
	core_http_server "github.com/LisLisich/RESTAPI/internal/core/transport/http/server"
)

type WalletService interface {
	GetWallet(ctx context.Context, userID int) (domain.Wallet, error)
}

type WalletHTTPHandler struct {
	service           WalletService
	sessionMiddleware core_http_middleware.Middleware
}

func NewWalletHTTPHandler(
	service WalletService,
	sessionMiddleware core_http_middleware.Middleware,
) *WalletHTTPHandler {
	return &WalletHTTPHandler{
		service:           service,
		sessionMiddleware: sessionMiddleware,
	}
}

func (handler *WalletHTTPHandler) Routes() []core_http_server.Route {
	middleware := []core_http_middleware.Middleware{}
	if handler.sessionMiddleware != nil {
		middleware = append(middleware, handler.sessionMiddleware)
	}
	return []core_http_server.Route{
		{
			Method:     http.MethodGet,
			Path:       "/wallet",
			Handler:    handler.GetWallet,
			Middleware: middleware,
		},
	}
}
