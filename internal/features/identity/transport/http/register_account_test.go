package identity_transport_http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

	verifyCalled bool
	rawToken     string
	verifyErr    error

	loginCalled bool
	loginInput  identity_service.LoginInput
	loginErr    error
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

func (s *fakeIdentityService) Login(
	_ context.Context,
	input identity_service.LoginInput,
) (identity_service.BrowserSession, error) {
	s.loginCalled = true
	s.loginInput = input
	if s.loginErr != nil {
		return identity_service.BrowserSession{}, s.loginErr
	}

	account := domain.Account{
		UserID: 42,
		Email:  "user@example.com",
		Status: domain.AccountStatusActive,
	}
	return identity_service.BrowserSession{
		Account:   account,
		Token:     "session-token",
		CSRFToken: "csrf-token",
		ExpiresAt: time.Now().Add(12 * time.Hour),
	}, nil
}

func (s *fakeIdentityService) VerifyEmail(
	_ context.Context,
	rawToken string,
) (domain.Account, error) {
	s.verifyCalled = true
	s.rawToken = rawToken
	if s.verifyErr != nil {
		return domain.Account{}, s.verifyErr
	}

	account, err := domain.NewAccountUninitialized(42, "user@example.com")
	if err != nil {
		return domain.Account{}, err
	}
	if err := account.VerifyEmail(time.Date(2026, time.July, 24, 14, 0, 0, 0, time.UTC)); err != nil {
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

	if len(routes) != 3 {
		t.Fatalf("expected three routes, got %d", len(routes))
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
