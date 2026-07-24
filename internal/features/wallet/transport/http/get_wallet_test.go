package wallet_transport_http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	identity_service "github.com/LisLisich/fintask/internal/features/identity/service"
	identity_http_middleware "github.com/LisLisich/fintask/internal/features/identity/transport/http/middleware"
	"go.uber.org/zap"
)

type fakeWalletService struct {
	userID int
}

func (service *fakeWalletService) GetWallet(
	_ context.Context,
	userID int,
) (domain.Wallet, error) {
	service.userID = userID
	return domain.NewWallet(
		userID,
		3,
		domain.Money{MinorUnits: 12500, Currency: domain.CurrencyRUB},
	)
}

func TestGetWalletReturnsIntegerMinorUnits(t *testing.T) {
	service := &fakeWalletService{}
	handler := NewWalletHTTPHandler(service, nil)
	request := httptest.NewRequest(http.MethodGet, "/wallet", nil)
	logger := &core_logger.Logger{Logger: zap.NewNop()}
	request = request.WithContext(core_logger.ToContext(request.Context(), logger))
	request = request.WithContext(
		identity_http_middleware.ContextWithPrincipal(
			request.Context(),
			identity_service.Principal{UserID: 42},
		),
	)
	response := httptest.NewRecorder()

	handler.GetWallet(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if service.userID != 42 {
		t.Fatalf("expected authenticated user 42, got %d", service.userID)
	}
	if !strings.Contains(response.Body.String(), `"balance_minor":12500`) {
		t.Fatalf("expected integer minor units, got %q", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"currency":"RUB"`) {
		t.Fatalf("expected RUB currency, got %q", response.Body.String())
	}
}
