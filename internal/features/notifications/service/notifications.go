package service

import (
	"context"
	"time"
)

type Notification struct {
	ID        int64
	Title     string
	Body      string
	ReadAt    *time.Time
	CreatedAt time.Time
}

type NotificationReader interface {
	ListNotifications(ctx context.Context, userID int) ([]Notification, error)
}
