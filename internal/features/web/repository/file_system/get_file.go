package web_fs_repository

import (
	"errors"
	"fmt"
	"os"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func (r *WebRepository) GetFile(filepath string) ([]byte, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf(
				"file: %s: %w",
				filepath,
				core_errors.ErrNotFound,
			)
		}
		return nil, fmt.Errorf(
			"get file: %s: %w",
			filepath,
			core_errors.ErrNotFound,
		)
	}
	return file, nil
}
