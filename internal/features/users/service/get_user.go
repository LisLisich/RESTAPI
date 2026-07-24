package users_service

import (
	"context"
	"fmt"

	"github.com/LisLisich/fintask/internal/core/domain"
)

func (s *UserService) GetUser(
	ctx context.Context,
	id int,
) (domain.User, error) {
	users, err := s.userRepository.GetUser(
		ctx,
		id,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user from repository: %w", err)
	}
	return users, nil
}
