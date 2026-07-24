package identity_postgres_repository

import core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"

type Pool interface {
	core_postgres_pool.Pool
	core_postgres_pool.Transactor
}

type IdentityRepository struct {
	pool Pool
}

func NewIdentityRepository(pool Pool) *IdentityRepository {
	return &IdentityRepository{pool: pool}
}
