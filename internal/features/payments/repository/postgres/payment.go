package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"
)

func (repository *PaymentRepository) ReservePayment(
	ctx context.Context,
	payment domain.Payment,
) (domain.Payment, error) {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()

	const query = `
		INSERT INTO todoapp.payments (
			id,
			account_user_id,
			idempotency_key,
			amount_minor,
			currency,
			created_at,
			updated_at
		)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (account_user_id, idempotency_key)
		DO UPDATE SET idempotency_key = EXCLUDED.idempotency_key
		RETURNING
			id::text,
			account_user_id,
			idempotency_key,
			COALESCE(provider_payment_id, ''),
			status,
			amount_minor,
			currency,
			COALESCE(confirmation_url, ''),
			created_at;
	`
	return scanPayment(repository.pool.QueryRow(
		ctx,
		query,
		payment.ID,
		payment.AccountUserID,
		payment.IdempotencyKey,
		payment.Amount.MinorUnits,
		payment.Amount.Currency,
		payment.CreatedAt,
	))
}

func (repository *PaymentRepository) AttachProviderPayment(
	ctx context.Context,
	localPaymentID string,
	providerPaymentID string,
	confirmationURL string,
	attachedAt time.Time,
) (domain.Payment, error) {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()

	const query = `
		UPDATE todoapp.payments
		SET
			provider_payment_id = $2,
			confirmation_url = $3,
			provider_attached_at = $4,
			updated_at = $4
		WHERE id = $1::uuid
			AND (provider_payment_id IS NULL OR provider_payment_id = $2)
		RETURNING
			id::text,
			account_user_id,
			idempotency_key,
			provider_payment_id,
			status,
			amount_minor,
			currency,
			confirmation_url,
			created_at;
	`
	payment, err := scanPayment(repository.pool.QueryRow(
		ctx,
		query,
		localPaymentID,
		providerPaymentID,
		confirmationURL,
		attachedAt,
	))
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return domain.Payment{}, fmt.Errorf(
			"payment not found or attached to another provider id: %w",
			core_errors.ErrConflict,
		)
	}
	return payment, err
}

func (repository *PaymentRepository) CreditSucceededPayment(
	ctx context.Context,
	providerPaymentID string,
	amount domain.Money,
	creditedAt time.Time,
) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()

	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin payment credit: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	const lockPayment = `
		SELECT id::text, account_user_id, amount_minor, currency, status
		FROM todoapp.payments
		WHERE provider_payment_id = $1
		FOR UPDATE;
	`
	var (
		localPaymentID string
		accountUserID  int
		amountMinor    int64
		currency       domain.Currency
		status         domain.PaymentStatus
	)
	if err := tx.QueryRow(ctx, lockPayment, providerPaymentID).Scan(
		&localPaymentID,
		&accountUserID,
		&amountMinor,
		&currency,
		&status,
	); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return false, fmt.Errorf("payment not found: %w", core_errors.ErrNotFound)
		}
		return false, fmt.Errorf("lock payment: %w", err)
	}
	if status == domain.PaymentStatusSucceeded {
		return false, nil
	}
	if amountMinor != amount.MinorUnits || currency != amount.Currency {
		return false, fmt.Errorf(
			"provider amount does not match reserved payment: %w",
			core_errors.ErrConflict,
		)
	}

	const updatePayment = `
		UPDATE todoapp.payments
		SET status = 'succeeded', credited_at = $2, updated_at = $2
		WHERE id = $1::uuid;
	`
	if _, err := tx.Exec(ctx, updatePayment, localPaymentID, creditedAt); err != nil {
		return false, fmt.Errorf("mark payment succeeded: %w", err)
	}

	const insertLedgerTransaction = `
		INSERT INTO todoapp.ledger_transactions (
			id, reference_type, reference_id, created_at
		)
		VALUES ($1::uuid, 'payment', $2, $3);
	`
	if _, err := tx.Exec(
		ctx,
		insertLedgerTransaction,
		localPaymentID,
		providerPaymentID,
		creditedAt,
	); err != nil {
		return false, fmt.Errorf("insert payment ledger transaction: %w", err)
	}

	const insertPostings = `
		INSERT INTO todoapp.ledger_postings (
			transaction_id, account_code, amount_minor, currency, created_at
		)
		VALUES
			($1::uuid, $2, $3, $4, $5),
			($1::uuid, 'yookassa:clearing', $6, $4, $5);
	`
	if _, err := tx.Exec(
		ctx,
		insertPostings,
		localPaymentID,
		"wallet:"+strconv.Itoa(accountUserID),
		amount.MinorUnits,
		amount.Currency,
		creditedAt,
		-amount.MinorUnits,
	); err != nil {
		return false, fmt.Errorf("insert payment ledger postings: %w", err)
	}

	const updateWallet = `
		UPDATE todoapp.wallets
		SET
			balance_minor = balance_minor + $2,
			version = version + 1,
			updated_at = $3
		WHERE account_user_id = $1;
	`
	tag, err := tx.Exec(ctx, updateWallet, accountUserID, amount.MinorUnits, creditedAt)
	if err != nil {
		return false, fmt.Errorf("credit wallet: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return false, fmt.Errorf("wallet not found: %w", core_errors.ErrNotFound)
	}

	payload, err := json.Marshal(map[string]any{
		"payment_id":   localPaymentID,
		"user_id":      accountUserID,
		"amount_minor": amount.MinorUnits,
		"currency":     amount.Currency,
	})
	if err != nil {
		return false, fmt.Errorf("marshal payment outbox payload: %w", err)
	}
	const insertOutbox = `
		INSERT INTO todoapp.outbox_events (
			topic, aggregate_type, aggregate_id, payload, available_at, created_at
		)
		VALUES (
			'payments.wallet_credited', 'payment', $1, $2, $3, $3
		);
	`
	if _, err := tx.Exec(ctx, insertOutbox, localPaymentID, payload, creditedAt); err != nil {
		return false, fmt.Errorf("insert payment outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit payment credit: %w", err)
	}
	committed = true
	return true, nil
}

func scanPayment(row core_postgres_pool.Row) (domain.Payment, error) {
	var payment domain.Payment
	var amountMinor int64
	var currency domain.Currency
	if err := row.Scan(
		&payment.ID,
		&payment.AccountUserID,
		&payment.IdempotencyKey,
		&payment.ProviderPaymentID,
		&payment.Status,
		&amountMinor,
		&currency,
		&payment.ConfirmationURL,
		&payment.CreatedAt,
	); err != nil {
		return domain.Payment{}, fmt.Errorf("scan payment: %w", err)
	}
	amount, err := domain.NewPaymentAmount(amountMinor)
	if err != nil {
		return domain.Payment{}, fmt.Errorf("invalid stored payment amount: %w", err)
	}
	if currency != domain.CurrencyRUB {
		return domain.Payment{}, fmt.Errorf(
			"invalid stored payment currency %q: %w",
			currency,
			core_errors.ErrInvalidArgument,
		)
	}
	payment.Amount = amount
	return payment, nil
}
