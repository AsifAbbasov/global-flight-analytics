package confidence

import (
	"errors"
	"testing"
)

func TestParseLevelNormalizesValue(t *testing.T) {
	level, err := ParseLevel(" Medium ")
	if err != nil {
		t.Fatalf("parse level: %v", err)
	}

	if level != LevelMedium {
		t.Fatalf(
			"level = %q, want %q",
			level,
			LevelMedium,
		)
	}
}

func TestParseLevelRejectsUnknownValue(t *testing.T) {
	_, err := ParseLevel("certain")
	if !errors.Is(err, ErrLevelInvalid) {
		t.Fatalf(
			"error = %v, want %v",
			err,
			ErrLevelInvalid,
		)
	}
}

func TestMinimumReturnsWeakerLevel(t *testing.T) {
	if got := Minimum(
		LevelHigh,
		LevelLow,
	); got != LevelLow {
		t.Fatalf(
			"minimum = %q, want %q",
			got,
			LevelLow,
		)
	}
}
