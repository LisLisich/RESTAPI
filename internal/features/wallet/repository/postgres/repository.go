package wallet_postgres_repository

import core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"

type WalletRepository struct {
	pool core_postgres_pool.Pool
}

func NewWalletRepository(pool core_postgres_pool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}
