package payments_transport_http

import (
	"fmt"
	"net/http"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_request "github.com/LisLisich/RESTAPI/internal/core/transport/http/request"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
	identity_http_middleware "github.com/LisLisich/RESTAPI/internal/features/identity/transport/http/middleware"
)

type CreatePaymentRequest struct {
	AmountMinor int64 `json:"amount_minor" validate:"required,gt=0"`
}

type PaymentResponse struct {
	ID                string `json:"id"`
	ProviderPaymentID string `json:"provider_payment_id"`
	Status            string `json:"status"`
	AmountMinor       int64  `json:"amount_minor"`
	Currency          string `json:"currency"`
	ConfirmationURL   string `json:"confirmation_url"`
}

func (handler *PaymentHTTPHandler) CreatePayment(
	rw http.ResponseWriter,
	request *http.Request,
) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(request.Context()),
		rw,
	)
	principal, ok := identity_http_middleware.PrincipalFromContext(request.Context())
	if !ok {
		responseHandler.ErrorResponse(
			fmt.Errorf("authenticated principal is required: %w", core_errors.ErrUnauthorized),
			"authentication required",
		)
		return
	}
	var body CreatePaymentRequest
	if err := core_http_request.DecodeAndValidateRequest(request, &body); err != nil {
		responseHandler.ErrorResponse(err, "invalid payment request")
		return
	}
	payment, err := handler.service.CreatePayment(
		request.Context(),
		principal.UserID,
		request.Header.Get("Idempotency-Key"),
		body.AmountMinor,
	)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create payment")
		return
	}
	responseHandler.JSONResponse(
		PaymentResponse{
			ID:                payment.ID,
			ProviderPaymentID: payment.ProviderPaymentID,
			Status:            string(payment.Status),
			AmountMinor:       payment.Amount.MinorUnits,
			Currency:          string(payment.Amount.Currency),
			ConfirmationURL:   payment.ConfirmationURL,
		},
		http.StatusCreated,
	)
}
