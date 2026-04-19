package users_servise

import (
	"context"
	"fmt"
)

func (s *UserService) DeleteUser(
	ctx context.Context,
	id int,
) error {
	if err := s.userRepository.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
