package web_transport_http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	"go.uber.org/zap"
)

type fakeWebService struct {
	getMainPageCalled bool
	getMainPageErr    error
}

var _ WebService = (*fakeWebService)(nil)

func (s *fakeWebService) GetMainPage() ([]byte, error) {
	s.getMainPageCalled = true

	if s.getMainPageErr != nil {
		return nil, s.getMainPageErr
	}
	return []byte("<html></html>"), nil
}

func TestGetMainPageReturnsHTML(t *testing.T) {
	service := &fakeWebService{}
	handler := NewWebHTTPHandler(service)
	request := newWebRequest()
	response := httptest.NewRecorder()

	handler.GetMainPage(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.getMainPageCalled {
		t.Fatal("expected service GetMainPage to be called")
	}
	if response.Body.String() != "<html></html>" {
		t.Fatalf("expected html body, got %q", response.Body.String())
	}
}

func TestGetMainPageReturnsInternalServerErrorWhenServiceFails(t *testing.T) {
	service := &fakeWebService{
		getMainPageErr: errors.New("storage unavailable"),
	}
	handler := NewWebHTTPHandler(service)
	request := newWebRequest()
	response := httptest.NewRecorder()

	handler.GetMainPage(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
	if !service.getMainPageCalled {
		t.Fatal("expected service GetMainPage to be called")
	}
}

func newWebRequest() *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	logger := &core_logger.Logger{
		Logger: zap.NewNop(),
	}
	return request.WithContext(core_logger.ToContext(request.Context(), logger))
}
