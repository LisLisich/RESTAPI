package identity_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
)

const passwordResetTopic = "identity.password_reset_requested"

type passwordResetPayload struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

func (r *IdentityRepository) RequestPasswordReset(
	ctx context.Context,
	request identity_service.PasswordResetRequest,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password reset request: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	const consumePreviousTokens = `
		UPDATE todoapp.account_tokens
		SET consumed_at = $2
		WHERE account_user_id = $1
			AND purpose = 'password_reset'
			AND consumed_at IS NULL;
	`
	if _, err := tx.Exec(
		ctx,
		consumePreviousTokens,
		request.AccountUserID,
		request.CreatedAt,
	); err != nil {
		return fmt.Errorf("consume previous password reset tokens: %w", err)
	}

	const insertToken = `
		INSERT INTO todoapp.account_tokens (
			account_user_id,
			purpose,
			token_hash,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5);
	`
	if _, err := tx.Exec(
		ctx,
		insertToken,
		request.AccountUserID,
		"password_reset",
		request.Token.Hash,
		request.ExpiresAt,
		request.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert password reset token: %w", err)
	}

	payload, err := json.Marshal(passwordResetPayload{
		UserID: request.AccountUserID,
		Email:  request.Email,
		Token:  request.Token.Raw,
	})
	if err != nil {
		return fmt.Errorf("marshal password reset payload: %w", err)
	}

	const insertOutbox = `
		INSERT INTO todoapp.outbox_events (
			topic,
			aggregate_type,
			aggregate_id,
			payload,
			available_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $5);
	`
	if _, err := tx.Exec(
		ctx,
		insertOutbox,
		passwordResetTopic,
		"account",
		strconv.Itoa(request.AccountUserID),
		payload,
		request.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert password reset outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit password reset request: %w", err)
	}
	committed = true
	return nil
}

func (r *IdentityRepository) ResetPassword(
	ctx context.Context,
	tokenHash []byte,
	passwordHash string,
	resetAt time.Time,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		WITH consumed_token AS (
			UPDATE todoapp.account_tokens
			SET consumed_at = $3
			WHERE token_hash = $1
				AND purpose = 'password_reset'
				AND consumed_at IS NULL
				AND expires_at > $3
			RETURNING account_user_id
		),
		updated_credential AS (
			UPDATE todoapp.password_credentials AS credential
			SET
				password_hash = $2,
				updated_at = $3
			FROM consumed_token
			WHERE credential.account_user_id = consumed_token.account_user_id
			RETURNING credential.account_user_id
		),
		revoked_sessions AS (
			UPDATE todoapp.sessions AS session
			SET revoked_at = $3
			FROM updated_credential
			WHERE session.account_user_id = updated_credential.account_user_id
				AND session.revoked_at IS NULL
			RETURNING session.id
		)
		SELECT account_user_id
		FROM updated_credential;
	`

	var accountUserID int
	err := r.pool.QueryRow(ctx, query, tokenHash, passwordHash, resetAt).Scan(&accountUserID)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return fmt.Errorf(
			"password reset token is invalid, expired, or already consumed: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	if err != nil {
		return fmt.Errorf("reset password credential: %w", err)
	}
	return nil
}
