package confidence

import "testing"

func TestMinimumFailsClosedForInvalidOperand(t *testing.T) {
	got := Minimum(Level("broken"), LevelHigh)
	if got != LevelNone {
		t.Fatalf("Minimum() = %q, want %q", got, LevelNone)
	}
}
