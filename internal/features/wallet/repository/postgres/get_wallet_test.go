package wallet_postgres_repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"
)

type fakePool struct {
	sql  string
	args []any
	row  core_postgres_pool.Row
}

func (*fakePool) Query(context.Context, string, ...any) (core_postgres_pool.Rows, error) {
	panic("unexpected Query")
}

func (pool *fakePool) QueryRow(_ context.Context, sql string, args ...any) core_postgres_pool.Row {
	pool.sql = sql
	pool.args = args
	return pool.row
}

func (*fakePool) Exec(context.Context, string, ...any) (core_postgres_pool.CommandTag, error) {
	panic("unexpected Exec")
}

func (*fakePool) Close() {}

func (*fakePool) OpTimeout() time.Duration {
	return time.Second
}

type fakeWalletRow struct {
	userID       int
	version      int
	balanceMinor int64
	currency     domain.Currency
	err          error
}

func (row fakeWalletRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	*dest[0].(*int) = row.userID
	*dest[1].(*int) = row.version
	*dest[2].(*int64) = row.balanceMinor
	*dest[3].(*domain.Currency) = row.currency
	return nil
}

func TestGetWalletLoadsWalletByAccountOwner(t *testing.T) {
	pool := &fakePool{
		row: fakeWalletRow{
			userID:       42,
			version:      3,
			balanceMinor: 12500,
			currency:     domain.CurrencyRUB,
		},
	}
	repository := NewWalletRepository(pool)

	wallet, err := repository.GetWallet(context.Background(), 42)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if wallet.AccountUserID != 42 || wallet.Balance.MinorUnits != 12500 {
		t.Fatalf("unexpected wallet %+v", wallet)
	}
	if !strings.Contains(pool.sql, "WHERE account_user_id=$1") {
		t.Fatalf("expected owner filter, got %q", pool.sql)
	}
}

func TestGetWalletMapsMissingWalletToNotFound(t *testing.T) {
	pool := &fakePool{row: fakeWalletRow{err: core_postgres_pool.ErrNoRows}}
	repository := NewWalletRepository(pool)

	_, err := repository.GetWallet(context.Background(), 42)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
