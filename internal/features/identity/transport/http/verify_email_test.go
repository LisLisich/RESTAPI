package identity_transport_http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerifyEmailReturnsNoContent(t *testing.T) {
	service := &fakeIdentityService{}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequestForPath(
		"/auth/verify-email",
		`{"token":"raw-verification-token"}`,
	)
	response := httptest.NewRecorder()

	handler.VerifyEmail(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if !service.verifyCalled || service.rawToken != "raw-verification-token" {
		t.Fatal("expected raw verification token to be passed to service")
	}
	if response.Body.Len() != 0 {
		t.Fatalf("expected empty response body, got %q", response.Body.String())
	}
}

func TestVerifyEmailRejectsMissingToken(t *testing.T) {
	service := &fakeIdentityService{}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequestForPath("/auth/verify-email", `{}`)
	response := httptest.NewRecorder()

	handler.VerifyEmail(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.verifyCalled {
		t.Fatal("service must not be called without token")
	}
}

func TestIdentityRoutesExposeEmailVerificationPath(t *testing.T) {
	handler := NewIdentityHTTPHandler(&fakeIdentityService{})
	routes := handler.Routes()

	if routes[1].Method != http.MethodPost || routes[1].Path != "/auth/verify-email" {
		t.Fatalf("unexpected verification route: %s %s", routes[1].Method, routes[1].Path)
	}
}

func newIdentityRequestForPath(path string, body string) *http.Request {
	request := newIdentityRequest(body)
	request.URL.Path = path
	request.Body = http.NoBody
	request.Body = ioNopCloser{Reader: strings.NewReader(body)}
	return request
}

type ioNopCloser struct {
	*strings.Reader
}

func (ioNopCloser) Close() error {
	return nil
}
