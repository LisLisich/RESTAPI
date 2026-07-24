package identity_transport_http

import (
	"net/http"

	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_request "github.com/LisLisich/RESTAPI/internal/core/transport/http/request"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
)

type APITokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// LoginAPI godoc
// @Summary Получить JWT для CLI
// @Description Возвращает Ed25519 access JWT и непрозрачный refresh token
// @Tags identity
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Учетные данные"
// @Success 200 {object} APITokenResponse "Пара токенов"
// @Failure 401 {object} core_http_response.ErrorResponse "Неверные учетные данные"
// @Router /api/v2/auth/token [post]
func (h *IdentityHTTPHandler) LoginAPI(rw http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(r.Context()),
		rw,
	)
	var request LoginRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate API login")
		return
	}
	pair, err := h.identityService.LoginAPI(
		r.Context(),
		identity_service.LoginInput{Email: request.Email, Password: request.Password},
	)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to issue API tokens")
		return
	}
	responseHandler.JSONResponse(apiTokenResponse(pair), http.StatusOK)
}

type RefreshAPIRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshAPI godoc
// @Summary Обновить JWT
// @Description Одноразово ротирует refresh token и выдает новую пару
// @Tags identity
// @Accept json
// @Produce json
// @Param request body RefreshAPIRequest true "Refresh token"
// @Success 200 {object} APITokenResponse "Новая пара токенов"
// @Failure 401 {object} core_http_response.ErrorResponse "Refresh token недействителен"
// @Router /api/v2/auth/token/refresh [post]
func (h *IdentityHTTPHandler) RefreshAPI(rw http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(r.Context()),
		rw,
	)
	var request RefreshAPIRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate refresh request")
		return
	}
	pair, err := h.identityService.RefreshAPI(r.Context(), request.RefreshToken)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to rotate API tokens")
		return
	}
	responseHandler.JSONResponse(apiTokenResponse(pair), http.StatusOK)
}

func apiTokenResponse(pair identity_service.TokenPair) APITokenResponse {
	return APITokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    pair.TokenType,
		ExpiresIn:    pair.ExpiresIn,
	}
}
