package users_service

import (
	"context"
	"errors"
	"testing"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestPatchUserAppliesPatchBeforeRepositoryUpdate(t *testing.T) {
	repository := &fakeUserRepository{
		storedUser: newServiceUser(15, "Ivan Ivanov", "+79998887766"),
	}
	service := NewUserService(repository)
	patch := domain.NewUserPatch(
		domain.Nullable[string]{Set: true, Value: serviceStringPtr("Ivan Updated")},
		domain.Nullable[string]{Set: true, Value: nil},
	)

	user, err := service.PatchUser(context.Background(), 15, patch)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.getUserCalled {
		t.Fatal("expected GetUser to be called before PatchUser")
	}
	if !repository.patchUserCalled {
		t.Fatal("expected PatchUser to be called")
	}
	if repository.patchedUserID != 15 {
		t.Fatalf("expected patched user id %d, got %d", 15, repository.patchedUserID)
	}
	if repository.patchedUser.FullName != "Ivan Updated" {
		t.Fatalf("expected patched full name %q, got %q", "Ivan Updated", repository.patchedUser.FullName)
	}
	if repository.patchedUser.PhoneNumber != nil {
		t.Fatalf("expected patched phone number to be nil, got %q", *repository.patchedUser.PhoneNumber)
	}
	if user.FullName != "Ivan Updated" {
		t.Fatalf("expected returned full name %q, got %q", "Ivan Updated", user.FullName)
	}
}

func TestPatchUserReturnsGetUserErrorBeforeApplyingPatch(t *testing.T) {
	repository := &fakeUserRepository{
		getUserErr: core_errors.ErrNotFound,
	}
	service := NewUserService(repository)
	patch := domain.NewUserPatch(
		domain.Nullable[string]{Set: true, Value: serviceStringPtr("Ivan Updated")},
		domain.Nullable[string]{},
	)

	_, err := service.PatchUser(context.Background(), 15, patch)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !repository.getUserCalled {
		t.Fatal("expected GetUser to be called")
	}
	if repository.patchUserCalled {
		t.Fatal("expected PatchUser not to be called when GetUser fails")
	}
}

func TestPatchUserRejectsInvalidPatchBeforeRepositoryUpdate(t *testing.T) {
	repository := &fakeUserRepository{
		storedUser: newServiceUser(15, "Ivan Ivanov", "+79998887766"),
	}
	service := NewUserService(repository)
	patch := domain.NewUserPatch(
		domain.Nullable[string]{Set: true, Value: serviceStringPtr("Iv")},
		domain.Nullable[string]{},
	)

	_, err := service.PatchUser(context.Background(), 15, patch)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if !repository.getUserCalled {
		t.Fatal("expected GetUser to be called")
	}
	if repository.patchUserCalled {
		t.Fatal("expected PatchUser not to be called for invalid patch")
	}
}

func TestPatchUserWrapsRepositoryPatchError(t *testing.T) {
	repository := &fakeUserRepository{
		storedUser:   newServiceUser(15, "Ivan Ivanov", "+79998887766"),
		patchUserErr: core_errors.ErrConflict,
	}
	service := NewUserService(repository)
	patch := domain.NewUserPatch(
		domain.Nullable[string]{Set: true, Value: serviceStringPtr("Ivan Updated")},
		domain.Nullable[string]{},
	)

	_, err := service.PatchUser(context.Background(), 15, patch)

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if !repository.patchUserCalled {
		t.Fatal("expected PatchUser to be called")
	}
}
