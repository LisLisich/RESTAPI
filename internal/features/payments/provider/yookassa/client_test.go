package yookassa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LisLisich/fintask/internal/core/domain"
)

func TestClientCreatesRedirectPayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "shop" || password != "secret" {
			t.Fatal("expected Basic Auth")
		}
		if got := r.Header.Get("Idempotence-Key"); got != "local-id" {
			t.Fatalf("unexpected idempotence key %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body["amount"].(map[string]any)["value"] != "125.50" {
			t.Fatalf("unexpected amount: %#v", body["amount"])
		}
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{
			"id":"provider-id",
			"status":"pending",
			"paid":false,
			"test":true,
			"amount":{"value":"125.50","currency":"RUB"},
			"confirmation":{"type":"redirect","confirmation_url":"https://pay.example/confirm"}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "shop", "secret", server.Client())
	payment, err := client.CreatePayment(
		t.Context(),
		CreatePaymentInput{
			IdempotenceKey: "local-id",
			Amount:         domain.Money{MinorUnits: 12550, Currency: domain.CurrencyRUB},
			ReturnURL:      "https://app.example/payments/return",
			Description:    "Пополнение FinTask",
		},
	)

	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if payment.ID != "provider-id" || payment.ConfirmationURL != "https://pay.example/confirm" {
		t.Fatalf("unexpected payment: %#v", payment)
	}
}

func TestClientGetsAndParsesSucceededPaymentExactly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/payments/provider-id" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{
			"id":"provider-id",
			"status":"succeeded",
			"paid":true,
			"test":true,
			"amount":{"value":"0.01","currency":"RUB"}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "shop", "secret", server.Client())
	payment, err := client.GetPayment(t.Context(), "provider-id")

	if err != nil {
		t.Fatalf("get payment: %v", err)
	}
	if payment.Amount.MinorUnits != 1 || !payment.Paid || !payment.Test {
		t.Fatalf("unexpected payment: %#v", payment)
	}
}
