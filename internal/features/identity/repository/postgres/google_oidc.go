package identity_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
)

func (repository *IdentityRepository) StoreGoogleLoginAttempt(
	ctx context.Context,
	attempt identity_service.GoogleLoginAttempt,
) error {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()
	const query = `
		INSERT INTO todoapp.google_login_attempts (
			state_hash, nonce, code_verifier, expires_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5);
	`
	if _, err := repository.pool.Exec(
		ctx, query, attempt.StateHash, attempt.Nonce, attempt.CodeVerifier,
		attempt.ExpiresAt, attempt.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert Google login attempt: %w", err)
	}
	return nil
}

func (repository *IdentityRepository) ConsumeGoogleLoginAttempt(
	ctx context.Context,
	stateHash []byte,
	consumedAt time.Time,
) (identity_service.GoogleLoginAttempt, error) {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()
	const query = `
		UPDATE todoapp.google_login_attempts
		SET consumed_at = $2
		WHERE state_hash = $1
			AND consumed_at IS NULL
			AND expires_at > $2
		RETURNING state_hash, nonce, code_verifier, expires_at, created_at;
	`
	var attempt identity_service.GoogleLoginAttempt
	if err := repository.pool.QueryRow(ctx, query, stateHash, consumedAt).Scan(
		&attempt.StateHash, &attempt.Nonce, &attempt.CodeVerifier,
		&attempt.ExpiresAt, &attempt.CreatedAt,
	); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return attempt, fmt.Errorf("Google login attempt is invalid: %w", core_errors.ErrForbidden)
		}
		return attempt, fmt.Errorf("consume Google login attempt: %w", err)
	}
	return attempt, nil
}

func (repository *IdentityRepository) ResolveGoogleIdentity(
	ctx context.Context,
	identity identity_service.GoogleIdentity,
	createdAt time.Time,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin Google identity resolution: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	const findIdentity = `
		SELECT account_user_id
		FROM todoapp.oauth_identities
		WHERE provider = 'google' AND subject = $1;
	`
	var userID int
	err = tx.QueryRow(ctx, findIdentity, identity.Subject).Scan(&userID)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return 0, fmt.Errorf("commit existing Google identity: %w", err)
		}
		committed = true
		return userID, nil
	}
	if !errors.Is(err, core_postgres_pool.ErrNoRows) {
		return 0, fmt.Errorf("find Google identity: %w", err)
	}

	const findEmail = `SELECT user_id FROM todoapp.accounts WHERE email = $1;`
	err = tx.QueryRow(ctx, findEmail, strings.ToLower(identity.Email)).Scan(&userID)
	if err == nil {
		return 0, fmt.Errorf(
			"email already belongs to another account; explicit linking is required: %w",
			core_errors.ErrConflict,
		)
	}
	if !errors.Is(err, core_postgres_pool.ErrNoRows) {
		return 0, fmt.Errorf("check existing account email: %w", err)
	}

	name := normalizedGoogleName(identity.Name)
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO todoapp.users (full_name) VALUES ($1) RETURNING id;`,
		name,
	).Scan(&userID); err != nil {
		return 0, fmt.Errorf("insert Google user: %w", err)
	}
	email := strings.ToLower(strings.TrimSpace(identity.Email))
	if _, err := tx.Exec(ctx, `
		INSERT INTO todoapp.accounts (
			user_id, email, status, email_verified_at, created_at, updated_at
		)
		VALUES ($1, $2, 'active', $3, $3, $3);
	`, userID, email, createdAt); err != nil {
		return 0, fmt.Errorf("insert Google account: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO todoapp.wallets (
			account_user_id, balance_minor, currency, created_at, updated_at
		)
		VALUES ($1, 0, $2, $3, $3);
	`, userID, domain.CurrencyRUB, createdAt); err != nil {
		return 0, fmt.Errorf("insert Google wallet: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO todoapp.oauth_identities (
			account_user_id, provider, subject, email_snapshot, created_at
		)
		VALUES ($1, 'google', $2, $3, $4);
	`, userID, identity.Subject, email, createdAt); err != nil {
		return 0, fmt.Errorf("insert Google identity: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit Google identity: %w", err)
	}
	committed = true
	return userID, nil
}

func normalizedGoogleName(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) < 3 {
		return "Google User"
	}
	if len(runes) > 100 {
		runes = runes[:100]
	}
	return string(runes)
}
