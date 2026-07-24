package wallet_service

import (
	"context"
	"errors"
	"testing"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

type fakeWalletRepository struct {
	called bool
	userID int
	wallet domain.Wallet
	err    error
}

func (repository *fakeWalletRepository) GetWallet(
	_ context.Context,
	userID int,
) (domain.Wallet, error) {
	repository.called = true
	repository.userID = userID
	if repository.err != nil {
		return domain.Wallet{}, repository.err
	}
	return repository.wallet, nil
}

func TestGetWalletReturnsOwnedWallet(t *testing.T) {
	wallet, err := domain.NewWallet(
		42,
		3,
		domain.Money{MinorUnits: 12500, Currency: domain.CurrencyRUB},
	)
	if err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	repository := &fakeWalletRepository{wallet: wallet}
	service := NewWalletService(repository)

	got, err := service.GetWallet(context.Background(), 42)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.called || repository.userID != 42 {
		t.Fatal("expected repository lookup for user 42")
	}
	if got.Balance.MinorUnits != 12500 {
		t.Fatalf("unexpected balance %d", got.Balance.MinorUnits)
	}
}

func TestGetWalletPreservesNotFound(t *testing.T) {
	service := NewWalletService(
		&fakeWalletRepository{err: core_errors.ErrNotFound},
	)

	_, err := service.GetWallet(context.Background(), 42)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
