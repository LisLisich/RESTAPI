package identity_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"
)

func (r *IdentityRepository) VerifyEmail(
	ctx context.Context,
	tokenHash []byte,
	verifiedAt time.Time,
) (domain.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		WITH consumed_token AS (
			UPDATE todoapp.account_tokens
			SET consumed_at = $2
			WHERE token_hash = $1
				AND purpose = 'email_verification'
				AND consumed_at IS NULL
				AND expires_at > $2
			RETURNING account_user_id
		)
		UPDATE todoapp.accounts AS account
		SET
			status = 'active',
			email_verified_at = $2,
			updated_at = $2,
			version = version + 1
		FROM consumed_token
		WHERE account.user_id = consumed_token.account_user_id
		RETURNING
			account.user_id,
			account.email,
			account.status,
			account.email_verified_at;
	`

	var account domain.Account
	err := r.pool.QueryRow(ctx, query, tokenHash, verifiedAt).Scan(
		&account.UserID,
		&account.Email,
		&account.Status,
		&account.EmailVerifiedAt,
	)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return domain.Account{}, fmt.Errorf(
			"verification token is invalid, expired, or already consumed: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("scan verified account: %w", err)
	}

	return account, nil
}
