package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "pending"
	OutboxStatusDeadLetter OutboxStatus = "dead_letter"
)

type OutboxEvent struct {
	ID            int64
	Topic         string
	AggregateType string
	AggregateID   string
	Payload       json.RawMessage
	Attempts      int
}

type OutboxRepository interface {
	ClaimNext(ctx context.Context, claimedAt time.Time) (*OutboxEvent, error)
	MarkPublished(ctx context.Context, id int64, publishedAt time.Time) error
	MarkFailed(
		ctx context.Context,
		id int64,
		status OutboxStatus,
		availableAt time.Time,
		lastError string,
	) error
}

type Dispatcher interface {
	Deliver(ctx context.Context, event OutboxEvent) error
}

type Processor struct {
	repository  OutboxRepository
	dispatcher  Dispatcher
	maxAttempts int
	retryBase   time.Duration
	now         func() time.Time
}

func NewProcessor(
	repository OutboxRepository,
	dispatcher Dispatcher,
	maxAttempts int,
	retryBase time.Duration,
	now func() time.Time,
) *Processor {
	return &Processor{
		repository:  repository,
		dispatcher:  dispatcher,
		maxAttempts: maxAttempts,
		retryBase:   retryBase,
		now:         now,
	}
}

func (processor *Processor) ProcessOne(ctx context.Context) (bool, error) {
	now := processor.now().UTC()
	event, err := processor.repository.ClaimNext(ctx, now)
	if err != nil {
		return false, fmt.Errorf("claim outbox event: %w", err)
	}
	if event == nil {
		return false, nil
	}
	if err := processor.dispatcher.Deliver(ctx, *event); err != nil {
		status := OutboxStatusPending
		availableAt := now.Add(processor.retryDelay(event.Attempts))
		if event.Attempts >= processor.maxAttempts {
			status = OutboxStatusDeadLetter
			availableAt = now
		}
		if markErr := processor.repository.MarkFailed(
			ctx,
			event.ID,
			status,
			availableAt,
			err.Error(),
		); markErr != nil {
			return true, fmt.Errorf(
				"deliver event: %v; mark event failed: %w",
				err,
				markErr,
			)
		}
		return true, fmt.Errorf("deliver outbox event %d: %w", event.ID, err)
	}
	if err := processor.repository.MarkPublished(ctx, event.ID, now); err != nil {
		return true, fmt.Errorf("mark outbox event published: %w", err)
	}
	return true, nil
}

func (processor *Processor) retryDelay(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	if attempts > 10 {
		attempts = 10
	}
	return processor.retryBase * time.Duration(1<<(attempts-1))
}
