package users_service

import (
	"context"
	"errors"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestGetUserPassesIDToRepository(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewUserService(repository)

	user, err := service.GetUser(context.Background(), 15)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.getUserCalled {
		t.Fatal("expected GetUser to be called")
	}
	if repository.gotUserID != 15 {
		t.Fatalf("expected user id %d, got %d", 15, repository.gotUserID)
	}
	if user.ID != 15 {
		t.Fatalf("expected returned user id %d, got %d", 15, user.ID)
	}
}

func TestGetUserWrapsRepositoryError(t *testing.T) {
	repository := &fakeUserRepository{
		getUserErr: core_errors.ErrNotFound,
	}
	service := NewUserService(repository)

	_, err := service.GetUser(context.Background(), 15)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !repository.getUserCalled {
		t.Fatal("expected GetUser to be called")
	}
}
