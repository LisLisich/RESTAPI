package core_http_request

import (
	"errors"
	"net/http/httptest"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestGetIntPathValueReturnsInteger(t *testing.T) {
	request := httptest.NewRequest("GET", "/users/15", nil)
	request.SetPathValue("id", "15")

	value, err := GetIntPathValue(request, "id")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != 15 {
		t.Fatalf("expected value %d, got %d", 15, value)
	}
}

func TestGetIntPathValueReturnsInvalidArgumentForMissingValue(t *testing.T) {
	request := httptest.NewRequest("GET", "/users/15", nil)

	_, err := GetIntPathValue(request, "id")

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestGetIntPathValueReturnsInvalidArgumentForNonInteger(t *testing.T) {
	request := httptest.NewRequest("GET", "/users/not-integer", nil)
	request.SetPathValue("id", "not-integer")

	_, err := GetIntPathValue(request, "id")

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
