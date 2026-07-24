package postgres

import core_postgres_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool"

type Pool interface {
	core_postgres_pool.Pool
	core_postgres_pool.Transactor
}

type PaymentRepository struct {
	pool Pool
}

func NewPaymentRepository(pool Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}
