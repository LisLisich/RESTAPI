package postgres

import core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"

type NotificationRepository struct {
	pool core_postgres_pool.Pool
}

func NewNotificationRepository(pool core_postgres_pool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}
