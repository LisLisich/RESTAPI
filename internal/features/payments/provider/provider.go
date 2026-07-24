package provider

import (
	"context"

	"github.com/LisLisich/fintask/internal/core/domain"
)

type CreatePaymentInput struct {
	IdempotenceKey string
	Amount         domain.Money
	ReturnURL      string
	Description    string
}

type Payment struct {
	ID              string
	Status          string
	Paid            bool
	Test            bool
	Amount          domain.Money
	ConfirmationURL string
}

type PaymentProvider interface {
	CreatePayment(ctx context.Context, input CreatePaymentInput) (Payment, error)
	GetPayment(ctx context.Context, paymentID string) (Payment, error)
}
