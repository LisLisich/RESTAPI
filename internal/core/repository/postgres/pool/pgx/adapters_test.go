package core_pgx_pool

import (
	"errors"
	"testing"

	core_postgres_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapErrorsMapsUniqueViolation(t *testing.T) {
	err := mapErrors(&pgconn.PgError{Code: "23505", Message: "duplicate key"})

	if !errors.Is(err, core_postgres_pool.ErrViolatesUnique) {
		t.Fatalf("expected ErrViolatesUnique, got %v", err)
	}
}

func TestMapErrorsMapsForeignKeyViolation(t *testing.T) {
	err := mapErrors(&pgconn.PgError{Code: "23503", Message: "foreign key violation"})

	if !errors.Is(err, core_postgres_pool.ErrViolatesForeignKey) {
		t.Fatalf("expected ErrViolatesForeignKey, got %v", err)
	}
}
