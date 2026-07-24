package users_transport_http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	"go.uber.org/zap"
)

type fakeUsersService struct {
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

	deleteUserCalled bool
	deletedUserID    int
	deleteUserErr    error

	patchUserCalled bool
	patchedUserID   int
	patchedPatch    domain.UserPatch
	patchUserErr    error
}

var _ UsersService = (*fakeUsersService)(nil)

func (s *fakeUsersService) CreateUser(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	s.createUserCalled = true
	s.createdUser = user

	if s.createUserErr != nil {
		return domain.User{}, s.createUserErr
	}

	return domain.NewUser(
		10,
		1,
		user.FullName,
		user.PhoneNumber,
	), nil
}

func (s *fakeUsersService) GetUsers(
	ctx context.Context,
	limit *int,
	offset *int,
) ([]domain.User, error) {
	s.getUsersCalled = true
	s.getUsersLimit = limit
	s.getUsersOffset = offset

	if s.getUsersErr != nil {
		return nil, s.getUsersErr
	}

	return []domain.User{
		newUser(10, "Ivan Ivanov", "+79998887766"),
		newUser(11, "Petr Petrov", "+71112223344"),
	}, nil
}

func (s *fakeUsersService) GetUser(
	ctx context.Context,
	id int,
) (domain.User, error) {
	s.getUserCalled = true
	s.gotUserID = id

	if s.getUserErr != nil {
		return domain.User{}, s.getUserErr
	}

	return newUser(id, "Ivan Ivanov", "+79998887766"), nil
}

func (s *fakeUsersService) DeleteUser(
	ctx context.Context,
	id int,
) error {
	s.deleteUserCalled = true
	s.deletedUserID = id

	if s.deleteUserErr != nil {
		return s.deleteUserErr
	}

	return nil
}

func (s *fakeUsersService) PatchUser(
	ctx context.Context,
	id int,
	patch domain.UserPatch,
) (domain.User, error) {
	s.patchUserCalled = true
	s.patchedUserID = id
	s.patchedPatch = patch

	if s.patchUserErr != nil {
		return domain.User{}, s.patchUserErr
	}

	return domain.NewUser(
		id,
		2,
		"Ivan Updated",
		nil,
	), nil
}

func TestCreateUserReturnsCreatedOnValidRequest(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newCreateUserRequest(`{
		"full_name": "Ivan Ivanov",
		"phone_number": "+79998887766"
	}`)
	response := httptest.NewRecorder()

	handler.CreateUser(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}
	if !service.createUserCalled {
		t.Fatal("expected service CreateUser to be called")
	}
	if service.createdUser.FullName != "Ivan Ivanov" {
		t.Fatalf("expected full name %q, got %q", "Ivan Ivanov", service.createdUser.FullName)
	}
	if service.createdUser.PhoneNumber == nil || *service.createdUser.PhoneNumber != "+79998887766" {
		t.Fatalf("expected phone number %q, got %v", "+79998887766", service.createdUser.PhoneNumber)
	}
}

func TestCreateUserReturnsBadRequestOnInvalidRequest(t *testing.T) {
	service := &fakeUsersService{}
	handler := NewUsersHTTPHandler(service)

	request := newCreateUserRequest(`{
		"full_name": "Iv",
		"phone_number": "+79998887766"
	}`)
	response := httptest.NewRecorder()

	handler.CreateUser(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.createUserCalled {
		t.Fatal("expected service CreateUser not to be called")
	}
}

func TestCreateUserReturnsConflictWhenServiceReturnsConflict(t *testing.T) {
	service := &fakeUsersService{
		createUserErr: fmt.Errorf("create user: %w", core_errors.ErrConflict),
	}
	handler := NewUsersHTTPHandler(service)

	request := newCreateUserRequest(`{
		"full_name": "Ivan Ivanov",
		"phone_number": "+79998887766"
	}`)
	response := httptest.NewRecorder()

	handler.CreateUser(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, response.Code)
	}
	if !service.createUserCalled {
		t.Fatal("expected service CreateUser to be called")
	}
}

func newCreateUserRequest(body string) *http.Request {
	return newUserRequestWithLogger(
		http.MethodPost,
		"/users",
		strings.NewReader(body),
	)
}

func newUserRequestWithLogger(method string, target string, body *strings.Reader) *http.Request {
	request := httptest.NewRequest(method, target, body)
	logger := &core_logger.Logger{
		Logger: zap.NewNop(),
	}
	return request.WithContext(core_logger.ToContext(request.Context(), logger))
}

func userPtr[T any](value T) *T {
	return &value
}

func newUser(id int, fullName string, phoneNumber string) domain.User {
	return domain.NewUser(
		id,
		1,
		fullName,
		userPtr(phoneNumber),
	)
}
