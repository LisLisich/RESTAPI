package identity_transport_http

import (
	"net/http"

	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_request "github.com/LisLisich/RESTAPI/internal/core/transport/http/request"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
)

type RequestPasswordResetRequest struct {
	Email string `json:"email" validate:"required,email,max=254" example:"user@example.com"`
}

type RequestPasswordResetResponse struct {
	Message string `json:"message"`
}

// RequestPasswordReset godoc
// @Summary Запросить сброс пароля
// @Description Создает одноразовый токен, если активный аккаунт существует
// @Tags identity
// @Accept json
// @Produce json
// @Param request body RequestPasswordResetRequest true "Email аккаунта"
// @Success 202 {object} RequestPasswordResetResponse "Запрос принят"
// @Failure 400 {object} core_http_response.ErrorResponse "Некорректный запрос"
// @Failure 500 {object} core_http_response.ErrorResponse "Внутренняя ошибка"
// @Router /api/v2/auth/password-reset/request [post]
func (h *IdentityHTTPHandler) RequestPasswordReset(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(r.Context()),
		rw,
	)
	var request RequestPasswordResetRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate password reset request")
		return
	}
	if err := h.identityService.RequestPasswordReset(r.Context(), request.Email); err != nil {
		responseHandler.ErrorResponse(err, "failed to request password reset")
		return
	}
	responseHandler.JSONResponse(
		RequestPasswordResetResponse{
			Message: "password reset instructions will be sent if the account exists",
		},
		http.StatusAccepted,
	)
}

type ConfirmPasswordResetRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=12,max=128"`
}

// ConfirmPasswordReset godoc
// @Summary Установить новый пароль
// @Description Потребляет одноразовый токен, меняет пароль и отзывает старые сессии
// @Tags identity
// @Accept json
// @Produce json
// @Param request body ConfirmPasswordResetRequest true "Токен и новый пароль"
// @Success 204 "Пароль изменен"
// @Failure 400 {object} core_http_response.ErrorResponse "Токен неверен, просрочен или использован"
// @Failure 500 {object} core_http_response.ErrorResponse "Внутренняя ошибка"
// @Router /api/v2/auth/password-reset/confirm [post]
func (h *IdentityHTTPHandler) ConfirmPasswordReset(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(r.Context()),
		rw,
	)
	var request ConfirmPasswordResetRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate password reset confirmation")
		return
	}
	if err := h.identityService.ResetPassword(
		r.Context(),
		request.Token,
		request.NewPassword,
	); err != nil {
		responseHandler.ErrorResponse(err, "failed to reset password")
		return
	}
	responseHandler.NoContentResponse()
}
