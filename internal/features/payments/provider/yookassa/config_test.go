package yookassa

import (
	"strings"
	"testing"
)

func TestNewConfigAllowsDisabledIntegrationWithoutSecrets(t *testing.T) {
	t.Setenv("YOOKASSA_ENABLED", "false")
	t.Setenv("YOOKASSA_SHOP_ID", "")
	t.Setenv("YOOKASSA_SECRET_KEY", "")
	t.Setenv("YOOKASSA_RETURN_URL", "")

	config, err := NewConfig()

	if err != nil {
		t.Fatalf("new config: %v", err)
	}
	if config.Enabled {
		t.Fatal("expected disabled integration")
	}
}

func TestNewConfigRequiresCredentialsWhenEnabled(t *testing.T) {
	t.Setenv("YOOKASSA_ENABLED", "true")
	t.Setenv("YOOKASSA_SHOP_ID", "")
	t.Setenv("YOOKASSA_SECRET_KEY", "")
	t.Setenv("YOOKASSA_RETURN_URL", "")

	_, err := NewConfig()

	if err == nil {
		t.Fatal("expected missing credentials error")
	}
}

func TestNewConfigRequiresHTTPSReturnURL(t *testing.T) {
	t.Setenv("YOOKASSA_ENABLED", "true")
	t.Setenv("YOOKASSA_SHOP_ID", "shop")
	t.Setenv("YOOKASSA_SECRET_KEY", "secret")
	t.Setenv("YOOKASSA_RETURN_URL", "http://app.example/payments/return")

	_, err := NewConfig()

	if err == nil || !strings.Contains(err.Error(), "HTTPS") {
		t.Fatalf("expected HTTPS error, got %v", err)
	}
}
