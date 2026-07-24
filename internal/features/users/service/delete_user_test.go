package users_service

import (
	"context"
	"errors"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestDeleteUserPassesIDToRepository(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewUserService(repository)

	err := service.DeleteUser(context.Background(), 15)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.deleteUserCalled {
		t.Fatal("expected DeleteUser to be called")
	}
	if repository.deletedUserID != 15 {
		t.Fatalf("expected user id %d, got %d", 15, repository.deletedUserID)
	}
}

func TestDeleteUserWrapsRepositoryError(t *testing.T) {
	repository := &fakeUserRepository{
		deleteUserErr: core_errors.ErrNotFound,
	}
	service := NewUserService(repository)

	err := service.DeleteUser(context.Background(), 15)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !repository.deleteUserCalled {
		t.Fatal("expected DeleteUser to be called")
	}
}
