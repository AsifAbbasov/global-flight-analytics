package postgres

import (
	"context"
	"errors"
	"testing"
)

func TestRequireRepositoryContextRejectsNil(t *testing.T) {
	t.Parallel()

	err := requireRepositoryContext(nil)
	if !errors.Is(err, ErrRepositoryContextRequired) {
		t.Fatalf("expected repository context error, got %v", err)
	}
}

func TestRequireRepositoryContextAcceptsCallerContext(t *testing.T) {
	t.Parallel()

	if err := requireRepositoryContext(context.Background()); err != nil {
		t.Fatalf("require repository context: %v", err)
	}
}
