package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeOutboxRepository struct {
	event          *OutboxEvent
	claimErr       error
	publishedID    int64
	failedID       int64
	failedStatus   OutboxStatus
	failedAt       time.Time
	failedError    string
	notificationID int64
}

func (repository *fakeOutboxRepository) ClaimNext(context.Context, time.Time) (*OutboxEvent, error) {
	return repository.event, repository.claimErr
}

func (repository *fakeOutboxRepository) MarkPublished(
	_ context.Context,
	id int64,
	_ time.Time,
) error {
	repository.publishedID = id
	return nil
}

func (repository *fakeOutboxRepository) MarkFailed(
	_ context.Context,
	id int64,
	status OutboxStatus,
	availableAt time.Time,
	lastError string,
) error {
	repository.failedID = id
	repository.failedStatus = status
	repository.failedAt = availableAt
	repository.failedError = lastError
	return nil
}

type fakeDispatcher struct {
	err error
}

func (dispatcher fakeDispatcher) Deliver(context.Context, OutboxEvent) error {
	return dispatcher.err
}

func TestProcessorPublishesDeliveredEvent(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	repository := &fakeOutboxRepository{event: &OutboxEvent{ID: 7, Attempts: 1}}
	processor := NewProcessor(repository, fakeDispatcher{}, 5, time.Minute, func() time.Time {
		return now
	})

	processed, err := processor.ProcessOne(t.Context())

	if err != nil {
		t.Fatalf("process event: %v", err)
	}
	if !processed || repository.publishedID != 7 {
		t.Fatal("expected event to be published")
	}
}

func TestProcessorSchedulesRetryAfterDeliveryFailure(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	repository := &fakeOutboxRepository{event: &OutboxEvent{ID: 7, Attempts: 2}}
	processor := NewProcessor(
		repository,
		fakeDispatcher{err: errors.New("smtp unavailable")},
		5,
		time.Minute,
		func() time.Time { return now },
	)

	processed, err := processor.ProcessOne(t.Context())

	if err == nil || !processed {
		t.Fatalf("expected delivery error, got processed=%v err=%v", processed, err)
	}
	if repository.failedStatus != OutboxStatusPending ||
		!repository.failedAt.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("unexpected retry: status=%s at=%s", repository.failedStatus, repository.failedAt)
	}
}

func TestProcessorMovesExhaustedEventToDeadLetter(t *testing.T) {
	repository := &fakeOutboxRepository{event: &OutboxEvent{ID: 7, Attempts: 5}}
	processor := NewProcessor(
		repository,
		fakeDispatcher{err: errors.New("permanent failure")},
		5,
		time.Minute,
		time.Now,
	)

	_, err := processor.ProcessOne(t.Context())

	if err == nil {
		t.Fatal("expected delivery error")
	}
	if repository.failedStatus != OutboxStatusDeadLetter {
		t.Fatalf("expected dead letter, got %s", repository.failedStatus)
	}
}
