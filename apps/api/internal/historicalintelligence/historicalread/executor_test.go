package historicalread

import (
	"errors"
	"testing"
)

func TestNewPostgresInTransactionRejectsNil(
	t *testing.T,
) {
	_, err := NewPostgresInTransaction(nil)
	if !errors.Is(
		err,
		ErrPostgresTransactionRequired,
	) {
		t.Fatalf(
			"error = %v, want %v",
			err,
			ErrPostgresTransactionRequired,
		)
	}
}
