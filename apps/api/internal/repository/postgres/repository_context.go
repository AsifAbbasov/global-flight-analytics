package postgres

import (
	"context"
	"errors"
)

var ErrRepositoryContextRequired = errors.New(
	"repository context is required",
)

func requireRepositoryContext(ctx context.Context) error {
	if ctx == nil {
		return ErrRepositoryContextRequired
	}
	return nil
}
