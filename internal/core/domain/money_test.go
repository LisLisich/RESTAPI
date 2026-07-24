package domain

import (
	"errors"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestNewMoneyStoresIntegerKopecks(t *testing.T) {
	money, err := NewMoney(12345, CurrencyRUB)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if money.MinorUnits != 12345 {
		t.Fatalf("expected 12345 kopecks, got %d", money.MinorUnits)
	}
	if money.Currency != CurrencyRUB {
		t.Fatalf("expected RUB, got %q", money.Currency)
	}
}

func TestNewMoneyRejectsUnsupportedCurrency(t *testing.T) {
	_, err := NewMoney(100, Currency("USD"))

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestNewPaymentAmountMustBePositive(t *testing.T) {
	for _, amount := range []int64{-1, 0} {
		_, err := NewPaymentAmount(amount)
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Fatalf("expected amount %d to be rejected, got %v", amount, err)
		}
	}

	amount, err := NewPaymentAmount(1)
	if err != nil {
		t.Fatalf("expected one kopeck to be valid, got %v", err)
	}
	if amount.MinorUnits != 1 || amount.Currency != CurrencyRUB {
		t.Fatalf("unexpected payment amount %+v", amount)
	}
}
