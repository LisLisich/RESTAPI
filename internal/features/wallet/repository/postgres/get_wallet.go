package wallet_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"
)

func (repository *WalletRepository) GetWallet(
	ctx context.Context,
	userID int,
) (domain.Wallet, error) {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()

	const query = `
		SELECT account_user_id, version, balance_minor, currency
		FROM todoapp.wallets
		WHERE account_user_id=$1;
	`
	var (
		accountUserID int
		version       int
		balanceMinor  int64
		currency      domain.Currency
	)
	err := repository.pool.QueryRow(ctx, query, userID).Scan(
		&accountUserID,
		&version,
		&balanceMinor,
		&currency,
	)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return domain.Wallet{}, fmt.Errorf(
			"wallet for user id='%d': %w",
			userID,
			core_errors.ErrNotFound,
		)
	}
	if err != nil {
		return domain.Wallet{}, fmt.Errorf("scan wallet: %w", err)
	}
	wallet, err := domain.NewWallet(
		accountUserID,
		version,
		domain.Money{MinorUnits: balanceMinor, Currency: currency},
	)
	if err != nil {
		return domain.Wallet{}, fmt.Errorf("restore wallet domain: %w", err)
	}
	return wallet, nil
}
