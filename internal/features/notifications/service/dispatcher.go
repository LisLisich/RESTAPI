package service

import (
	"context"
	"encoding/json"
	"fmt"
)

type NotificationRepository interface {
	SaveInApp(
		ctx context.Context,
		eventID int64,
		userID int,
		title string,
		body string,
	) error
	GetAccountEmail(ctx context.Context, userID int) (string, error)
}

type EmailSender interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

type NotificationDispatcher struct {
	repository NotificationRepository
	email      EmailSender
}

func NewNotificationDispatcher(
	repository NotificationRepository,
	email EmailSender,
) *NotificationDispatcher {
	return &NotificationDispatcher{repository: repository, email: email}
}

func (dispatcher *NotificationDispatcher) Deliver(
	ctx context.Context,
	event OutboxEvent,
) error {
	switch event.Topic {
	case "payments.wallet_credited":
		return dispatcher.deliverWalletCredit(ctx, event)
	case "identity.email_verification_requested":
		return dispatcher.deliverIdentityEmail(
			ctx,
			event,
			"Подтверждение email FinTask",
			"Код подтверждения: ",
		)
	case "identity.password_reset_requested":
		return dispatcher.deliverIdentityEmail(
			ctx,
			event,
			"Сброс пароля FinTask",
			"Код сброса пароля: ",
		)
	default:
		return fmt.Errorf("unsupported outbox topic %q", event.Topic)
	}
}

func (dispatcher *NotificationDispatcher) deliverWalletCredit(
	ctx context.Context,
	event OutboxEvent,
) error {
	var payload struct {
		UserID      int    `json:"user_id"`
		AmountMinor int64  `json:"amount_minor"`
		Currency    string `json:"currency"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode wallet credit payload: %w", err)
	}
	if payload.UserID <= 0 || payload.AmountMinor <= 0 || payload.Currency != "RUB" {
		return fmt.Errorf("invalid wallet credit payload")
	}
	title := "Кошелек пополнен"
	body := fmt.Sprintf(
		"На кошелек зачислено %s RUB.",
		formatMinorUnits(payload.AmountMinor),
	)
	if err := dispatcher.repository.SaveInApp(
		ctx,
		event.ID,
		payload.UserID,
		title,
		body,
	); err != nil {
		return fmt.Errorf("save in-app notification: %w", err)
	}
	email, err := dispatcher.repository.GetAccountEmail(ctx, payload.UserID)
	if err != nil {
		return fmt.Errorf("get notification email: %w", err)
	}
	if err := dispatcher.email.Send(ctx, email, title, body); err != nil {
		return fmt.Errorf("send wallet credit email: %w", err)
	}
	return nil
}

func (dispatcher *NotificationDispatcher) deliverIdentityEmail(
	ctx context.Context,
	event OutboxEvent,
	subject string,
	prefix string,
) error {
	var payload struct {
		UserID int    `json:"user_id"`
		Email  string `json:"email"`
		Token  string `json:"token"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode identity notification payload: %w", err)
	}
	if payload.UserID <= 0 || payload.Email == "" || payload.Token == "" {
		return fmt.Errorf("invalid identity notification payload")
	}
	if err := dispatcher.email.Send(
		ctx,
		payload.Email,
		subject,
		prefix+payload.Token,
	); err != nil {
		return fmt.Errorf("send identity email: %w", err)
	}
	return nil
}

func formatMinorUnits(value int64) string {
	return fmt.Sprintf("%d.%02d", value/100, value%100)
}
