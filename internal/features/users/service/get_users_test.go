package users_service

import (
	"context"
	"errors"
	"testing"

	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

func TestGetUsersPassesQueryParamsToRepository(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewUserService(repository)
	limit := serviceIntPtr(20)
	offset := serviceIntPtr(40)

	users, err := service.GetUsers(context.Background(), limit, offset)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if !repository.getUsersCalled {
		t.Fatal("expected GetUsers to be called")
	}
	if repository.getUsersLimit == nil || *repository.getUsersLimit != 20 {
		t.Fatalf("expected limit %d, got %v", 20, repository.getUsersLimit)
	}
	if repository.getUsersOffset == nil || *repository.getUsersOffset != 40 {
		t.Fatalf("expected offset %d, got %v", 40, repository.getUsersOffset)
	}
}

func TestGetUsersRejectsNegativeLimitBeforeRepository(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewUserService(repository)

	_, err := service.GetUsers(context.Background(), serviceIntPtr(-1), nil)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if repository.getUsersCalled {
		t.Fatal("expected GetUsers not to be called for negative limit")
	}
}

func TestGetUsersRejectsNegativeOffsetBeforeRepository(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewUserService(repository)

	_, err := service.GetUsers(context.Background(), nil, serviceIntPtr(-1))

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if repository.getUsersCalled {
		t.Fatal("expected GetUsers not to be called for negative offset")
	}
}

func TestGetUsersWrapsRepositoryError(t *testing.T) {
	repository := &fakeUserRepository{
		getUsersErr: core_errors.ErrNotFound,
	}
	service := NewUserService(repository)

	_, err := service.GetUsers(context.Background(), nil, nil)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !repository.getUsersCalled {
		t.Fatal("expected GetUsers to be called")
	}
}
