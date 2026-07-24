package identity_transport_http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoginAPIReturnsBearerTokenPair(t *testing.T) {
	service := &fakeIdentityService{}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequestForPath(
		"/auth/token",
		`{"email":"user@example.com","password":"strong-password"}`,
	)
	response := httptest.NewRecorder()

	handler.LoginAPI(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.loginAPICalled {
		t.Fatal("expected LoginAPI call")
	}
	if !strings.Contains(response.Body.String(), `"access_token":"access-token"`) ||
		!strings.Contains(response.Body.String(), `"refresh_token":"refresh-token"`) {
		t.Fatalf("unexpected token response %q", response.Body.String())
	}
}

func TestRefreshAPIReturnsRotatedPair(t *testing.T) {
	service := &fakeIdentityService{}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequestForPath(
		"/auth/token/refresh",
		`{"refresh_token":"old-refresh-token"}`,
	)
	response := httptest.NewRecorder()

	handler.RefreshAPI(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.refreshAPICalled || service.refreshToken != "old-refresh-token" {
		t.Fatal("expected refresh token rotation")
	}
	if !strings.Contains(response.Body.String(), `"refresh_token":"new-refresh-token"`) {
		t.Fatalf("unexpected token response %q", response.Body.String())
	}
}
