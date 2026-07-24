package domain

import (
	"fmt"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

type Wallet struct {
	AccountUserID int
	Version       int
	Balance       Money
}

func NewWallet(
	accountUserID int,
	version int,
	balance Money,
) (Wallet, error) {
	if accountUserID <= 0 {
		return Wallet{}, fmt.Errorf("wallet account id must be positive: %w", core_errors.ErrInvalidArgument)
	}
	if version <= 0 {
		return Wallet{}, fmt.Errorf("wallet version must be positive: %w", core_errors.ErrInvalidArgument)
	}
	if balance.Currency != CurrencyRUB {
		return Wallet{}, fmt.Errorf(
			"wallet currency must be RUB: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	if balance.MinorUnits < 0 {
		return Wallet{}, fmt.Errorf(
			"wallet balance cannot be negative: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	return Wallet{
		AccountUserID: accountUserID,
		Version:       version,
		Balance:       balance,
	}, nil
}
