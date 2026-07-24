package identity_transport_http

import (
	"net/http"

	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	core_http_request "github.com/LisLisich/fintask/internal/core/transport/http/request"
	core_http_response "github.com/LisLisich/fintask/internal/core/transport/http/response"
)

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required" example:"opaque-verification-token"`
}

// VerifyEmail godoc
// @Summary Подтвердить email
// @Description Активирует аккаунт по одноразовому токену подтверждения
// @Tags identity
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Токен подтверждения"
// @Success 204 "Email подтвержден"
// @Failure 400 {object} core_http_response.ErrorResponse "Токен неверен, просрочен или уже использован"
// @Failure 500 {object} core_http_response.ErrorResponse "Внутренняя ошибка"
// @Router /api/v2/auth/verify-email [post]
func (h *IdentityHTTPHandler) VerifyEmail(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request VerifyEmailRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate email verification request")
		return
	}

	if _, err := h.identityService.VerifyEmail(ctx, request.Token); err != nil {
		responseHandler.ErrorResponse(err, "failed to verify email")
		return
	}

	responseHandler.NoContentResponse()
}
