package identity_transport_http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestLoginSetsProtectedSessionAndCSRFCookies(t *testing.T) {
	service := &fakeIdentityService{}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequestForPath(
		"/auth/login",
		`{"email":"user@example.com","password":"strong-password"}`,
	)
	response := httptest.NewRecorder()

	handler.Login(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.loginCalled {
		t.Fatal("expected Login to be called")
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected two cookies, got %d", len(cookies))
	}

	sessionCookie := findCookie(t, cookies, SessionCookieName)
	if sessionCookie.Value != "session-token" {
		t.Fatalf("expected session token cookie, got %q", sessionCookie.Value)
	}
	if !sessionCookie.HttpOnly || !sessionCookie.Secure {
		t.Fatal("session cookie must be HttpOnly and Secure")
	}
	if sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax, got %v", sessionCookie.SameSite)
	}

	csrfCookie := findCookie(t, cookies, CSRFCookieName)
	if csrfCookie.Value != "csrf-token" {
		t.Fatalf("expected csrf token cookie, got %q", csrfCookie.Value)
	}
	if csrfCookie.HttpOnly {
		t.Fatal("csrf cookie must be readable by frontend")
	}
	if !csrfCookie.Secure || csrfCookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("csrf cookie must be Secure and SameSite=Strict")
	}
}

func TestLoginReturnsUnauthorizedWithoutCookies(t *testing.T) {
	service := &fakeIdentityService{
		loginErr: fmt.Errorf("login: %w", core_errors.ErrUnauthorized),
	}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequestForPath(
		"/auth/login",
		`{"email":"user@example.com","password":"wrong-password"}`,
	)
	response := httptest.NewRecorder()

	handler.Login(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
	if len(response.Result().Cookies()) != 0 {
		t.Fatal("failed login must not set cookies")
	}
}

func TestIdentityRoutesExposeLoginPath(t *testing.T) {
	handler := NewIdentityHTTPHandler(&fakeIdentityService{})
	routes := handler.Routes()

	if routes[2].Method != http.MethodPost || routes[2].Path != "/auth/login" {
		t.Fatalf("unexpected login route: %s %s", routes[2].Method, routes[2].Path)
	}
}

func findCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}
