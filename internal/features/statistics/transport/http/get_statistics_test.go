package statistics_transport_http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	"go.uber.org/zap"
)

type fakeStatisticsService struct {
	getStatisticsCalled bool
	gotUserID           *int
	gotFrom             *time.Time
	gotTo               *time.Time
	getStatisticsErr    error
}

var _ StatisticsService = (*fakeStatisticsService)(nil)

func (s *fakeStatisticsService) GetStatistics(
	ctx context.Context,
	userID *int,
	from *time.Time,
	to *time.Time,
) (domain.Statistics, error) {
	s.getStatisticsCalled = true
	s.gotUserID = userID
	s.gotFrom = from
	s.gotTo = to

	if s.getStatisticsErr != nil {
		return domain.Statistics{}, s.getStatisticsErr
	}
	rate := 50.0
	avg := 2 * time.Hour
	return domain.NewStatistics(4, 2, &rate, &avg), nil
}

func TestGetStatisticsReturnsOKOnValidQueryParams(t *testing.T) {
	service := &fakeStatisticsService{}
	handler := NewStatisticsHTTPHandler(service)

	request := newStatisticsRequest("/statistics?user_id=7&from=2026-07-01&to=2026-07-05")
	response := httptest.NewRecorder()

	handler.GetStatistics(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.getStatisticsCalled {
		t.Fatal("expected service GetStatistics to be called")
	}
	if service.gotUserID == nil || *service.gotUserID != 7 {
		t.Fatalf("expected user id %d, got %v", 7, service.gotUserID)
	}
	if service.gotFrom == nil || !service.gotFrom.Equal(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected from date 2026-07-01, got %v", service.gotFrom)
	}
	if service.gotTo == nil || !service.gotTo.Equal(time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected to date 2026-07-05, got %v", service.gotTo)
	}
}

func TestGetStatisticsReturnsBadRequestOnInvalidQueryParam(t *testing.T) {
	service := &fakeStatisticsService{}
	handler := NewStatisticsHTTPHandler(service)

	request := newStatisticsRequest("/statistics?from=invalid-date")
	response := httptest.NewRecorder()

	handler.GetStatistics(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.getStatisticsCalled {
		t.Fatal("expected service GetStatistics not to be called")
	}
}

func TestGetStatisticsReturnsInternalServerErrorWhenServiceFails(t *testing.T) {
	service := &fakeStatisticsService{
		getStatisticsErr: errors.New("storage unavailable"),
	}
	handler := NewStatisticsHTTPHandler(service)

	request := newStatisticsRequest("/statistics")
	response := httptest.NewRecorder()

	handler.GetStatistics(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
	if !service.getStatisticsCalled {
		t.Fatal("expected service GetStatistics to be called")
	}
}

func newStatisticsRequest(target string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
	logger := &core_logger.Logger{
		Logger: zap.NewNop(),
	}
	return request.WithContext(core_logger.ToContext(request.Context(), logger))
}
