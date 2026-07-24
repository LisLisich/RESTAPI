package identity_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
)

const emailVerificationTopic = "identity.email_verification_requested"

type emailVerificationPayload struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

func (r *IdentityRepository) Register(
	ctx context.Context,
	registration identity_service.Registration,
) (domain.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Account{}, fmt.Errorf("begin identity registration: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	const insertUser = `
		INSERT INTO todoapp.users (full_name, phone_number)
		VALUES ($1, $2)
		RETURNING id;
	`
	var userID int
	if err := tx.QueryRow(
		ctx,
		insertUser,
		registration.User.FullName,
		registration.User.PhoneNumber,
	).Scan(&userID); err != nil {
		return domain.Account{}, mapRegistrationError("insert user", err)
	}

	const insertAccount = `
		INSERT INTO todoapp.accounts (
			user_id,
			email,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $4);
	`
	if _, err := tx.Exec(
		ctx,
		insertAccount,
		userID,
		registration.Account.Email,
		registration.Account.Status,
		registration.CreatedAt,
	); err != nil {
		return domain.Account{}, mapRegistrationError("insert account", err)
	}

	const insertPasswordCredential = `
		INSERT INTO todoapp.password_credentials (
			account_user_id,
			password_hash,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $3);
	`
	if _, err := tx.Exec(
		ctx,
		insertPasswordCredential,
		userID,
		registration.PasswordHash,
		registration.CreatedAt,
	); err != nil {
		return domain.Account{}, mapRegistrationError("insert password credential", err)
	}

	const insertVerificationToken = `
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
		insertVerificationToken,
		userID,
		"email_verification",
		registration.VerificationToken.Hash,
		registration.VerificationExpiresAt,
		registration.CreatedAt,
	); err != nil {
		return domain.Account{}, mapRegistrationError("insert verification token", err)
	}

	payload, err := json.Marshal(emailVerificationPayload{
		UserID: userID,
		Email:  registration.Account.Email,
		Token:  registration.VerificationToken.Raw,
	})
	if err != nil {
		return domain.Account{}, fmt.Errorf("marshal email verification outbox payload: %w", err)
	}

	const insertOutboxEvent = `
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
		insertOutboxEvent,
		emailVerificationTopic,
		"account",
		strconv.Itoa(userID),
		payload,
		registration.CreatedAt,
	); err != nil {
		return domain.Account{}, mapRegistrationError("insert verification outbox event", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Account{}, fmt.Errorf("commit identity registration: %w", err)
	}
	committed = true

	account := registration.Account
	account.UserID = userID
	return account, nil
}

func mapRegistrationError(operation string, err error) error {
	if errors.Is(err, core_postgres_pool.ErrViolatesUnique) {
		return fmt.Errorf("%s: account already exists: %w", operation, core_errors.ErrConflict)
	}
	return fmt.Errorf("%s: %w", operation, err)
}
