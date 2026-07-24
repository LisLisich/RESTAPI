package identity_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"
	identity_service "github.com/LisLisich/fintask/internal/features/identity/service"
)

func (r *IdentityRepository) CreateRefreshToken(
	ctx context.Context,
	token identity_service.StoredRefreshToken,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		INSERT INTO todoapp.refresh_tokens (
			account_user_id,
			family_id,
			token_hash,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5);
	`
	if _, err := r.pool.Exec(
		ctx,
		query,
		token.AccountUserID,
		token.FamilyID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

func (r *IdentityRepository) RotateRefreshToken(
	ctx context.Context,
	oldTokenHash []byte,
	newToken identity_service.StoredRefreshToken,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		WITH consumed_token AS (
			UPDATE todoapp.refresh_tokens
			SET used_at = $4
			WHERE token_hash = $1
				AND used_at IS NULL
				AND revoked_at IS NULL
				AND expires_at > $4
			RETURNING account_user_id, family_id
		),
		inserted_token AS (
			INSERT INTO todoapp.refresh_tokens (
				account_user_id,
				family_id,
				token_hash,
				expires_at,
				created_at
			)
			SELECT account_user_id, family_id, $2, $3, $4
			FROM consumed_token
			RETURNING account_user_id
		)
		SELECT account_user_id
		FROM inserted_token;
	`
	var userID int
	err := r.pool.QueryRow(
		ctx,
		query,
		oldTokenHash,
		newToken.TokenHash,
		newToken.ExpiresAt,
		newToken.CreatedAt,
	).Scan(&userID)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return 0, fmt.Errorf(
			"refresh token is invalid, expired, or already used: %w",
			core_errors.ErrUnauthorized,
		)
	}
	if err != nil {
		return 0, fmt.Errorf("rotate refresh token: %w", err)
	}
	return userID, nil
}
