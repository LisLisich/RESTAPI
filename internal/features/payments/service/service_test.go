package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	payments_provider "github.com/LisLisich/fintask/internal/features/payments/provider"
)

type fakeRepository struct {
	reserved       domain.Payment
	reserveErr     error
	attached       domain.Payment
	attachErr      error
	credited       bool
	creditErr      error
	creditProvider string
	creditAmount   domain.Money
}

func (repository *fakeRepository) ReservePayment(
	_ context.Context,
	_ domain.Payment,
) (domain.Payment, error) {
	return repository.reserved, repository.reserveErr
}

func (repository *fakeRepository) AttachProviderPayment(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ time.Time,
) (domain.Payment, error) {
	return repository.attached, repository.attachErr
}

func (repository *fakeRepository) CreditSucceededPayment(
	_ context.Context,
	providerPaymentID string,
	amount domain.Money,
	_ time.Time,
) (bool, error) {
	repository.creditProvider = providerPaymentID
	repository.creditAmount = amount
	return repository.credited, repository.creditErr
}

type fakeProvider struct {
	created     payments_provider.Payment
	createErr   error
	createInput payments_provider.CreatePaymentInput
	fetched     payments_provider.Payment
	fetchErr    error
}

func (provider *fakeProvider) CreatePayment(
	_ context.Context,
	input payments_provider.CreatePaymentInput,
) (payments_provider.Payment, error) {
	provider.createInput = input
	return provider.created, provider.createErr
}

func (provider *fakeProvider) GetPayment(
	_ context.Context,
	_ string,
) (payments_provider.Payment, error) {
	return provider.fetched, provider.fetchErr
}

func TestCreatePaymentCallsProviderWithStableLocalID(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	reserved, _ := domain.NewPayment(
		"local-id",
		42,
		"client-key",
		domain.Money{MinorUnits: 12550, Currency: domain.CurrencyRUB},
		now,
	)
	attached := reserved
	attached.ProviderPaymentID = "provider-id"
	attached.ConfirmationURL = "https://pay.example/confirm"
	repository := &fakeRepository{reserved: reserved, attached: attached}
	provider := &fakeProvider{created: payments_provider.Payment{
		ID:              "provider-id",
		Status:          "pending",
		Test:            true,
		Amount:          reserved.Amount,
		ConfirmationURL: attached.ConfirmationURL,
	}}
	service := NewPaymentService(repository, provider, "https://app.example/return", func() time.Time {
		return now
	})

	payment, err := service.CreatePayment(t.Context(), 42, "client-key", 12550)

	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if payment.ProviderPaymentID != "provider-id" {
		t.Fatalf("unexpected payment: %#v", payment)
	}
	if provider.createInput.IdempotenceKey != "local-id" ||
		provider.createInput.Amount.MinorUnits != 12550 {
		t.Fatalf("unexpected provider input: %#v", provider.createInput)
	}
}

func TestCreatePaymentReturnsExistingAttachedPaymentWithoutProviderCall(t *testing.T) {
	now := time.Now().UTC()
	reserved, _ := domain.NewPayment(
		"local-id",
		42,
		"client-key",
		domain.Money{MinorUnits: 100, Currency: domain.CurrencyRUB},
		now,
	)
	reserved.ProviderPaymentID = "provider-id"
	repository := &fakeRepository{reserved: reserved}
	provider := &fakeProvider{}
	service := NewPaymentService(repository, provider, "https://app.example/return", time.Now)

	payment, err := service.CreatePayment(t.Context(), 42, "client-key", 100)

	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if payment.ProviderPaymentID != "provider-id" {
		t.Fatalf("unexpected payment: %#v", payment)
	}
	if provider.createInput.IdempotenceKey != "" {
		t.Fatal("provider must not be called for attached payment")
	}
}

func TestCreatePaymentRejectsChangedAmountForSameKey(t *testing.T) {
	reserved, _ := domain.NewPayment(
		"local-id",
		42,
		"client-key",
		domain.Money{MinorUnits: 100, Currency: domain.CurrencyRUB},
		time.Now(),
	)
	service := NewPaymentService(
		&fakeRepository{reserved: reserved},
		&fakeProvider{},
		"https://app.example/return",
		time.Now,
	)

	_, err := service.CreatePayment(t.Context(), 42, "client-key", 200)

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestHandleSucceededWebhookCreditsVerifiedSandboxPayment(t *testing.T) {
	repository := &fakeRepository{credited: true}
	provider := &fakeProvider{fetched: payments_provider.Payment{
		ID:     "provider-id",
		Status: "succeeded",
		Paid:   true,
		Test:   true,
		Amount: domain.Money{MinorUnits: 12550, Currency: domain.CurrencyRUB},
	}}
	service := NewPaymentService(repository, provider, "https://app.example/return", time.Now)

	credited, err := service.HandleSucceededWebhook(t.Context(), "provider-id")

	if err != nil {
		t.Fatalf("handle webhook: %v", err)
	}
	if !credited || repository.creditProvider != "provider-id" ||
		repository.creditAmount.MinorUnits != 12550 {
		t.Fatal("expected verified payment credit")
	}
}

func TestHandleSucceededWebhookRejectsUnverifiedStatus(t *testing.T) {
	provider := &fakeProvider{fetched: payments_provider.Payment{
		ID:     "provider-id",
		Status: "pending",
		Paid:   false,
		Test:   true,
		Amount: domain.Money{MinorUnits: 100, Currency: domain.CurrencyRUB},
	}}
	service := NewPaymentService(
		&fakeRepository{},
		provider,
		"https://app.example/return",
		time.Now,
	)

	_, err := service.HandleSucceededWebhook(t.Context(), "provider-id")

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}
