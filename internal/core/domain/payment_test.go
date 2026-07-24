package domain

import (
	"errors"
	"testing"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestNewPaymentCreatesPendingRUBPayment(t *testing.T) {
	payment, err := NewPayment(
		"local-id",
		42,
		"idempotency-key",
		Money{MinorUnits: 12550, Currency: CurrencyRUB},
		time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("new payment: %v", err)
	}
	if payment.Status != PaymentStatusPending || payment.Amount.MinorUnits != 12550 {
		t.Fatalf("unexpected payment: %#v", payment)
	}
}

func TestNewPaymentRejectsInvalidIdentityAndAmount(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		userID         int
		idempotencyKey string
		amount         Money
	}{
		{name: "empty id", userID: 42, idempotencyKey: "key", amount: Money{1, CurrencyRUB}},
		{name: "invalid user", id: "id", idempotencyKey: "key", amount: Money{1, CurrencyRUB}},
		{name: "empty key", id: "id", userID: 42, amount: Money{1, CurrencyRUB}},
		{name: "zero amount", id: "id", userID: 42, idempotencyKey: "key", amount: Money{0, CurrencyRUB}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewPayment(test.id, test.userID, test.idempotencyKey, test.amount, time.Now())
			if !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected invalid argument, got %v", err)
			}
		})
	}
}
