package identity_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
)

func (r *IdentityRepository) GetPasswordAccount(
	ctx context.Context,
	email string,
) (identity_service.PasswordAccount, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		SELECT
			account.user_id,
			account.email,
			account.status,
			account.email_verified_at,
			credential.password_hash
		FROM todoapp.accounts AS account
		INNER JOIN todoapp.password_credentials AS credential
			ON credential.account_user_id = account.user_id
		WHERE account.email = $1;
	`

	var passwordAccount identity_service.PasswordAccount
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&passwordAccount.Account.UserID,
		&passwordAccount.Account.Email,
		&passwordAccount.Account.Status,
		&passwordAccount.Account.EmailVerifiedAt,
		&passwordAccount.PasswordHash,
	)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return identity_service.PasswordAccount{}, fmt.Errorf(
			"account with email not found: %w",
			core_errors.ErrNotFound,
		)
	}
	if err != nil {
		return identity_service.PasswordAccount{}, fmt.Errorf("scan password account: %w", err)
	}

	return passwordAccount, nil
}

func (r *IdentityRepository) CreateSession(
	ctx context.Context,
	session identity_service.StoredSession,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		INSERT INTO todoapp.sessions (
			account_user_id,
			token_hash,
			csrf_hash,
			expires_at,
			created_at,
			last_seen_at
		)
		VALUES ($1, $2, $3, $4, $5, $5);
	`
	if _, err := r.pool.Exec(
		ctx,
		query,
		session.AccountUserID,
		session.TokenHash,
		session.CSRFHash,
		session.ExpiresAt,
		session.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert browser session: %w", err)
	}

	return nil
}

func (r *IdentityRepository) GetSession(
	ctx context.Context,
	tokenHash []byte,
	now time.Time,
) (identity_service.StoredAuthenticationSession, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		SELECT
			session.account_user_id,
			session.csrf_hash
		FROM todoapp.sessions AS session
		INNER JOIN todoapp.accounts AS account
			ON account.user_id = session.account_user_id
		WHERE session.token_hash = $1
			AND session.revoked_at IS NULL
			AND session.expires_at > $2
			AND account.status = 'active';
	`

	var session identity_service.StoredAuthenticationSession
	err := r.pool.QueryRow(ctx, query, tokenHash, now).Scan(
		&session.AccountUserID,
		&session.CSRFHash,
	)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return identity_service.StoredAuthenticationSession{}, fmt.Errorf(
			"active session not found: %w",
			core_errors.ErrNotFound,
		)
	}
	if err != nil {
		return identity_service.StoredAuthenticationSession{}, fmt.Errorf(
			"scan browser session: %w",
			err,
		)
	}

	return session, nil
}

func (r *IdentityRepository) RevokeSession(
	ctx context.Context,
	tokenHash []byte,
	revokedAt time.Time,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		UPDATE todoapp.sessions
		SET revoked_at = $2
		WHERE token_hash = $1
			AND revoked_at IS NULL;
	`
	if _, err := r.pool.Exec(ctx, query, tokenHash, revokedAt); err != nil {
		return fmt.Errorf("revoke browser session: %w", err)
	}
	return nil
}
