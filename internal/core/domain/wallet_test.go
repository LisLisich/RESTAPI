package domain

import (
	"errors"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestNewWalletAcceptsNonNegativeRUBBalance(t *testing.T) {
	wallet, err := NewWallet(42, 3, Money{MinorUnits: 12500, Currency: CurrencyRUB})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if wallet.AccountUserID != 42 || wallet.Version != 3 {
		t.Fatalf("unexpected wallet identity %+v", wallet)
	}
	if wallet.Balance.MinorUnits != 12500 {
		t.Fatalf("expected balance 12500, got %d", wallet.Balance.MinorUnits)
	}
}

func TestNewWalletRejectsNegativeOrForeignCurrencyBalance(t *testing.T) {
	tests := []Money{
		{MinorUnits: -1, Currency: CurrencyRUB},
		{MinorUnits: 100, Currency: Currency("USD")},
	}

	for _, balance := range tests {
		_, err := NewWallet(42, 1, balance)
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Fatalf("expected ErrInvalidArgument for %+v, got %v", balance, err)
		}
	}
}
