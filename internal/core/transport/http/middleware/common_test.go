package core_http_middleware

import (
	"net/http/httptest"
	"testing"
)

func TestRequestPathForLogDoesNotContainQuery(t *testing.T) {
	request := httptest.NewRequest(
		"GET",
		"https://example.test/api/v2/auth/google/callback?code=secret&state=csrf",
		nil,
	)

	got := requestPathForLog(request)

	if got != "/api/v2/auth/google/callback" {
		t.Fatalf("expected path without query, got %q", got)
	}
}

func TestRequestPathForLogPreservesEscapedPath(t *testing.T) {
	request := httptest.NewRequest("GET", "https://example.test/tasks/a%2Fb", nil)

	got := requestPathForLog(request)

	if got != "/tasks/a%2Fb" {
		t.Fatalf("expected escaped path, got %q", got)
	}
}
