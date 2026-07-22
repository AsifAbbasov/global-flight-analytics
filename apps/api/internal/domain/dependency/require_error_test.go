package dependency

import (
	"errors"
	"testing"
)

func TestRequireReturnsTypedError(t *testing.T) {
	err := Require("repository", nil)
	if !errors.Is(err, ErrRequired) {
		t.Fatalf("Require() error = %v", err)
	}
}
