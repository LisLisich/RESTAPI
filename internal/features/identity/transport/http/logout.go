package identity_transport_http

import (
	"net/http"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
)

// Logout godoc
// @Summary Завершить текущую сессию
// @Description Отзывает серверную сессию и очищает session/CSRF cookies
// @Tags identity
// @Produce json
// @Success 204 "Сессия завершена"
// @Failure 401 {object} core_http_response.ErrorResponse "Требуется аутентификация"
// @Failure 403 {object} core_http_response.ErrorResponse "Некорректный CSRF-токен"
// @Failure 500 {object} core_http_response.ErrorResponse "Внутренняя ошибка"
// @Router /auth/logout [post]
func (h *IdentityHTTPHandler) Logout(rw http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(r.Context()),
		rw,
	)
	sessionCookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		responseHandler.ErrorResponse(
			core_errors.ErrUnauthorized,
			"authentication required",
		)
		return
	}
	if err := h.identityService.Logout(r.Context(), sessionCookie.Value); err != nil {
		responseHandler.ErrorResponse(err, "failed to logout")
		return
	}

	clearAuthCookie(rw, SessionCookieName, true, http.SameSiteLaxMode)
	clearAuthCookie(rw, CSRFCookieName, false, http.SameSiteStrictMode)
	responseHandler.NoContentResponse()
}

func clearAuthCookie(
	rw http.ResponseWriter,
	name string,
	httpOnly bool,
	sameSite http.SameSite,
) {
	http.SetCookie(
		rw,
		&http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0).UTC(),
			MaxAge:   -1,
			HttpOnly: httpOnly,
			Secure:   true,
			SameSite: sameSite,
		},
	)
}
