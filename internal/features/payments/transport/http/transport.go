package payments_transport_http

import (
	"context"
	"net/http"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_http_middleware "github.com/LisLisich/RESTAPI/internal/core/transport/http/middleware"
	core_http_server "github.com/LisLisich/RESTAPI/internal/core/transport/http/server"
)

type PaymentService interface {
	CreatePayment(
		ctx context.Context,
		accountUserID int,
		idempotencyKey string,
		amountMinor int64,
	) (domain.Payment, error)
	HandleSucceededWebhook(
		ctx context.Context,
		providerPaymentID string,
	) (bool, error)
}

type PaymentHTTPHandler struct {
	service                  PaymentService
	authenticationMiddleware core_http_middleware.Middleware
}

func NewPaymentHTTPHandler(
	service PaymentService,
	authenticationMiddleware core_http_middleware.Middleware,
) *PaymentHTTPHandler {
	return &PaymentHTTPHandler{
		service:                  service,
		authenticationMiddleware: authenticationMiddleware,
	}
}

func (handler *PaymentHTTPHandler) Routes() []core_http_server.Route {
	protectedMiddleware := []core_http_middleware.Middleware{}
	if handler.authenticationMiddleware != nil {
		protectedMiddleware = append(
			protectedMiddleware,
			handler.authenticationMiddleware,
		)
	}
	return []core_http_server.Route{
		{
			Method:     http.MethodPost,
			Path:       "/payments",
			Handler:    handler.CreatePayment,
			Middleware: protectedMiddleware,
		},
		{
			Method:  http.MethodPost,
			Path:    "/payments/webhook/yookassa",
			Handler: handler.YooKassaWebhook,
		},
	}
}
