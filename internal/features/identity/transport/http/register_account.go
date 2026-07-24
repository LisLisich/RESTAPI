package identity_transport_http

import (
	"net/http"

	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	core_http_request "github.com/LisLisich/fintask/internal/core/transport/http/request"
	core_http_response "github.com/LisLisich/fintask/internal/core/transport/http/response"
	identity_service "github.com/LisLisich/fintask/internal/features/identity/service"
)

type RegisterAccountRequest struct {
	FullName string `json:"full_name" validate:"required,min=3,max=100" example:"Ivan Ivanov"`
	Email    string `json:"email" validate:"required,email,max=254" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=12,max=128" example:"strong-password"`
}

type RegisterAccountResponse struct {
	UserID int    `json:"user_id" example:"42"`
	Email  string `json:"email" example:"user@example.com"`
	Status string `json:"status" example:"pending_verification"`
}

// RegisterAccount godoc
// @Summary Зарегистрировать аккаунт
// @Description Создает аккаунт и отправляет запрос на подтверждение email через outbox
// @Tags identity
// @Accept json
// @Produce json
// @Param request body RegisterAccountRequest true "Данные регистрации"
// @Success 201 {object} RegisterAccountResponse "Аккаунт ожидает подтверждения email"
// @Failure 400 {object} core_http_response.ErrorResponse "Некорректный запрос"
// @Failure 409 {object} core_http_response.ErrorResponse "Email уже зарегистрирован"
// @Failure 500 {object} core_http_response.ErrorResponse "Внутренняя ошибка"
// @Router /api/v2/auth/register [post]
func (h *IdentityHTTPHandler) RegisterAccount(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request RegisterAccountRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate registration request")
		return
	}

	account, err := h.identityService.RegisterAccount(
		ctx,
		identity_service.RegisterAccountInput{
			FullName: request.FullName,
			Email:    request.Email,
			Password: request.Password,
		},
	)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to register account")
		return
	}

	responseHandler.JSONResponse(
		RegisterAccountResponse{
			UserID: account.UserID,
			Email:  account.Email,
			Status: string(account.Status),
		},
		http.StatusCreated,
	)
}
