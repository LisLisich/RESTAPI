package domain

import (
	"fmt"
	"strings"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusCanceled  PaymentStatus = "canceled"
)

type Payment struct {
	ID                 string
	AccountUserID      int
	IdempotencyKey     string
	ProviderPaymentID  string
	Status             PaymentStatus
	Amount             Money
	ConfirmationURL    string
	CreatedAt          time.Time
	ProviderAttachedAt *time.Time
	CreditedAt         *time.Time
}

func NewPayment(
	id string,
	accountUserID int,
	idempotencyKey string,
	amount Money,
	createdAt time.Time,
) (Payment, error) {
	id = strings.TrimSpace(id)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if id == "" || accountUserID <= 0 || idempotencyKey == "" {
		return Payment{}, fmt.Errorf(
			"payment identity is invalid: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	if _, err := NewPaymentAmount(amount.MinorUnits); err != nil ||
		amount.Currency != CurrencyRUB {
		return Payment{}, fmt.Errorf(
			"payment amount is invalid: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	if createdAt.IsZero() {
		return Payment{}, fmt.Errorf(
			"payment creation time is required: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	return Payment{
		ID:             id,
		AccountUserID:  accountUserID,
		IdempotencyKey: idempotencyKey,
		Status:         PaymentStatusPending,
		Amount:         amount,
		CreatedAt:      createdAt.UTC(),
	}, nil
}
