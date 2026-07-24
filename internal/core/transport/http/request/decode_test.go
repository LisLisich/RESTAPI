package core_http_request

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

type taggedRequest struct {
	Name string `json:"name" validate:"required"`
}

type customValidatableRequest struct {
	Name string `json:"name"`
}

func (r *customValidatableRequest) Validate() error {
	if r.Name == "bad" {
		return fmt.Errorf("bad name")
	}
	return nil
}

func TestDecodeAndValidateRequestDecodesValidBody(t *testing.T) {
	request := httptest.NewRequest("POST", "/users", strings.NewReader(`{"name":"Ivan"}`))
	var body taggedRequest

	err := DecodeAndValidateRequest(request, &body)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if body.Name != "Ivan" {
		t.Fatalf("expected name %q, got %q", "Ivan", body.Name)
	}
}

func TestDecodeAndValidateRequestReturnsInvalidArgumentForInvalidJSON(t *testing.T) {
	request := httptest.NewRequest("POST", "/users", strings.NewReader(`{`))
	var body taggedRequest

	err := DecodeAndValidateRequest(request, &body)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestDecodeAndValidateRequestUsesStructTags(t *testing.T) {
	request := httptest.NewRequest("POST", "/users", strings.NewReader(`{}`))
	var body taggedRequest

	err := DecodeAndValidateRequest(request, &body)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestDecodeAndValidateRequestUsesCustomValidate(t *testing.T) {
	request := httptest.NewRequest("POST", "/users", strings.NewReader(`{"name":"bad"}`))
	var body customValidatableRequest

	err := DecodeAndValidateRequest(request, &body)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
