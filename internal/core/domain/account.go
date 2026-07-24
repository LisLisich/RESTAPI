package domain

import (
	"fmt"
	"net/mail"
	"strings"
	"time"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

type AccountStatus string

const (
	AccountStatusPendingVerification AccountStatus = "pending_verification"
	AccountStatusActive              AccountStatus = "active"
)

type Account struct {
	UserID          int
	Email           string
	Status          AccountStatus
	EmailVerifiedAt *time.Time
}

func NewAccountUninitialized(userID int, email string) (Account, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return Account{}, err
	}

	return Account{
		UserID: userID,
		Email:  normalizedEmail,
		Status: AccountStatusPendingVerification,
	}, nil
}

func (a *Account) VerifyEmail(verifiedAt time.Time) error {
	if verifiedAt.IsZero() {
		return fmt.Errorf("verification time is required: %w", core_errors.ErrInvalidArgument)
	}

	verifiedAt = verifiedAt.UTC()
	a.EmailVerifiedAt = &verifiedAt
	a.Status = AccountStatusActive
	return nil
}

func ValidatePassword(password string) error {
	passwordLength := len([]rune(password))
	if passwordLength < 12 || passwordLength > 128 {
		return fmt.Errorf(
			"invalid password length %d: %w",
			passwordLength,
			core_errors.ErrInvalidArgument,
		)
	}
	return nil
}

func normalizeEmail(email string) (string, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if len(normalizedEmail) == 0 || len(normalizedEmail) > 254 {
		return "", fmt.Errorf("invalid email length: %w", core_errors.ErrInvalidArgument)
	}

	address, err := mail.ParseAddress(normalizedEmail)
	if err != nil || address.Address != normalizedEmail {
		return "", fmt.Errorf("invalid email format: %w", core_errors.ErrInvalidArgument)
	}

	return normalizedEmail, nil
}
