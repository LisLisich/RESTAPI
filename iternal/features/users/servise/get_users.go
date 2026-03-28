package users_servise

import (
	"context"
	"fmt"

	"github.com/LisLisich/RESTAPI/iternal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/iternal/core/errors"
)

func (s *UserService) GetUsers(
	ctx context.Context,
	limit *int,
	offset *int,
) ([]domain.User, error) {
	if limit != nil && *limit < 0 {
		return nil, fmt.Errorf(
			"limit must be non-negativ: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	if offset != nil && *offset < 0 {
		return nil, fmt.Errorf(
			"offset must be non-negativ: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	users, err := s.userRepository.GetUsers(
		ctx,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("get users from repository: %w", err)
	}
	return users, nil
}
