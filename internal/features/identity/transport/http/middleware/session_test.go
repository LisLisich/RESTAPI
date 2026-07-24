package identity_http_middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
	"go.uber.org/zap"
)

type fakeSessionAuthenticator struct {
	called          bool
	sessionToken    string
	csrfToken       string
	requireCSRF     bool
	principal       identity_service.Principal
	authenticateErr error
}

func (a *fakeSessionAuthenticator) AuthenticateSession(
	_ context.Context,
	rawSessionToken string,
	rawCSRFToken string,
	requireCSRF bool,
) (identity_service.Principal, error) {
	a.called = true
	a.sessionToken = rawSessionToken
	a.csrfToken = rawCSRFToken
	a.requireCSRF = requireCSRF
	if a.authenticateErr != nil {
		return identity_service.Principal{}, a.authenticateErr
	}
	return a.principal, nil
}

func TestSessionMiddlewareAddsPrincipalForSafeRequest(t *testing.T) {
	authenticator := &fakeSessionAuthenticator{
		principal: identity_service.Principal{UserID: 42},
	}
	var gotUserID int
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		principal, ok := PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("expected principal in context")
		}
		gotUserID = principal.UserID
		rw.WriteHeader(http.StatusNoContent)
	})
	handler := Session(authenticator, "session-cookie")(next)
	request := newMiddlewareRequest(http.MethodGet)
	request.AddCookie(&http.Cookie{Name: "session-cookie", Value: "raw-session"})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if gotUserID != 42 {
		t.Fatalf("expected user id 42, got %d", gotUserID)
	}
	if authenticator.requireCSRF {
		t.Fatal("safe GET request must not require CSRF token")
	}
}

func TestSessionMiddlewareRequiresCSRFForMutation(t *testing.T) {
	authenticator := &fakeSessionAuthenticator{
		principal: identity_service.Principal{UserID: 42},
	}
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not run after CSRF rejection")
	})
	handler := Session(authenticator, "session-cookie")(next)
	request := newMiddlewareRequest(http.MethodPost)
	request.AddCookie(&http.Cookie{Name: "session-cookie", Value: "raw-session"})
	request.Header.Set(CSRFHeaderName, "raw-csrf")
	response := httptest.NewRecorder()

	authenticator.authenticateErr = fmt.Errorf("csrf: %w", core_errors.ErrForbidden)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
	if !authenticator.requireCSRF || authenticator.csrfToken != "raw-csrf" {
		t.Fatal("expected mutation to require CSRF header")
	}
}

func TestSessionMiddlewareRejectsMissingCookie(t *testing.T) {
	authenticator := &fakeSessionAuthenticator{}
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not run without session")
	})
	handler := Session(authenticator, "session-cookie")(next)
	request := newMiddlewareRequest(http.MethodGet)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
	if authenticator.called {
		t.Fatal("authenticator must not be called without session cookie")
	}
}

func newMiddlewareRequest(method string) *http.Request {
	request := httptest.NewRequest(method, "/protected", nil)
	logger := &core_logger.Logger{Logger: zap.NewNop()}
	return request.WithContext(core_logger.ToContext(request.Context(), logger))
}
