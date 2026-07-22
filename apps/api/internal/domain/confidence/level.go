// Package confidence defines the shared ordinal confidence value object.
//
// Context-specific structures remain responsible for explaining what the
// confidence means. This package only owns the stable ordinal vocabulary.
package confidence

import (
	"errors"
	"fmt"
	"strings"
)

type Level string

const (
	LevelNone   Level = "none"
	LevelLow    Level = "low"
	LevelMedium Level = "medium"
	LevelHigh   Level = "high"
)

var ErrLevelInvalid = errors.New(
	"confidence level is invalid",
)

func ParseLevel(
	value string,
) (Level, error) {
	level := Level(
		strings.ToLower(
			strings.TrimSpace(value),
		),
	)

	if err := level.Validate(); err != nil {
		return LevelNone, err
	}

	return level, nil
}

func (level Level) Validate() error {
	switch level {
	case LevelNone,
		LevelLow,
		LevelMedium,
		LevelHigh:
		return nil
	default:
		return fmt.Errorf(
			"%w: %q",
			ErrLevelInvalid,
			level,
		)
	}
}

func (level Level) Rank() int {
	switch level {
	case LevelHigh:
		return 3
	case LevelMedium:
		return 2
	case LevelLow:
		return 1
	default:
		return 0
	}
}

func Minimum(
	left Level,
	right Level,
) Level {
	if left.Validate() != nil || right.Validate() != nil {
		return LevelNone
	}

	if left.Rank() <= right.Rank() {
		return left
	}

	return right
}
