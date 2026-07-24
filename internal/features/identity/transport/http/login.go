package identity_transport_http

import (
	"net/http"
	"time"

	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	core_http_request "github.com/LisLisich/fintask/internal/core/transport/http/request"
	core_http_response "github.com/LisLisich/fintask/internal/core/transport/http/response"
	identity_service "github.com/LisLisich/fintask/internal/features/identity/service"
)

const (
	SessionCookieName = "__Host-fintask_session"
	CSRFCookieName    = "__Host-fintask_csrf"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=254" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=12,max=128" example:"strong-password"`
}

type LoginResponse struct {
	UserID int    `json:"user_id" example:"42"`
	Email  string `json:"email" example:"user@example.com"`
}

// Login godoc
// @Summary Войти по email и паролю
// @Description Создает серверную cookie-сессию и отдельный CSRF-токен
// @Tags identity
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Учетные данные"
// @Success 200 {object} LoginResponse "Вход выполнен"
// @Failure 400 {object} core_http_response.ErrorResponse "Некорректный запрос"
// @Failure 401 {object} core_http_response.ErrorResponse "Неверные учетные данные"
// @Failure 500 {object} core_http_response.ErrorResponse "Внутренняя ошибка"
// @Router /api/v2/auth/login [post]
func (h *IdentityHTTPHandler) Login(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request LoginRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate login request")
		return
	}

	session, err := h.identityService.Login(
		ctx,
		identity_service.LoginInput{
			Email:    request.Email,
			Password: request.Password,
		},
	)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to login")
		return
	}

	maxAge := max(int(time.Until(session.ExpiresAt).Seconds()), 0)
	http.SetCookie(
		rw,
		&http.Cookie{
			Name:     SessionCookieName,
			Value:    session.Token,
			Path:     "/",
			Expires:  session.ExpiresAt,
			MaxAge:   maxAge,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		},
	)
	http.SetCookie(
		rw,
		&http.Cookie{
			Name:     CSRFCookieName,
			Value:    session.CSRFToken,
			Path:     "/",
			Expires:  session.ExpiresAt,
			MaxAge:   maxAge,
			HttpOnly: false,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		},
	)

	responseHandler.JSONResponse(
		LoginResponse{
			UserID: session.Account.UserID,
			Email:  session.Account.Email,
		},
		http.StatusOK,
	)
}
