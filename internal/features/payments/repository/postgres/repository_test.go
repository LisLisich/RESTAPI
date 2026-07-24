package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"
)

type fakePool struct {
	row core_postgres_pool.Row
	tx  *fakeTx
}

func (*fakePool) Query(context.Context, string, ...any) (core_postgres_pool.Rows, error) {
	panic("unexpected Query")
}

func (pool *fakePool) QueryRow(context.Context, string, ...any) core_postgres_pool.Row {
	return pool.row
}

func (*fakePool) Exec(context.Context, string, ...any) (core_postgres_pool.CommandTag, error) {
	panic("unexpected Exec")
}

func (pool *fakePool) Begin(context.Context) (core_postgres_pool.Tx, error) {
	return pool.tx, nil
}

func (*fakePool) Close() {}

func (*fakePool) OpTimeout() time.Duration { return time.Second }

type paymentRow struct {
	id              string
	userID          int
	idempotencyKey  string
	providerID      string
	status          domain.PaymentStatus
	amountMinor     int64
	currency        domain.Currency
	confirmationURL string
	createdAt       time.Time
}

func (row paymentRow) Scan(dest ...any) error {
	*dest[0].(*string) = row.id
	*dest[1].(*int) = row.userID
	*dest[2].(*string) = row.idempotencyKey
	*dest[3].(*string) = row.providerID
	*dest[4].(*domain.PaymentStatus) = row.status
	*dest[5].(*int64) = row.amountMinor
	*dest[6].(*domain.Currency) = row.currency
	*dest[7].(*string) = row.confirmationURL
	*dest[8].(*time.Time) = row.createdAt
	return nil
}

type creditRow struct {
	id          string
	userID      int
	amountMinor int64
	currency    domain.Currency
	status      domain.PaymentStatus
}

func (row creditRow) Scan(dest ...any) error {
	*dest[0].(*string) = row.id
	*dest[1].(*int) = row.userID
	*dest[2].(*int64) = row.amountMinor
	*dest[3].(*domain.Currency) = row.currency
	*dest[4].(*domain.PaymentStatus) = row.status
	return nil
}

type fakeTx struct {
	row        core_postgres_pool.Row
	querySQL   string
	execSQL    []string
	committed  bool
	rolledBack bool
}

func (*fakeTx) Query(context.Context, string, ...any) (core_postgres_pool.Rows, error) {
	panic("unexpected Query")
}

func (tx *fakeTx) QueryRow(_ context.Context, sql string, _ ...any) core_postgres_pool.Row {
	tx.querySQL = sql
	return tx.row
}

func (tx *fakeTx) Exec(
	_ context.Context,
	sql string,
	_ ...any,
) (core_postgres_pool.CommandTag, error) {
	tx.execSQL = append(tx.execSQL, sql)
	return fakeCommandTag{}, nil
}

func (tx *fakeTx) Commit(context.Context) error {
	tx.committed = true
	return nil
}

func (tx *fakeTx) Rollback(context.Context) error {
	tx.rolledBack = true
	return nil
}

type fakeCommandTag struct{}

func (fakeCommandTag) RowsAffected() int64 { return 1 }

func TestReservePaymentUsesAccountScopedIdempotency(t *testing.T) {
	now := time.Now().UTC()
	pool := &fakePool{row: paymentRow{
		id:             "0eea6811-6d64-43e5-b6ec-eca3d206dddf",
		userID:         42,
		idempotencyKey: "client-key",
		status:         domain.PaymentStatusPending,
		amountMinor:    12550,
		currency:       domain.CurrencyRUB,
		createdAt:      now,
	}}
	repository := NewPaymentRepository(pool)
	payment, _ := domain.NewPayment(
		"0eea6811-6d64-43e5-b6ec-eca3d206dddf",
		42,
		"client-key",
		domain.Money{MinorUnits: 12550, Currency: domain.CurrencyRUB},
		now,
	)

	reserved, err := repository.ReservePayment(t.Context(), payment)

	if err != nil {
		t.Fatalf("reserve payment: %v", err)
	}
	if reserved.ID != payment.ID || reserved.Amount != payment.Amount {
		t.Fatalf("unexpected payment: %#v", reserved)
	}
}

func TestCreditSucceededPaymentWritesAtomicLedgerAndOutbox(t *testing.T) {
	tx := &fakeTx{row: creditRow{
		id:          "0eea6811-6d64-43e5-b6ec-eca3d206dddf",
		userID:      42,
		amountMinor: 12550,
		currency:    domain.CurrencyRUB,
		status:      domain.PaymentStatusPending,
	}}
	repository := NewPaymentRepository(&fakePool{tx: tx})

	credited, err := repository.CreditSucceededPayment(
		t.Context(),
		"provider-id",
		domain.Money{MinorUnits: 12550, Currency: domain.CurrencyRUB},
		time.Now(),
	)

	if err != nil {
		t.Fatalf("credit payment: %v", err)
	}
	if !credited || !tx.committed {
		t.Fatal("expected committed credit")
	}
	if !strings.Contains(tx.querySQL, "FOR UPDATE") {
		t.Fatalf("expected row lock, got %q", tx.querySQL)
	}
	expectedTables := []string{
		"todoapp.payments",
		"todoapp.ledger_transactions",
		"todoapp.ledger_postings",
		"todoapp.wallets",
		"todoapp.outbox_events",
	}
	if len(tx.execSQL) != len(expectedTables) {
		t.Fatalf("expected %d writes, got %d", len(expectedTables), len(tx.execSQL))
	}
	for index, table := range expectedTables {
		if !strings.Contains(tx.execSQL[index], table) {
			t.Fatalf("expected write to %s, got %q", table, tx.execSQL[index])
		}
	}
}

func TestCreditSucceededPaymentIgnoresAlreadyCreditedPayment(t *testing.T) {
	tx := &fakeTx{row: creditRow{
		id:          "0eea6811-6d64-43e5-b6ec-eca3d206dddf",
		userID:      42,
		amountMinor: 12550,
		currency:    domain.CurrencyRUB,
		status:      domain.PaymentStatusSucceeded,
	}}
	repository := NewPaymentRepository(&fakePool{tx: tx})

	credited, err := repository.CreditSucceededPayment(
		t.Context(),
		"provider-id",
		domain.Money{MinorUnits: 12550, Currency: domain.CurrencyRUB},
		time.Now(),
	)

	if err != nil {
		t.Fatalf("credit payment: %v", err)
	}
	if credited || len(tx.execSQL) != 0 {
		t.Fatal("duplicate webhook must not write again")
	}
}
