package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/RESTAPI/internal/core/repository/postgres/pool"
	notifications_service "github.com/LisLisich/RESTAPI/internal/features/notifications/service"
)

func (repository *NotificationRepository) ClaimNext(
	ctx context.Context,
	claimedAt time.Time,
) (*notifications_service.OutboxEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()
	const query = `
		WITH candidate AS (
			SELECT id
			FROM todoapp.outbox_events
			WHERE status IN ('pending', 'processing')
				AND available_at <= $1
			ORDER BY id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE todoapp.outbox_events AS event
		SET
			status = 'processing',
			attempts = attempts + 1,
			available_at = $1 + INTERVAL '5 minutes'
		FROM candidate
		WHERE event.id = candidate.id
		RETURNING
			event.id,
			event.topic,
			event.aggregate_type,
			event.aggregate_id,
			event.payload,
			event.attempts;
	`
	var event notifications_service.OutboxEvent
	if err := repository.pool.QueryRow(ctx, query, claimedAt).Scan(
		&event.ID,
		&event.Topic,
		&event.AggregateType,
		&event.AggregateID,
		&event.Payload,
		&event.Attempts,
	); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("claim outbox event: %w", err)
	}
	return &event, nil
}

func (repository *NotificationRepository) MarkPublished(
	ctx context.Context,
	id int64,
	publishedAt time.Time,
) error {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()
	const query = `
		UPDATE todoapp.outbox_events
		SET status = 'published', published_at = $2, last_error = NULL
		WHERE id = $1 AND status = 'processing';
	`
	tag, err := repository.pool.Exec(ctx, query, id, publishedAt)
	if err != nil {
		return fmt.Errorf("mark outbox published: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("processing outbox event not found: %w", core_errors.ErrNotFound)
	}
	return nil
}

func (repository *NotificationRepository) MarkFailed(
	ctx context.Context,
	id int64,
	status notifications_service.OutboxStatus,
	availableAt time.Time,
	lastError string,
) error {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()
	const query = `
		UPDATE todoapp.outbox_events
		SET status = $2, available_at = $3, last_error = $4
		WHERE id = $1 AND status = 'processing';
	`
	tag, err := repository.pool.Exec(ctx, query, id, status, availableAt, lastError)
	if err != nil {
		return fmt.Errorf("mark outbox failed: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("processing outbox event not found: %w", core_errors.ErrNotFound)
	}
	return nil
}
