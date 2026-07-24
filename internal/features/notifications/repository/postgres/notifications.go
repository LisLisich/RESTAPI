package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_postgres_pool "github.com/LisLisich/fintask/internal/core/repository/postgres/pool"
	notifications_service "github.com/LisLisich/fintask/internal/features/notifications/service"
)

func (repository *NotificationRepository) SaveInApp(
	ctx context.Context,
	eventID int64,
	userID int,
	title string,
	body string,
) error {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()
	const query = `
		INSERT INTO todoapp.notifications (
			account_user_id, outbox_event_id, title, body
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (account_user_id, outbox_event_id) DO NOTHING;
	`
	if _, err := repository.pool.Exec(ctx, query, userID, eventID, title, body); err != nil {
		return fmt.Errorf("save notification: %w", err)
	}
	return nil
}

func (repository *NotificationRepository) GetAccountEmail(
	ctx context.Context,
	userID int,
) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()
	const query = `SELECT email FROM todoapp.accounts WHERE user_id = $1;`
	var email string
	if err := repository.pool.QueryRow(ctx, query, userID).Scan(&email); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return "", fmt.Errorf("account not found: %w", core_errors.ErrNotFound)
		}
		return "", fmt.Errorf("get account email: %w", err)
	}
	return email, nil
}

func (repository *NotificationRepository) ListNotifications(
	ctx context.Context,
	userID int,
) ([]notifications_service.Notification, error) {
	ctx, cancel := context.WithTimeout(ctx, repository.pool.OpTimeout())
	defer cancel()
	const query = `
		SELECT id, title, body, read_at, created_at
		FROM todoapp.notifications
		WHERE account_user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 100;
	`
	rows, err := repository.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	notifications := make([]notifications_service.Notification, 0)
	for rows.Next() {
		var notification notifications_service.Notification
		var readAt *time.Time
		if err := rows.Scan(
			&notification.ID,
			&notification.Title,
			&notification.Body,
			&readAt,
			&notification.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notification.ReadAt = readAt
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}
	return notifications, nil
}
