package payments_transport_http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	identity_service "github.com/LisLisich/fintask/internal/features/identity/service"
	identity_http_middleware "github.com/LisLisich/fintask/internal/features/identity/transport/http/middleware"
	"go.uber.org/zap"
)

type fakePaymentService struct {
	payment           domain.Payment
	createUserID      int
	createKey         string
	createAmountMinor int64
	webhookPaymentID  string
}

func (service *fakePaymentService) CreatePayment(
	_ context.Context,
	userID int,
	idempotencyKey string,
	amountMinor int64,
) (domain.Payment, error) {
	service.createUserID = userID
	service.createKey = idempotencyKey
	service.createAmountMinor = amountMinor
	return service.payment, nil
}

func (service *fakePaymentService) HandleSucceededWebhook(
	_ context.Context,
	providerPaymentID string,
) (bool, error) {
	service.webhookPaymentID = providerPaymentID
	return true, nil
}

func TestCreatePaymentUsesPrincipalAndIdempotencyHeader(t *testing.T) {
	service := &fakePaymentService{payment: domain.Payment{
		ID:                "local-id",
		ProviderPaymentID: "provider-id",
		Status:            domain.PaymentStatusPending,
		Amount:            domain.Money{MinorUnits: 12550, Currency: domain.CurrencyRUB},
		ConfirmationURL:   "https://pay.example/confirm",
	}}
	handler := NewPaymentHTTPHandler(service, nil)
	request := newRequest(
		http.MethodPost,
		"/payments",
		[]byte(`{"amount_minor":12550}`),
	)
	request.Header.Set("Idempotency-Key", "client-key")
	request = request.WithContext(identity_http_middleware.ContextWithPrincipal(
		request.Context(),
		identity_service.Principal{UserID: 42},
	))
	response := httptest.NewRecorder()

	handler.CreatePayment(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body)
	}
	if service.createUserID != 42 ||
		service.createKey != "client-key" ||
		service.createAmountMinor != 12550 {
		t.Fatal("unexpected service input")
	}
}

func TestWebhookProcessesPaymentSucceeded(t *testing.T) {
	service := &fakePaymentService{}
	handler := NewPaymentHTTPHandler(service, nil)
	request := newRequest(
		http.MethodPost,
		"/payments/webhook/yookassa",
		[]byte(`{
			"type":"notification",
			"event":"payment.succeeded",
			"object":{"id":"provider-id"}
		}`),
	)
	response := httptest.NewRecorder()

	handler.YooKassaWebhook(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body)
	}
	if service.webhookPaymentID != "provider-id" {
		t.Fatalf("unexpected provider payment id %q", service.webhookPaymentID)
	}
}

func newRequest(method string, path string, body []byte) *http.Request {
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	logger := &core_logger.Logger{Logger: zap.NewNop()}
	return request.WithContext(core_logger.ToContext(request.Context(), logger))
}
