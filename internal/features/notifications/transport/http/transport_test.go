package notifications_transport_http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	identity_service "github.com/LisLisich/fintask/internal/features/identity/service"
	identity_http_middleware "github.com/LisLisich/fintask/internal/features/identity/transport/http/middleware"
	notifications_service "github.com/LisLisich/fintask/internal/features/notifications/service"
	"go.uber.org/zap"
)

type fakeReader struct {
	userID        int
	notifications []notifications_service.Notification
}

func (reader *fakeReader) ListNotifications(
	_ context.Context,
	userID int,
) ([]notifications_service.Notification, error) {
	reader.userID = userID
	return reader.notifications, nil
}

func TestListNotificationsUsesAuthenticatedOwner(t *testing.T) {
	reader := &fakeReader{notifications: []notifications_service.Notification{{
		ID:        1,
		Title:     "Кошелек пополнен",
		Body:      "На кошелек зачислено 1.00 RUB.",
		CreatedAt: time.Now(),
	}}}
	handler := NewNotificationHTTPHandler(reader, nil)
	request := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	logger := &core_logger.Logger{Logger: zap.NewNop()}
	request = request.WithContext(core_logger.ToContext(request.Context(), logger))
	request = request.WithContext(identity_http_middleware.ContextWithPrincipal(
		request.Context(),
		identity_service.Principal{UserID: 42},
	))
	response := httptest.NewRecorder()

	handler.ListNotifications(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
	}
	if reader.userID != 42 {
		t.Fatalf("expected user 42, got %d", reader.userID)
	}
}
