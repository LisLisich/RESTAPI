package identity_transport_http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
	"go.uber.org/zap"
)

type fakeIdentityService struct {
	registerCalled bool
	input          identity_service.RegisterAccountInput
	err            error
}

func (s *fakeIdentityService) RegisterAccount(
	_ context.Context,
	input identity_service.RegisterAccountInput,
) (domain.Account, error) {
	s.registerCalled = true
	s.input = input
	if s.err != nil {
		return domain.Account{}, s.err
	}

	account, err := domain.NewAccountUninitialized(42, input.Email)
	if err != nil {
		return domain.Account{}, err
	}
	return account, nil
}

func TestRegisterAccountReturnsCreatedWithoutPassword(t *testing.T) {
	service := &fakeIdentityService{}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequest(`{
		"full_name": "Ivan Ivanov",
		"email": "user@example.com",
		"password": "strong-password"
	}`)
	response := httptest.NewRecorder()

	handler.RegisterAccount(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}
	if !service.registerCalled {
		t.Fatal("expected RegisterAccount to be called")
	}
	if service.input.Email != "user@example.com" {
		t.Fatalf("expected email %q, got %q", "user@example.com", service.input.Email)
	}
	if strings.Contains(response.Body.String(), "strong-password") {
		t.Fatal("password must not be returned")
	}
	if !strings.Contains(response.Body.String(), `"status":"pending_verification"`) {
		t.Fatalf("expected pending verification response, got %q", response.Body.String())
	}
}

func TestRegisterAccountRejectsMalformedRequest(t *testing.T) {
	service := &fakeIdentityService{}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequest(`{"email":`)
	response := httptest.NewRecorder()

	handler.RegisterAccount(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.registerCalled {
		t.Fatal("service must not be called for malformed request")
	}
}

func TestRegisterAccountReturnsConflict(t *testing.T) {
	service := &fakeIdentityService{
		err: fmt.Errorf("register: %w", core_errors.ErrConflict),
	}
	handler := NewIdentityHTTPHandler(service)
	request := newIdentityRequest(`{
		"full_name": "Ivan Ivanov",
		"email": "user@example.com",
		"password": "strong-password"
	}`)
	response := httptest.NewRecorder()

	handler.RegisterAccount(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, response.Code)
	}
}

func TestIdentityRoutesExposeV2RegistrationPath(t *testing.T) {
	handler := NewIdentityHTTPHandler(&fakeIdentityService{})
	routes := handler.Routes()

	if len(routes) != 1 {
		t.Fatalf("expected one route, got %d", len(routes))
	}
	if routes[0].Method != http.MethodPost || routes[0].Path != "/auth/register" {
		t.Fatalf("unexpected registration route: %s %s", routes[0].Method, routes[0].Path)
	}
}

func newIdentityRequest(body string) *http.Request {
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)
	logger := &core_logger.Logger{Logger: zap.NewNop()}
	return request.WithContext(core_logger.ToContext(request.Context(), logger))
}
