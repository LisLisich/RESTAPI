package users_service

import (
	"context"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
)

type UserService struct {
	userRepository UserRepository
}

type UserRepository interface {
	CreateUser(
		ctx context.Context,
		user domain.User,
	) (domain.User, error)
	GetUsers(
		ctx context.Context,
		limit *int,
		offset *int,
	) ([]domain.User, error)
	GetUser(
		ctx context.Context,
		id int,
	) (domain.User, error)
	DeleteUser(
		ctx context.Context,
		id int,
	) error
	PatchUser(
		ctx context.Context,
		id int,
		patch domain.User,
	) (domain.User, error)
}

func NewUserService(
	userRepository UserRepository,
) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}
