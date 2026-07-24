package domain

import (
	"errors"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestNewLedgerTransactionAcceptsBalancedPostings(t *testing.T) {
	transaction, err := NewLedgerTransaction(
		"payment-42",
		[]LedgerPosting{
			{Account: "cash:provider", Money: Money{MinorUnits: -15000, Currency: CurrencyRUB}},
			{Account: "wallet:42", Money: Money{MinorUnits: 15000, Currency: CurrencyRUB}},
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transaction.ID != "payment-42" {
		t.Fatalf("expected transaction id, got %q", transaction.ID)
	}
	if len(transaction.Postings) != 2 {
		t.Fatalf("expected two postings, got %d", len(transaction.Postings))
	}
}

func TestNewLedgerTransactionRejectsUnbalancedPostings(t *testing.T) {
	_, err := NewLedgerTransaction(
		"payment-42",
		[]LedgerPosting{
			{Account: "cash:provider", Money: Money{MinorUnits: -15000, Currency: CurrencyRUB}},
			{Account: "wallet:42", Money: Money{MinorUnits: 14999, Currency: CurrencyRUB}},
		},
	)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestNewLedgerTransactionRejectsSingleZeroOrForeignCurrencyPosting(t *testing.T) {
	tests := []struct {
		name     string
		postings []LedgerPosting
	}{
		{
			name: "single posting",
			postings: []LedgerPosting{
				{Account: "wallet:42", Money: Money{MinorUnits: 100, Currency: CurrencyRUB}},
			},
		},
		{
			name: "zero posting",
			postings: []LedgerPosting{
				{Account: "cash:provider", Money: Money{MinorUnits: 0, Currency: CurrencyRUB}},
				{Account: "wallet:42", Money: Money{MinorUnits: 0, Currency: CurrencyRUB}},
			},
		},
		{
			name: "unsupported currency",
			postings: []LedgerPosting{
				{Account: "cash:provider", Money: Money{MinorUnits: -100, Currency: Currency("USD")}},
				{Account: "wallet:42", Money: Money{MinorUnits: 100, Currency: Currency("USD")}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLedgerTransaction("transaction", tt.postings)
			if !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
		})
	}
}

func TestNewLedgerTransactionCopiesPostings(t *testing.T) {
	postings := []LedgerPosting{
		{Account: "cash:provider", Money: Money{MinorUnits: -100, Currency: CurrencyRUB}},
		{Account: "wallet:42", Money: Money{MinorUnits: 100, Currency: CurrencyRUB}},
	}
	transaction, err := NewLedgerTransaction("transaction", postings)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	postings[0].Money.MinorUnits = -200

	if transaction.Postings[0].Money.MinorUnits != -100 {
		t.Fatal("ledger transaction must own an immutable copy of postings")
	}
}
