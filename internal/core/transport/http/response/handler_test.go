package core_http_response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	"go.uber.org/zap"
)

func TestErrorResponseMapsSentinelErrorsToStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "invalid argument",
			err:        core_errors.ErrInvalidArgument,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found",
			err:        core_errors.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "conflict",
			err:        core_errors.ErrConflict,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "unauthorized",
			err:        core_errors.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "forbidden",
			err:        core_errors.ErrForbidden,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "unknown",
			err:        errors.New("unknown"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler := NewHTTPResponseHandler(newNopLogger(), response)

			handler.ErrorResponse(tt.err, "message")

			if response.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, response.Code)
			}
		})
	}
}

func TestJSONResponseWritesStatusAndBody(t *testing.T) {
	response := httptest.NewRecorder()
	handler := NewHTTPResponseHandler(newNopLogger(), response)

	handler.JSONResponse(map[string]string{"status": "ok"}, http.StatusCreated)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}
	if response.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("expected json body, got %q", response.Body.String())
	}
}

func TestErrorResponseDoesNotExposeInternalError(t *testing.T) {
	response := httptest.NewRecorder()
	handler := NewHTTPResponseHandler(newNopLogger(), response)

	handler.ErrorResponse(errors.New("password=secret db=production"), "request failed")

	var body ErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("expected public status text, got %q", body.Error)
	}
	if body.Message != "request failed" {
		t.Fatalf("expected public message, got %q", body.Message)
	}
	if body.Error == "password=secret db=production" {
		t.Fatal("internal error was exposed")
	}
}

func TestResponseWriterTracksStatusCode(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewResponseWriter(recorder)

	if writer.GetStatusCode() != http.StatusOK {
		t.Fatalf("expected default status %d, got %d", http.StatusOK, writer.GetStatusCode())
	}

	writer.WriteHeader(http.StatusCreated)

	if writer.GetStatusCode() != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, writer.GetStatusCode())
	}
}

func newNopLogger() *core_logger.Logger {
	return &core_logger.Logger{
		Logger: zap.NewNop(),
	}
}
