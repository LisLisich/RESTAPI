package identity_transport_http

import (
	"fmt"
	"net/http"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
)

const googleStateCookieName = "__Host-fintask_google_state"

func (h *IdentityHTTPHandler) StartGoogleLogin(
	rw http.ResponseWriter,
	request *http.Request,
) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(request.Context()),
		rw,
	)
	if h.googleLoginService == nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("Google login is disabled: %w", core_errors.ErrNotFound),
			"Google login is disabled",
		)
		return
	}
	start, err := h.googleLoginService.Begin(request.Context())
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to start Google login")
		return
	}
	http.SetCookie(rw, &http.Cookie{
		Name: googleStateCookieName, Value: start.State,
		Path: "/", Expires: start.ExpiresAt,
		MaxAge:   max(int(time.Until(start.ExpiresAt).Seconds()), 0),
		HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(rw, request, start.AuthorizationURL, http.StatusFound)
}

func (h *IdentityHTTPHandler) CompleteGoogleLogin(
	rw http.ResponseWriter,
	request *http.Request,
) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(request.Context()),
		rw,
	)
	stateCookie, err := request.Cookie(googleStateCookieName)
	if err != nil {
		responseHandler.ErrorResponse(
			fmt.Errorf("Google state cookie is required: %w", core_errors.ErrForbidden),
			"invalid Google login state",
		)
		return
	}
	if request.URL.Query().Get("iss") != "https://accounts.google.com" {
		responseHandler.ErrorResponse(
			fmt.Errorf("unexpected Google issuer: %w", core_errors.ErrForbidden),
			"invalid Google issuer",
		)
		return
	}
	session, err := h.googleLoginService.Complete(
		request.Context(),
		request.URL.Query().Get("code"),
		request.URL.Query().Get("state"),
		stateCookie.Value,
	)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to complete Google login")
		return
	}
	http.SetCookie(rw, &http.Cookie{
		Name: SessionCookieName, Value: session.Token, Path: "/",
		Expires:  session.ExpiresAt,
		MaxAge:   max(int(time.Until(session.ExpiresAt).Seconds()), 0),
		HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(rw, &http.Cookie{
		Name: CSRFCookieName, Value: session.CSRFToken, Path: "/",
		Expires: session.ExpiresAt,
		MaxAge:  max(int(time.Until(session.ExpiresAt).Seconds()), 0),
		Secure:  true, SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(rw, &http.Cookie{
		Name: googleStateCookieName, Value: "", Path: "/",
		MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
	})
	responseHandler.JSONResponse(LoginResponse{
		UserID: session.Account.UserID,
		Email:  session.Account.Email,
	}, http.StatusOK)
}
