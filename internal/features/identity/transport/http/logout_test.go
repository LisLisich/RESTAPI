package identity_transport_http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogoutRevokesSessionAndClearsCookies(t *testing.T) {
	service := &fakeIdentityService{}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequestForPath("/auth/logout", `{}`)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "raw-session-token"})
	response := httptest.NewRecorder()

	handler.Logout(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if !service.logoutCalled || service.logoutToken != "raw-session-token" {
		t.Fatal("expected current session revocation")
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected two cleared cookies, got %d", len(cookies))
	}
	for _, cookie := range cookies {
		if cookie.MaxAge != -1 || cookie.Value != "" {
			t.Fatalf("expected cleared cookie, got %+v", cookie)
		}
	}
}

func TestLogoutRouteUsesProtectedMiddleware(t *testing.T) {
	middlewareCalled := false
	protectedMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(rw, r)
		})
	}
	service := &fakeIdentityService{}
	handler := NewIdentityHTTPHandler(service, protectedMiddleware)
	routes := handler.Routes()
	request := newIdentityRequestForPath("/auth/logout", `{}`)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "raw-session-token"})
	response := httptest.NewRecorder()

	routes[5].WithMiddleware().ServeHTTP(response, request)

	if !middlewareCalled {
		t.Fatal("expected protected middleware on logout")
	}
}
