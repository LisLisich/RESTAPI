package users_service

import (
	"context"
	"errors"
	"testing"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
)

type fakeUserRepository struct {
	createUserCalled bool
	createdUser      domain.User
	createUserErr    error

	getUsersCalled bool
	getUsersLimit  *int
	getUsersOffset *int
	getUsersErr    error

	getUserCalled bool
	gotUserID     int
	getUserErr    error
	storedUser    domain.User

	deleteUserCalled bool
	deletedUserID    int
	deleteUserErr    error

	patchUserCalled bool
	patchedUserID   int
	patchedUser     domain.User
	patchUserErr    error
}

var _ UserRepository = (*fakeUserRepository)(nil)

func (f *fakeUserRepository) CreateUser(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	f.createUserCalled = true
	f.createdUser = user

	if f.createUserErr != nil {
		return domain.User{}, f.createUserErr
	}

	return user, nil
}

func (f *fakeUserRepository) GetUsers(
	ctx context.Context,
	limit *int,
	offset *int,
) ([]domain.User, error) {
	f.getUsersCalled = true
	f.getUsersLimit = limit
	f.getUsersOffset = offset

	if f.getUsersErr != nil {
		return nil, f.getUsersErr
	}

	return []domain.User{
		newServiceUser(10, "Ivan Ivanov", "+79998887766"),
	}, nil
}

func (f *fakeUserRepository) GetUser(
	ctx context.Context,
	id int,
) (domain.User, error) {
	f.getUserCalled = true
	f.gotUserID = id

	if f.getUserErr != nil {
		return domain.User{}, f.getUserErr
	}
	if f.storedUser.ID != 0 {
		return f.storedUser, nil
	}

	return newServiceUser(id, "Ivan Ivanov", "+79998887766"), nil
}

func (f *fakeUserRepository) DeleteUser(
	ctx context.Context,
	id int,
) error {
	f.deleteUserCalled = true
	f.deletedUserID = id

	if f.deleteUserErr != nil {
		return f.deleteUserErr
	}

	return nil
}

func (f *fakeUserRepository) PatchUser(
	ctx context.Context,
	id int,
	user domain.User,
) (domain.User, error) {
	f.patchUserCalled = true
	f.patchedUserID = id
	f.patchedUser = user

	if f.patchUserErr != nil {
		return domain.User{}, f.patchUserErr
	}

	return user, nil
}

func TestCreateUserRejectsInvalidDomainBeforeRepository(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewUserService(repository)
	user := domain.NewUserUninitialized("Iv", nil)

	_, err := service.CreateUser(context.Background(), user)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if repository.createUserCalled {
		t.Fatal("expected CreateUser not to be called for invalid user")
	}
}

func TestCreateUserPassesValidDomainToRepository(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewUserService(repository)
	user := domain.NewUserUninitialized("Ivan Ivanov", serviceStringPtr("+79998887766"))

	createdUser, err := service.CreateUser(context.Background(), user)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repository.createUserCalled {
		t.Fatal("expected CreateUser to be called")
	}
	if repository.createdUser.FullName != "Ivan Ivanov" {
		t.Fatalf("expected full name %q, got %q", "Ivan Ivanov", repository.createdUser.FullName)
	}
	if createdUser.FullName != "Ivan Ivanov" {
		t.Fatalf("expected created user full name %q, got %q", "Ivan Ivanov", createdUser.FullName)
	}
}

func TestCreateUserWrapsRepositoryError(t *testing.T) {
	repository := &fakeUserRepository{
		createUserErr: core_errors.ErrConflict,
	}
	service := NewUserService(repository)
	user := domain.NewUserUninitialized("Ivan Ivanov", serviceStringPtr("+79998887766"))

	_, err := service.CreateUser(context.Background(), user)

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if !repository.createUserCalled {
		t.Fatal("expected CreateUser to be called")
	}
}

func serviceStringPtr(value string) *string {
	return &value
}

func serviceIntPtr(value int) *int {
	return &value
}

func newServiceUser(id int, fullName string, phoneNumber string) domain.User {
	return domain.NewUser(
		id,
		1,
		fullName,
		serviceStringPtr(phoneNumber),
	)
}
