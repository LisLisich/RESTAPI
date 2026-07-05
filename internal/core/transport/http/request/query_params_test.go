package core_http_request

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestGetIntQueryParamReturnsNilWhenAbsent(t *testing.T) {
	request := httptest.NewRequest("GET", "/tasks", nil)

	value, err := GetIntQueryParam(request, "limit")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != nil {
		t.Fatalf("expected nil value, got %d", *value)
	}
}

func TestGetIntQueryParamReturnsInteger(t *testing.T) {
	request := httptest.NewRequest("GET", "/tasks?limit=20", nil)

	value, err := GetIntQueryParam(request, "limit")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value == nil || *value != 20 {
		t.Fatalf("expected value %d, got %v", 20, value)
	}
}

func TestGetIntQueryParamReturnsInvalidArgumentForNonInteger(t *testing.T) {
	request := httptest.NewRequest("GET", "/tasks?limit=bad", nil)

	_, err := GetIntQueryParam(request, "limit")

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestGetDateQueryParamReturnsDate(t *testing.T) {
	request := httptest.NewRequest("GET", "/statistics?from=2026-07-01", nil)

	value, err := GetDateQueryParam(request, "from")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if value == nil || !value.Equal(want) {
		t.Fatalf("expected date %v, got %v", want, value)
	}
}

func TestGetDateQueryParamReturnsInvalidArgumentForInvalidDate(t *testing.T) {
	request := httptest.NewRequest("GET", "/statistics?from=2026-99-99", nil)

	_, err := GetDateQueryParam(request, "from")

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
