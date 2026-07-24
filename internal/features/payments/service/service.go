package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	payments_provider "github.com/LisLisich/RESTAPI/internal/features/payments/provider"
	"github.com/gofrs/uuid"
)

const maxIdempotencyKeyLength = 64

type PaymentRepository interface {
	ReservePayment(ctx context.Context, payment domain.Payment) (domain.Payment, error)
	AttachProviderPayment(
		ctx context.Context,
		localPaymentID string,
		providerPaymentID string,
		confirmationURL string,
		attachedAt time.Time,
	) (domain.Payment, error)
	CreditSucceededPayment(
		ctx context.Context,
		providerPaymentID string,
		amount domain.Money,
		creditedAt time.Time,
	) (bool, error)
}

type PaymentService struct {
	repository PaymentRepository
	provider   payments_provider.PaymentProvider
	returnURL  string
	now        func() time.Time
}

func NewPaymentService(
	repository PaymentRepository,
	provider payments_provider.PaymentProvider,
	returnURL string,
	now func() time.Time,
) *PaymentService {
	return &PaymentService{
		repository: repository,
		provider:   provider,
		returnURL:  returnURL,
		now:        now,
	}
}

func (service *PaymentService) CreatePayment(
	ctx context.Context,
	accountUserID int,
	idempotencyKey string,
	amountMinor int64,
) (domain.Payment, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) == 0 || len(idempotencyKey) > maxIdempotencyKeyLength {
		return domain.Payment{}, fmt.Errorf(
			"idempotency key must contain 1..%d characters: %w",
			maxIdempotencyKeyLength,
			core_errors.ErrInvalidArgument,
		)
	}
	amount, err := domain.NewPaymentAmount(amountMinor)
	if err != nil {
		return domain.Payment{}, err
	}
	localID, err := uuid.NewV4()
	if err != nil {
		return domain.Payment{}, fmt.Errorf("generate payment id: %w", err)
	}
	payment, err := domain.NewPayment(
		localID.String(),
		accountUserID,
		idempotencyKey,
		amount,
		service.now(),
	)
	if err != nil {
		return domain.Payment{}, err
	}
	reserved, err := service.repository.ReservePayment(ctx, payment)
	if err != nil {
		return domain.Payment{}, fmt.Errorf("reserve payment: %w", err)
	}
	if reserved.Amount != amount {
		return domain.Payment{}, fmt.Errorf(
			"idempotency key already belongs to another amount: %w",
			core_errors.ErrConflict,
		)
	}
	if reserved.ProviderPaymentID != "" {
		return reserved, nil
	}

	providerPayment, err := service.provider.CreatePayment(
		ctx,
		payments_provider.CreatePaymentInput{
			IdempotenceKey: reserved.ID,
			Amount:         reserved.Amount,
			ReturnURL:      service.returnURL,
			Description:    "Пополнение кошелька FinTask",
		},
	)
	if err != nil {
		return domain.Payment{}, fmt.Errorf("create YooKassa payment: %w", err)
	}
	if !providerPayment.Test ||
		providerPayment.Amount != reserved.Amount ||
		providerPayment.ID == "" ||
		providerPayment.ConfirmationURL == "" {
		return domain.Payment{}, fmt.Errorf(
			"YooKassa returned an invalid sandbox payment: %w",
			core_errors.ErrConflict,
		)
	}
	attached, err := service.repository.AttachProviderPayment(
		ctx,
		reserved.ID,
		providerPayment.ID,
		providerPayment.ConfirmationURL,
		service.now(),
	)
	if err != nil {
		return domain.Payment{}, fmt.Errorf("attach YooKassa payment: %w", err)
	}
	return attached, nil
}

func (service *PaymentService) HandleSucceededWebhook(
	ctx context.Context,
	providerPaymentID string,
) (bool, error) {
	providerPaymentID = strings.TrimSpace(providerPaymentID)
	if providerPaymentID == "" {
		return false, fmt.Errorf("provider payment id is required: %w", core_errors.ErrInvalidArgument)
	}
	payment, err := service.provider.GetPayment(ctx, providerPaymentID)
	if err != nil {
		return false, fmt.Errorf("verify YooKassa payment: %w", err)
	}
	if payment.ID != providerPaymentID ||
		payment.Status != string(domain.PaymentStatusSucceeded) ||
		!payment.Paid ||
		!payment.Test {
		return false, fmt.Errorf(
			"YooKassa payment is not a succeeded sandbox payment: %w",
			core_errors.ErrConflict,
		)
	}
	credited, err := service.repository.CreditSucceededPayment(
		ctx,
		payment.ID,
		payment.Amount,
		service.now(),
	)
	if err != nil {
		return false, fmt.Errorf("credit succeeded payment: %w", err)
	}
	return credited, nil
}
