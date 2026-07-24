package payments_transport_http

import (
	"fmt"
	"net/http"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_request "github.com/LisLisich/RESTAPI/internal/core/transport/http/request"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
)

type YooKassaWebhookRequest struct {
	Type   string `json:"type" validate:"required"`
	Event  string `json:"event" validate:"required"`
	Object struct {
		ID string `json:"id" validate:"required"`
	} `json:"object" validate:"required"`
}

func (handler *PaymentHTTPHandler) YooKassaWebhook(
	rw http.ResponseWriter,
	request *http.Request,
) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(request.Context()),
		rw,
	)
	var body YooKassaWebhookRequest
	if err := core_http_request.DecodeAndValidateRequest(request, &body); err != nil {
		responseHandler.ErrorResponse(err, "invalid YooKassa webhook")
		return
	}
	if body.Type != "notification" || body.Event != "payment.succeeded" {
		responseHandler.ErrorResponse(
			fmt.Errorf("unsupported YooKassa event: %w", core_errors.ErrInvalidArgument),
			"unsupported YooKassa event",
		)
		return
	}
	credited, err := handler.service.HandleSucceededWebhook(
		request.Context(),
		body.Object.ID,
	)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to process YooKassa webhook")
		return
	}
	responseHandler.JSONResponse(
		map[string]bool{"credited": credited},
		http.StatusOK,
	)
}
