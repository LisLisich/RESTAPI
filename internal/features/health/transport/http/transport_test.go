package health_transport_http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	"go.uber.org/zap"
)

func TestHealthReturnsOK(t *testing.T) {
	handler := NewHealthHTTPHandler()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request = request.WithContext(core_logger.ToContext(
		request.Context(),
		&core_logger.Logger{Logger: zap.NewNop()},
	))
	response := httptest.NewRecorder()

	handler.Health(response, request)

	if response.Code != http.StatusOK || response.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected response %d %s", response.Code, response.Body)
	}
}
