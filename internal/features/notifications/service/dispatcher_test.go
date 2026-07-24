package service

import (
	"context"
	"testing"
)

type fakeNotificationRepository struct {
	eventID int64
	userID  int
	title   string
	body    string
	email   string
}

func (repository *fakeNotificationRepository) SaveInApp(
	_ context.Context,
	eventID int64,
	userID int,
	title string,
	body string,
) error {
	repository.eventID = eventID
	repository.userID = userID
	repository.title = title
	repository.body = body
	return nil
}

func (repository *fakeNotificationRepository) GetAccountEmail(
	context.Context,
	int,
) (string, error) {
	return repository.email, nil
}

type fakeEmailSender struct {
	to      string
	subject string
	body    string
}

func (sender *fakeEmailSender) Send(
	_ context.Context,
	to string,
	subject string,
	body string,
) error {
	sender.to = to
	sender.subject = subject
	sender.body = body
	return nil
}

func TestNotificationDispatcherPersistsAndSendsWalletCredit(t *testing.T) {
	repository := &fakeNotificationRepository{email: "user@example.com"}
	sender := &fakeEmailSender{}
	dispatcher := NewNotificationDispatcher(repository, sender)
	event := OutboxEvent{
		ID:    7,
		Topic: "payments.wallet_credited",
		Payload: []byte(`{
			"user_id":42,
			"amount_minor":12550,
			"currency":"RUB"
		}`),
	}

	err := dispatcher.Deliver(t.Context(), event)

	if err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if repository.eventID != 7 || repository.userID != 42 {
		t.Fatal("expected in-app notification")
	}
	if sender.to != "user@example.com" || sender.subject == "" || sender.body == "" {
		t.Fatal("expected email notification")
	}
}
